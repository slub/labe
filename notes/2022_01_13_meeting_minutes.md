# Meeting Minutes

> 2022-01-06, 1300, JN, CR, MC, TW

## Checklist

* [ ] **AP1 Prozessierungspipeline**

Implemented eight tasks for the processing pipeline:

```
CombinedUpdate
IdMappingDatabase
IdMappingTable
OpenCitationsDatabase
OpenCitationsDownload
OpenCitationsSingleFile
SolrDatabase
SolrFetchDocs
```

There is a single command line tool: `labe.pyz` that allows to run tasks.

```
Command line interface to run luigi tasks for labe project ≋
(https://github.com/grossweber/labe).

Examples:

    List tasks:

        $ labe.pyz -l
        CombinedUpdate
        IdMappingDatabase
        IdMappingTable
        OpenCitationsDatabase
        OpenCitationsDownload
        OpenCitationsSingleFile
        SolrDatabase
        SolrFetchDocs

    Run task:

        $ labe.pyz -r SolrDatabase --name main

    Show task output location:

        $ labe.pyz -O OpenCitationsDatabase
        /usr/share/labe/OpenCitationsDatabase/c90e82e35c9d02c00f81bee6d1f34b132953398c.db

Symlinks point to the current version of a task output. They will only be
updated, if the task ran successfully. This way we can identify outdated files:

    $ labe.pyz --list-deletable

Use cron job to schedule tasks:

    0 10 * * * rm -rf $(labe.pyz --list-deletable)
    0 30 * * * labe.pyz -r CombinedUpdate

Relevant configuration files:

    /etc/luigi/luigi.cfg
    /etc/luigi/logging.ini
    /etc/labe/labe.cfg

```


* [ ] **AP2 Abruf der aktuellen Dumps**

There is a module
[`oci.py`](https://github.com/GROSSWEBER/labe/blob/main/python/labe/oci.py)
which checks Open Citataions website for new dumps. The most recent download
URL can be queried from the command line:

```
$ ./labe.pyz -L
https://figshare.com/ndownloader/articles/6741422/versions/12
```


* [ ] **AP3 Verarbeiten und Reduktion der Daten**

Data is prepared and put into a querable form by utilitizing sqlite3.

* we turn OCI dump into sqlite3
* we turn solr index documents into sqlite3
* we generate an id-to-doi mapping and store it in sqlite3

A standalone tool,
[`makta`](https://github.com/GROSSWEBER/labe/tree/main/go/ckit#makta) can turn
a TSV file into sqlite3 database (scale to billions of rows). A small tool,
[`tabjson`](https://github.com/GROSSWEBER/labe/tree/main/go/ckit#tabjson) turns
a JSON document with an ID into a TSV file for further generation of a
key-value style sqlite3 database.

Instead of batch processing, we use these sqlite3 databases in conjunction with
the API server to server queries.

* [ ] **AP4 Speichern der Daten im Datahub**

We save all files on a dedicated machine ("sdvlabe"). There is one data
directory, which contains all downloaded files and final databases.

* [ ] **AP5 Einspielen der Daten in einen Index**

Originally, the idea of the index was to be a data store for the merged data.
We use sqlite3 databases as data store. The
[`makta`](https://github.com/GROSSWEBER/labe/tree/main/go/ckit#makta) helps us to move tabular data quickly into sqlite3.

* [ ] **AP6 Bereitstellung der Daten als REST API**

The [`labed`](https://github.com/GROSSWEBER/labe/tree/main/go/ckit#labed) tool
is a standalone HTTP server that serves queries for inbound and outbound
citations for a given ID (e.g. ai-49-28bf812...) or DOI (10.123/123...). The
labed server will utilize the id-mapping, citations and index cache databases
to fuse a single JSON response containing information about inbound, outbound
citations as well as citations currently not found in the catalog data.

The server is lean (~1K LOC) and focusses on performance. It is possible to
trade memory for speed by using a builtin cache (which caches expensive
responses on first query). Via middleware, the server supports gzip
compression, logging and query tracing.

* [ ] **AP7 Optional: Automatischer Delta Report**

A delta report generator is currently in progress. It's design will piggyback
on existing UNIX facilities, such as
[`diff(1)`](https://man7.org/linux/man-pages/man1/diff.1.html).

## Additional work items

### ID-DOI mapping

> Um zu erfahren, welche Werke im SLUB Bestand sind, kann eine Abfrage der DOIs
> gegen einen Solr Index stattfinden.

Not all DOI were readily available in indices; we wrote a specific tool,
[`doisniffer`](https://github.com/GROSSWEBER/labe/tree/main/go/ckit#doisniffer)
that allows to augment existing JSON lines SOLR files with DOI information.

### Performance reports

In order to test our design, we conducted regular performance tests. A final
report is outstanding, but the TL;DR is:

* labed server can serve 100+ rps sustained with moderate load
* while some queries (documents with many edges) may take longer than 1s, the majority of requests come back in less than 500ms

```
$ curl -sL https://is.gd/xGqzsg | \
    zstd -dc |
    parallel -j 40 "curl -s http://localhost:8000/id/{}" | \
    grep -v "no rows in result set" | \
    jq -rc '[.id, .doi, .extra.citing_count, .extra.cited_count, .extra.took] | @tsv'
```

### Lightning Talk

* lightning talk on the Go parts of the projects at [Leipzig
  Gophers](https://golangleipzig.space/)
[#23](https://golangleipzig.space/posts/meetup-23-wrapup/):
[Slides](https://github.com/miku/dwstalk)

## Extensibility

* in order to add more id-doi mappings, the data needs to be included in [`IdMappingTable`](https://github.com/GROSSWEBER/labe/blob/c67474c272cbbc51405bf53eb22d656622547c38/python/labe/tasks.py#L275-L325) task
* labed supports multiple index data stores to get catalog metadata from (via `-Q` flag)

## Maintenance

Currently (01/2022) we expect the project to be able to deployed and run with only little intervention for a 12-24 months. A list of a few maintenance issues are:

* update python *dependencies* (adjust `setup.py`)
* update Go *dependencies* (`go get -u -v`)

Currently, we use about 300G per version on a 1T disk drive, hence we can
accommodate two versions side-by-side. As citation data and index data is
expected to grow, the *disk size* may be a limiting factor in 12-24 months.

* OCI links are currently scraped, which means that as soon as OCI changes their webpage layout, the scraper module needs to be adjusted

More maintenance notes:

* crontab documentation is included
* a cleanup script and mechanism to only keep the most recent version of the data is included
* a basic ansible setup is included
* systemd service file is included
* both the Go and Python parts of the project can be packaged into a debian package (via `make deb` target)
* limitation: the Python version on the development machine needs to match the Python version on the target machine (see: [`Makefile`](https://github.com/GROSSWEBER/labe/blob/c67474c272cbbc51405bf53eb22d656622547c38/python/Makefile#L17-L21))

Additional notes on maintenance.

* total SLOC count (Python, Go) as of 2022-01-12: 2596 (3240 with blanks) in 25 files
* 4 separate Go tools, independent of each other
* Python package, with a regalar package appearance (via `setup.py`)
* Makefile target for formatting Python (via `make fmt`)
* Code documentation for each function and most modules
