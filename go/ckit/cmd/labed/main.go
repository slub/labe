// An API server for catalog and citation data.
//
// TODO: http: URL query contains semicolon, which is no longer a supported separator; parts of the query may be stripped when parsed; see golang.org/issue/25192
// TODO: use globalconf for flags

//go:build linux

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/miku/labe/go/ckit"
	"github.com/miku/labe/go/ckit/xflag"
)

var (
	identifierDatabasePath = flag.String("I", "i.db", "identifier database path")
	ociDatabasePath        = flag.String("O", "o.db", "oci as a datbase path")
	blobServerURL          = flag.String("bs", "", "blob server URL")
	solrBlobPath           = flag.String("S", "", "solr blob URL")
	listenAddr             = flag.String("l", "localhost:3000", "host and port to listen on")
	enableStopWatch        = flag.Bool("W", false, "enable stopwatch")
	enableGzip             = flag.Bool("z", false, "enable gzip compression")
	enableLogging          = flag.Bool("L", false, "enable logging")
	enableCache            = flag.Bool("C", false, "enable in-memory caching of expensive responses")
	cacheCleanupInterval   = flag.Duration("Ct", 8*time.Hour, "cache cleanup interval")
	cacheTriggerDuration   = flag.Duration("Cg", 250*time.Millisecond, "cache trigger duration")
	cacheDefaultExpiration = flag.Duration("Cx", 72*time.Hour, "cache default expiration")
	showVersion            = flag.Bool("version", false, "show version")
	logFile                = flag.String("logfile", "", "file to log to")

	sqliteBlobPath xflag.Array // allows to specify multiple database to get catalog metadata from

	Version   string // set by makefile
	Buildtime string // set by makefile
	Help      string = `usage: labed [OPTION]

labed is an api server for the labe project; it works with three data stores
(sizes as of 01/2022):

* (1) [-I] an sqlite3 catalog id to doi translation database (around 13GB)
* (2) [-O] an sqlite3 version of OCI/COCI (around 150GB)
* (3) [-Q] a key-value mapping from catalog ids to catalog entities; these
           can be repeated (size depends on how much metadata is included)

Each database may be updated separately, with separate processes.

Examples

- http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA3My9wbmFzLjg1LjguMjQ0NA
- http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwMS9qYW1hLjI4Mi4xNi4xNTE5
- http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwNi9qbXJlLjE5OTkuMTcxNQ
- http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE3Ny8xMDQ5NzMyMzA1Mjc2Njg3
- http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU
- http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxNC9hb3MvMTE3NjM0Nzk2Mw
- http://{{ .listenAddr }}/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMjMwNy8yMDk1NTIx

Bulk requests

    $ curl -sL https://git.io/JzVmJ |
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
	flag.Var(&sqliteBlobPath, "Q", "blob sqlite3 server (repeatable)")
	flag.Usage = func() {
		fmt.Printf(strings.Replace(Help, `{{ .listenAddr }}`, *listenAddr, -1))
		fmt.Println("Flags")
		fmt.Println()
		flag.PrintDefaults()
	}
	flag.Parse()
	if *showVersion {
		fmt.Printf("labed %v %v\n", Version, Buildtime)
		os.Exit(0)
	}
	var logWriter io.Writer = os.Stdout
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("could not open log file: %v", err)
		}
		defer f.Close()
		logWriter = f
	}
	log.SetOutput(logWriter)
	// Setup database connections.
	if _, err := os.Stat(*identifierDatabasePath); os.IsNotExist(err) {
		log.Fatal(err)
	}
	if _, err := os.Stat(*ociDatabasePath); os.IsNotExist(err) {
		log.Fatal(err)
	}
	identifierDatabase, err := sqlx.Open("sqlite3", ckit.WithReadOnly(*identifierDatabasePath))
	if err != nil {
		log.Fatal(err)
	}
	ociDatabase, err := sqlx.Open("sqlite3", ckit.WithReadOnly(*ociDatabasePath))
	if err != nil {
		log.Fatal(err)
	}
	// Setup index data fetcher.
	var fetcher ckit.Fetcher
	switch {
	case *solrBlobPath != "":
		fetcher = &ckit.SolrBlob{BaseURL: *solrBlobPath}
	case *blobServerURL != "":
		fetcher = &ckit.BlobServer{BaseURL: *blobServerURL}
	case len(sqliteBlobPath) > 0:
		group := &ckit.FetchGroup{}
		if err := group.FromDatabaseFiles(sqliteBlobPath...); err != nil {
			log.Fatal(err)
		}
		fetcher = group
		log.Printf("setup group fetcher over %d databases: %v", len(group.Backends), group.Backends)
	default:
		log.Fatal("need blob server (-bs), sqlite3 database (-Q) or solr (-S)")
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
	// Setup signal handler; TODO: actually reload connections.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	go func() {
		for sig := range c {
			println(sig)
			log.Printf("TODO: reloading database connections")
		}
	}()
	log.Fatal(http.ListenAndServe(*listenAddr, h))
}
