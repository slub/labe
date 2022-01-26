# README

Helpers to fetch and build the databases required for LABE API server. Specifically we need:

* [x] checks for new OCI/COCI releases, from [https://opencitations.net/download](https://opencitations.net/download)
* [x] regular dumps of internal SOLR indices and conversion into sqlite3 databases, using [solrdump](https://github.com/ubleipzig/solrdump)

The [luigi](https://github.com/spotify/luigi) task orchestrator is already used by
[SLUB Dresden](https://www.slub-dresden.de/), so we'll use it for modelling the
dependency graph of tasks.

## Deploy Option

We build a single executable with [shiv](https://github.com/linkedin/shiv) (see
LI ENG [blog
post](https://engineering.linkedin.com/blog/2018/05/introducing-and-open-sourcing-shiv)).

> shiv allows us to create a single binary artifact from a Python project that
> includes all of its dependencies. The only thing required to run a
> full-fledged Python application is an interpreter.

## Tasks

* [x] check for new OCI dump
* [x] download OCI dump
* [x] turn OCI dump into a more managable format (single zstd file)
* [x] fetch SOLR index copy, e.g. via [solrdump](https://github.com/ubleipzig/solrdump)
* [x] turn SOLR JSON files into (id, doc) TSV
* [x] create sqlite3 database from TSV with [makta](https://github.com/miku/labe/tree/main/go/ckit#makta)
* [x] move databases into place

Additionally, we want:

* [ ] monitoring, if a task fails (to a service email)
* [ ] a delta report
* [x] cache warmup, if necessary (yes, it is; currently squashed into the "cron" [line](https://github.com/slub/labe/blob/9b27980ebc3bfaf72358287c514681b8c4126803/ansible/roles/labe/tasks/main.yml#L117))
* [x] cleanup of obsolete tasks

Constraints: We will only have disk space for a single update. We may want to
reduce the index data size, e.g. reduce in a pipe while dumping from solr or
select a number of fields (`solrdump -fl ...`) -- **Update**: we abridge the SOLR
documents ("short"), and save disk space; one complete update with intermediate
artifacts occupies about 310G as of 01/2022.

## Deployment

* run "labe.pyz -r CombinedUpdate", e.g. daily
* run "rm -f $(labe.pyz --list-deletable)", e.g. daily

Example cron: [roles/labe/tasks/main.yml](https://github.com/slub/labe/blob/9b27980ebc3bfaf72358287c514681b8c4126803/ansible/roles/labe/tasks/main.yml#L99-L118)

## Directory layout

* 13G + 150G + 42G + 5.5G + 2.5G = 213G in databases
* a current set of outputs and intermediate files: 308G
* total disk space: 1007G

Currently (01/2022) we can accommodate two full copies at the same time. We
have headroom for a 30% increase in data size, before we need to take
additional measures (e.g. delete intermediate artifacts, too).

```
$ tree -sh
.
├── [4.0K]  IdMappingDatabase
│   ├── [  67]  current -> /usr/local/share/labe/IdMappingDatabase/date-2022-01-10.db
│   └── [ 13G]  date-2022-01-10.db
├── [4.0K]  IdMappingTable
│   ├── [  69]  current -> /usr/local/share/labe/IdMappingTable/date-2022-01-10.tsv.zst
│   └── [452M]  date-2022-01-10.tsv.zst
├── [4.0K]  OpenCitationsDatabase
│   ├── [150G]  c90e82e35c9d02c00f81bee6d1f34b132953398c.db
│   └── [  96]  current -> /usr/local/share/labe/OpenCitationsDatabase/c90e82e35c9d02c00f81bee6d1f34b132953398c.db
├── [4.0K]  OpenCitationsDownload
│   ├── [ 31G]  c90e82e35c9d02c00f81bee6d1f34b132953398c.zip
│   └── [  97]  current -> /usr/local/share/labe/OpenCitationsDownload/c90e82e35c9d02c00f81bee6d1f34b132953398c.zip
├── [4.0K]  OpenCitationsSingleFile
│   ├── [ 31G]  c90e82e35c9d02c00f81bee6d1f34b132953398c.zst
│   └── [  99]  current -> /usr/local/share/labe/OpenCitationsSingleFile/c90e82e35c9d02c00f81bee6d1f34b132953398c.zst
├── [4.0K]  SolrDatabase
│   ├── [  76]  current-ai-short -> /usr/local/share/labe/SolrDatabase/date-2022-01-10-name-ai-short.db
│   ├── [  78]  current-main -> /usr/local/share/labe/SolrDatabase/date-2022-01-09-name-main-short.db
│   ├── [  78]  current-main-short -> /usr/local/share/labe/SolrDatabase/date-2022-01-10-name-main-short.db
│   ├── [  83]  current-slub-production -> /usr/local/share/labe/SolrDatabase/date-2022-01-10-name-slub-production.db
│   ├── [5.5G]  date-2022-01-09-name-main-short.db
│   ├── [ 42G]  date-2022-01-10-name-ai-short.db
│   ├── [5.5G]  date-2022-01-10-name-main-short.db
│   └── [2.5G]  date-2022-01-10-name-slub-production.db
└── [4.0K]  SolrFetchDocs
    ├── [  78]  current-ai-short -> /usr/local/share/labe/SolrFetchDocs/date-2022-01-10-name-ai-short.zst
    ├── [  74]  current-main -> /usr/local/share/labe/SolrFetchDocs/date-2022-01-10-name-main.zst
    ├── [  80]  current-main-short -> /usr/local/share/labe/SolrFetchDocs/date-2022-01-10-name-main-short.zst
    ├── [  85]  current-slub-production -> /usr/local/share/labe/SolrFetchDocs/date-2022-01-10-name-slub-production.zst
    ├── [5.7G]  date-2022-01-10-name-ai-short.zst
    ├── [968M]  date-2022-01-10-name-main-short.zst
    ├── [ 20G]  date-2022-01-10-name-main.zst
    └── [231M]  date-2022-01-10-name-slub-production.zst

7 directories, 26 files
```
