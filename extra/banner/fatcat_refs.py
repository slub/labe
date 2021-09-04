#!/usr/bin/env python

"""
Fetch references from fatcat ref search index.

> https://search.fatcat.wiki/fatcat_ref/_search?q=source_release_ident:3jyo6iz5yrc4vbczch3mx4zfvy
"""

import requests
import json
from urllib.parse import urljoin
import logging
import backoff

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
    For a ident, return all docs matching criteria.
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
        for doc in resp["hits"]["hits"]:
            yield doc["_source"]
        relation = resp["hits"]["total"]["relation"]
        if relation != "eq":
            raise ValueError(
                "expected eq for total relation, found: {}".format(relation)
            )
        total = resp["hits"]["total"]["value"]
        if total < (offset + size):
            break

def fetch_releases(ids, api="https://api.fatcat.wiki/v0/release/"):
    """
    Fetch a batch of releases by identifier from fatcat api.
    """
    for id in ids:
        url = urljoin(api, id)
        logger.debug(url)
        yield requests.get(url).json()

def bfs(field="target_release_ident", ident="3jyo6iz5yrc4vbczch3mx4zfvy", max_docs=100):
    other = (
        "source_release_ident"
        if field == "target_release_ident"
        else "target_release_ident"
    )
    seen = set()
    queue = [ident]
    while queue:
        logger.debug("bfs: queue size: %s, seen: %s", len(queue), len(seen))
        ident = queue.pop()
        seen.add(ident)
        if len(seen) == max_docs:
            break
        for doc in fetch_docs(field=field, ident=ident):
            queue.append(doc[other])
    return seen


if __name__ == "__main__":
    data = list(fetch_docs())
    print(data)
    ids = bfs(max_docs=3)
    print(ids)
    print(list(fetch_releases(ids)))
