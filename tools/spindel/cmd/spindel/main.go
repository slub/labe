// An experimental API server for catalog and citation data.
//
// Notes
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
// In a testrun with 100K ids, 99% of the request finish in less than 0.18s,
// 95% of the requests finish in less than 0.09s.
//
// In [14]: pd.read_csv("fixtures/t.tsv", header=None, names=["t"]).quantile([1, 0.99, 0.95, 0.75, 0.5, 0.25, 0])
// Out[14]:
//              t
// 1.00  7.140881
// 0.99  0.188159
// 0.95  0.091745
// 0.75  0.034356
// 0.50  0.016230
// 0.25  0.006562
// 0.00  0.000510
//
// In [15]: pd.read_csv("fixtures/t.tsv", header=None, names=["t"]).describe()
// Out[15]:
//                   t
// count  64860.000000
// mean       0.029028
// std        0.060780
// min        0.000510
// 25%        0.006562
// 50%        0.016230
// 75%        0.034356
// max        7.140881
//
// Another way to see performance.
//
// $ cat fixtures/100K.ids | parallel -j 40 "curl -s http://localhost:3000/q/{}" | pv -l > /dev/null
//
// Alternative sqlite3 index store. Even unoptimized slightly faster.
//
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE0Ni9hbm51cmV2LmJpLjY0LjA3MDE5NS4wMDA1MjU  0    1128  0.699212362
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8xMDQwLTM1OTAuNC4xLjI2                 0    1350  0.825507965
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAyMS9jcjA1MDk5Mng                          558  1356  0.860744752
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8wMDMzLTI5MDkuODcuMi4yNDU              0    1461  0.879090792
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzOC80MzQ2Ng                               20   1060  0.89260545
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAyMS9qYTk4MzQ5NHg                          0    1599  0.956337087
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAxNi8wMDIyLTI4MzYoNzQpOTAwMzEteA           26   1473  0.99648035
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA4Ni8yNjE3MDM                              0    1953  1.188313433
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxNC9hb3MvMTE3NjM0Nzk2Mw                   0    1774  1.412984889
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMjMwNy8yMDk1NTIx                             0    2197  1.455840794
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU                      0    2567  1.818761813
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA3My9wbmFzLjg1LjguMjQ0NA                   0    4461  2.337560631
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE3Ny8xMDQ5NzMyMzA1Mjc2Njg3                 11   8881  6.080010918
//
// One sql query.
//
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwMi9pamMuMjkxMDU0MDQxMw                   665   10   0.626366921
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA1Ni9uZWptb2EwNjE4OTQ                      905   17   0.634581836
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTEzNy8wMTA1MDAz                             587   2    0.679079214
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAyMS9qcDk3MzE4MjE                          1119  24   0.697614807
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTEyNi9zY2llbmNlLjI3Ny41MzMyLjE2NTk          928   0    0.724675178
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE0Ni9hbm51cmV2LmJpLjY0LjA3MDE5NS4wMDA1MjU  1128  0    0.741423984
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAyMS9jcjA1MDk5Mng                          1356  558  0.871165314
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8xMDQwLTM1OTAuNC4xLjI2                 1350  0    0.960690559
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzOC80MzQ2Ng                               1060  20   0.968587647
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAyMS9qYTk4MzQ5NHg                          1599  0    1.004478166
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8wMDMzLTI5MDkuODcuMi4yNDU              1461  0    1.009825454
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAxNi8wMDIyLTI4MzYoNzQpOTAwMzEteA           1473  26   1.03845787
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA4Ni8yNjE3MDM                              1953  0    1.209649958
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxNC9hb3MvMTE3NjM0Nzk2Mw                   1774  0    1.3164763449999999
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU                      2567  0    1.840246227
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA3My9wbmFzLjg1LjguMjQ0NA                   4461  0    2.690623817
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE3Ny8xMDQ5NzMyMzA1Mjc2Njg3                 8881  11   5.718691781
//
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/miku/labe/tools/spindel"
)

