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
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/slub/labe/go/ckit"
	"github.com/slub/labe/go/ckit/cache"
	"github.com/slub/labe/go/ckit/tabutils"
	"github.com/slub/labe/go/ckit/xflag"
	"github.com/thoas/stats"
)

var (
	listenAddr             = flag.String("addr", "localhost:8000", "host and port to listen on")
	identifierDatabasePath = flag.String("i", "", "identifier database path (id-doi mapping)")
	ociDatabasePath        = flag.String("o", "", "oci as a datbase path (citations)")
	enableStopWatch        = flag.Bool("stopwatch", false, "enable stopwatch")
	enableGzip             = flag.Bool("z", false, "enable gzip compression")
	enableCache            = flag.Bool("c", false, "enable caching of expensive responses")
	cacheTriggerDuration   = flag.Duration("t", 250*time.Millisecond, "cache trigger duration")
	cacheMaxFileSize       = flag.Int64("cx", 1<<36, "maximum filesize cache in bytes")
	showVersion            = flag.Bool("version", false, "show version")
	accessLogFile          = flag.String("a", "", "path to access log file, do not write access log if empty")
	logFile                = flag.String("logfile", "", "application log file (stderr if empty)")
	quiet                  = flag.Bool("q", false, "no application logging at all")

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
	var (
		logWriter                       io.Writer = os.Stderr
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
		log.Printf("[ok] setup group fetcher over %d database(s): %v",
			len(g.Backends), sqliteFetcherPaths)
	default:
		log.Fatal("need at least one sqlite3 metadata index database (-m)")
	}
	// Setup server.
	srv := &ckit.Server{
		IdentifierDatabase: identifierDatabase,
		OciDatabase:        ociDatabase,
		IndexData:          fetcher,
		Router:             mux.NewRouter(),
		StopWatchEnabled:   *enableStopWatch,
		Stats:              stats.New(),
	}
	// Setup caching. Albeit the cache will be persistant, treat it like an
	// emphemeral thing, e.g. the cache file does not survive the process.
	if *enableCache {
		f, err := ioutil.TempFile("", "labed-cache-")
		if err != nil {
			log.Fatal(err)
		}
		// Cleanup on exit and ...
		defer func() {
			f.Close()
			os.Remove(f.Name())
		}()
		// ... also react to SIGTERM (e.g. via systemd restart) with removal of
		// old file.
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-ch
			log.Printf("[..] attempting graceful shutdown")
			cerr := f.Close()
			rerr := os.Remove(f.Name())
			switch {
			case cerr != nil || rerr != nil:
				log.Printf("[xx] cleanup failed: %v %v", cerr, rerr)
				os.Exit(1)
			default:
				log.Printf("[ok] shutdown successful")
				os.Exit(0)
			}
		}()
		// Setup cache and attach to our handler.
		c, err := cache.New(f.Name())
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close()
		c.MaxFileSize = *cacheMaxFileSize
		srv.Cache = c
		srv.CacheTriggerDuration = *cacheTriggerDuration
	}
	// Setup routes.
	srv.Routes()
	// Basic reachability checks.
	if err := srv.Ping(); err != nil {
		log.Fatal(err)
	}
	// Print banner.
	fmt.Fprintln(os.Stderr, strings.Replace(Banner, `{{ .listenAddr }}`, *listenAddr, -1))
	log.Printf("[ok] labed ≋ starting %s %s http://%s", Version, Buildtime, *listenAddr)
	// Our handler.
	var h http.Handler = srv
	// Add middleware.
	if *enableGzip {
		h = handlers.CompressHandler(srv)
	}
	if *accessLogFile != "" {
		f, err := os.OpenFile(*accessLogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		h = handlers.LoggingHandler(f, h)
	}
	log.Fatal(http.ListenAndServe(*listenAddr, srv.Stats.Handler(h)))
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
