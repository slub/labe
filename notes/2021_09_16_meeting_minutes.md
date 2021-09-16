# Meeting Minutes

> 2021-09-16, 1300, JN, CR, MC

## Topics

* [ ] DOI issue and numbers so far: 8.2% of main, about 89.6% (est); total
  currently: 737872 + (62666501 - 6502749) = 56901624 -- using an explicit
  "doi" field would *much better* (see [data/README](../data/README.md))
* [ ] deployment sizing (start with 0.5T, M/L machine, 8x/16G); [zstd](https://github.com/facebook/zstd) will help

## Notes

* `doi_str_mv` is a [recommended dynamic field](https://vufind.org/wiki/development:architecture:solr_index_schema)
* for the resulting service: "one query should suffice" - how much information do we need to carry around?

## TODO

* [ ] verify DOI provenance, e.g. which field it comes from; MARC DOI, e.g. "024", `$2`, ...
* [ ] fuse OCI and index data
* [ ] setup intranet connectivity


