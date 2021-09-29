package spindel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/miku/labe/tools/spindel/set"
	"github.com/segmentio/encoding/json"
)

// Server wraps three data sources required for index and citation data fusion.
// The IdentifierDatabase is a map from local identifier (e.g. 0-1238201) to
// DOI, the OciDatabase contains citing and cited relationsships from OCI data
// dump and IndexData allows to fetch a metadata blob from a service, e.g. a
// key value store like microblob.
type Server struct {
	IdentifierDatabase *sqlx.DB
	OciDatabase        *sqlx.DB
	IndexData          Fetcher
	Router             *mux.Router

	// Testing feature, fixed for now; just for development. TODO: remove this.
	FeatureFetchSet bool
}

// Routes sets up route.
func (s *Server) Routes() {
	s.Router.HandleFunc("/", s.handleIndex())
	s.Router.HandleFunc("/q/{id}", s.handleQuery())
}

// ServeHTTP turns the server into an HTTP handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "spindel")
	}
}

// handleQuery does all the lookups, but that should elsewhere, in a more
// testable place. Also, reuse some existing stats library. Also TODO: optimize
// backend requests and think up schema for delivery.
func (s *Server) handleQuery() http.HandlerFunc {
	// Map is a generic lookup table. We use it together with sqlite3.
	type Map struct {
		Key   string `db:"k"`
		Value string `db:"v"`
	}
	// Response contains a subset of index data fused with citation data.
	// Citing and cited documents are unparsed. For unmatched docs, we keep
	// only transmit the DOI, e.g. as {"doi": "10.123/123"}.
	type Response struct {
		ID        string            `json:"id"`
		DOI       string            `json:"doi"`
		Citing    []json.RawMessage `json:"citing,omitempty"`
		Cited     []json.RawMessage `json:"cited,omitempty"`
		Unmatched struct {
			Citing []json.RawMessage `json:"citing,omitempty"`
			Cited  []json.RawMessage `json:"cited,omitempty"`
		} `json:"unmatched"`
		Extra struct {
			Took                 float64 `json:"took"`
			UnmatchedCitingCount int     `json:"unmatched_citing_count"`
			UnmatchedCitedCount  int     `json:"unmatched_cited_count"`
			CitingCount          int     `json:"citing_count"`
			CitedCount           int     `json:"cited_count"`
		} `json:"extra"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		// (1) resolve id to doi
		// (2) lookup related doi via oci
		// (3) resolve doi to ids
		// (4) lookup all ids
		// (5) include unmatched ids
		// (6) assemble result
		var (
			ctx          = r.Context()
			started      = time.Now()
			vars         = mux.Vars(r)
			citing       []Map // rows containing paper references (value is the reference item)
			cited        []Map // rows containing inbound relations (key is the citing id)
			ids          []Map // local identifiers
			outbound     = set.New()
			inbound      = set.New()
			matched      []string    // dangling DOI that have no local id
			unmatchedSet = set.New() // all unmatched doi
			response     = &Response{
				ID: vars["id"], // the local identifier
			}
		)
		// (1) Get the DOI for the local id; or get out.
		if err := s.IdentifierDatabase.GetContext(ctx, &response.DOI,
			"SELECT v FROM map WHERE k = ?", response.ID); err != nil {
			httpErrLog(w, err)
			return
		}
		// (2) With the DOI, find outbound (citing) and inbound (cited)
		// references in the OCI database.
		if err := s.OciDatabase.SelectContext(ctx, &citing,
			"SELECT * FROM map WHERE k = ?", response.DOI); err != nil {
			httpErrLog(w, err)
			return
		}
		if err := s.OciDatabase.SelectContext(ctx, &cited,
			"SELECT * FROM map WHERE v = ?", response.DOI); err != nil {
			httpErrLog(w, err)
			return
		}
		// (3) We want to collect the unique set of DOI to get the complete
		// indexed documents.
		for _, v := range citing {
			outbound.Add(v.Value)
		}
		for _, v := range cited {
			inbound.Add(v.Key)
		}
		ss := outbound.Union(inbound)
		if ss.IsEmpty() {
			// This is where the difference in the benchmark runs comes from,
			// e.g. 64860/100000; estimated ratio 64% of records with DOI will
			// have some reference information. TODO: dig a bit deeper.
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// (4) We now need to map back the DOI to the internal identifiers. That's
		// probably a more expensive query.
		query, args, err := sqlx.In("SELECT * FROM map WHERE v IN (?)", ss.Slice())
		if err != nil {
			httpErrLog(w, err)
			return
		}
		query = s.IdentifierDatabase.Rebind(query)
		if err := s.IdentifierDatabase.SelectContext(ctx, &ids, query, args...); err != nil {
			httpErrLog(w, err)
			return
		}
		// (5) Here, we can find unmatched items.
		for _, v := range ids {
			matched = append(matched, v.Value)
		}
		unmatchedSet = ss.Difference(set.FromSlice(matched))
		for k := range unmatchedSet {
			// We shortcut and do not use a proper JSON marshaller to save a
			// bit of time. TODO: may switch to proper JSON encoding, if other
			// parts are more optimized.
			b := []byte(fmt.Sprintf(`{"doi": %q}`, k))
			if outbound.Contains(k) {
				response.Unmatched.Cited = append(response.Unmatched.Cited, b)
			} else {
				response.Unmatched.Citing = append(response.Unmatched.Citing, b)
			}
		}
		// (6) At this point, we need to assemble the result. For each
		// identifier we want the full metadata. We use an local copy of the
		// index. We could also ask a live index here.
		for _, v := range ids {
			// Access the data, here we use the blob, but we could ask SOLR, too.
			b, err := s.IndexData.Fetch(v.Key)
			if errors.Is(err, ErrBlobNotFound) {
				continue
			}
			if err != nil {
				httpErrLog(w, err)
				return
			}
			// We have the blob and the {k: local, v: doi} values, so all we
			// should need.
			switch {
			case outbound.Contains(v.Value):
				response.Citing = append(response.Citing, b)
			case inbound.Contains(v.Value):
				response.Cited = append(response.Cited, b)
			}
		}
		// Fill extra fields.
		response.Extra.CitingCount = len(response.Citing)
		response.Extra.CitedCount = len(response.Cited)
		response.Extra.UnmatchedCitingCount = len(response.Unmatched.Citing)
		response.Extra.UnmatchedCitedCount = len(response.Unmatched.Cited)
		response.Extra.Took = time.Since(started).Seconds()
		// Put it on the wire.
		w.Header().Add("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		if err := enc.Encode(response); err != nil {
			httpErrLog(w, err)
			return
		}
	}
}

// Ping returns an error, if any of the datastores are not available.
func (s *Server) Ping() error {
	if err := s.IdentifierDatabase.Ping(); err != nil {
		return err
	}
	if err := s.OciDatabase.Ping(); err != nil {
		return err
	}
	if pinger, ok := s.IndexData.(Pinger); ok {
		if err := pinger.Ping(); err != nil {
			return fmt.Errorf("could not reach index data service: %w", err)
		}
	} else {
		log.Printf("index data service: unknown status")
	}
	return nil
}

// httpErrLogStatus logs the error and returns.
func httpErrLogStatus(w http.ResponseWriter, err error, status int) {
	log.Printf("failed [%d]: %v", status, err)
	http.Error(w, err.Error(), status)
}

// httpErrLog tries to infer an appropriate status code.
func httpErrLog(w http.ResponseWriter, err error) {
	var status = http.StatusInternalServerError
	if errors.Is(err, context.Canceled) {
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		status = http.StatusNotFound
	}
	httpErrLogStatus(w, err, status)
}
