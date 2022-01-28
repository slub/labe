package ckit

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/icholy/replace"
	"github.com/jmoiron/sqlx"
	"github.com/klauspost/compress/zstd"
	"github.com/segmentio/encoding/json"
	"github.com/slub/labe/go/ckit/cache"
	"github.com/slub/labe/go/ckit/set"
	"github.com/thoas/stats"
	"golang.org/x/text/transform"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// Server wraps three data sources required for index and citation data fusion.
// The IdentifierDatabase is a map from local identifier (e.g. 0-1238201) to
// DOI, the OciDatabase contains citing and cited relationships from OCI/COCI
// citation corpus and IndexData allows to fetch a metadata blob from a backing store.
//
// A performance data point: On a 8 core 16G RAM machine we can keep a
// sustained load will flat out at about 12K SQL qps, 150MB/s reads off disk.
// Total size of databases involved at about 224GB plus 7 GB cache (ie. at most
// 6% of the data can be held in memory at any time).
//
// Under load requesting the most costly 150K docs (with 32 threads) the server
// will hover at around 4% RAM (or about 640MB).
type Server struct {
	// IdentifierDatabase maps local ids to DOI. The expected schema documented
	// here: https://github.com/miku/labe/tree/main/go/ckit#makta
	IdentifierDatabase *sqlx.DB
	// OciDatabase contains DOI to DOI mappings representing a citation
	// relationship. The expected schema documented here:
	// https://github.com/miku/labe/tree/main/go/ckit#makta
	OciDatabase *sqlx.DB
	// IndexData allows to fetch a metadata blob given an identifier. This is
	// an interface that in the past has been implemented by types wrapping
	// microblob, SOLR and sqlite3, as well as a FetchGroup, that allows to
	// query multiple backends. We settled on sqlite3 and FetchGroup, the other
	// implementation are now gone.
	IndexData Fetcher
	// Router to register routes on.
	Router *mux.Router
	// StopWatchEnabled enabled the stopwatch, a builtin, simplistic request tracer.
	StopWatchEnabled bool
	// Cache for expensive items.
	Cache *cache.Cache
	// CacheTriggerDuration determines which items to cache.
	CacheTriggerDuration time.Duration
	// Stats, like request counts and status codes.
	Stats *stats.Stats
}

// Map is a generic lookup table. We use it together with sqlite3. This
// corresponds to the format generated by the makta command line tool:
// https://github.com/miku/labe/tree/main/go/ckit#makta.
type Map struct {
	Key   string `db:"k"`
	Value string `db:"v"`
}

// ErrorMessage from failed requests.
type ErrorMessage struct {
	Status int   `json:"status,omitempty"`
	Err    error `json:"err,omitempty"`
}

// Response contains a subset of index data fused with citation data.  Citing
// and cited documents are raw bytes, but typically will contain JSON. For
// unmatched docs, we only transmit the DOI, e.g. as {"doi": "10.123/123"}.
type Response struct {
	ID        string            `json:"id"`
	DOI       string            `json:"doi"`
	Citing    []json.RawMessage `json:"citing,omitempty"`
	Cited     []json.RawMessage `json:"cited,omitempty"`
	Unmatched struct {
		Citing []json.RawMessage `json:"citing,omitempty"`
		Cited  []json.RawMessage `json:"cited,omitempty"`
	} `json:"unmatched,omitempty"`
	Extra struct {
		Took                 float64 `json:"took"` // seconds
		UnmatchedCitingCount int     `json:"unmatched_citing_count"`
		UnmatchedCitedCount  int     `json:"unmatched_cited_count"`
		CitingCount          int     `json:"citing_count"`
		CitedCount           int     `json:"cited_count"`
		Cached               bool    `json:"cached"`
	} `json:"extra"`
}

// updateCounts updates extra fields containing counts.
func (r *Response) updateCounts() {
	r.Extra.CitingCount = len(r.Citing)
	r.Extra.CitedCount = len(r.Cited)
	r.Extra.UnmatchedCitingCount = len(r.Unmatched.Citing)
	r.Extra.UnmatchedCitedCount = len(r.Unmatched.Cited)
}

// Routes sets up route.
func (s *Server) Routes() {
	s.Router.HandleFunc("/", s.handleIndex()).Methods("GET")
	s.Router.HandleFunc("/cache", s.handleCacheInfo()).Methods("GET")
	s.Router.HandleFunc("/cache", s.handleCachePurge()).Methods("DELETE")
	s.Router.HandleFunc("/id/{id}", s.handleLocalIdentifier()).Methods("GET")
	s.Router.HandleFunc("/doi/{doi:.*}", s.handleDOI()).Methods("GET")
	s.Router.HandleFunc("/stats", s.handleStats()).Methods("GET")
}

