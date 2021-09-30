# Meeting Minutes

> 2021-09-20, 1300, ...

## Topics

* [ ] abstract "id-to-index-document" a bit more; layered approach, e.g. live
  (all to solr); prefetch (solrdump/microblob/sqlite3/pg); caching (after each
  request); with a TTL; maybe some url with a wildcard for id, e.g.
  "http://hello.world/r/{{ id }}"

> We have 3+1 backends, sqlite3, microblob, solr; can add mysql or pg as well
> with a little adapter; sqlite3 (fetchall) seems the most performant so far;
> we could also group backends for cascading or hedged requests

* [ ] support non-source items, e.g.

```json
{
    ...
    "matched": {          |
        "citing": ...     | ... index schema (any, could be just an id)
        "cited": ...      |
    },                    |

    "unmatched": {        | maybe PDA, etc
        "citing": ...     | ... doi
        "cited": ...      |
    }                     |
    ...
}
```

> implemented, a current example: [https://i.imgur.com/uJAWsAE.png](https://i.imgur.com/uJAWsAE.png)

![](uJAWsAE.png)

* [ ] prepare deployment; VM in progress; 8; 16; 500G; debian 10; systemd unit
    * [ ] tools to generate suitable data stores (e.g. via OCI, a copy of solr, etc.)
    * [ ] a service unit for the server
    * [ ] implement HTTP cache policy
    * [ ] implement some logging strategy
    * [ ] 12-factor app considerations, e.g. switch to envconfig
    * [ ] some frontend like nginx
    * [ ] an ansible playbook for tools and service

## Notes

...

## S312-2021

* [ ] fast rebuild from a local copy: via set of local sqlite3 databases

> from OCI dump to database, we need about 2h with `mkocidb`, getting an
> offline version of SOLR with `solrdump` can take a day or more; DOI
> annotation can take 2h; the derived databases may take an hour; a fresh setup
> every week may be possible

* [ ] automatic pipeline: put commands into cronjob; run commands manually
* [ ] query index for DOI: not supported, we use workarounds
* [ ] data reduction: on the fly, included matched and unmatched items
* [ ] open citations now 1B+ in size (about 30% larger)
* [ ] metadata from OCI: but we use metadata from index or just the DOI (for the unmatched items)
* [ ] definition of `slub_id` [3] - is it just `id`
* [ ] REST API: query via `doi` or `id` possible - do we need the metadata of the queries record, too?
* [ ] delta: do reports per setup (e.g. new OCI dump), then a separate program to compare to reports

Some generic tools we may get out of this:

* [ ] a fast sqlite3 key-value database builder; inserting and indexing 10M entries takes
      27s - about 370K rows/s; currently mkocidb, but could be made more generic

