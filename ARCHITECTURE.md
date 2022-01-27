# Architecture

[Luigi](https://github.com/spotify/luigi) for orchestration (but trying to keep
task code minimal). Structured directories and filenames for data artifacts.

![](static/labe-tree.png)

The result are three [sqlite3](https://sqlite.org/) databases:

* a) id-doi "mapping" database
* b) "citations" databases (doi-doi)
* c) index "metadata" key value store (id-doc)

There is one "mapping" and one "citations" database, but there can be one or
more "metadata" databases.

A server assembles fused results from these databases on the fly (and caches
expesive requests).

![](static/Labe-Sequence.png)

This way we should get a good balance:

* we need *little preprocessing*, we mostly turn CSV or SOLR JSON into sqlite databases
* we still can *be fast* through caching, which can be done forehandedly ("cache
  warming") or as data is actually requests

The server delivers JSON responses, which can be included in catalog frontends.

```json
$ curl -sL "http://localhost:8000/doi/10.1016/s0273-1177(97)00070-7" | jq .
{
  "id": "ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAxNi9zMDI3My0xMTc3KDk3KTAwMDcwLTc",
  "doi": "10.1016/s0273-1177(97)00070-7",
  "cited": [
    {
      "author": [
        "Meljac, Claire",
        "Voyazopoulos, Robert"
      ],
      "doi_str_mv": [
        "10.4000/rechercheseducations.819"
      ],
      "format": [
        "ElectronicArticle"
      ],
      "id": "ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuNDAwMC9yZWNoZXJjaGVzZWR1Y2F0aW9ucy44MTk",
      "institution": [
        "DE-Zi4",
        "DE-14",
        "DE-Ch1",
        "DE-Gla1",
        "DE-D161",
        "DE-Brt1",
        "DE-Pl11",
        "DE-Rs1",
        "DE-82",
        "DE-D275",
        "DE-15",
        "DE-105",
        "DE-L229",
        "DE-Bn3",
        "DE-Zwi2"
      ],
      "title": "Binet, citoyen indigne ?",
      "url": [
        "http://dx.doi.org/10.4000/rechercheseducations.819"
      ]
    }
  ],
  "unmatched": {},
  "extra": {
    "took": 0.001490343,
    "unmatched_citing_count": 0,
    "unmatched_cited_count": 0,
    "citing_count": 0,
    "cited_count": 1,
    "cached": false
  }
}
```