// ServeHTTP turns the server into an HTTP handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

// handleIndex handles the root route.
func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		docs := `
    ___       ___       ___       ___       ___
   /\__\     /\  \     /\  \     /\  \     /\  \
  /:/  /    /::\  \   /::\  \   /::\  \   /::\  \
 /:/__/    /::\:\__\ /::\:\__\ /::\:\__\ /:/\:\__\
 \:\  \    \/\::/  / \:\::/  / \:\:\/  / \:\/:/  /
  \:\__\     /:/  /   \::/  /   \:\/  /   \::/  /
   \/__/     \/__/     \/__/     \/__/     \/__/

Pid: %d

Available endpoints:

    /              GET
    /cache         DELETE
    /cache         GET
    /doi/{doi}     GET
    /id/{id}       GET
    /stats         GET

Examples (hostport may be different):

  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA3My9wbmFzLjg1LjguMjQ0NA
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwMS9qYW1hLjI4Mi4xNi4xNTE5
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwNi9qbXJlLjE5OTkuMTcxNQ
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE3Ny8xMDQ5NzMyMzA1Mjc2Njg3
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxNC9hb3MvMTE3NjM0Nzk2Mw
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMjMwNy8yMDk1NTIx

`
		fmt.Fprintf(w, docs, os.Getpid())
	}
}

// handleCacheInfo returns the number of currently cached items.
func (s *Server) handleCacheInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.Cache != nil {
			count, err := s.Cache.ItemCount()
			if err != nil {
				httpErrLog(w, http.StatusInternalServerError, err)
				return
			}
			err = json.NewEncoder(w).Encode(map[string]interface{}{
				"count": count,
				"path":  s.Cache.Path,
			})
			if err != nil {
				httpErrLog(w, http.StatusInternalServerError, err)
				return
			}
		}
	}
}

// handleCachePurge empties the cache.
func (s *Server) handleCachePurge() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.Cache != nil {
			s.Cache.Flush()
			log.Println("flushed cached")
		}
	}
}

// handleStats renders a JSON overview of server metrics.
func (s *Server) handleStats() http.HandlerFunc {
	s.Stats.MetricsCounts = make(map[string]int)
	s.Stats.MetricsTimers = make(map[string]time.Time)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		if err := enc.Encode(s.Stats.Data()); err != nil {
			httpErrLog(w, http.StatusInternalServerError, err)
			return
		}
	}
}

// handleDOI currently only redirects to the local id handler.
func (s *Server) handleDOI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx      = r.Context()
			vars     = mux.Vars(r)
			response = &Response{
				DOI: vars["doi"],
			}
		)
		err := s.IdentifierDatabase.GetContext(ctx, &response.ID, "SELECT k FROM map WHERE v = ?", response.DOI)
		if err != nil {
			switch {
			case err == context.Canceled:
				log.Println(err)
			default:
				http.Error(w, `{"msg": "no id found", "status": 404}`, http.StatusNotFound)
			}
		} else {
			target := fmt.Sprintf("/id/%s", response.ID)
			w.Header().Set("Content-Type", "text/plain") // disable http snippet
			http.Redirect(w, r, target, http.StatusTemporaryRedirect)
		}
	}
}

