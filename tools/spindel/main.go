// $ time spindel -info | jq .
// 2021/09/22 13:56:54 ⚑ querying three data stores ...
// {
//   "identifier_database_count": 56879665,
//   "oci_database_count": 1119201441,
//   "index_data_count": 61529978
// }
//
// real    0m20.467s
// user    0m2.366s
// sys     0m18.477s
//
// Some stats; 179 rps, random requests, no parallel requests within the
// server (e.g. no parallel index data requests).
//
// 256G index data, minimal caching.
//
// $ pcstat index.data
// +------------+----------------+------------+-----------+---------+
// | Name       | Size (bytes)   | Pages      | Cached    | Percent |
// |------------+----------------+------------+-----------+---------|
// | index.data | 256360643273   | 62588048   | 885414    | 001.415 |
// +------------+----------------+------------+-----------+---------+
//
// $ time cat 100K.ids | parallel -j 10 "curl -s http://localhost:3000/q/{}" | jq -rc '[.blob_count, .elapsed_s.total, .id] | @tsv'
//
// ...
// 3       0.007706688     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAyMy9iOmZyYWMuMDAwMDAzMTA5My44NTA2MS5jOQ
// 2       0.003119393     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE0My9wdHBzLjE1NC4xNTQ
// 31      0.032643698     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAxNi9zMDAwMy0zNDcyKDgwKTgwMDI2LTE
// 74      0.056817144     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwMi9hZG1hLjIwMDUwMDk2Mw
// 24      0.026795839     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwNi9qbXJlLjE5OTkuMTcxNQ
// 11      0.011584241     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTEzNC9zMDAzNzQ0NjYxNTAyMDAyMA
// 9       0.020058694     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA2My8xLjQ5MjIxNDI
// 5       0.006531584     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA4MC8wMjc3MzgxODkwODA1MDMwNw
// 828     1.042512224     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwMS9qYW1hLjI4Mi4xNi4xNTE5
// 166     0.210055634     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA4Ni8xNDMzMjM
// 35      0.049543236     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA4OC8xNzQ4LTMxOTAvYWJhYmIw
// 12      0.01719839      ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA4OC8xNzU1LTEzMTUvNTg4LzMvMDMyMDIx
// 15      0.024281162     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMjMwNy8yMzk5MTc4
// 18      0.018429985     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAxNi8wMzc4LTUxNzMoODApOTAxMzgtNg
// ...
//
// real    9m18.799s
// user    16m58.769s
// sys     14m19.567s
//
// Somewhat predictable performance, with the slowest out of 100K requests:
//
// 696     0.7669515       ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAxNi8wMDAzLTQ5MTYoNjMpOTAwNzgtMg
// 990     0.767538368     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAyMS9jcjk2MDAxN3Q
// 947     0.782944785     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTEwMy9waHlzcmV2bGV0dC4xMy40Nzk
// 877     0.792816803     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTEyNi9zY2llbmNlLjc3MzIzODI
// 1020    0.805743207     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwMi9zaW0uNDA4NQ
// 926     0.826975227     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA3My9wbmFzLjEwMzI5MTMxMDA
// 956     0.841350091     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTM2NC9qb3NhYi4xMy4wMDA0ODE
// 923     0.861391752     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA1Ni9uZWptb2EwNjE4OTQ
// 1137    0.878462042     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTEwMy9waHlzcmV2YS40My4yMDQ2
// 929     0.934591869     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTEyNi9zY2llbmNlLjI3Ny41MzMyLjE2NTk
// 1144    1.032953765     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAyMS9qcDk3MzE4MjE
// 1500    1.163330572     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAxNi8wMDIyLTI4MzYoNzQpOTAwMzEteA
// 1352    1.208945268     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8xMDQwLTM1OTAuNC4xLjI2
// 1129    1.230243451     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE0Ni9hbm51cmV2LmJpLjY0LjA3MDE5NS4wMDA1MjU
// 1463    1.276817506     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8wMDMzLTI5MDkuODcuMi4yNDU
// 1600    1.413687315     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAyMS9qYTk4MzQ5NHg
// 1915    1.426443495     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAyMS9jcjA1MDk5Mng
// 1954    1.529097444     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA4Ni8yNjE3MDM
// 2198    1.973452902     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMjMwNy8yMDk1NTIx
// 1775    2.225739231     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxNC9hb3MvMTE3NjM0Nzk2Mw
// 2568    2.449228354     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU
// 4462    3.477182936     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA3My9wbmFzLjg1LjguMjQ0NA
// 8893    8.461666468     ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE3Ny8xMDQ5NzMyMzA1Mjc2Njg3
//
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/sync/errgroup"
)

