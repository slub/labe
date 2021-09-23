package spindel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/miku/labe/tools/spindel/set"
	"github.com/segmentio/encoding/json"
	"golang.org/x/sync/errgroup"
)

// Server wraps various data stores.
type Server struct {
	IdentifierDatabase *sqlx.DB
	OciDatabase        *sqlx.DB
	IndexDataService   string // TODO: make this slightly more abstract.
	Router             *mux.Router
}

func (s *Server) Info(ctx context.Context) error {
	var (
		info = struct {
			IdentifierDatabaseCount int `json:"identifier_database_count"`
			OciDatabaseCount        int `json:"oci_database_count"`
			IndexDataCount          int `json:"index_data_count"`
		}{}
		row    *sql.Row
		client = http.Client{
			Timeout: 120 * time.Second,
		}
		funcs = []func() error{
			func() error {
				row = s.IdentifierDatabase.QueryRowContext(ctx, "SELECT count(*) FROM map")
				return row.Scan(&info.IdentifierDatabaseCount)
			},
			func() error {
				row = s.OciDatabase.QueryRowContext(ctx, "SELECT count(*) FROM map")
				return row.Scan(&info.OciDatabaseCount)
			},
			func() error {
				req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/count", s.IndexDataService), nil)
				if err != nil {
					return err
				}
				resp, err := client.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				dec := json.NewDecoder(resp.Body)
				var countResp = struct {
					Count int `json:"count"`
				}{}
				if err := dec.Decode(&countResp); err != nil {
					return err
				}
				info.IndexDataCount = countResp.Count
				return nil
			},
		}
	)
	g, ctx := errgroup.WithContext(ctx)
	for _, f := range funcs {
		g.Go(f)
	}
	log.Println("⚑ querying three data stores ...")
	if err := g.Wait(); err != nil {
		return err
	}
	b, err := json.Marshal(info)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func (s *Server) Routes() {
	s.Router.HandleFunc("/", s.handleIndex())
	s.Router.HandleFunc("/q/{id}", s.handleQuery())
}

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "spindel")
	}
}

// httpErrLogStatus logs the error and returns.
func httpErrLogStatus(w http.ResponseWriter, err error, status int) {
	log.Printf("failed [%d]: %v", status, err)
	http.Error(w, err.Error(), status)
}

// httpErrLog tries to infer an appropriate status code.
func httpErrLog(w http.ResponseWriter, err error) {
	var status = http.StatusInternalServerError
	if errors.Is(err, sql.ErrNoRows) {
		status = http.StatusNotFound
	}
	httpErrLogStatus(w, err, status)
}

// handleQuery does all the lookups, but that should elsewhere, in a more
// testable place. Also, reuse some existing stats library. Also TODO:
// parallelize all backend requests and think up schema for delivery.
func (s *Server) handleQuery() http.HandlerFunc {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	// Map is a generic lookup table. We use it together with sqlite3.
	type Map struct {
		Key   string `db:"k"`
		Value string `db:"v"`
	}
	// Response contains a subset of index data fused with citation data.
	type Response struct {
		// The local identifier.
		ID string `json:"id"`
		// The DOI for the local identifier.
		DOI string `json:"doi"`
		// We want to safe time not doing any serialization, if we do not need it.
		Citing []json.RawMessage `json:"citing"`
		Cited  []json.RawMessage `json:"cited"`
		// Some extra information for now, may not need these.
		Extra struct {
			Took        float64 `json:"took"`
			CitingCount int     `json:"citing_count"`
			CitedCount  int     `json:"cited_count"`
		} `json:"extra"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		// Outline
		// (1) resolve id to doi
		// (2) lookup related doi via oci
		// (3) resolve doi to ids
		// (4) lookup all ids
		// (5) assemble result
		var (
			ctx      = r.Context()
			started  = time.Now()
			vars     = mux.Vars(r)
			citing   []Map // rows containing paper references (value is the reference item)
			cited    []Map // rows containing inbound relations (key is the citing id)
			ids      []Map // local identifiers
			outbound = set.New()
			inbound  = set.New()
			response = Response{
				ID:     vars["id"], // the local identifier
				Citing: []json.RawMessage{},
				Cited:  []json.RawMessage{},
			}
		)
		// (1) Get the DOI for the local id; or get out.
		if err := s.IdentifierDatabase.GetContext(ctx, &response.DOI, "SELECT v FROM map WHERE k = ?", response.ID); err != nil {
			httpErrLog(w, err)
			return
		}
		// (2) With the DOI, find outbound (citing) and inbound (cited)
		// references in the OCI database.
		if err := s.OciDatabase.SelectContext(ctx, &citing, "SELECT * FROM map WHERE k = ?", response.DOI); err != nil {
			httpErrLog(w, err)
			return
		}
		if err := s.OciDatabase.SelectContext(ctx, &cited, "SELECT * FROM map WHERE v = ?", response.DOI); err != nil {
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
		// (5) At this point, we need to assemble the result. For each
		// identifier we want the full metadata. We use an local copy of the
		// index. We could also ask a life index here.
		for _, v := range ids {
			// Access the data, here we use the blob, but we could ask SOLR, too.
			link := fmt.Sprintf("%s/%s", s.IndexDataService, v.Key)
			resp, err := client.Get(link) // TODO: a better client
			if err != nil {
				httpErrLog(w, err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				continue
			}
			b, err := ioutil.ReadAll(resp.Body)
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
		response.Extra.CitingCount = len(response.Citing)
		response.Extra.CitedCount = len(response.Cited)
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

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

func (s *Server) Ping() error {
	if err := s.IdentifierDatabase.Ping(); err != nil {
		return err
	}
	if err := s.OciDatabase.Ping(); err != nil {
		return err
	}
	resp, err := http.Get(s.IndexDataService)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("index data service: %s", resp.Status)
	}
	return nil
}
