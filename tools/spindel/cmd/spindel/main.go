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
// $ time cat 100K.ids | parallel -j 10 "curl -s http://localhost:3000/id/{}" | jq -rc '[.blob_count, .elapsed_s.total, .id] | @tsv'
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
// $ cat fixtures/100K.ids | parallel -j 40 "curl -s http://localhost:3000/id/{}" | pv -l > /dev/null
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
// One sql query with "in" clause; seemingly not too that much of a difference.
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
// Solr id queries are surprisingly expensive.
//
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwMi9pamMuMjkxMDU0MDQxMw    10      665     40.102183975
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwNi9hYmlvLjIwMDAuNDc1Mw    17      425     40.284360471
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAxNi9qLm1zZXIuMjAwOS4wMy4wMDE       0       447     40.311937977
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE2MS8wMS5jaXIuOTMuMS43      3       568     40.401925034
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA2My8xLjQ4NjQ3Nzg   20      819     40.764142995
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA4OC8wOTU0LTM4OTkvNDIvMTIvMTI1MDA1  87      9       43.384139584
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTEwMy9waHlzcmV2bGV0dC42Ny45Mzc       15      610     44.236279892
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAxNi9zMDE0MC02NzM2KDAyKTA5MDg4LTg   21      636     44.343999438
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTEwMy9waHlzcmV2bGV0dC4xMy40Nzk       0       946     45.504724379
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAyMS9jcjk2MDAxN3Q   288     701     45.76471719
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTEyNi9zY2llbmNlLjE1NzM5MjYw  0       759     46.569540665
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA4MC8wMTYyMTQ1OS4xOTk0LjEwNDc2ODE4  31      652     56.894786746
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzOC80MzQ2Ng        20      1060    69.66376108
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTEwMy9waHlzcmV2YS40My4yMDQ2  10      1126    73.831349012
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAxNi8wMDIyLTI4MzYoNzQpOTAwMzEteA    26      1473    77.451487135
//
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/miku/labe/tools/spindel"
)

var (
	identifierDatabasePath = flag.String("I", "i.db", "identifier database path")
	ociDatabasePath        = flag.String("O", "o.db", "oci as a datbase path")
	blobServerURL          = flag.String("bs", "", "blob server URL")
	sqliteBlobPath         = flag.String("Q", "", "sqlite3 blob index path")
	solrBlobPath           = flag.String("S", "", "solr blob URL")
	listenAddr             = flag.String("l", "localhost:3000", "host and port to listen on")
	enableStopWatch        = flag.Bool("W", false, "enable stopwatch")
	enableGzip             = flag.Bool("g", false, "enable gzip compression")
	enableLogging          = flag.Bool("L", false, "enable logging")
	showVersion            = flag.Bool("version", false, "show version")

	Version   string
	Buildtime string
	Help      string = `usage: spindel [OPTION]

spindel is an experimental api server for labe; it works with three data stores.

* (1) an sqlite3 catalog id to doi translation table (11GB)
* (2) an sqlite3 version of OCI (145GB)
* (3) a key-value store mapping catalog ids to catalog entities (two
      implementations: 256GB microblob, 353GB sqlite3)

Each database may be updated separately, with separate processes; e.g.
currently we use the experimental mkocidb command turn (k, v) TSV files into
sqlite3 lookup databases.

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

   _|_|_|            _|                  _|            _|
 _|        _|_|_|        _|_|_|      _|_|_|    _|_|    _|
   _|_|    _|    _|  _|  _|    _|  _|    _|  _|_|_|_|  _|
       _|  _|    _|  _|  _|    _|  _|    _|  _|        _|
 _|_|_|    _|_|_|    _|  _|    _|    _|_|_|    _|_|_|  _|
           _|
           _|

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

func withReadOnly(path string) string {
	return fmt.Sprintf("file:%s?mode=ro", path)
}

func main() {
	flag.Usage = func() {
		fmt.Printf(strings.Replace(Help, `{{ .listenAddr }}`, *listenAddr, -1))
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
	identifierDatabase, err := sqlx.Open("sqlite3", withReadOnly(*identifierDatabasePath))
	if err != nil {
		log.Fatal(err)
	}
	ociDatabase, err := sqlx.Open("sqlite3", withReadOnly(*ociDatabasePath))
	if err != nil {
		log.Fatal(err)
	}
	var fetcher spindel.Fetcher
	switch {
	case *solrBlobPath != "":
		fetcher = &spindel.SolrBlob{BaseURL: *solrBlobPath}
	case *blobServerURL != "":
		fetcher = &spindel.BlobServer{BaseURL: *blobServerURL}
	case *sqliteBlobPath != "":
		if _, err := os.Stat(*sqliteBlobPath); os.IsNotExist(err) {
			log.Fatal(err)
		}
		indexDatabase, err := sqlx.Open("sqlite3", withReadOnly(*sqliteBlobPath))
		if err != nil {
			log.Fatal(err)
		}
		fetcher = &spindel.SqliteBlob{DB: indexDatabase}
	default:
		log.Fatal("need blob server (-bs), sqlite3 database (-Q) or solr (-S)")
	}
	srv := &spindel.Server{
		IdentifierDatabase: identifierDatabase,
		OciDatabase:        ociDatabase,
		IndexData:          fetcher,
		Router:             mux.NewRouter(),
		StopWatchEnabled:   *enableStopWatch,
	}
	srv.Routes()
	if err := srv.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, strings.Replace(Banner, `{{ .listenAddr }}`, *listenAddr, -1))
	log.Printf("spindel starting %s %s http://%s",
		Version, Buildtime, *listenAddr)
	// Setup middleware.
	var h http.Handler = srv
	if *enableGzip {
		h = handlers.CompressHandler(srv)
	}
	if *enableLogging {
		h = handlers.LoggingHandler(os.Stdout, h)
	}
	log.Fatal(http.ListenAndServe(*listenAddr, h))
}
