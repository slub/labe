# THE LOCAL SUBSET

Which DOI are relevant? We need a list of all DOI in the catalog.

```
$ unpigz -c is.gz | grep '"DE-14"' | jq -r '["DE-14", .doi] | @tsv'
```