var (
	identifierDatabasePath = flag.String("I", "i.db", "identifier database path")
	ociDatabasePath        = flag.String("O", "o.db", "oci as a datbase path")
	blobServerURL          = flag.String("bs", "", "blob server url")
	sqliteBlobPath         = flag.String("Q", "", "sqlite3 blob index path")
	listenAddr             = flag.String("l", "localhost:3000", "host and port to listen on")
	showVersion            = flag.Bool("version", false, "show version")

	Version   string
	Buildtime string
	Help      string = `usage: spindel [-I FILE] [-O FILE] [-bs URL] [-Q FILE] [-l ADDR] [-version]

spindel is an experimental api server for labe; it works with three data stores.

* (1) an sqlite3 catalog id to doi translation table (11GB)
* (2) an sqlite3 version of OCI (145GB)
* (3) a key-value store mapping catalog ids to catalog entities (two
      implementations: 256GB microblob, 353GB sqlite3)

Each database may be updated separately, with separate processes; e.g.
currently we use the experimental mkocidb command turn (k, v) TSV files into
sqlite3 lookup databases.

Examples

- http://%s/q/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwNi9qbXJlLjE5OTkuMTcxNQ
- http://%s/q/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwMS9qYW1hLjI4Mi4xNi4xNTE5

Bulk requests

    $ curl -sL https://git.io/JzVmJ |
    parallel -j 40 "curl -s http://%s/q/{}" |
    jq -rc '[.id, .doi, .extra.citing_count, .extra.cited_count, .extra.took] | @tsv'

`

	Banner string = `

   _|_|_|            _|                  _|            _|
 _|        _|_|_|        _|_|_|      _|_|_|    _|_|    _|
   _|_|    _|    _|  _|  _|    _|  _|    _|  _|_|_|_|  _|
       _|  _|    _|  _|  _|    _|  _|    _|  _|        _|
 _|_|_|    _|_|_|    _|  _|    _|    _|_|_|    _|_|_|  _|
           _|
           _|

Examples

- http://{{ .listenAddr }}/q/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA3My9wbmFzLjg1LjguMjQ0NA
- http://{{ .listenAddr }}/q/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwMS9qYW1hLjI4Mi4xNi4xNTE5
- http://{{ .listenAddr }}/q/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwNi9qbXJlLjE5OTkuMTcxNQ
- http://{{ .listenAddr }}/q/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE3Ny8xMDQ5NzMyMzA1Mjc2Njg3
- http://{{ .listenAddr }}/q/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU
- http://{{ .listenAddr }}/q/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxNC9hb3MvMTE3NjM0Nzk2Mw
- http://{{ .listenAddr }}/q/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMjMwNy8yMDk1NTIx
`
)

func main() {
	flag.Usage = func() {
		fmt.Printf(Help, *listenAddr, *listenAddr, *listenAddr)
		fmt.Println("Flags\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *showVersion {
		fmt.Printf("spindel %v %v\n", Version, Buildtime)
		os.Exit(0)
	}
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
	var fetcher spindel.Fetcher
	switch {
	case *blobServerURL != "":
		fetcher = &spindel.BlobServer{BaseURL: *blobServerURL}
	case *sqliteBlobPath != "":
		fetcher = &spindel.SqliteBlob{Path: *sqliteBlobPath}
	default:
		log.Fatal("need blob server (-bs) or sqlite3 database (-Q)")
	}
	srv := &spindel.Server{
		IdentifierDatabase: identifierDatabase,
		OciDatabase:        ociDatabase,
		IndexData:          fetcher,
		Router:             mux.NewRouter(),
	}
	srv.Routes()
	if err := srv.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, strings.Replace(Banner, `{{ .listenAddr }}`, *listenAddr, -1))
	log.Printf("spindel starting %s %s http://%s",
		Version, Buildtime, *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, srv))
}
