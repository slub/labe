"""
Sniff out DOI from finc/vufind SOLR documents.
"""

import collections
import re

import marcx

Match = collections.namedtuple("Match", "key value match")


def field_match(key, value, pattern, parse_marc=True):
    """
    Given a root document, return all matches of pattern per field. Also parse
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
            print("[skip] cannot create MARC record: {}".format(exc),
                  file=sys.stderr)
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


def sniff_doi(fobj,
              writer=sys.stdout,
              aggressive=False,
              tab=True,
              update=False,
              parse_marc=True,
              doi_pattern=r'10[.][0-9]{2,6}/[^ "]{3,}'):
    """
    Sniff out DOI in a newline delimited JSON file. Output to stdout by
    default. If aggressive is True, try to filter out invalid DOI more
    aggressively. Arguments tab and update control whether we want to list the
    DOI or update the document.
    """
    # Q: does a DOI allow slashes in the non-prefix, e.g. "10.123/abc/epdf"
    # A: yes, rare, but that's ok! 10.6094/UNIFR/13040, 10.1051/hel/2019018, ...
    pat_doi = re.compile(doi_pattern)

    for i, line in enumerate(fobj):
        doc = json.loads(line)
        sniffed = set()
        for match in field_match(None, doc, pat_doi, parse_marc=parse_marc):
            if not match:
                continue
            value = match.match.group(0)
            # some exceptions
            if value.endswith("/epdf"):
                value = value[:len(value) - 5]
            if value.endswith(".") or value.endswith("*"):
                # 10.1007/978-3-322-84738-6.
                # 10.4028/www.scientific.net/AMR.429*
                value = value[:-1]
            # can rule out barcode directly
            if aggressive and ("barcode" in match.key or "dewey" in match.key):
                print("[skip] {}:{}".format(match.key, value), file=sys.stderr)
                sniffed = set()  # do we need this
                break
            sniffed.add(value)
            if tab:
                writer.write("{}\t{}\t{}\n".format(
                    doc["id"],
                    match.key,
                    value,
                ))
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


def test_field_match():
    Case = collections.namedtuple("Case", "key value count")
    cases = (
        Case(key=None, value={}, count=0),
        Case(key=None, value={"a": "abc"}, count=0),
        Case(key=None, value={"a": "10.123/123"}, count=1),
    )
    pat_doi = re.compile(r'10[.][0-9]{2,6}/[^ "]{3,}')
    for c in cases:
        result = list(field_match(c.key, c.value, pat_doi))
        assert c.count == len(result), c
