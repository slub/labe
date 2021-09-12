#!/usr/bin/env python

"""
Fetch references from fatcat ref search index.

> https://search.fatcat.wiki/fatcat_ref/_search?q=source_release_ident:3jyo6iz5yrc4vbczch3mx4zfvy
"""

import argparse
import json
import logging
from urllib.parse import urljoin

import backoff
import requests

logger = logging.getLogger("simple_example")
logger.setLevel(logging.DEBUG)

ch = logging.StreamHandler()
ch.setLevel(logging.DEBUG)

formatter = logging.Formatter(
    "[%(asctime)s][%(name)s][%(levelname)-8s] %(message)s", datefmt="%Y-%m-%d %H:%M:%S"
)
ch.setFormatter(formatter)

logger.addHandler(ch)


def fetch_docs(
    field="target_release_ident",
    ident="3jyo6iz5yrc4vbczch3mx4zfvy",
    server="https://search.fatcat.wiki",
    index="fatcat_ref",
    size=100,
):
    """
    For a ident, yield all documents.
    """
    offset = 0
    while True:
        query = {
            "track_total_hits": True,
            "from": offset,
            "size": size,
            "query": {"term": {field: ident,}},
        }
        search_url = urljoin(server, "{}/_search".format(index))
        logger.debug(search_url)
        resp = requests.get(search_url, json=query)
        if resp.status_code >= 400:
            raise RuntimeError(
                "got {} for {} {}".format(resp.status_code, search_url, query)
            )
        resp = resp.json()
        relation = resp["hits"]["total"]["relation"]
        if relation != "eq":
            raise ValueError(
                "expected eq for total relation, found: {}".format(relation)
            )

        for doc in resp["hits"]["hits"]:
            yield doc["_source"]

        total = resp["hits"]["total"]["value"]
        if total < (offset + size):
            break
        else:
            offset = offset + size


def fetch_releases(ids, api="https://api.fatcat.wiki/v0/release/"):
    """
    Fetch a batch of releases by identifier from fatcat api.
    """
    for id in ids:
        url = urljoin(api, id)
        logger.debug(url)
        yield requests.get(url).json()


def edge_id_pair_to_release(tup):
    releases = list(fetch_releases(tup))
    return (releases[0], releases[1])


def bfs(
    field="target_release_ident", ident="3jyo6iz5yrc4vbczch3mx4zfvy", max_edges=100
):
    """
    TODO: we want to record the edges, not just the ids.
    """
    other = (
        "source_release_ident"
        if field == "target_release_ident"
        else "target_release_ident"
    )
    edges = set()
    queue = [ident]
    while queue:
        logger.debug("bfs: queue size: %s, edges: %s", len(queue), len(edges))
        ident = queue.pop()
        if len(edges) > max_edges:
            break
        for doc in fetch_docs(field=field, ident=ident):
            queue.append(doc[other])
            edges.add((ident, doc[other]))
    return edges


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter
    )
    parser.add_argument(
        "--max-edges",
        default=10,
        type=int,
        help="consider at most this number of edges (approximately)",
    )
    parser.add_argument(
        "--field", default="target_release_ident", type=str, help="which field to query"
    )
    parser.add_argument(
        "--ident",
        default="d4oy6i3aa5eethsoxh7onpmzuq",
        type=str,
        help="id to start with",
    )
    parser.add_argument(
        "--compact-table",
        default=False,
        action="store_true",
        help="more compact output",
    )
    args = parser.parse_args()
    data = list(fetch_docs(field=args.field, ident=args.ident))
    logger.debug(data)
    edges = bfs(max_edges=args.max_edges)
    logger.debug(edges)
    for a, b in [edge_id_pair_to_release(t) for t in edges]:
        if args.compact_table:
            fields = (
                b["title"][:32],
                str(b.get("release_year", "MISS")),
                "=>",
                a["title"][:32],
                str(a.get("release_year", "MISS")),
            )
            print("  ".join(fields))
        else:
            fields = (
                b["ident"],
                b["title"][:32],
                str(b.get("release_year", "NA")),
                "=>",
                a["ident"],
                a["title"][:32],
                str(a.get("release_year", "NA")),
            )
            print("\t".join(fields))
