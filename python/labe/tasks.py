"""

Derivation tasks.

$ tree -sh .local/share/labe
.local/share/labe
├── [4.0K]  OpenCitationsDatabase
│   └── [144G]  a6d28e3e04bc206d3c4021a3381deb5c4443104b.db
├── [4.0K]  OpenCitationsDownload
│   └── [  28]  a6d28e3e04bc206d3c4021a3381deb5c4443104b.zip -> /home/czygan/tmp/6741422.zip
├── [4.0K]  OpenCitationsSingleFile
│   └── [ 29G]  a6d28e3e04bc206d3c4021a3381deb5c4443104b.zst
└── [4.0K]  SolrFetchDocs
    ├── [5.5G]  date-2021-11-25-name-ai.zst
    ├── [968M]  date-2021-11-25-name-main.zst
    └── [ 23M]  date-2021-11-25-name-slub-production.zst

4 directories, 6 files

"""

import datetime
import functools
import hashlib
import os
import tempfile

import luigi

from labe.base import BaseTask, shellout
from labe.oci import OpenCitationsDataset

__all__ = [
    'IdMappingDatabase',
    'OpenCitationsDatabase',
    'OpenCitationsDownload',
    'OpenCitationsSingleFile',
    'SolrDatabase',
    'SolrFetchDocs',
    'Task',
]


class Task(BaseTask):
    """
    Superclass for labe tasks.
    """
    TAG = "labe"
    BASE = os.path.join(tempfile.gettempdir())

    @functools.lru_cache(maxsize=None)
    def open_citations_url(self):
        ds = OpenCitationsDataset()
        return ds.most_recent_download_url()

    @functools.lru_cache(maxsize=None)
    def open_citations_url_hash(self):
        url = self.open_citations_url()
        return hashlib.sha1(url.encode("utf-8")).hexdigest()


class OpenCitationsDownload(Task):
    """
    Download open citations corpus, currently via figshare.
    """
    def run(self):
        output = shellout("""curl -sL "{url}" > {output}""",
                          url=self.open_citations_url)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        filename = "{}.zip".format(self.open_citations_url_hash())
        return luigi.LocalTarget(path=self.path(filename=filename))


class OpenCitationsSingleFile(Task):
    """
    Turn nested zip files in to a single, zstd compressed flat file.

    As of 11/2021 OpenCitations download is a zip of zip files, but it is
    easier to work with a single CSV file.

    Also, figshare does not support HTTP range requests, which would allow us
    to convert zip files on the fly.
    """
    def requires(self):
        return OpenCitationsDownload()

    def run(self):
        output = shellout("""
                 T=$(mktemp -d) && unzip -d $T {file} &&
                 for f in $(find "$T" -name "*zip"); do
                     unzip -p "$f"
                 done | grep -vF 'oci,citing' | zstd -c -T0 > {output} &&
                 rm -rf "$T"
                 """,
                          file=self.input().path,
                          preserve_whitespace=True)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        filename = "{}.zst".format(self.open_citations_url_hash())
        return luigi.LocalTarget(path=self.path(filename=filename))


class OpenCitationsDatabase(Task):
    """
    Convert CSV to TSV and turn it into a sqlite3 database. The final
    conversion step. Requires the "makta" tool.

    TODO: exit code 1 on after ~50GB written, but not obvious why. [A
    wonderfully subtle bug, caused by a sloppy checking if a file exists as
    evidence the database schema exists as well; interestingly import does not
    fail, which is surprising; only at index creation time].

    2021/11/25 13:21:13 [io] written 57.6G · 24.1M/s
    2021/11/25 13:21:13 exit status 1

    ...

    sqlite> select count(*) from map;
    1104185948

    Task takes about 95m3.050s.
    """
    def requires(self):
        return OpenCitationsSingleFile()

    def run(self):
        output = shellout(r"""
                          zstdcat -T0 {input} |
                          cut -d, -f2,3 |
                          sed -e 's@,@\t@' |
                          makta -init -o {output} -I 3
                          """,
                          input=self.input().path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        filename = "{}.db".format(self.open_citations_url_hash())
        return luigi.LocalTarget(path=self.path(filename=filename))


class SolrFetchDocs(Task):
    """
    Fetch JSON data from solr, store compressed; using solrdump
    (https://git.io/J1pxG).

    Some timings: 190min for "main", 32s for "slub-production", 1374min for "ai" (22h).
    """
    date = luigi.DateParameter(default=datetime.date.today())
    name = luigi.Parameter(
        default="main", description="index name, url lookup up from a config")

    def run(self):
        # This should live elsewhere.
        urlmap = {
            "main":
            "https://index.ubl-proxy.slub-dresden.de/solr/biblio",
            "ai":
            "https://ai.ubl-proxy.slub-dresden.de/solr/biblio",
            "slub-production":
            "https://slubidx.ubl-proxy.slub-dresden.de/solr/slub-production",
        }
        try:
            url = urlmap[self.name]
        except KeyError:
            raise LookupError('cannot map name to solr url')
        output = shellout("""
                          solrdump -verbose -server {server} -rows 2000 -fl 'id,title,author,format,url' |
                          zstd -c -T0 > {output}
                          """,
                          server=url)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="zst"))


class SolrDatabase(Task):
    """
    Convert SOLR JSON into an sqlite database.

    Timings:

    main             1m48.112s
    slub-production  0m3.451s
    ai              12m47.890s
    """
    date = luigi.DateParameter(default=datetime.date.today())
    name = luigi.Parameter(
        default="main", description="index name, url lookup up from a config")

    def requires(self):
        return SolrFetchDocs(date=self.date, name=self.name)

    def run(self):
        output = shellout("""
                          zstdcat -T0 {input} |
                          tabjson |
                          makta -init -I 1 -o {output}
                          """,
                          input=self.input().path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="db"))


class IdMappingDatabase(Task):
    """
    We need to sniff out DOI from all index data and build a (id, doi) database.
    """
    date = luigi.DateParameter(default=datetime.date.today())

    def requires(self):
        return [
            SolrFetchDocs(date=self.date, name="main"),
            SolrFetchDocs(date=self.date, name="ai"),
        ]

    def run(self):
        raise NotImplementedError
        #
        # we need to extract (id, doi) pairs a bit differently from "main" and "ai"
        #
        # input_paths = " ".join([t.path for t in self.input()])
        # output = shellout("""
        #              zstdcat -T0 {inputs} |
        #              doisniffer |
        #              jq -rc '[.id, .doi_str_mv[0]] | @tsv' |
        #              makta -init -o {output} -I 3
        #          """,
        #                   inputs=input_paths)
        # luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="db"))