var (
	identifierDatabasePath = flag.String("I", "i.db", "identifier database path")
	ociDatabasePath        = flag.String("O", "o.db", "oci as a datbase path")
	indexDataBaseURL       = flag.String("D", "http://localhost:8820", "index data lookup base URL")
	listen                 = flag.String("l", "localhost:3000", "host and port to listen on")
	showInfo               = flag.Bool("info", false, "show db info only")
)

// Map is a generic lookup table.
type Map struct {
	//
	// Somewhat predictable performance, with the slowest out of 100K requests:

	Key   string `db:"k"`
	Value string `db:"v"`
}

type server struct {
	identifierDatabase *sqlx.DB
	ociDatabase        *sqlx.DB
	indexDataService   string
	router             *mux.Router
}

func (s *server) Info() error {
	var (
		info = struct {
			IdentifierDatabaseCount int `json:"identifier_database_count"`
			OciDatabaseCount        int `json:"oci_database_count"`
			IndexDataCount          int `json:"index_data_count"`
		}{}
		row *sql.Row
		g   errgroup.Group
	)

	var funcs = []func() error{
		func() error {
			row = s.identifierDatabase.QueryRow("SELECT count(*) FROM map")
			return row.Scan(&info.IdentifierDatabaseCount)
		},
		func() error {
			row = s.ociDatabase.QueryRow("SELECT count(*) FROM map")
			return row.Scan(&info.OciDatabaseCount)
		},
		func() error {
			resp, err := http.Get(fmt.Sprintf("%s/count", s.indexDataService))
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

func (s *server) routes() {
	s.router.HandleFunc("/", s.handleIndex())
	s.router.HandleFunc("/q/{id}", s.handleQuery())
}

func (s *server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "spindel")
	}
}

func (s *server) handleQuery() http.HandlerFunc {
	type benchStat = struct {
		Identifier     string    `json:"id"`
		Started        time.Time `json:"started"`
		BlobCount      int       `json:"blob_count"`
		ElapsedSeconds struct {
			IdentifierDatabaseLookup float64 `json:"identifier_database"`
			OciDatabaseLookup        float64 `json:"oci_database"`
			IndexDataLookup          float64 `json:"index_data"`
			Total                    float64 `json:"total"`
		} `json:"elapsed_s"`
		ElapsedRatio struct {
			IdentifierDatabaseLookup float64 `json:"identifier_database"`
			OciDatabaseLookup        float64 `json:"oci_database"`
			IndexDataLookup          float64 `json:"index_data"`
		} `json:"elapsed_r"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		// (1) resolve id to doi
		// (2) lookup related doi via oci
		// (3) resolve doi to ids
		// (4) lookup all ids
		// (5) assemble result
		started := time.Now()
		stat := benchStat{Started: started}

		id := vars["id"]
		stat.Identifier = id
		// (1)
		var m Map
		if err := s.identifierDatabase.Get(&m, "SELECT * FROM map WHERE k = ?", id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// (2)
		var (
			doi    = m.Value
			citing []Map
			cited  []Map
		)
		if err := s.ociDatabase.Select(&citing, "SELECT * FROM map WHERE k = ?", doi); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := s.ociDatabase.Select(&cited, "SELECT * FROM map WHERE v = ?", doi); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		stat.ElapsedSeconds.IdentifierDatabaseLookup = time.Since(started).Seconds()
		// log.Println(m)
		// log.Println(citing)
		// log.Println(cited)
		// log.Println(time.Since(started)) // 3-12ms

		// (3)
		var dois []string
		for _, v := range citing {
			dois = append(dois, v.Key)
			dois = append(dois, v.Value)
		}
		for _, v := range cited {
			dois = append(dois, v.Key)
			dois = append(dois, v.Value)
		}
		ss := FromSlice(dois)
		// log.Printf("%d dois to lookup", ss.Len())
		if ss.IsEmpty() {
			return
		}
		query, args, err := sqlx.In("SELECT * FROM map WHERE v IN (?)", ss.Slice())
		if err != nil {
			http.Error(w, "in: "+err.Error(), http.StatusInternalServerError)
			return
		}
		query = s.identifierDatabase.Rebind(query)
		var ids []Map
		if err := s.identifierDatabase.Select(&ids, query, args...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		stat.ElapsedSeconds.OciDatabaseLookup = time.Since(started).Seconds()
		// log.Println(ids) // the keys are our local ids
		var blobs []string
		for _, v := range ids {
			link := fmt.Sprintf("%s/%s", s.indexDataService, v.Key)
			resp, err := http.Get(link)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			blobs = append(blobs, string(b))
		}
		stat.BlobCount = len(blobs)
		stat.ElapsedSeconds.IndexDataLookup = time.Since(started).Seconds()
		// log.Printf("collected index data for %s [%d] in %v", id, len(blobs), time.Since(started))
		// XXX: calculate ratio
		stat.ElapsedSeconds.Total = time.Since(started).Seconds()
		stat.ElapsedRatio.IdentifierDatabaseLookup = stat.ElapsedSeconds.IdentifierDatabaseLookup / stat.ElapsedSeconds.Total
		stat.ElapsedRatio.OciDatabaseLookup = (stat.ElapsedSeconds.OciDatabaseLookup -
			stat.ElapsedSeconds.IdentifierDatabaseLookup) / stat.ElapsedSeconds.Total
		stat.ElapsedRatio.IndexDataLookup = (stat.ElapsedSeconds.IndexDataLookup -
			stat.ElapsedSeconds.OciDatabaseLookup) / stat.ElapsedSeconds.Total
		enc := json.NewEncoder(w)
		if err := enc.Encode(stat); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) Ping() error {
	if err := s.identifierDatabase.Ping(); err != nil {
		return err
	}
	if err := s.ociDatabase.Ping(); err != nil {
		return err
	}
	resp, err := http.Get(s.indexDataService)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("index data service: %s", resp.Status)
	}
	return nil
}

// Set implements basic string set operations, not thread-safe.
type Set map[string]struct{}

// New creates a new set.
func NewSet() Set {
	var s = make(Set)
	return s
}

// FromSlice initializes a set from a slice.
func FromSlice(vs []string) Set {
	s := NewSet()
	for _, v := range vs {
		s.Add(v)
	}
	return s
}

// Clear removes all elements.
func (s Set) Clear() {
	for k := range s {
		delete(s, k)
	}
}

// Add adds an element.
func (s Set) Add(v string) Set {
	s[v] = struct{}{}
	return s
}

// Len returns number of elements in set.
func (s Set) Len() int {
	return len(s)
}

// IsEmpty returns if set has zero elements.
func (s Set) IsEmpty() bool {
	return s.Len() == 0
}

// Equals returns true, if sets contain the same elements.
func (s Set) Equals(t Set) bool {
	for k := range s {
		if !t.Contains(k) {
			return false
		}
	}
	return s.Len() == t.Len()
}

// Contains returns membership status.
func (s Set) Contains(v string) bool {
	_, ok := (s)[v]
	return ok
}

// Intersection returns a new set containing all elements found in both sets.
func (s Set) Intersection(t Set) Set {
	u := NewSet()
	for k := range s {
		if t.Contains(k) {
			u.Add(k)
		}
	}
	return u
}

// Union returns the union of two sets.
func (s Set) Union(t Set) Set {
	u := NewSet()
	for k := range s {
		u.Add(k)
	}
	for k := range t {
		u.Add(k)
	}
	return u
}

// Slice returns all elements as a slice.
func (s Set) Slice() (result []string) {
	for k := range s {
		result = append(result, k)
	}
	return
}

// Sorted returns all elements as a slice, sorted.
func (s Set) Sorted() (result []string) {
	for k := range s {
		result = append(result, k)
	}
	sort.Strings(result)
	return
}

func main() {
	flag.Parse()
	if _, err := os.Stat(*identifierDatabasePath); os.IsNotExist(err) {
		log.Fatal(err)
	}
	if _, err := os.Stat(*ociDatabasePath); os.IsNotExist(err) {
		log.Fatal(err)
	}
	identifierDatabase, err := sqlx.Open("sqlite3", *identifierDatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	ociDatabase, err := sqlx.Open("sqlite3", *ociDatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	srv := &server{
		identifierDatabase: identifierDatabase,
		ociDatabase:        ociDatabase,
		indexDataService:   *indexDataBaseURL,
		router:             mux.NewRouter(),
	}
	if err := srv.Ping(); err != nil {
		log.Fatal(err)
	}
	if *showInfo {
		if err := srv.Info(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}
	srv.routes()
	log.Printf("spindel http://%s", *listen)
	log.Fatal(http.ListenAndServe(*listen, srv))
}
