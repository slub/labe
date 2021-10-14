# Meeting Minutes

> 2021-10-14, 1300, JN, CR, MC

## Topics

* [x] factored out [slikv](https://github.com/miku/slikv) sqlite3 as key-value store; replaces mkocidb
* [ ] prepare deployment; VM in progress; 8; 16; 500G; debian 10; systemd unit
    * [x] collect ideas on deployment
    * [ ] tools to generate suitable data stores (e.g. via OCI, a copy of solr, etc.)
    * [ ] a service unit for the server
    * [ ] implement HTTP cache policy
    * [ ] implement some logging strategy
    * [ ] 12-factor app considerations, e.g. switch to envconfig
    * [ ] some frontend like nginx
    * [ ] an ansible playbook for tools and service
* [ ] deployment, maybe next week; test-frontend
* [ ] db backup on a network mount
* [ ] SSHKEY (sent)
* [ ] data acquisition; packaging and deployment; "DEB", systemd, /etc/...; [1TB]
* [ ] optimization ideas: ETag, Cache-Control, ...; sqlite3; proxy: nginx
* [ ] data analysis, e.g. OCI

## Notes

* [ ] machine online; via UBL net
* [ ] HTTP port; HTTPS port

Next steps:

* [ ] separate update pipelines for [oci], [main], [ai] and [id-doi] databases and documentation

Misc.

* possibly inclusion of dataset repositories

## S312-2021

* [ ] fast rebuild from a local copy: via set of local sqlite3 databases

> from OCI dump to database, we need about 2h with `mkocidb`, getting an
> offline version of SOLR with `solrdump` can take a day or more; DOI
> annotation can take 2h; the derived databases may take an hour; a fresh setup
> every week may be possible

* [ ] automatic pipeline: put commands into cronjob; run commands manually
* [ ] query index for DOI: not supported, we use workarounds
* [ ] data reduction: on the fly, included matched and unmatched items
* [x] open citations now 1B+ in size (about 30% larger)
* [ ] metadata from OCI: but we use metadata from index or just the DOI (for the unmatched items)
* [x] definition of `slub_id` [3] - is it just `id`
* [x] REST API: query via `doi` or `id` possible - do we need the metadata of the queries record, too?
* [ ] delta: do reports per setup (e.g. new OCI dump), then a separate program to compare to reports

Some generic tools we may get out of this:

* [ ] a fast sqlite3 key-value database builder; inserting and indexing 10M entries takes
      27s - about 370K rows/s; currently mkocidb, but could be made more generic

> miku/mkocidb ... mkkvdb
