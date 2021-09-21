package main

import (
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
)

var (
	identifierDatabasePath = flag.String("I", "i.db", "identifier database path")
	ociDatabasePath        = flag.String("O", "o.db", "oci as a datbase path")
	indexDataBaseURL       = flag.String("D", "http://localhost:8820", "index data lookup base URL")
	listen                 = flag.String("l", "localhost:3000", "host and port to listen on")
)

// Map is a generic lookup table.
type Map struct {
	Key   string `db:"k"`
	Value string `db:"v"`
}

type server struct {
	identifierDatabase *sqlx.DB
	ociDatabase        *sqlx.DB
	indexDataService   string
	router             *mux.Router
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
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		// (1) resolve id to doi
		// (2) lookup related doi via oci
		// (3) resolve doi to ids
		// (4) lookup all ids
		// (5) assemble result
		started := time.Now()

		id := vars["id"]
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
		log.Printf("collected index data for %s [%d] in %v", id, len(blobs), time.Since(started))
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
	srv.routes()
	log.Printf("spindel http://%s", *listen)
	log.Fatal(http.ListenAndServe(*listen, srv))
}
