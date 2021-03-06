#!/usr/bin/env python

"""
Sniff out DOI from newline delimited JSON (from SOLR). Workaround the fact,
that there is currently no explicit DOI field in the index data.

Example usage:

    $ zstdcat -T0 ma.json.zst | parallel --pipe -j 8 --block 10M 'python doisniffer.py -t'
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

Slow, about 2 krps. Issues remain:

* 10.1787/agr-outl-data-en上查询。
"""

import argparse
import collections
import fileinput
import os
import re
import sys

import marcx
import orjson as json

Match = collections.namedtuple("Match", "key value match")


def field_match(key, value, pattern, parse_marc=True):
    """
    Given a root document, all matches of pattern per field. Also parse
    "fullrecord" into MARC21 and search each field for the pattern.
    """
    if key == "fullrecord" and parse_marc:
        try:
            record = marcx.FatRecord(data=value.encode("utf-8"))
            for v in record.flatten():
                match = pattern.search(v)
                if not match:
                    continue
                yield Match(key + ":marc", value, match)
        except UnicodeDecodeError as exc:
            print("[skip] cannot create MARC record: {}".format(exc), file=sys.stderr)
            yield None
    elif isinstance(value, str):
        for match in pattern.finditer(value):
            yield Match(key, value, match)
    elif isinstance(value, list):
        for v in value:
            yield from field_match(key, v, pattern, parse_marc=parse_marc)
    elif isinstance(value, dict):
        for k, v in value.items():
            yield from field_match(k, v, pattern, parse_marc=parse_marc)
    else:
        yield None


def sniff_doi(
    fobj, writer=sys.stdout, aggressive=False, tab=True, update=False, parse_marc=True
):
    """
    Sniff out DOI in a newline delimited JSON file. Output to stdout. If
    aggressive is True, try to filter out invalid DOI more aggressively.
    Arguments tab and update control whether we want to list the DOI or update
    the document.
    """
    # Q: does a DOI allow slashes in the non-prefix, e.g. "10.123/abc/epdf"
    # A: yes, rare, but that's ok! 10.6094/UNIFR/13040, 10.1051/hel/2019018, ...
    pat_doi = re.compile(r'10[.][0-9]{2,6}/[^ "]{3,}')

    for i, line in enumerate(fobj):
        doc = json.loads(line)
        sniffed = set()
        for match in field_match(None, doc, pat_doi, parse_marc=parse_marc):
            if not match:
                continue
            value = match.match.group(0)
            # some exceptions
            if value.endswith("/epdf"):
                value = value[: len(value) - 5]
            if value.endswith(".") or value.endswith("*"):
                # 10.1007/978-3-322-84738-6.
                # 10.4028/www.scientific.net/AMR.429*
                value = value[:-1]
            # can rule out barcode directly
            if aggressive and ("barcode" in match.key or "dewey" in match.key):
                print("[skip] {}:{}".format(match.key, value), file=sys.stderr)
                sniffed = set()
                break
            sniffed.add(value)
            if tab:
                writer.write("{}\t{}\t{}\n".format(doc["id"], match.key, value,))
        if update:
            # <dynamicField name="*_str_mv" type="string" indexed="true"
            # stored="true" multiValued="true" docValues="true"/>
            if len(sniffed) > 1:
                print(
                    "[warn] multiple doi: {}".format(", ".join(sniffed)),
                    file=sys.stderr,
                )
            doc["doi_str_mv"] = list(sniffed)
            writer.write(json.dumps(doc).decode("utf-8"))
            writer.write("\n")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter
    )
    parser.add_argument(
        "-i", "--input-file", nargs="?", type=argparse.FileType("r"), default=sys.stdin
    )
    parser.add_argument("-t", "--tab", action="store_true", help="generate table")
    parser.add_argument(
        "-u", "--update", action="store_true", help="update record inline"
    )
    parser.add_argument(
        "--aggressive", action="store_true", help="filter out mostly like bad doi"
    )
    parser.add_argument(
        "-m", "--parse-marc", action="store_true", help="parse fullrecord as marc"
    )
    parser.add_argument(
        "-q", "--quiet", action="store_true", help="suppress any output"
    )
    args = parser.parse_args()
    writer = os.devnull if args.quiet else sys.stdout
    sniff_doi(
        args.input_file,
        writer=writer,
        aggressive=args.aggressive,
        tab=args.tab,
        update=args.update,
        parse_marc=args.parse_marc,
    )
