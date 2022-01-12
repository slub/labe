# Meeting Minutes

> 2022-01-13, 1300, JN, CR, MC, TW

## Checklist

All code and notes are available under a permissive free software license; the
git repository can be found at
[https://github.com/GROSSWEBER/labe](https://github.com/GROSSWEBER/labe).

* [ ] **AP1 Prozessierungspipeline**

Implemented eight
[tasks](https://github.com/GROSSWEBER/labe/blob/main/python/labe/tasks.py) for
the processing pipeline (using [luigi](https://github.com/spotify/luigi) as
pipeline framework):

```
CombinedUpdate
IdMappingDatabase         *
IdMappingTable
OpenCitationsDatabase     *
OpenCitationsDownload
OpenCitationsSingleFile
SolrDatabase              *
SolrFetchDocs
```

The `*` tasks are sqlite3 databases used by the API server.

There is a single command line tool, named
[`labe.pyz`](https://github.com/GROSSWEBER/labe/blob/801c700dcec4dbca864e022176f275f1acbc31a1/python/Makefile#L23-L26)
that allows to run tasks (this is a single file packed Python project, built
with [shiv](https://shiv.readthedocs.io/)).

> Shiv is a command line utility for building fully self-contained Python
> zipapps as outlined in [PEP 441](https://www.python.org/dev/peps/pep-0441/)
> but with all their dependencies included!

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
which checks [Open Citations website](https://opencitations.net/download) for
new dumps. The most recent download URL can be queried from the command line:

```
$ labe.pyz -L
https://figshare.com/ndownloader/articles/6741422/versions/12
```

* [ ] **AP3 Verarbeiten und Reduktion der Daten**

Data is prepared and put into a queryable form by utilitizing [sqlite3](https://sqlite.org).

* we convert the OCI CSV data dump into an sqlite3 database
* we turn solr index documents into sqlite3 key value store (key: id, value: doc)
* we generate an id-to-doi mapping and store it in sqlite3

A standalone tool -
[`makta`](https://github.com/GROSSWEBER/labe/tree/main/go/ckit#makta) - can turn
a TSV file into an sqlite3 database (scales to billions of rows). A small tool -
[`tabjson`](https://github.com/GROSSWEBER/labe/tree/main/go/ckit#tabjson) - turns
a JSON document with an ID field into a TSV file (to be used in a key-value style).

Instead of batch processing, we use these sqlite3 databases in conjunction with
the API server to serve queries.

* [ ] **AP4 Speichern der Daten im Datahub**

We save all files on a dedicated host ("sdvlabe", *Intel [Xeon Gold
5218](https://ark.intel.com/content/www/us/en/ark/products/192444/intel-xeon-gold-5218-processor-22m-cache-2-30-ghz.html)
(8) @ 2.294GHz*, 1T disk). There is one data directory, which contains all
downloaded files and final databases. Naming of files and directories is
regular.

* [ ] **AP5 Einspielen der Daten in einen Index**

Originally, an index was to be used as a data store for the merged data.
However, we can also use sqlite3 databases as our backing data stores. The
[`makta`](https://github.com/GROSSWEBER/labe/tree/main/go/ckit#makta) tool
helps us to move tabular data into sqlite3 quickly.

* [ ] **AP6 Bereitstellung der Daten als REST API**

The [`labed`](https://github.com/GROSSWEBER/labe/tree/main/go/ckit#labed)
program is a standalone HTTP server that serves queries for inbound and
outbound citations for a given ID (e.g. ai-49-28bf812...) or DOI
(10.123/123...).

The labed server will utilize the id-mapping, citations and
index cache databases to fuse a single JSON response containing information
about inbound, outbound citations as well as citations currently not found in
the catalog data.

![](../static/Labe-Sequence.png)

The server is minimalistic (~1K LOC) and focusses on performance. It is
possible to trade memory for speed by using a built-in cache (which caches
expensive responses on first query). Via middleware, the server supports gzip
compression, logging and query tracing.

* [ ] **AP7 Optional: Automatischer Delta Report**

A delta report generator is currently in progress. Its design will piggyback
on existing UNIX facilities, such as
[`diff(1)`](https://man7.org/linux/man-pages/man1/diff.1.html).

## Additional work items

### ID-DOI mapping

> Um zu erfahren, welche Werke im SLUB Bestand sind, kann eine Abfrage der DOIs
> gegen einen Solr Index stattfinden.

Not all DOI were readily available in indices; we wrote a specific tool -
[`doisniffer`](https://github.com/GROSSWEBER/labe/tree/main/go/ckit#doisniffer) - that allows to augment existing JSON lines SOLR files with DOI information.

```shell
$ echo '{"title": "example title", "notes": "see also: 10.123/123"}' | doisniffer | jq .
{
  "doi_str_mv": [
    "10.123/123"
  ],
  "notes": "see also: 10.123/123",
  "title": "example title"
}
```

### Performance reports

In order to test our design, we conducted regular performance tests. A final
report is outstanding, but the TL;DR is:

* labed server can serve 100+ rps sustained with moderate load
* while some queries (documents with many edges) may take longer than 1s, the
  majority of requests come back in less than 500ms

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
[#23](https://golangleipzig.space/posts/meetup-23-wrapup/) (2021-11-23), titled
[A data web service](https://github.com/miku/dwstalk)

## Extensibility

* in order to add more id-doi mappings, the data needs to be included in
  [`IdMappingTable`](https://github.com/GROSSWEBER/labe/blob/c67474c272cbbc51405bf53eb22d656622547c38/python/labe/tasks.py#L275-L325) task
* labed supports multiple index data stores to get catalog metadata from (via `-Q` flag)

## Maintenance

Currently (01/2022) we think the project can run with minimal intervention for
a 12-24 months. A list of a few maintenance issues are:

* update python *dependencies* (adjust `setup.py`)
* update Go *dependencies* (`go get -u -v`)

Currently, we use about 300G per update cycle on a 1T disk drive, hence we can
accommodate at most two versions at the same time. As citation data and index
data is expected to grow, the *disk size* may be a limiting factor in 12-24
months.

* OCI links are currently scraped, which means that as soon as OCI changes
  their webpage layout, the scraper module needs to be adjusted

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

