#!/usr/bin/env python

"""
Sniff out DOI from newline delimited JSON.
"""

import fileinput
import re
import collections
import json

Match = collections.namedtuple("Match", "key value match")

def field_match(key, value, pattern):
    """
    Given a root document, all matches of pattern per field.
    """
    if isinstance(value, list):
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



if __name__ == '__main__':
    pat_doi = re.compile(r'10[.][0-9]{2,6}/[^ "]{3,}')
    pad = 30
    for i, line in enumerate(fileinput.input()):
        doc = json.loads(line)
        for match in field_match(None, doc, pat_doi):
            if not match:
                continue
            print("{}\t{}\t{}".format(i, match.key, match.match.group(0)))

# $ zstdcat -T0 ma.json.zst | python doisniffer.py
# 3855    fullrecord      10.73/0941qOCLC219
# 3855    dewey-full      10.73/0941
# 3855    dewey-raw       10.73/0941
# 3855    dewey-search    10.73/0941
# 8566    barcode_de105   10.2626/18.
# 13374   barcode_de105   10.5763/18.
# 19330   fullrecord      10.1007/978-1-4613-0893-52doi8
# 19330   ismn    10.1007/978-1-4613-0893-5
# 19330   marc024a_ct_mv  10.1007/978-1-4613-0893-5
# 21195   barcode_de105   10.1424/18.
# 21997   fullrecord      10.73/028/5qOCLC21910aNursing
# 21997   dewey-full      10.73/028/5
# 21997   dewey-raw       10.73/028/5
# 21997   dewey-search    10.73/028/5
# 22603   fullrecord      10.71/1qOCLC219

