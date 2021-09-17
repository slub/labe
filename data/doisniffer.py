#!/usr/bin/env python

"""
Sniff out DOI from newline delimited JSON. This is part of a workaround the
fact, that there is no explicit DOI field in the index data.

Example usage:

    $ zstdcat -T0 ma.json.zst | parallel --pipe -j 8 --block 10M 'python doisniffer.py'
    0-011506326     fullrecord:marc 10.73/0941
    0-011506326     dewey-full      10.73/0941
    0-011506326     dewey-raw       10.73/0941
    0-011506326     dewey-search    10.73/0941
    0-013497936     barcode_de105   10.2626/18.
    0-015609626     barcode_de105   10.5763/18.
    0-016998340     fullrecord:marc 10.1007/978-1-4613-0893-5
    0-016998340     ismn    10.1007/978-1-4613-0893-5
    0-016998340     marc024a_ct_mv  10.1007/978-1-4613-0893-5
    0-017646340     barcode_de105   10.1424/18.
    0-017964148     fullrecord:marc 10.73/028/5
    0-017964148     dewey-full      10.73/028/5
    0-017964148     dewey-raw       10.73/028/5
    0-017964148     dewey-search    10.73/028/5
    0-018767389     fullrecord:marc 10.69/52/019
    0-018767389     dewey-full      10.69/52/019
    0-018767389     dewey-raw       10.69/52/019
    0-018767389     dewey-search    10.69/52/019
    0-020574460     fullrecord:marc 10.1007/978-3-642-65371-1
    0-020574460     spelling        10.1007/978-3-642-65371-1
    0-020596286     fullrecord:marc 10.1007/978-3-322-88793-1
    0-020596286     ismn    10.1007/978-3-322-88793-1
    0-020596286     marc024a_ct_mv  10.1007/978-3-322-88793-1
    0-021232083     fullrecord:marc 10.13109/9783666532573
    ...

Slow, about 2 krps.

"""

import collections
import fileinput
import json
import re

import marcx

Match = collections.namedtuple("Match", "key value match")


def field_match(key, value, pattern):
    """
    Given a root document, all matches of pattern per field. Also parse
    "fullrecord" into MARC21 and search each field for the pattern.
    """
    if key == "fullrecord":
        record = marcx.FatRecord(data=value.encode("utf-8"))
        for v in record.flatten():
            match = pattern.search(v)
            if not match:
                continue
            yield Match(key + ":marc", value, match)
    elif isinstance(value, list):
        for v in value:
            yield from field_match(key, v, pattern)
    elif isinstance(value, str):
        matches = list(pattern.finditer(value))
        for match in matches:
            yield Match(key, value, match)
    elif isinstance(value, dict):
        for k, v in value.items():
            yield from field_match(k, v, pattern)
    else:
        yield None


if __name__ == "__main__":
    pat_doi = re.compile(r'10[.][0-9]{2,6}/[^ "]{3,}')
    pad = 30
    for i, line in enumerate(fileinput.input()):
        doc = json.loads(line)
        for match in field_match(None, doc, pat_doi):
            if not match:
                continue
            print(
                "{}\t{}\t{}".format(doc["id"], match.key, match.match.group(0))
            )

