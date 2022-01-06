# README

Helpers to fetch and build the databases required for LABE API server. Specifically we need:

* [ ] checks for new OCI/COCI releases
* [ ] regular dumps of internal SOLR indices and conversion into sqlite3 databases

The [luigi](https://github.com/spotify/luigi) orchestrator is already used by
[SLUB Dresden](https://www.slub-dresden.de/), so we'll use it for modelling the
dependency graph of tasks.

## Deploy Option

We build a single executable with [shiv](https://github.com/linkedin/shiv).

## Tasks

* [x] check for new OCI dump
* [x] download OCI dump
* [x] turn OCI dump into a more managable format (single zstd file)
* [x] fetch SOLR index copy, e.g. via [solrdump](https://github.com/ubleipzig/solrdump)
* [x] turn SOLR JSON files into (id, doc) TSV
* [x] create sqlite3 database from TSV with [makta](https://github.com/miku/labe/tree/main/go/ckit#makta)
* [ ] move databases into place
* [ ] inform API server to reset database connections via [SIGHUP](https://en.wikipedia.org/wiki/SIGHUP)

Additionally, we want:

* [ ] monitoring, if a task fails (to a service email)
* [ ] a delta report
* [ ] cache warmup, if necessary
* [ ] cleanup of obsolete tasks

Constraints: We will only have disk space for a single update. We may want to
reduce the index data size, e.g. reduce in a pipe while dumping from solr or
select a number of fields (`solrdump -fl ...`).
