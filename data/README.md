# Data Folder

## Raw index data

* complete SOLR data, as is; to inspect and decide whether it already contains
  the information required for matching with COCI corpus

```
$ solrdump -server $SOLR -q 'institution:DE-14' -verbose | zstd -c -T0 > index.json.zst
```

## Questions

* [ ] can we get the DOI from all raw indexed data via regex?

## Observations

* a few DOAJ records do not record DOI, but actually point to articles, which
  have one; example: https://doaj.org/article/0000128283cc43b79edeaaa33e826a58,
https://doaj.org/api/v2/articles/0000128283cc43b79edeaaa33e826a58,
https://journals.library.columbia.edu/index.php/cswr/article/view/1975 - has:
https://doi.org/10.7916/cswr.v6i1.1975

However, about 80% of DOAJ seems to have a DOI within the DOAJ metadata already:

```
$ taskcat DOAJIntermediateSchema| jq -rc .doi | head -1000000 | grep -c "^null"
219615
```
