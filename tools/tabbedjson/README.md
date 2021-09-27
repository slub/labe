# Non-generic JSON to TSV converter

```sh
$ tabbedjson -h
Usage of ./tabbedjson:
  -C    compress value; gz+b64
  -T    emit table showing possible savings through compression
```

Turns jsonlines with an "id" field into (id, doc) TSV.

```
$ head -1 ../../data/index.data | tabbedjson
ai-49-aHR0...uMi4xNTU       {"access_facet":"Electronic Resourc ... }
```

Around 80K docs/s. Supports some basic compression schema.

```sh
$ head -1 ../../data/index.data | tabbedjson -C
ai-49-aHR0...uMi4xNTU       H4sIAAAAAAAA/6xWTY/bNhC991cQPCWArJVUK177...
```

Display savings by using compression.

```sh
$ head ../../data/index.data | tabbedjson -C -T | column -t
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8wMDIxLTkwMTAuNjIuMi4xNTU  2223  1005  0.452092
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA3MDY1OQ               2217  993   0.447903
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8wMDIxLTkwMTAuNjIuMi4xNDY  2172  969   0.446133
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA0OTMxNQ               2325  1037  0.446022
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA3NDk0Ng               2340  1037  0.443162
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzOC9zai9sZXUvMjQwMTI4Mg       1841  973   0.528517
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA3Mzk5Mg               2118  937   0.442398
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzOC9zai9sZXUvMjQwMTI4Mw       1841  973   0.528517
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzOC9zai9sZXUvMjQwMTI4NA       1841  973   0.528517
ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA0MDIzMg               2148  989   0.460428
```

We want this to put our index data into a key value store. Compression (with
gzip) seems 3-4 slower than using no compression. Also, compression will need
extra handling at request time.
