# ckit

Citation graph kit for the LABE project at [SLUB
Dresden](https://www.slub-dresden.de/). This subproject contains a few
standalone command lines programs and servers. The task orchestration part lives under [labe/python](../../python).

* [doisniffer](#doisniffer), [filter](https://en.wikipedia.org/wiki/Filter_(software)) to find DOI by patterns in Solr VuFind JSON documents
* [labed](#labed), an HTTP server serving Open Citations data fused with catalog metadata
* [tabjson](#tabjson), turn JSON into TSV
* [makta](#makta), turn TSV files into sqlite3 databases

To build all binaries, run:

```
$ make -j
```

To build a debian package, run:

```
$ make -j deb
```

To cleanup all artifacts, run:

```
$ make clean
```

----

## doisniffer

Tool to turn a [VuFind
Solr](https://vufind.org/wiki/development:architecture:solr_index_schema)
schema without DOI field into one with - `doi_str_mv` - by sniffing out
potential DOI from other fields.

```
$ cat index.ndj | doisniffer > augmented.ndj
```

By default, only documents are passed through, which actually contain a DOI.

```
Usage of doisniffer:
  -K string
        ignore keys (regexp), comma separated (default "barcode,dewey")
  -S    do not skip unmatched documents
  -b int
        batch size (default 5000)
  -i string
        identifier key (default "id")
  -k string
        update key (default "doi_str_mv")
  -version
        show version and exit
  -w int
        number of workers (default 8)
```

----

## labed

![](static/45582_reading_lg.gif)

HTTP API server, takes requests for a given id and returns a result fused from
OCI citations and index data.

It currently works with three types of sqlite3 databases:

* id-to-doi mapping
* OCI citations
* index metadata

### Usage

```sh
usage: labed [OPTION]

labed is an web service fusing Open Citation and Library Catalog data (SLUB);
it works with three types of databases:

(1) [-i] an sqlite3 catalog-id-to-doi translation database (10G+)
(2) [-o] an sqlite3 version of OCI/COCI (150GB+)
(3) [-m] an sqlite3 mapping from catalog ids to (json) metadata; this can be repeated
         (size depends on index size and on how much metadata is included) (40-350G)

Each database may be updated separately, with separate processes.

Examples

  $ labed -c -z -addr localhost:1234 -i i.db -o o.db -m d.db

  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA3My9wbmFzLjg1LjguMjQ0NA
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE3Ny8xMDQ5NzMyMzA1Mjc2Njg3
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxNC9hb3MvMTE3NjM0Nzk2Mw
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMjMwNy8yMDk1NTIx

Bulk requests

  $ curl -sL https://is.gd/xGqzsg | zstd -dc -T0 |
    parallel -j 40 "curl -s http://localhost:8000/id/{}" |
    jq -rc '[.id, .doi, .extra.citing_count, .extra.cited_count, .extra.took] | @tsv'

Flags

  -addr string
        host and port to listen on (default "localhost:8000")
  -c    enable in-memory caching of expensive responses
  -cg duration
        cache trigger duration (default 250ms)
  -ct duration
        cache cleanup interval (default 8h0m0s)
  -cx duration
        cache default expiration (default 72h0m0s)
  -i string
        identifier database path (id-doi mapping)
  -l    enable logging
  -logfile string
        file to log to
  -m value
        index metadata cache sqlite3 path (repeatable)
  -o string
        oci as a datbase path (citations)
  -q    no output at all
  -stopwatch
        enable stopwatch
  -version
        show version
  -z    enable gzip compression
```

### Using a stopwatch

Experimental `-stopwatch` flag to trace duration of various operations.

```sh
$ labed -stopwatch -c -z -i i.db -o o.db -m index.db
2022/01/13 12:26:30 setup group fetcher over 1 databases: [index.db]

    ___       ___       ___       ___       ___
   /\__\     /\  \     /\  \     /\  \     /\  \
  /:/  /    /::\  \   /::\  \   /::\  \   /::\  \
 /:/__/    /::\:\__\ /::\:\__\ /::\:\__\ /:/\:\__\
 \:\  \    \/\::/  / \:\::/  / \:\:\/  / \:\/:/  /
  \:\__\     /:/  /   \::/  /   \:\/  /   \::/  /
   \/__/     \/__/     \/__/     \/__/     \/__/

Examples:

  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA3My9wbmFzLjg1LjguMjQ0NA
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwMS9qYW1hLjI4Mi4xNi4xNTE5
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAwNi9qbXJlLjE5OTkuMTcxNQ
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTE3Ny8xMDQ5NzMyMzA1Mjc2Njg3
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxNC9hb3MvMTE3NjM0Nzk2Mw
  http://localhost:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMjMwNy8yMDk1NTIx

2022/01/13 12:26:30 labed starting 4a89db4 2022-01-13T11:23:31Z http://localhost:8000

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

* [ ] a detailed performance report

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
