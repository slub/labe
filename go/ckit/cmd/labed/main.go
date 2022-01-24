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
	"github.com/slub/labe/go/ckit"
	"github.com/slub/labe/go/ckit/cache"
	"github.com/slub/labe/go/ckit/tabutils"
	"github.com/slub/labe/go/ckit/xflag"
)

var (
	identifierDatabasePath = flag.String("i", "", "identifier database path (id-doi mapping)")
	ociDatabasePath        = flag.String("o", "", "oci as a datbase path (citations)")
	listenAddr             = flag.String("addr", "localhost:8000", "host and port to listen on")
	enableStopWatch        = flag.Bool("stopwatch", false, "enable stopwatch")
	enableGzip             = flag.Bool("z", false, "enable gzip compression")
	enableLogging          = flag.Bool("l", false, "enable logging")
	enableCache            = flag.Bool("c", false, "enable caching of expensive responses")
	cacheTriggerDuration   = flag.Duration("t", 250*time.Millisecond, "cache trigger duration")
	showVersion            = flag.Bool("version", false, "show version")
	logFile                = flag.String("logfile", "", "file to log to")
	quiet                  = flag.Bool("q", false, "no output at all")
	warmCache              = flag.Bool("warm-cache", false, "warm cache, read one DOI per line from stdin")

	sqliteFetcherPaths xflag.Array // allows to specify multiple database to get catalog metadata from

	Version   string // set by makefile
	Buildtime string // set by makefile
	Help      string = `usage: labed [OPTION]

labed is an web service fusing Open Citation and Library Catalog data (SLUB);
it works with three types of databases:

(1) [-i] an sqlite3 catalog-id-to-doi translation database (10G+)
(2) [-o] an sqlite3 version of OCI/COCI (150GB+)
(3) [-m] an sqlite3 mapping from catalog ids to (json) metadata; this can be repeated
         (size depends on index size and on how much metadata is included) (40-350G)

Each database may be updated separately, with separate processes.

Examples

  $ labed -c -z -addr localhost:1234 -i i.db -o o.db -m d.db

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

Examples:

  http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA3My9wbmFzLjg1LjguMjQ0NA
  http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwMS9qYW1hLjI4Mi4xNi4xNTE5
  http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwNi9qbXJlLjE5OTkuMTcxNQ
  http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE3Ny8xMDQ5NzMyMzA1Mjc2Njg3
  http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU
  http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxNC9hb3MvMTE3NjM0Nzk2Mw
  http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMjMwNy8yMDk1NTIx
`
)

func main() {
	flag.Var(&sqliteFetcherPaths, "m", "index metadata cache sqlite3 path (repeatable)")
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
	if *warmCache {
		if err := ckit.WarmCache(os.Stdin, *listenAddr); err != nil {
			log.Fatal(err)
		}
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
	case *quiet:
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
		IdentifierDatabase: identifierDatabase,
		OciDatabase:        ociDatabase,
		IndexData:          fetcher,
		Router:             mux.NewRouter(),
		StopWatchEnabled:   *enableStopWatch,
	}
	// Setup caching.
	if *enableCache {
		f, err := ioutil.TempFile("", "labed-cache-")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		c, err := cache.New(f.Name())
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close()
		srv.Cache = c
		srv.CacheTriggerDuration = *cacheTriggerDuration
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
// read-only connection.
func openDatabase(filename string) (*sqlx.DB, error) {
	if len(filename) == 0 {
		return nil, fmt.Errorf("empty file")
	}
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filename)
	}
	return sqlx.Open("sqlite3", tabutils.WithReadOnly(filename))
}
