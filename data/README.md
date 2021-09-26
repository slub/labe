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

## DOI sniffing

Sniff out DOI, per field.

```
$ zstdcat -T0 ma.json.zst | parallel --pipe -j 8 --block 10M 'python doisniffer.py'
```

We get a 3-column file, with ID, field and value.

```
$ head ma.doi.sniffed.tsv
0-011506326     fullrecord:marc 10.73/0941
0-011506326     dewey-full      10.73/0941
0-011506326     dewey-raw       10.73/0941
0-011506326     dewey-search    10.73/0941
0-013497936     barcode_de105   10.2626/18
0-015609626     barcode_de105   10.5763/18
0-017646340     barcode_de105   10.1424/18
0-016998340     fullrecord:marc 10.1007/978-1-4613-0893-5
0-016998340     ismn    10.1007/978-1-4613-0893-5
0-016998340     marc024a_ct_mv  10.1007/978-1-4613-0893-5
```

Found:

* 6247608 entries across 996345 docs; 61 different fields

```
$ wc -l ma.doi.sniffed.tsv
6247608

$ cut -f1 ma.doi.sniffed.tsv | sort -u -S50% | wc -l
996345
```

There are 959485 unique DOI like strings:

```
$ cut -f3 ma.doi.sniffed.tsv | sort -u -S50% | wc -l
959485
```

Top 30 fields with doi-like strings:

```sh
$ cut -f2 ma.doi.sniffed.tsv | sort -S50% | uniq -c | sort -nr | head -30
2529777 fullrecord:marc
1573904 spelling
 938883 ismn
 644602 url
 538635 marc024a_ct_mv
  11091 ctrlnum
   4836 barcode_de105
   1726 isbn
   1346 footnote
    588 dateSpan
    389 spellingShingle
    371 signatur
    356 contents
    266 container_reference
    106 title_in_hierarchy
     99 dewey-search
     99 dewey-raw
     99 dewey-full
     79 hierarchy_sequence
     48 multipart_part
     39 container_title
     23 title_list_str
     22 title_full_unstemmed
     22 title_fullStr
     22 title_full
     18 title
     17 title_auth
     17 mab_dech1_str_mv
     12 title_sub
      8 is_hierarchy_title
```

Mostly, we have one DOI per ID, only for 14830 records, we have multiple DOI per record.

```sh
$ zstdcat -T0 ma.doi.sniffed.tsv.zst | python are_doi_unique_per_record.py | shuf -n 10
0-1734561769 {'10.4028/www.scientific.net/MSF.894', '10.4028/www.scientific.net/MSF.894*'}
0-1667801406 {'10.5771/9783845261614/dramas-of-reconciliation', '10.5771/9783845261614'}
0-173458324X {'10.4028/www.scientific.net/SSP.97-98*', '10.4028/www.scientific.net/SSP.97-98'}
0-1692685171 {'10.15480/882.2699', '10.1115/1.4045625'}
0-1734566264 {'10.4028/www.scientific.net/AMM.24-25', '10.4028/www.scientific.net/AMM.24-25*'}
0-1748014994 {'10.15480/882.2933', '10.1016/j.foodhyd.2020.106132'}
0-1663386293 {'10.5771/2509-9485-2018-2-313', '10.5771/2509-9485-2018-2-313/methodische-herausforderungen-quantitativer-befragungen-von-gefluechteten-am-beispiel-einer-vorstudie-in-sachsen-jahrgang-2-2018-heft-2'}
0-1734566701 {'10.4028/www.scientific.net/AST.76*', '10.4028/www.scientific.net/AST.76'}
0-1689836717 {'10.1186/s40497-019-0168-0', '10.1186/s40497-019-0168-0.pdf'}
0-1046496972 {'10.26504/rs58.pdf', '10.26504/rs58'}
```

## Enhance dataset with DOI

```
$ zstdcat -T0 ma.json.zst | parallel --pipe -j 8 --block 10M \
    'python doisniffer.py --parse-marc --aggressive --update' | \
    zstd -c -T0 > ma_with_doi.json.zst
```

## Joining datasets

Options:

We could use an online system, e.g. elasticsearch, to index the citation data
and do a request for each metadata item. At 100 rps, we would need something
like 1 week to iterate over 60M docs. Plus we need extra hardware to run the
index - although this would be a one time cost.

Offline version would allow to stream through the input file and amend the
metadata records on the fly. No extra component, but OCI needs to be prepared
to support this approach, e.g.

* [ ] get oci dump; citing-cited
* [ ] sort input by DOI
* [ ] sort by citing; these will be the outbound refs
* [ ] sort by cited; these will be the inbound refs

...

Other rough ideas.

> Target schema.

```json
{
    "id": "0-123",
    "doi": "10.123/123",
    "citing": [{}, {}, ...],
    "cited": [{}, {}, ...],
}
```

For `citing` and `cited` we want not just the DOI, but a few more fields:
title, authors, year.

May require multiple passes, if no external lookup is used.

* (1) turn OCI dump into an annotated dump, that is for each DOI add the extra information we want
* (2) iterate over both input and enhanced OCI dump and generate "citing" field
* (3) iterate over both input and enhanced OCI dump and generate "cited" field

Various sort ops necessary:

* sort OCI dump by citing
* sort OCI dump by cited
* sort index data by DOI

So in total: 3 sorts over 100G+ data, 3 passes.

We may trim down OCI dump before hand to column 2 and 3 only.

----

Another approach: turn OCI dump into a minimal, indexed sqlite3 lookup
database; might be ok to do 70M queries.

