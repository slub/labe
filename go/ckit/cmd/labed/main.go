// An API server for catalog and citation data.
//
// TODO: http: URL query contains semicolon, which is no longer a supported
// separator; parts of the query may be stripped when parsed; see
// golang.org/issue/25192 TODO: use globalconf for flags

//go:build linux

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/miku/labe/go/ckit"
	"github.com/miku/labe/go/ckit/xflag"
)

var (
	identifierDatabasePath = flag.String("I", "", "identifier database path (id-doi mapping)")
	ociDatabasePath        = flag.String("O", "", "oci as a datbase path (citations)")
	listenAddr             = flag.String("l", "localhost:8000", "host and port to listen on")
	enableStopWatch        = flag.Bool("W", false, "enable stopwatch")
	enableGzip             = flag.Bool("z", false, "enable gzip compression")
	enableLogging          = flag.Bool("L", false, "enable logging")
	enableCache            = flag.Bool("C", false, "enable in-memory caching of expensive responses")
	cacheCleanupInterval   = flag.Duration("Ct", 8*time.Hour, "cache cleanup interval")
	cacheTriggerDuration   = flag.Duration("Cg", 250*time.Millisecond, "cache trigger duration")
	cacheDefaultExpiration = flag.Duration("Cx", 72*time.Hour, "cache default expiration")
	showVersion            = flag.Bool("version", false, "show version")
	logFile                = flag.String("logfile", "", "file to log to")
	quite                  = flag.Bool("q", false, "no output at all")

	sqliteFetcherPaths xflag.Array // allows to specify multiple database to get catalog metadata from

	Version   string // set by makefile
	Buildtime string // set by makefile
	Help      string = `usage: labed [OPTION]

labed is an web service fusing Open Citations and Library Catalog data (SLUB);
it works with three types of database files:

(1) [-I] an sqlite3 catalog id-to-doi translation database (10+G)
(2) [-O] an sqlite3 version of OCI/COCI (around 150+GB)
(3) [-Q] an sqlite3 mapping from catalog ids to catalog entities; these
         can be repeated (size depends on index size and on how much metadata is included)

Each database may be updated separately, with separate processes.

Examples

  $ labed -C -z -l localhost:1234 -I i.db -O o.db -Q d.db

  http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA3My9wbmFzLjg1LjguMjQ0NA
  http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE3Ny8xMDQ5NzMyMzA1Mjc2Njg3
  http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU
  http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxNC9hb3MvMTE3NjM0Nzk2Mw
  http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMjMwNy8yMDk1NTIx

Bulk requests

  $ curl -sL https://is.gd/xGqzsg | zstd -dc -T0 |
    parallel -j 40 "curl -s http://{{ .listenAddr }}/id/{}" |
    jq -rc '[.id, .doi, .extra.citing_count, .extra.cited_count, .extra.took] | @tsv'

`

	Banner string = `

    ___       ___       ___       ___       ___
   /\__\     /\  \     /\  \     /\  \     /\  \
  /:/  /    /::\  \   /::\  \   /::\  \   /::\  \
 /:/__/    /::\:\__\ /::\:\__\ /::\:\__\ /:/\:\__\
 \:\  \    \/\::/  / \:\::/  / \:\:\/  / \:\/:/  /
  \:\__\     /:/  /   \::/  /   \:\/  /   \::/  /
   \/__/     \/__/     \/__/     \/__/     \/__/

Examples

- http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA3My9wbmFzLjg1LjguMjQ0NA
- http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwMS9qYW1hLjI4Mi4xNi4xNTE5
- http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwNi9qbXJlLjE5OTkuMTcxNQ
- http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE3Ny8xMDQ5NzMyMzA1Mjc2Njg3
- http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU
- http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxNC9hb3MvMTE3NjM0Nzk2Mw
- http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMjMwNy8yMDk1NTIx
`
)

func main() {
	flag.Var(&sqliteFetcherPaths, "Q", "index metadata cache sqlite3 path (repeatable)")
	flag.Usage = func() {
		fmt.Printf(strings.Replace(Help, `{{ .listenAddr }}`, *listenAddr, -1))
		fmt.Println("Flags")
		fmt.Println()
		flag.PrintDefaults()
	}
	flag.Parse()
	// Show version.
	if *showVersion {
		fmt.Printf("labed %v %v\n", Version, Buildtime)
		os.Exit(0)
	}
	var (
		logWriter                       io.Writer = os.Stdout
		identifierDatabase, ociDatabase *sqlx.DB
		fetcher                         ckit.Fetcher
		err                             error
	)
	// Setup logging and log output.
	switch {
	case *quite:
		log.SetOutput(ioutil.Discard)
	default:
		if *logFile != "" {
			f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatalf("could not open log file: %v", err)
			}
			defer f.Close()
			logWriter = f
		}
		log.SetOutput(logWriter)
	}
	// Setup database connections.
	if identifierDatabase, err = openDatabase(*identifierDatabasePath); err != nil {
		log.Fatal(err)
	}
	if ociDatabase, err = openDatabase(*ociDatabasePath); err != nil {
		log.Fatal(err)
	}
	// Setup index data fetcher.
	switch {
	case len(sqliteFetcherPaths) > 0:
		g := &ckit.FetchGroup{}
		if err := g.FromFiles(sqliteFetcherPaths...); err != nil {
			log.Fatal(err)
		}
		fetcher = g
		log.Printf("setup group fetcher over %d databases: %v",
			len(g.Backends), sqliteFetcherPaths)
	default:
		log.Fatal("need sqlite3 metadata index database (-Q)")
	}
	// Setup server.
	srv := &ckit.Server{
		IdentifierDatabase:     identifierDatabase,
		OciDatabase:            ociDatabase,
		IndexData:              fetcher,
		Router:                 mux.NewRouter(),
		StopWatchEnabled:       *enableStopWatch,
		CacheEnabled:           *enableCache,
		CacheTriggerDuration:   *cacheTriggerDuration,
		CacheDefaultExpiration: *cacheDefaultExpiration,
		CacheCleanupInterval:   *cacheCleanupInterval,
	}
	srv.Routes()
	// Basic reachability checks.
	if err := srv.Ping(); err != nil {
		log.Fatal(err)
	}
	// Print banner.
	fmt.Fprintln(os.Stderr, strings.Replace(Banner, `{{ .listenAddr }}`, *listenAddr, -1))
	log.Printf("labed starting %s %s http://%s", Version, Buildtime, *listenAddr)
	// Add middleware.
	var h http.Handler = srv
	if *enableGzip {
		h = handlers.CompressHandler(srv)
	}
	if *enableLogging {
		h = handlers.LoggingHandler(logWriter, h)
	}
	log.Fatal(http.ListenAndServe(*listenAddr, h))
}

// openDatabase first ensures a file does actually exists, then create as
// read-only database connection.
func openDatabase(filename string) (*sqlx.DB, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filename)
	}
	return sqlx.Open("sqlite3", ckit.WithReadOnly(filename))
}
