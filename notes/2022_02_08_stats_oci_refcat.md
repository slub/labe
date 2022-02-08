# Short note on "coverage" of OCI/COCO and OCI/COCI + refcat

> 2022-02-08

## TL;DR

Of the 56647568 DOI we find in SLUB metadata:

* 37941309 or 66.98% have at least one edge in the OCI graph
* 39953592 or 70.53% have at least one edge in the combined OCI/refcat graph

## OCI SLUB (DE-14)

```json
{
  "version": "2",
  "date": "2022-02-08",
  "institution": {
    "DE-14": {
      "num_mapped_doi": 56647568,
      "num_common_doi": 37941309,
      "ratio_common_mapped": 0.6697782506744155,
      "ratio_common_exp": 0.5318556377724355
    }
  },
  "index": {
    "num_mapped_doi": 66614507,
    "num_common_doi": 44310359,
    "ratio_common_mapped": 0.6651758152319585,
    "ratio_common_exp": 0.621136035287306
  },
  "oci": {
    "num_edges": 1271360866,
    "num_doi": 71337608,
    "stats_inbound": {
      "count": 58110382,
      "mean": 21.8783773612,
      "std": 104.0590729122,
      "min": 1,
      "0%": 1,
      "10%": 1,
      "25%": 2,
      "50%": 7,
      "75%": 20,
      "95%": 81,
      "99%": 220,
      "99.9%": 802,
      "100%": 200934,
      "max": 200934
    },
    "stats_outbound": {
      "count": 53437627,
      "mean": 23.7914918265,
      "std": 32.4106185876,
      "min": 1,
      "0%": 1,
      "10%": 2,
      "25%": 6,
      "50%": 16,
      "75%": 32,
      "95%": 68,
      "99%": 133,
      "99.9%": 307,
      "100%": 10557,
      "max": 10557
    }
  }
}
```

## OCI + REFCAT SLUB (DE-14)

```json
{
  "version": "2",
  "date": "2022-02-08",
  "institution": {
    "DE-14": {
      "num_mapped_doi": 56647568,
      "num_common_doi": 39953592,
      "ratio_common_mapped": 0.7053010995988389,
      "ratio_common_exp": 0.5057546348498173
    }
  },
  "index": {
    "num_mapped_doi": 66614507,
    "num_common_doi": 46710582,
    "ratio_common_mapped": 0.7012073511254838,
    "ratio_common_exp": 0.5912883463152062
  },
  "exp": {
    "num_edges": 1565087463,
    "num_doi": 78997975,
    "stats_inbound": {
      "count": 62011707,
      "mean": 25.2385805635,
      "std": 302.9702253606,
      "min": 1,
      "0%": 1,
      "10%": 1,
      "25%": 2,
      "50%": 7,
      "75%": 20,
      "95%": 84,
      "99%": 238,
      "99.9%": 1232,
      "100%": 541765,
      "max": 541765
    },
    "stats_outbound": {
      "count": 62253064,
      "mean": 25.1407298282,
      "std": 177.4013692967,
      "min": 1,
      "0%": 1,
      "10%": 2,
      "25%": 5,
      "50%": 14,
      "75%": 31,
      "95%": 68,
      "99%": 144,
      "99.9%": 516,
      "100%": 39160,
      "max": 39160
    }
  }
}
```
