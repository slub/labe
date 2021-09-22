package spindel

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/miku/labe/tools/spindel/set"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	IdentifierDatabase *sqlx.DB
	OciDatabase        *sqlx.DB
	IndexDataService   string
	Router             *mux.Router
}

func (s *Server) Info() error {
	var (
		info = struct {
			IdentifierDatabaseCount int `json:"identifier_database_count"`
			OciDatabaseCount        int `json:"oci_database_count"`
			IndexDataCount          int `json:"index_data_count"`
		}{}
		row   *sql.Row
		g     errgroup.Group
		funcs = []func() error{
			func() error {
				row = s.IdentifierDatabase.QueryRow("SELECT count(*) FROM map")
				return row.Scan(&info.IdentifierDatabaseCount)
			},
			func() error {
				row = s.OciDatabase.QueryRow("SELECT count(*) FROM map")
				return row.Scan(&info.OciDatabaseCount)
			},
			func() error {
				resp, err := http.Get(fmt.Sprintf("%s/count", s.IndexDataService))
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
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		// Outline
		// (1) resolve id to doi
		// (2) lookup related doi via oci
		// (3) resolve doi to ids
		// (4) lookup all ids
		// (5) assemble result
		var (
			started = time.Now()
			id      = vars["id"] // the local identifier
			doi     string       // the DOI for a given identifier
			citing  []Map        // rows containing paper references (value is the reference item)
			cited   []Map        // rows containing inbound relations (key is the citing id)
			dois    []string     // all related doi
			ids     []Map        // local identifiers
			blobs   []string     // index data blobs
		)
		// (1) Get the DOI for the local id; or get out.
		if err := s.IdentifierDatabase.Get(&doi, "SELECT v FROM map WHERE k = ?", id); err != nil {
			httpErrLog(w, err)
			return
		}
		// (2) With the DOI, find outbound (citing) and inbound (cited)
		// references in the OCI database.
		if err := s.OciDatabase.Select(&citing, "SELECT * FROM map WHERE k = ?", doi); err != nil {
			httpErrLog(w, err)
			return
		}
		if err := s.OciDatabase.Select(&cited, "SELECT * FROM map WHERE v = ?", doi); err != nil {
			httpErrLog(w, err)
			return
		}
		// (3) We want to collect the unique set of DOI to get the complete
		// indexed documents.
		for _, v := range citing {
			dois = append(dois, []string{v.Key, v.Value}...)
		}
		for _, v := range cited {
			dois = append(dois, []string{v.Key, v.Value}...)
		}
		ss := set.FromSlice(dois)
		if ss.IsEmpty() {
			// This is where the difference in the benchmark runs comes from,
			// e.g. 64860/100000; estimated ratio 64% of records with DOI will
			// have some reference information.
			return
		}
		// We now need to map back the DOI to the internal identifiers. That's
		// probably a more expensive query.
		query, args, err := sqlx.In("SELECT * FROM map WHERE v IN (?)", ss.Slice())
		if err != nil {
			httpErrLog(w, err)
			return
		}
		query = s.IdentifierDatabase.Rebind(query)
		if err := s.IdentifierDatabase.Select(&ids, query, args...); err != nil {
			httpErrLog(w, err)
			return
		}
		// For each identifier we want the full metadata. We use an local copy
		// of the index. We could also ask a life index here.
		blobStarted := time.Now()
		for _, v := range ids {
			link := fmt.Sprintf("%s/%s", s.IndexDataService, v.Key)
			resp, err := http.Get(link)
			if err != nil {
				httpErrLog(w, err)
				return
			}
			defer resp.Body.Close()
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				httpErrLog(w, err)
				return
			}
			blobs = append(blobs, string(b))
		}
		log.Printf("blob lookup took: %v", time.Since(blobStarted))
		enc := json.NewEncoder(w)
		if err := enc.Encode(map[string]interface{}{
			"blobs":     len(blobs),
			"elapsed_s": time.Since(started).Seconds(),
		}); err != nil {
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