// handleLocalIdentifier does all the lookups and assembles a JSON response.
func (s *Server) handleLocalIdentifier() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// (0) check for cached value
		// (1) resolve id to doi
		// (2) lookup related doi via oci
		// (3) resolve doi to ids
		// (4) lookup all ids
		// (5) include unmatched ids
		// (6) assemble result
		// (7) cache, if request was expensive
		var (
			ctx          = r.Context()
			started      = time.Now()
			vars         = mux.Vars(r)
			ids          []Map
			outbound     = set.New()
			inbound      = set.New()
			matched      []string
			unmatchedSet = set.New()
			response     = &Response{
				ID: vars["id"],
			}
			sw StopWatch
		)
		sw.SetEnabled(s.StopWatchEnabled)
		sw.Recordf("started query for: %s", response.ID)
		// Ganz sicher application/json.
		w.Header().Add("Content-Type", "application/json")
		// (0) Check cache first.
		if s.Cache != nil {
			b, err := s.Cache.Get(response.ID)
			if err == nil {
				t := time.Now()
				sw.Record("retrieved value from cache")
				r, err := zstd.NewReader(bytes.NewReader(b))
				if err != nil {
					httpErrLogf(w, http.StatusInternalServerError, "cache decompress: %w", err)
					return
				}
				var (
					took     = fmt.Sprintf(`"took":%f`, time.Since(started).Seconds())
					replacer = transform.NewReader(r, replace.RegexpString(
						regexp.MustCompile(`"took":[0-9.]+`), took),
					)
				)
				if _, err := io.Copy(w, replacer); err != nil {
					httpErrLogf(w, http.StatusInternalServerError, "cache copy: %w", err)
					return
				}
				r.Close()
				s.Stats.MeasureSinceWithLabels("cache_hit", t, nil)
				sw.Record("used cached value")
				sw.LogTable()
				return
			}
		}
		// (1) Get the DOI for the local id; or get out.
		t := time.Now()
		if err := s.IdentifierDatabase.GetContext(
			ctx, &response.DOI, "SELECT v FROM map WHERE k = ?", response.ID); err != nil {
			switch {
			case err == sql.ErrNoRows:
				log.Printf("no doi for local identifier (%s)", response.ID)
				httpErrLogf(w, http.StatusNotFound, "select id: %w", err)
			case err == context.Canceled:
				log.Printf("select id: %v", err)
				return
			default:
				httpErrLogf(w, http.StatusInternalServerError, "select id: %w", err)
			}
			return
		}
		s.Stats.MeasureSinceWithLabels("sql_query", t, nil)
		sw.Recordf("found doi for id: %s", response.DOI)
		// (2) Get outbound and inbound edges.
		citing, cited, err := s.edges(ctx, response.DOI)
		if err != nil {
			switch {
			case err == context.Canceled:
				log.Println(err)
			default:
				httpErrLogf(w, http.StatusInternalServerError, "edges: %w", err)
			}
			return
		}
		sw.Recordf("found %d outbound and %d inbound edges", len(citing), len(cited))
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
			log.Printf("no citations found for %s", response.ID)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// (4) Map relevant DOI back to local identifiers.
		if ids, err = s.mapToLocal(ctx, ss.Slice()); err != nil {
			switch {
			case err == context.Canceled:
				log.Println(err)
			default:
				httpErrLogf(w, http.StatusInternalServerError, "map: %w", err)
			}
			return
		}
		sw.Recordf("mapped %d dois back to ids", ss.Len())
		// (5) Here, we can find unmatched items, via DOI.
		for _, v := range ids {
			matched = append(matched, v.Value)
		}
		unmatchedSet = ss.Difference(set.FromSlice(matched))
		for k := range unmatchedSet {
			// We shortcut and do not use a proper JSON marshaller to save a
			// bit of time. TODO: may switch to proper JSON encoding, if other
			// parts are more optimized.
			b := []byte(fmt.Sprintf(`{"doi": %q}`, k))
			switch {
			case outbound.Contains(k):
				response.Unmatched.Citing = append(response.Unmatched.Citing, b)
			case inbound.Contains(k):
				response.Unmatched.Cited = append(response.Unmatched.Cited, b)
			default:
				// If this happens, the content of either inbound, outbound or
				// their union changed in-flight, which should not happen.
				panic("in-flight change of inbound or outbound values")
			}
		}
		sw.Record("recorded unmatched ids")
		// (6) At this point, we need to assemble the result. For each
		// identifier we want the full metadata. We use an local copy of the
		// index. We could also ask a live index here. This is agnostic to the
		// index data content, it can contain the full metadata record, or just
		// a few fields.
		for _, v := range ids {
			t := time.Now()
			b, err := s.IndexData.Fetch(v.Key)
			s.Stats.MeasureSinceWithLabels("index_data_fetch", t, nil)
			if errors.Is(err, ErrBlobNotFound) {
				continue
			}
			if err != nil {
				httpErrLogf(w, http.StatusInternalServerError, "index data fetch: %w", err)
				return
			}
			switch {
			case outbound.Contains(v.Value):
				response.Citing = append(response.Citing, b)
			case inbound.Contains(v.Value):
				response.Cited = append(response.Cited, b)
			}
		}
		sw.Recordf("fetched %d blob from index data store", len(ids))
		response.updateCounts()
		response.Extra.Took = time.Since(started).Seconds()
		switch {
		case s.Cache != nil && time.Since(started) > s.CacheTriggerDuration:
			// (7a) If this request was expensive, cache it.
			t := time.Now()
			response.Extra.Cached = true
			zbuf := bufPool.Get().(*bytes.Buffer)
			zbuf.Reset()
			// Wrap cache handling, so we can use defer to reclaim the buffer.
			wrap := func() error {
				defer bufPool.Put(zbuf)
				zw, err := zstd.NewWriter(zbuf)
				if err != nil {
					return fmt.Errorf("cache compress: %w", err)
				}
				var (
					mw  = io.MultiWriter(w, zw)
					enc = json.NewEncoder(mw)
				)
				if err := enc.Encode(response); err != nil {
					return fmt.Errorf("cache json encode: %w", err)
				}
				sw.Record("encoded json")
				if err := zw.Close(); err != nil {
					return fmt.Errorf("cache close: %w", err)
				}
				if err := s.Cache.Set(response.ID, zbuf.Bytes()); err != nil {
					log.Printf("failed to cache value for %s: %v", response.ID, err)
				} else {
					sw.Record("cached value")
				}
				s.Stats.MeasureSinceWithLabels("cached", t, nil)
				return nil
			}
			if err := wrap(); err != nil {
				httpErrLog(w, http.StatusInternalServerError, err)
				return
			}
		default:
			// (7b) Request was fast, no need to cache.
			enc := json.NewEncoder(w)
			if err := enc.Encode(response); err != nil {
				httpErrLogf(w, http.StatusInternalServerError, "encode: %w", err)
				return
			}
			sw.Record("encoded JSON")
		}
		sw.LogTable()
	}
}

