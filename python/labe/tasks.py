"""

Derivation tasks.

$ tree -sh ~/.local/share/labe
/home/czygan/.local/share/labe
├── [4.0K]  IdMappingDatabase
│   └── [  29]  date-2021-11-25.db -> /home/czygan/tmp/id_to_doi.db
├── [4.0K]  OpenCitationsDatabase
│   ├── [144G]  a6d28e3e04bc206d3c4021a3381deb5c4443104b.db
│   └── [150G]  c90e82e35c9d02c00f81bee6d1f34b132953398c.db
├── [4.0K]  OpenCitationsDownload
│   ├── [  28]  a6d28e3e04bc206d3c4021a3381deb5c4443104b.zip -> /home/czygan/tmp/6741422.zip
│   └── [ 31G]  c90e82e35c9d02c00f81bee6d1f34b132953398c.zip
├── [4.0K]  OpenCitationsSingleFile
│   ├── [ 29G]  a6d28e3e04bc206d3c4021a3381deb5c4443104b.zst
│   └── [ 31G]  c90e82e35c9d02c00f81bee6d1f34b132953398c.zst
├── [4.0K]  SolrDatabase
│   ├── [ 40G]  date-2021-11-25-name-ai.db
│   ├── [5.5G]  date-2021-11-25-name-main.db
│   └── [217M]  date-2021-11-25-name-slub-production.db
└── [4.0K]  SolrFetchDocs
    ├── [5.5G]  date-2021-11-25-name-ai.zst
    ├── [968M]  date-2021-11-25-name-main.zst
    └── [ 23M]  date-2021-11-25-name-slub-production.zst

6 directories, 13 files

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


@functools.lru_cache(maxsize=None)
def open_citations_most_recent_download_url():
    ds = OpenCitationsDataset()
    return ds.most_recent_download_url()


class Task(BaseTask):
    """
    Superclass for labe tasks.
    """
    TAG = "labe"
    BASE = os.path.join(tempfile.gettempdir())

    # If the downloaded file is smaller than this, trigger an error. A basic
    # sanity check to notice, if download method has changed.
    OPEN_CITATION_DOWNLOAD_SIZE_THRESHOLD = 25_000_000_000

    def open_citations_url(self):
        """
        Open citations download url.
        """
        return open_citations_most_recent_download_url()

    def open_citations_url_hash(self):
        """
        We use the sha1 of the URL to understand whether task need to be
        regenerated.
        """
        url = self.open_citations_url()
        return hashlib.sha1(url.encode("utf-8")).hexdigest()


class OpenCitationsDownload(Task):
    """
    Download open citations corpus, currently via figshare.
    """
    def run(self):
        output = shellout("""curl --fail -sL "{url}" > {output}""",
                          url=self.open_citations_url())
        # do a basic sanity check right here, e.g. in 12/2021 filesizes were
        # about 30GB; we fail if the file size seems too small
        filesize = os.path.getsize(output)
        if filesize < self.OPEN_CITATION_DOWNLOAD_SIZE_THRESHOLD:
            raise RuntimeError(
                "open citations download is suspicously small: {}, want at least {}"
                .format(filesize, OPEN_CITATION_DOWNLOAD_SIZE_THRESHOLD))
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        filename = "{}.zip".format(self.open_citations_url_hash())
        return luigi.LocalTarget(path=self.path(filename=filename))


class OpenCitationsSingleFile(Task):
    """
    Turn nested zip files in to a single, zstd compressed flat file.

    As of 11/2021 OpenCitations download is a zip of zip files, but it is
    easier to work with a single compressed CSV file.

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

    A full run (e.g. download, single file, database) may take 2-3h: 143m33.560s
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
    Fetch JSON data from solr, store compressed; uses solrdump
    (https://git.io/J1pxG).

    Some timings: 190min for "main", 32s for "slub-production", 1374min for
    "ai" (22h); with -rows 50000: eta about 2.5h (134m27.012s).
    """
    date = luigi.DateParameter(default=datetime.date.today())
    name = luigi.Parameter(
        default="main", description="index name, url lookup up from a config")

    def run(self):
        try:
            url = self.config["indices"][self.name]
        except KeyError:
            raise LookupError('cannot map name to solr url')
        output = shellout("""
                          solrdump -verbose -server {server} -rows 50000 -fl 'id,title,author,format,url,doi_str_mv' |
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


class IdMappingTable(Task):
    """
    Generate a two column TSV mapping local identifier to DOI.
    """
    date = luigi.DateParameter(default=datetime.date.today())

    def requires(self):
        return {
            "main": SolrFetchDocs(date=self.date, name="main"),
            "ai": SolrFetchDocs(date=self.date, name="ai"),
        }

    def run(self):
        output = shellout(""" zstdcat -T0 {input} |
                              doisniffer |
                              jq -rc '[.id, .doi_str_mv[0]] | @tsv' |
                              zstd -c -T0 >> {output}
                          """,
                          input=self.input().get("main").path)
        shellout(""" zstdcat -T0 {input} |
                     jq -rc 'select(.doi_str_mv | length > 0) | [.id, .doi_str_mv[0]] | @tsv' |
                     zstd -T0 -c >> {output}
                 """,
                 input=self.input().get("ai").path,
                 output=output)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"))


class IdMappingDatabase(Task):
    """
    We need to sniff out DOI from all index data and build a (id, doi) database.
    """
    date = luigi.DateParameter(default=datetime.date.today())

    def requires(self):
        return IdMappingTable(date=self.date)

    def run(self):
        output = shellout(""" zstd -q -d -c -T0 {inputs} |
                              makta -init -o {output} -I 3
                          """,
                          inputs=self.input().path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="db"))