A first run with `sqlite3` and 100M rows: 12min to insert (14G); so maybe 5h
and 150GB database size including indexes. We would need the extra blobs as
well. But any additional information would blow up the db size (considerably,
since titles and authors may amount for 10x or more the size of the DOI only).

Note: index creation naturally slows import down; w/o index we get a sustained
30M/s or higher insert speed; about 1B rows inserted w/o index in 28m31.370s.

With post insert index creation, we need 6m36.173s to create a database of 100M
(99057347, actually) rows, size: 13GB; with sqlite3 memory usage hovering
around 70% (not sure, if that is an internal limit -- like -S -- or dependent
on the data, which is larger). Ratio: about 2/5 min for insert/index. Expecting
80mins for complete dataset.

----

How about using a cache-first approach? Setup some more expensive operation,
then cache all results and eventually request all data to build up the cache.

> request for identifier -> find DOI -> find related DOI -> find related
  metadata -> merge -> cache -> serve

Would lower the computation burden for preprocessing and with a warm cache
would be just as fast.


## OCI dump and sizing

* 6741422v11.zip is 32GB zip compressed; may need a quick transfer into "single file zstd"

```sh
16,081,617,819 2019-10-21T22_41_20_1-63.zip
   764,296,250 2020-01-13T19_31_19_1-4.zip
 1,231,127,880 2020-04-25T04_48_36_1-5.zip
   518,631,764 2020-06-13T18_18_05_1-2.zip
   319,781,656 2020-08-20T18_12_28_1-2.zip
   706,940,736 2020-11-22T17_48_01_1-3.zip
 2,283,982,086 2021-01-27T23_11_15_1-9.zip
 6,936,692,157 2021-07-06T163745_0-4_1-5.zip
 2,764,959,519 2021-08-14T222043_0-4_1-2.zip
```

Zip to zst dance.

```sh
$ mkdir 6741422v11
$ unzip -d 6741422v11 6741422v11.zip
$ for f in $(find 6741422v11 -name "*zip"); do unzip -p $f; done | LC_ALL=C grep -vF 'oci,citing' | zstd -c -T0 > 6741422v11.zst
# 28m33.617s
```

Trim down the OCI dump, with [hck](https://github.com/sstadick/hck):

```sh
$ time zstdcat -T0 6741422v11.zst | pv -l | hck -d, -f2,3 | zstd -c -T0 > 6741422v11s.zst
# 9m4.413s
```

Resulting file is 12G only. Plain iteration:

```sh
$ time zstdcat -T0 6741422v11s.zst | pv -l > /dev/null
1.19G 0:02:23 [8.28M/s] [                                                                                                                                                                                        <=>                          ]

real    2m23.432s
user    2m28.104s
sys     0m40.814s
```

Sorting test runs:

```sh
$ time zstdcat -T0 6741422v11s.zst | LC_ALL=C sort -S70% -t $'\t' -k1,1 | zstd -c -T0 > 6741422v11s1.zst
...

$ time zstdcat -T0 6741422v11s.zst | LC_ALL=C sort -S70% -t $'\t' -k2,2 | zstd -c -T0 > 6741422v11s2.zst

real    37m18.037s
user    97m5.164s
sys     6m32.350s
```

## Design Sketch

Three-db setup:

* id to doi mapping (sqlite, mublob)
* doi to inbound and outbound doi (mkocidb)
* doi to data (mublob)

If we could query the indices directly for the doi, we could get rid of two
databases.

1. Create id to doi mapping from raw index data.

```sh
$ zstdcat -T0 ai.json.zst | jq -rc 'select(.url | length > 0) | [.id, .url[0]] | @tsv' | grep "doi.org" | \
    sed -e 's@http://dx.doi.org/@@;s@https://dx.doi.org/@@;s@http://doi.org/@@;s@https://doi.org/@@;s@\/\/@\/@' > ai_id_doi.tsv
```

Or:

```sh
$ zstdcat -T0 ma_with_doi.json.zst| jq -rc 'select(.doi_str_mv | length > 0) | [.id, .doi_str_mv[]] | @tsv' > ma_id_doi.tsv
```

Parallel:

```sh
$ time zstdcat -T0 ai.json.zst | pv -l | parallel -j 8 --pipe --block 10M \
    "jq -rc 'select(.url | length > 0) | [.id, .url[0]] | @tsv'" | grep "doi.org" | \
    sed -e 's@http://dx.doi.org/@@;s@https://dx.doi.org/@@;s@http://doi.org/@@;s@https://doi.org/@@;s@\/\/@\/@' > id_to_doi.tsv
```

2. Run mkocidb over reduced OCI data.

```sh
$ time zstdcat -T0 ../../data/6741422v11s1.zst | ./mkocidb
2021/09/20 17:34:55 [ok] initialized database -- data.db
written 57.6G -- 41.4M/s
2021/09/20 17:58:40 db setup done
2021/09/20 17:58:40 creating index
2021/09/20 19:07:53 [ok] created index -- data.db

real    92m58.885s
user    79m56.305s
sys     10m0.852s
```

3. Create doi to data lookup service.

```sh
$ time microblob -create-db-only -key id index.data
INFO[0000] creating db index.data.832a9151.db ...
  97% |███████████████████████████████████████ | [46s:1s]            Killed

real    46m17.090s
user    124m13.662s
sys     5m50.253s
```

## Alternative index data stores

First run with [microblob](https://github.com/miku/microblob), which turned out
to be ok, given that for some items it does dozens or hundreds.

Other options:

* sqlite3
* mysql
* pg
