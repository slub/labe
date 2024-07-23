"""
Helper to calculate diff.
"""

import datetime


def stats_diff(a, b):
    """
    Given two dictionaries, generate a dictionary with diffs between numeric
    values.

    {
      "version": "1",
      "date": "2022-02-02",
      "slub": {
        "num_mapped_doi": 56622824,
        "num_common_doi": 37942194,
        "ratio": 0.5318680435710712
      },
      "index": {
        "num_mapped_doi": 66586997,
        "num_common_doi": 44311206,
        "ratio": 0.6211479084075822
      },
      "oci": {
        "num_edges": 1271360866,
        "num_doi": 71337608,
        "stats_inbound": {
            ...
        },
        "stats_outbound": {
            ...
        }
      }
    }
    """
    diff = {
        "version": "1",
        "date": str(datetime.date.today()),
    }
    keys = ["index", "slub"]
    for key in keys:
        if key in a and key in b:
            diff["num_mapped_doi"] = b[key]["num_mapped_doi"] - a[key]["num_mapped_doi"]
            diff["num_common_doi"] = b[key]["num_common_doi"] - a[key]["num_common_doi"]
            diff["ratio"] = b[key]["ratio"] - a[key]["ratio"]

    return diff
