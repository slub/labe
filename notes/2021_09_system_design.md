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
<<<<<<< HEAD

=======
>>>>>>> dataset download stub
