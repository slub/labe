# Data Folder

## Raw index data

* complete SOLR data, as is; to inspect and decide whether it already contains
  the information required for matching with COCI corpus

```
$ solrdump -server $SOLR -q 'institution:DE-14' -verbose | zstd -c -T0 > index.json.zst
```

Note: grepping 59M docs takes 11m48.377s (with compression)

## Questions

* [ ] can we get the DOI from all raw indexed data via regex?

## Observations

* a few DOAJ records do not record DOI, but actually point to articles, which
  have one; example: https://doaj.org/article/0000128283cc43b79edeaaa33e826a58,
  https://doaj.org/api/v2/articles/0000128283cc43b79edeaaa33e826a58,
  https://journals.library.columbia.edu/index.php/cswr/article/view/1975 - has:
  https://doi.org/10.7916/cswr.v6i1.1975

Another example:

* https://doaj.org/api/v2/articles/000020ccd46f45b59f7ebbf88614b7f1,
  https://www.scielo.br/j/pab/a/fHJX5BmcWv67BH36jSx3qVx/,
  https://doi.org/10.1590/S0100-204X2001000100025

Could exact title match on SFC, e.g.
[https://is.gd/eGapKf](https://is.gd/eGapKf) to get an additional DOI for these
cases.

However, about 80% of DOAJ seems to have a DOI within the DOAJ metadata already:

```
$ taskcat DOAJIntermediateSchema | jq -rc .doi | head -1000000 | grep -c "^null"
219615
```

## Index contents

* DE-14: 9023521 docs
* 1470424 docs that contain something that looks like a DOI, but only 868085 unique

Checking a sample (5000):

```
time shuf -n 5000 ma.doi_unique.tsv | \
    awk '{print "https://doi.org/"$0}' | \
    clinker -w 256 -verbose

$ jq -rc .status ma.doi_unique_link_sample.json | sort | uniq -c | sort -nr
   4296 200
    469 null
    103 404
     95 403
     32 500
      5 400
```

About 0.85 valid, so maybe about 737872 valid DOI in a set of 9023521 - about 8.2% overall.

* DE-14: 62388840
* 64066526 URL, 62666501 unique

Non DOI URL: 6502749 (unique)

```
3530302 www.jstor.org
1602794 ieeexplore.ieee.org
1332640 doaj.org
  35627 pqdtopen.proquest.com
    694
    306 muse.jhu.edu
    148 www3.interscience.wiley.com
     75 www.historycooperative.org
     60 www.bioone.org
     54 www.sciencemag.org
     21 www.ingentaconnect.com
     19 bmrj.press.illinois.edu
      7 onlinelibrary.wiley.com
      1 www.mlajournals.org
      1 booksandjournals.brillonline.com
```

No DOI in metadata, but has DOI example:

* https://www.jstor.org/stable/523435, https://doi.org/10.2307/523435
* does this pattern apply to all of JSTOR? no, e.g. 1904 article does not have a DOI: https://www.jstor.org/stable/2375834
* ieee example: https://ieeexplore.ieee.org/document/1135671, http://doi.org/10.1109/tchmt.1980.1135671

