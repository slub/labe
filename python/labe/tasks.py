"""
Derivation tasks.
"""

import configparser
import datetime
import functools
import hashlib
import os
import tempfile
import zipfile

import luigi

from labe.base import BaseTask, shellout
from labe.oci import OpenCitationsDataset

__all__ = [
    'CombinedUpdate',
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
    # Put all task outputs under BASE/TAG/...
    TAG = "labe"

    # Where all output will go by default.
    BASE = os.path.join(tempfile.gettempdir())

    # As a basic sanity check, fail, if output file fall below certain file sizes (in bytes).
    expected_output_file_sizes = {
        "IdMappingDatabase": 12_000_000_000,
        "IdMappingTable": 400_000_000,
        "OpenCitationsDatabase": 150_000_000_000,
        "OpenCitationsDownload": 25_000_000_000,
        "SolrDatabase-ai-True": 40_000_000_000,
        "SolrDatabase-main-False": 160_000_000_000,
        "SolrDatabase-main-True": 4_000_000_000,
        "SolrDatabase-slub-production-False": 2_000_000_000,
    }

    # Name of the server process. We need this in order to inform the server to
    # reload the database connections, when updates are available.
    labed_server_process_name = "labed"

    @functools.lru_cache(maxsize=None)
    def open_citations_url(self):
        """
        Open citations download url.
        """
        direct_download_url = None
        try:
            direct_download_url = self.config["oci"]["direct"]
        except (configparser.NoSectionError, configparser.NoOptionError, KeyError):
            pass
        finally:
            ds = OpenCitationsDataset(direct_download_url=direct_download_url)
            return ds.most_recent_download_url()

    def open_citations_url_hash(self):
        """
        We use the sha1 of the URL to understand whether task need to be
        regenerated. This assumes that a new download lives under a new URL.
        """
        url = self.open_citations_url()
        return hashlib.sha1(url.encode("utf-8")).hexdigest()

    def validate_output_filesize(self, filename, minimum_size):
        """
        Raise an exception, if filename is smaller than filesize.
        """
        filesize = os.path.getsize(filename)
        if filesize < minimum_size:
            raise RuntimeError("{}: unexpected file size, got {}, want at least {}".format(
                filename, filesize, minimum_size))


class OpenCitationsDownload(Task):
    """
    Download open citations corpus, currently via figshare.
    """
    def run(self):
        url = self.open_citations_url()
        output = shellout("""
                          curl --fail -sL "{url}" > {output}
                          """,
                          url=url)

        # Do a basic sanity check right here, e.g. in 12/2021 filesize was
        # about 30GB; we fail if the file size seems too small.
        self.validate_output_filesize(output, self.expected_output_file_sizes["OpenCitationsDownload"])
        # We excect a zip file.
        if not zipfile.is_zipfile(output):
            raise RuntimeError("not a zip: {}".format(output))

        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        filename = "{}.zip".format(self.open_citations_url_hash())
        return luigi.LocalTarget(path=self.path(filename=filename))

    def on_success(self):
        self.create_symlink(name="current")


class OpenCitationsSingleFile(Task):
    """
    Turn nested zip files into a single, undecorated, flat zstd compressed
    file. As of 11/2021 OpenCitations download is a zip of zip files, but it is
    easier to work with a single compressed CSV file.

    Also, figshare does not support HTTP range requests, which would allow us
    to convert zip files on the fly. Pity.
    """
    def requires(self):
        return OpenCitationsDownload()

    def run(self):
        """
        Decompress, find zips, decompress, remove decoration, compress,
        cleanup.
        """
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
        fingerprint = self.open_citations_url_hash()
        filename = "{}.zst".format(fingerprint)
        return luigi.LocalTarget(path=self.path(filename=filename))

    def on_success(self):
        self.create_symlink(name="current")


class OpenCitationsDatabase(Task):
    """
    Convert CSV to TSV and turn it into a sqlite3 database. The final
    conversion step. Requires the "makta" tool (https://git.io/J9LRT).

    In 12/2021, task took about 95m3.050s. A full sequence (e.g. download,
    single file, database) can take 2-3h (143m33.560s).
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
        self.validate_output_filesize(output, self.expected_output_file_sizes["OpenCitationsDatabase"])
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        fingerprint = self.open_citations_url_hash()
        filename = "{}.db".format(fingerprint)
        return luigi.LocalTarget(path=self.path(filename=filename))

    def on_success(self):
        self.create_symlink(name="current")


class SolrFetchDocs(Task):
    """
    Fetch JSON data from solr; uses solrdump (https://git.io/J1pxG).

    Some timings: 190min for "main", 32s for "slub-production", 1374min for
    "ai" (22h); with "-rows 50000" eta about 2.5h (134m27.012s).
    """
    date = luigi.DateParameter(default=datetime.date.today())
    name = luigi.Parameter(default="main", description="index name, url lookup up from a config")
    short = luigi.BoolParameter(description="only fetch id,title,author,format,url,doi_str_mv fields, e.g. for ai")

    def run(self):
        try:
            indices = self.config["indices"]
            url = indices[self.name]
        except KeyError:
            raise LookupError('cannot map name to solr url, available indices: {}'.format(indices.keys()))
        extra_opts = ''
        if self.short:
            extra_opts = "-fl 'id,title,author,format,url,doi_str_mv,institution'"
        output = shellout("""
                          solrdump -verbose -server {server} -rows 50000 {extra_opts} |
                          zstd -c -T0 > {output}
                          """,
                          server=url,
                          extra_opts=extra_opts)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="zst"))

    def on_success(self):
        name = "{}-short".format(self.name) if self.short else self.name
        self.create_symlink(name="current", suffix=name)


class SolrDatabase(Task):
    """
    Convert SOLR JSON documents into an sqlite database to allow lookup of
    documents by key. Requires the small "tabjson" tool (https://git.io/J9LRH).

    Some timings: ai 12m47.890s, main 1m48.112s, slub-production 0m3.451s.
    """
    date = luigi.DateParameter(default=datetime.date.today())
    name = luigi.Parameter(default="main", description="index name, url lookup up from a config")
    short = luigi.BoolParameter(description="only fetch id,title,author,format,url,doi_str_mv fields, e.g. for ai")

    def requires(self):
        return SolrFetchDocs(date=self.date, name=self.name, short=self.short)

    def run(self):
        output = shellout("""
                          zstdcat -T0 {input} |
                          tabjson |
                          makta -init -I 1 -o {output}
                          """,
                          input=self.input().path)
        self.validate_output_filesize(
            output, self.expected_output_file_sizes["SolrDatabase-{}-{}".format(self.name, self.short)])
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="db"))

    def on_success(self):
        name = "{}-short".format(self.name) if self.short else self.name
        self.create_symlink(name="current", suffix=name)


class IdMappingTable(Task):
    """
    Generate a two column TSV mapping local identifiers to their DOI. May
    require the "doisniffer" tool (https://git.io/J9L0D) and GNU parallel,
    about 15min.
    """
    date = luigi.DateParameter(default=datetime.date.today())

    def requires(self):
        return {
            "slub-production": SolrFetchDocs(date=self.date, name="slub-production", short=False),
            "main": SolrFetchDocs(date=self.date, name="main", short=False),
            "ai": SolrFetchDocs(date=self.date, name="ai", short=True),
        }

    def run(self):
        # In 01/2022, for "main", we still need to apply "doisniffer", but that
        # may change in the future.
        output = shellout(""" zstd -q -d -c -T0 {input} |
                              doisniffer |
                              jq -rc '[.id, .doi_str_mv[0]] | @tsv' |
                              zstd -c -T0 >> {output}
                          """,
                          input=self.input().get("main").path)

        # In 01/2022, we use "doisniffer" for slub-production as well.
        shellout(""" zstd -q -d -c -T0 {input} |
                              doisniffer |
                              jq -rc '[.id, .doi_str_mv[0]] | @tsv' |
                              zstd -c -T0 >> {output}
                 """,
                 output=output,
                 input=self.input().get("slub-production").path)

        # In 01/2022, the "doi_str_mv" field is included in "ai" - with 73881207 values.
        shellout(""" zstd -q -d -c -T0 {input} |
                     parallel -j 8 --pipe --block 10M "jq -rc 'select(.doi_str_mv | length > 0) | [.id, .doi_str_mv[0]] | @tsv'" |
                     zstd -T0 -c >> {output}
                 """,
                 output=output,
                 input=self.input().get("ai").path)

        self.validate_output_filesize(output, self.expected_output_file_sizes["IdMappingTable"])
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"))

    def on_success(self):
        self.create_symlink(name="current")


class IdMappingDatabase(Task):
    """
    Generate a (id, doi) mapping database (5min).
    """
    date = luigi.DateParameter(default=datetime.date.today())

    def requires(self):
        return IdMappingTable(date=self.date)

    def run(self):
        output = shellout(""" zstd -q -d -c -T0 {input} |
                              makta -init -o {output} -I 3
                          """,
                          input=self.input().path)
        self.validate_output_filesize(output, self.expected_output_file_sizes["IdMappingDatabase"])
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="db"))

    def on_success(self):
        self.create_symlink(name="current")


class CombinedUpdate(luigi.WrapperTask):
    """
    Wrapper around generation of the the various databases required for the labed service.
    """
    date = luigi.DateParameter(default=datetime.date.today())

    def requires(self):
        yield SolrDatabase(date=self.date, name="ai", short=True)
        yield SolrDatabase(date=self.date, name="main", short=True)
        yield SolrDatabase(date=self.date, name="slub-production", short=False)
        yield IdMappingDatabase(date=self.date)
        yield OpenCitationsDatabase()
