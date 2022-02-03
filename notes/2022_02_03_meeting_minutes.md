# Meeting Minutes

> 2022-01-27, 1200, JN, MC

## TOP

* [x] ansible deployment on sdvlabe, config setup, auto-mode
* [ ] diff report

## Misc

* [ ] performance tests: [2022_01_30_performance_report.md](https://github.com/slub/labe/blob/main/notes/2022_01_30_performance_report.md)
* [ ] server tweaks, memory consumption
* [ ] disk-based cache (ephemeral, simple, tweakable)
* [ ] per institution filtering

```
$ curl -s http://0.0.0.0:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU | jq -r .extra
{
  "took": 0.000624,
  "unmatched_citing_count": 0,
  "unmatched_cited_count": 2045,
  "citing_count": 0,
  "cited_count": 3608,
  "cached": true
}

$ curl -s http://0.0.0.0:8000/id/ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU?i=DE-14 | jq -r .extra
{
  "took": 0.000559,
  "unmatched_citing_count": 0,
  "unmatched_cited_count": 3072,
  "citing_count": 0,
  "cited_count": 2581,
  "cached": true,
  "institution": "DE-14"
}
```

* [ ] stats tasks, daily stats generation

## TODO

* [ ] set notification email (and do not override config (cf. https://unix.stackexchange.com/a/511865/376)

----


## Notes (last meeting)

Potential other DOI-DOI citation sources.

> *OpenAlex*, MAG, crossref, unpaywall, DOAJ, orcid, http://openalex.org/, arxiv,
> zenodo, ...

* [https://archive.org/details/openalex_2021-10-11_beta](https://archive.org/details/openalex_2021-10-11_beta)

![](mag-entity-relationship-diagram.png)

### Misc

* licensing vs source (OA); openalex; OA status (green, gold, ...)
* unpaywall, metrics; "citation impact"
