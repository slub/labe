# ckit

Citation graph kit for the LABE project at [SLUB Dresden](https://www.slub-dresden.de/).

* [labed](#labed), an HTTP server serving Open Citations data fused with catalog metadata
* [tabjson](#tabjson), turn JSON into TSV
* [makta](#makta), turn TSV files into sqlite3 databases

To build all binaries, run:

```
$ make
```

To build a debian package, run:

```
$ make deb
```

To cleanup all artifacts, run:

```
$ make clean
```

## labed

![](static/45582_reading_lg.gif)

Experimental API server, takes requests for a given id and returns a
result fused from OCI citations and index data.

It currently needs three sqlite3 database.

```
$ ls -lLh i.db o.db index.db
-rw-r--r-- 1 tir tir  11G Sep 20 23:13 i.db
-rw-r--r-- 1 tir tir 330G Sep 26 22:32 index.db
-r--r--r-- 1 tir tir 145G Sep 21 00:36 o.db
```

May become 4 or 5 databases in the future with various update cycles:

* main index copy (W)
* main index id-doi mapping (W)
* ai index copy (M)
* ai index id-doi mapping (M)
* oci copy (as updated)

Random timings (TODO: performance report):

```
ai-49-aHR0cD...TAuMTEwNC9wc...    10.1104/pp.88.4.1411           0   33   0.011371553
ai-49-aHR0cD...TAuMTc1NzYva...    10.17576/jsm-2019-4808-23      0   3    0.002403981
ai-49-aHR0cD...TAuMTYxNC93d...    10.1614/wt-08-045.1            19  12   0.006658463
ai-49-aHR0cD...TAuMzg5Ny96b...    10.3897/zookeys.449.6813       0   1    0.000609854
ai-49-aHR0cD...TAuMTA4OC8xN...    10.1088/1757-899x/768/5/052105 2   0    0.000913447
ai-49-aHR0cD...TAuNTgxMS9jc...    10.5811/cpcem.2019.7.43632     1   0    0.047257667
ai-49-aHR0cD...TAuMTEwMy9wa...    10.1103/physrevc.49.3061       27  4    0.008262996
ai-49-aHR0cD...TAuMTM3MS9qb...    10.1371/journal.pone.0077786   38  15   0.018779194
ai-49-aHR0cD...TAuMTAwMi9sZ...    10.1002/ldr.3400040418         2   0    0.000982242
ai-49-aHR0cD...TAuMTEwMy9wa...    10.1103/physrevlett.81.3187    15  14   0.007743473
ai-49-aHR0cD...TAuMTAwMi9ub...    10.1002/nme.1620300822         7   6    0.004755116
ai-49-aHR0cD...TAuMTM3MS9qb...    10.1371/journal.pcbi.1002234   54  4    0.018582831
ai-49-aHR0cD...TAuMTAxNi8wM...    10.1016/0165-4896(94)00731-4   5   4    0.004127696
ai-49-aHR0cD...TAuMTA5My9qe...    10.1093/jxb/49.318.21          0   0    0.000267756
ai-49-aHR0cD...TAuMTE0Mi9zM...    10.1142/s0218126619500051      22  2    0.006445901
ai-49-aHR0cD...TAuNzg2MS9jb...    10.7861/clinmedicine.17-4-332  13  8    0.005840636
ai-49-aHR0cD...TAuMTM3My9jb...    10.1373/clinchem.2013.204446   20  11   0.011903923
ai-49-aHR0cD...TAuMTE0My9qa...    10.1143/jjap.9.958             0   7    0.002963267
ai-49-aHR0cD...TAuMTAyMS9hb...    10.1021/am8001605              29  64   0.022973696
ai-49-aHR0cD...TAuMTIwNy9zM...    10.1207/s15326934crj1401_1     0   21   0.056867545
```

### Usage

```sh
usage: labed [OPTION]

labed is an experimental api server for labe; it works with three data stores.

* (1) an sqlite3 catalog id to doi translation table (11GB)
* (2) an sqlite3 version of OCI (145GB)
* (3) a key-value store mapping catalog ids to catalog entities (two
      implementations: 256GB microblob, 353GB sqlite3)

Each database may be updated separately, with separate processes; e.g.
currently we use the experimental mkocidb command turn (k, v) TSV files into
sqlite3 lookup databases.

Examples

- http://localhost:3000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA3My9wbmFzLjg1LjguMjQ0NA
- http://localhost:3000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwMS9qYW1hLjI4Mi4xNi4xNTE5
- http://localhost:3000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwNi9qbXJlLjE5OTkuMTcxNQ
- http://localhost:3000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE3Ny8xMDQ5NzMyMzA1Mjc2Njg3
- http://localhost:3000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU
- http://localhost:3000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxNC9hb3MvMTE3NjM0Nzk2Mw
- http://localhost:3000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMjMwNy8yMDk1NTIx

Bulk requests

    $ curl -sL https://git.io/JzVmJ |
    parallel -j 40 "curl -s http://localhost:3000/id/{}" |
    jq -rc '[.id, .doi, .extra.citing_count, .extra.cited_count, .extra.took] | @tsv'

Flags

  -C    enable in-memory caching of expensive responses
  -Cg duration
        cache trigger duration (default 250ms)
  -Ct duration
        cache ttl (default 8h0m0s)
  -Cx duration
        cache default expiration (default 72h0m0s)
  -I string
        identifier database path (default "i.db")
  -L    enable logging
  -O string
        oci as a datbase path (default "o.db")
  -Q string
        sqlite3 blob index path
  -S string
        solr blob URL
  -W    enable stopwatch
  -bs string
        blob server URL
  -l string
        host and port to listen on (default "localhost:3000")
  -version
        show version
  -z    enable gzip compression
```

### Using a stopwatch

Experimental `-W` flag to trace duration of various operations.

```sh
$ labed -W -bs http://localhost:8820

    ___       ___       ___       ___       ___
   /\__\     /\  \     /\  \     /\  \     /\  \
  /:/  /    /::\  \   /::\  \   /::\  \   /::\  \
 /:/__/    /::\:\__\ /::\:\__\ /::\:\__\ /:/\:\__\
 \:\  \    \/\::/  / \:\::/  / \:\:\/  / \:\/:/  /
  \:\__\     /:/  /   \::/  /   \:\/  /   \::/  /
   \/__/     \/__/     \/__/     \/__/     \/__/

Examples

- http://localhost:3000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA3My9wbmFzLjg1LjguMjQ0NA
- http://localhost:3000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwMS9qYW1hLjI4Mi4xNi4xNTE5
- http://localhost:3000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwNi9qbXJlLjE5OTkuMTcxNQ
- http://localhost:3000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE3Ny8xMDQ5NzMyMzA1Mjc2Njg3
- http://localhost:3000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU
- http://localhost:3000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxNC9hb3MvMTE3NjM0Nzk2Mw
- http://localhost:3000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMjMwNy8yMDk1NTIx

2021/09/29 17:35:19 labed starting 3870a68 2021-09-29T15:34:00Z http://localhost:3000
2021/09/29 17:35:20 timings for XVlB

> XVlB    0    0s             0.00    started query for: ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA5OC9yc3BhLjE5OTguMDE2NA
> XVlB    1    397.191µs      0.00    found doi for id: 10.1098/rspa.1998.0164
> XVlB    2    481.676µs      0.01    found 8 citing items
> XVlB    3    18.984627ms    0.23    found 456 cited items
> XVlB    4    13.421306ms    0.16    mapped 464 dois back to ids
> XVlB    5    494.163µs      0.01    recorded unmatched ids
> XVlB    6    44.093361ms    0.52    fetched 302 blob from index data store
> XVlB    7    6.422462ms     0.08    encoded JSON
> XVlB    -    -              -       -
> XVlB    S    84.294786ms    1.0     total
```


### TODO

* [ ] a better name, e.g. labesrv, labesvc, cdfuse, catfuse, labed, ...
* [ ] a detailed performance report
* [ ] tools or scripts to generate the input database from scratch

----

## tabjson

> A non-generic, quick JSON to TSV converter

Turns jsonlines with an `id` field into `(id, doc)` TSV. We want to take an
index snapshot, extract the id, create a TSV and then put it into sqlite3 so we
can serve queries.

This is nothing, `jq` could not do, e.g. via (might need escaping).

```
$ jq -rc '[.id, (. | tojson)] | @tsv'
```

TODO: remove or use compression.


```sh
$ tabjson -h
Usage of tabjson:
  -C    compress value; gz+b64
  -T    emit table showing possible savings through compression
```

Examples.

```
$ head -1 ../../data/index.data | tabjson
ai-49-aHR0...uMi4xNTU       {"access_facet":"Electronic Resourc ... }
```

Around 80K docs/s. Supports some basic compression schema.

```sh
$ head -1 ../../data/index.data | tabjson -C
ai-49-aHR0...uMi4xNTU       H4sIAAAAAAAA/6xWTY/bNhC991cQPCWArJVUK177...
```

Display savings by using compression.

```sh
$ head ../../data/index.data | tabjson -C -T | column -t
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8wMDIxLTkwMTAuNjIuMi4xNTU  2223  1005  0.452092
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA3MDY1OQ               2217  993   0.447903
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8wMDIxLTkwMTAuNjIuMi4xNDY  2172  969   0.446133
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA0OTMxNQ               2325  1037  0.446022
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA3NDk0Ng               2340  1037  0.443162
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzOC9zai9sZXUvMjQwMTI4Mg       1841  973   0.528517
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA3Mzk5Mg               2118  937   0.442398
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzOC9zai9sZXUvMjQwMTI4Mw       1841  973   0.528517
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzOC9zai9sZXUvMjQwMTI4NA       1841  973   0.528517
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA0MDIzMg               2148  989   0.460428
```

We want this to put our index data into a key value store. Compression (with
gzip) seems 3-4 slower than using no compression. Also, compression will need
extra handling at request time.

----

## makta

> **mak**e a database from **ta**bular data

Turn [tabular data](https://en.wikipedia.org/wiki/Tab-separated_values) into a
lookup table using [sqlite3](https://sqlite.org/). This is a working PROTOTYPE
with limitations, e.g. no customizations, the table definition is fixed, etc.

> CREATE TABLE IF NOT EXISTS map (k TEXT, v TEXT)

As a performance data point, an example dataset with 1B+ rows can be inserted
and indexed in less than two hours (on a [recent
CPU](https://ark.intel.com/content/www/us/en/ark/products/122589/intel-core-i7-8550u-processor-8m-cache-up-to-4-00-ghz.html)
and an [nvme](https://en.wikipedia.org/wiki/NVM_Express) drive; database file
size: 400G).

![](static/443238.gif)

### Installation

> [https://github.com/miku/labe/releases](https://github.com/miku/labe/releases) (wip)

```sh
$ go install github.com/miku/labe/go/ckit/cmd/makta@latest
```

### How it works

Data is chopped up into smaller chunks (defaults to about 64MB) and imported with
the `.import` [command](https://www.sqlite.org/cli.html). Indexes are created
only after all data has been imported.

### Example

```sh
$ cat fixtures/sample-xs.tsv | column -t
10.1001/10-v4n2-hsf10003                    10.1177/003335490912400218
10.1001/10-v4n2-hsf10003                    10.1097/01.bcr.0000155527.76205.a2
10.1001/amaguidesnewsletters.1996.novdec01  10.1056/nejm199312303292707
10.1001/amaguidesnewsletters.1996.novdec01  10.1016/s0363-5023(05)80265-5
10.1001/amaguidesnewsletters.1996.novdec01  10.1001/jama.1994.03510440069036
10.1001/amaguidesnewsletters.1997.julaug01  10.1097/00007632-199612150-00003
10.1001/amaguidesnewsletters.1997.mayjun01  10.1164/ajrccm/147.4.1056
10.1001/amaguidesnewsletters.1997.mayjun01  10.1136/thx.38.10.760
10.1001/amaguidesnewsletters.1997.mayjun01  10.1056/nejm199507133330207
10.1001/amaguidesnewsletters.1997.mayjun01  10.1378/chest.88.3.376

$ makta -o xs.db < fixtures/sample-xs.tsv
2021/10/04 16:13:06 [ok] initialized database · xs.db
2021/10/04 16:13:06 [io] written 679B · 361.3K/s
2021/10/04 16:13:06 [ok] 1/2 created index · xs.db
2021/10/04 16:13:06 [ok] 2/2 created index · xs.db

$ sqlite3 xs.db 'select * from map'
10.1001/10-v4n2-hsf10003|10.1177/003335490912400218
10.1001/10-v4n2-hsf10003|10.1097/01.bcr.0000155527.76205.a2
10.1001/amaguidesnewsletters.1996.novdec01|10.1056/nejm199312303292707
10.1001/amaguidesnewsletters.1996.novdec01|10.1016/s0363-5023(05)80265-5
10.1001/amaguidesnewsletters.1996.novdec01|10.1001/jama.1994.03510440069036
10.1001/amaguidesnewsletters.1997.julaug01|10.1097/00007632-199612150-00003
10.1001/amaguidesnewsletters.1997.mayjun01|10.1164/ajrccm/147.4.1056
10.1001/amaguidesnewsletters.1997.mayjun01|10.1136/thx.38.10.760
10.1001/amaguidesnewsletters.1997.mayjun01|10.1056/nejm199507133330207
10.1001/amaguidesnewsletters.1997.mayjun01|10.1378/chest.88.3.376

$ sqlite3 xs.db 'select * from map where k = "10.1001/amaguidesnewsletters.1997.mayjun01" '
10.1001/amaguidesnewsletters.1997.mayjun01|10.1164/ajrccm/147.4.1056
10.1001/amaguidesnewsletters.1997.mayjun01|10.1136/thx.38.10.760
10.1001/amaguidesnewsletters.1997.mayjun01|10.1056/nejm199507133330207
10.1001/amaguidesnewsletters.1997.mayjun01|10.1378/chest.88.3.376
```

### Motivation

> SQLite is likely used more than all other database engines combined. Billions
> and billions of copies of SQLite exist in the wild. -- [https://www.sqlite.org/mostdeployed.html](https://www.sqlite.org/mostdeployed.html)

Sometimes, programs need lookup tables to map values between two domains. A
[dictionary](https://xlinux.nist.gov/dads/HTML/dictionary.html) is a perfect
data structure as long as the data fits in memory. For larger sets (hundreds of
millions of entries), a dictionary may not work.

The *makta* tool currently takes a two-column TSV and turns it into an sqlite3
database, which you can query in your program. Depending on a couple of
factors, you maybe be able to query the lookup database with about 1-50K
queries per second.

Finally, sqlite3 is just an awesome database and [recommeded storage
format](https://www.sqlite.org/locrsf.html).

### Usage

```sh
$ makta -h
Usage of makta:
  -B int
        buffer size (default 67108864)
  -C int
        sqlite3 cache size, needs memory = C x page size (default 1000000)
  -I int
        index mode: 0=none, 1=k, 2=v, 3=kv (default 3)
  -o string
        output filename (default "data.db")
  -version
        show version and exit
```

### Performance

```sh
$ wc -l fixtures/sample-10m.tsv
10000000 fixtures/sample-10m.tsv

$ stat --format "%s" fixtures/sample-10m.tsv
548384897

$ time makta < fixtures/sample-10m.tsv
2021/09/30 16:58:07 [ok] initialized database -- data.db
2021/09/30 16:58:17 [io] written 523M · 56.6M/s
2021/09/30 16:58:21 [ok] 1/2 created index -- data.db
2021/09/30 16:58:34 [ok] 2/2 created index -- data.db

real    0m26.267s
user    0m24.122s
sys     0m3.224s
```

* 10M rows stored, with indexed keys and values in 27s, 370370 rows/s

### TODO

* [ ] allow tab-importing to be done programmatically, for any number of columns
* [x] a better name: mktabdb, mktabs, dbize - go with makta for now
* [ ] could write a tool for *burst* queries, e.g. split data into N shard,
      create N databases and distribute queries across files - e.g. `dbize db.json`
      with the same repl, etc. -- if we've seen 300K inserts per db, we may see 0.X x CPU x 300K, maybe millions/s.

As shortcut, we could add commands that turn an index into a "sqlite3" database
in one command, e.g.

```shell
$ mkindexdb -server ... -key-field id -o index.db
```

### Design ideas

A design that works with 50M rows per database, e.g. 20 files for 1B rows;
grouped under a single directory. Every interaction only involves the
directory, not the individual files.

----

Clip art from [ClipArt ETC](https://etc.usf.edu/clipart/).
