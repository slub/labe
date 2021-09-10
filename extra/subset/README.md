# THE LOCAL SUBSET

Which DOI are relevant? We need a list of all DOI in the catalog.

```
$ unpigz -c is.gz | grep '"DE-14"' | jq -r '["DE-14", .doi] | @tsv'
```

About 8000 unique prefixes:

```
$ head prefix.freq
9316875 10.1016
3820156 10.1002
2848038 10.1109
2842013 10.1080
2694827 10.1111
2056527 10.1093
2021384 10.2307
1788098 10.1177
1582397 10.1097
1505757 10.1021
```
