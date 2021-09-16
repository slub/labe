# Ideas on system design

Raw inputs.

```
$ labe status
```

List downloaded files and possible updates. We need one directory locally, e.g.
under `XDG_DATA_HOME`, and subdirs for the various tasks. Directory should be
fully managed by the program, may contain a managed sqlite3 database recording
runs and actions (e.g. for delta, etc).

A separate folder for solr downloads; use script or
[solrdump](https://github.com/ubleipzig/solrdump).

Every task should be runnable separately, or at once with dependency
resolution.

```
$ labe-run SlubSolrExport
$ labe-run CociDownload
$ labe-run SlubFiltered
```

Other commands.

```
$ labe-status
$ labe-tree
```

Alternative third-party projects.

* [DVC](https://dvc.org/) for handling versions of the raw inputs and pipelines
* [taskfile.dev](https://taskfile.dev/#/)

## Design issues

* [ ] we need SOLR data with DOI; might need external service to have best DOI
  coverage; how much is actually lost in the SOLR schema?
* [ ] one operation that applies an own data subset (e.g. by DOI) to open
  citations to produce a "local" version of citation links

Possible interaction, e.g. a sort of fusion:

```
$ ocifuse -oci coci.csv -our data.json -doi-field-name doi > fused.json
```

Take OCI and local file, output will be local file with additional fields for
inbound and outbound references, e.g. like (id refers to local id):

```json
{
    "id": "id-432",
    ...
    "citing": ["id-230", "id-123", ...],
    "cited": ["id-729", "id-192", ...],
    ...
}
```

This file should be servable per HTTP for catalog frontend or other systems.
May contain more information about the cited and citing entities (e.g. title,
authors, year, ...) to minimize additional requests. In fact: we want *only
one* request to get information about all linkage (inbound, outbound) - this
may be a few or a few thousand records. Opportinities for caching.