// Ping returns an error, if any of the datastores is not available.
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

// edges returns citing (outbound) and cited (inbound) edges for a given DOI.
func (s *Server) edges(ctx context.Context, doi string) (citing, cited []Map, err error) {
	var t time.Time
	t = time.Now()
	if err := s.OciDatabase.SelectContext(
		ctx, &citing, "SELECT * FROM map WHERE k = ?", doi); err != nil {
		return nil, nil, err
	}
	s.Stats.MeasureSinceWithLabels("sql_query", t, nil)
	t = time.Now()
	if err := s.OciDatabase.SelectContext(
		ctx, &cited, "SELECT * FROM map WHERE v = ?", doi); err != nil {
		return nil, nil, err
	}
	s.Stats.MeasureSinceWithLabels("sql_query", t, nil)
	return citing, cited, nil
}

// mapToLocal takes a list of DOI and returns a slice of Maps containing the
// local id and DOI.
func (s *Server) mapToLocal(ctx context.Context, dois []string) (ids []Map, err error) {
	// sqlite has a limit on the variable count, which at most is 999; it may
	// lead to "too many SQL variables", SQLITE_LIMIT_VARIABLE_NUMBER (default
	// value: 999, cf:
	// https://www.daemon-systems.org/man/sqlite3_bind_blob.3.html).
	//
	//   The NNN value must be between 1 and the sqlite3_limit() parameter
	//   SQLITE_LIMIT_VARIABLE_NUMBER (default value: 999)
	//
	// I cannot say, what the optimal batch size here would be.
	for _, batch := range batchedStrings(dois, 500) {
		t := time.Now()
		query, args, err := sqlx.In("SELECT * FROM map WHERE v IN (?)", batch)
		if err != nil {
			return nil, fmt.Errorf("query (%d): %v", len(dois), err)
		}
		var result []Map
		query = s.IdentifierDatabase.Rebind(query)
		if err := s.IdentifierDatabase.SelectContext(ctx, &result, query, args...); err != nil {
			return nil, fmt.Errorf("select (%d): %v", len(dois), err)
		}
		s.Stats.MeasureSinceWithLabels("sql_query", t, nil)
		for _, r := range result {
			ids = append(ids, r)
		}
	}
	return ids, nil
}

// batchedStrings batches one string slice into a potentially smaller number of
// strings slices with size at most n.
func batchedStrings(ss []string, n int) (result [][]string) {
	b, e := 0, n
	for {
		if len(ss) <= e {
			result = append(result, ss[b:])
			return
		} else {
			result = append(result, ss[b:e])
			b, e = e, e+n
		}
	}
	return
}

// httpErrLogf is a log formatting helper.
func httpErrLogf(w http.ResponseWriter, status int, s string, a ...interface{}) {
	httpErrLog(w, status, fmt.Errorf(s, a...))
}

// httpErrLogStatus returns an error to the client and logs the error.
func httpErrLog(w http.ResponseWriter, status int, err error) {
	log.Printf("failed [%d]: %v", status, err)
	b, err := json.Marshal(&ErrorMessage{
		Status: status,
		Err:    err,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Error(w, string(b), status)
}
