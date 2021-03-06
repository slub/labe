# Meeting Minutes

> 2021-09-23, 1300, JN, MC

## Topics

* [x] status of DOI coverage in index data
* [x] first fusion with OCI, coverage ratio
* [x] first design of fused records for rest api

## Notes

* [ ] abstract "id-to-index-document" a bit more; layered approach, e.g. live
  (all to solr); prefetch (solrdump/microblob/sqlite3/pg); caching (after each
  request); with a TTL; maybe some url with a wildcard for id, e.g.
  "http://hello.world/r/{{ id }}"

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

* [ ] prepare deployment; VM in progress; 8; 16; 500G; debian 10; systemd unit

