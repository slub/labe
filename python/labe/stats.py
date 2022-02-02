"""
Stats related tasks. Mainly:

* [ ] number of DOI in citation corpus
* [ ] edge count distibution (across the whole graph)
* [ ] size of overlap with the index
"""

import datetime
import os
import tempfile

import luigi
import pandas as pd

from labe.base import T, shellout
from labe.tasks import IdMappingTable, OpenCitationsSingleFile, Task

__all__ = [
    'IndexMappedDOI',
    'OpenCitationsCitedCount',
    'OpenCitationsCitingCount',
    'OpenCitationsInboundStats',
    'OpenCitationsOutboundStats',
    'OpenCitationsSourceDOI',
    'OpenCitationsTargetDOI',
    'OpenCitationsUniqueDOI',
    'StatsCommonDOI',
    'StatsReportData',
]


class OpenCitationsSourceDOI(Task):
    """
    List of DOI that are source of a citation edge. Normalized and sorted.
    18m54.167s.

    Issues: We still have '"' in DOI.
    """

    def requires(self):
        return OpenCitationsSingleFile()

    def run(self):
        output = shellout("""
                          zstdcat -T0 {input} |
                          LC_ALL=C cut -d, -f 2 |
                          LC_ALL=C tr [:upper:] [:lower:] |
                          LC_ALL=C sort -S50% |
                          zstd -c -T0 > {output}
                          """,
                          input=self.input().path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        fingerprint = self.open_citations_url_hash()
        filename = "{}.tsv.zst".format(fingerprint)
        return luigi.LocalTarget(path=self.path(filename=filename))

    def on_success(self):
        self.create_symlink(name="current")


class OpenCitationsTargetDOI(Task):
    """
    List of DOI that are target of a citation edge. Normalized and sorted.
    """

    def requires(self):
        return OpenCitationsSingleFile()

    def run(self):
        output = shellout("""
                          zstdcat -T0 {input} |
                          LC_ALL=C cut -d, -f 3 |
                          LC_ALL=C tr [:upper:] [:lower:] |
                          LC_ALL=C sort -S50% |
                          zstd -c -T0 > {output}
                          """,
                          input=self.input().path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        fingerprint = self.open_citations_url_hash()
        filename = "{}.tsv.zst".format(fingerprint)
        return luigi.LocalTarget(path=self.path(filename=filename))

    def on_success(self):
        self.create_symlink(name="current")


class OpenCitationsCitedCount(Task):
    """
    Generate a table with two columns: inbound link count and DOI.
    """

    def requires(self):
        return OpenCitationsTargetDOI()

    def run(self):
        output = shellout(r"""
                          zstdcat -T0 {input} |
                          LC_ALL=C uniq -c |
                          LC_ALL=C sort -S 40% -nr |
                          LC_ALL=C sed -e 's@^[ ]*@@;s@ @\t@' |
                          zstd -c -T0 > {output}
                          """,
                          input=self.input().path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        fingerprint = self.open_citations_url_hash()
        filename = "{}.tsv.zst".format(fingerprint)
        return luigi.LocalTarget(path=self.path(filename=filename))

    def on_success(self):
        self.create_symlink(name="current")


class OpenCitationsCitingCount(Task):
    """
    Generate a table with two columns: outbound link count and DOI.
    """

    def requires(self):
        return OpenCitationsSourceDOI()

    def run(self):
        output = shellout(r"""
                          zstdcat -T0 {input} |
                          LC_ALL=C uniq -c |
                          LC_ALL=C sort -S 40% -nr |
                          LC_ALL=C sed -e 's@^[ ]*@@;s@ @\t@' |
                          zstd -c -T0 > {output}
                          """,
                          input=self.input().path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        fingerprint = self.open_citations_url_hash()
        filename = "{}.tsv.zst".format(fingerprint)
        return luigi.LocalTarget(path=self.path(filename=filename))

    def on_success(self):
        self.create_symlink(name="current")


class OpenCitationsInboundStats(Task):
    """
    Inbound edge count distribution.

      {
        "inbound_edges": {
          "count": 58110382,
          "mean": 21.8783773612,
          "std": 104.0590729122,
          "min": 1,
          "0%": 1,
          "10%": 1,
          "25%": 2,
          "50%": 7,
          "75%": 20,
          "95%": 81,
          "99%": 220,
          "99.9%": 802,
          "100%": 200934,
          "max": 200934
        }
      }
    """

    def requires(self):
        return OpenCitationsCitedCount()

    def run(self):
        output = shellout("zstdcat -T0 {input} | cut -f1 > {output}", input=self.input().path)
        df = pd.read_csv(output, header=None, names=["inbound_edges"], skip_blank_lines=True)
        percentiles = [0, 0.1, 0.25, 0.5, 0.75, 0.95, 0.99, 0.999, 1]
        with tempfile.NamedTemporaryFile(mode="wb", delete=False) as f:
            df.describe(percentiles=percentiles).to_json(f)
        luigi.LocalTarget(f.name).move(self.output().path)
        os.remove(output)

    def output(self):
        fingerprint = self.open_citations_url_hash()
        filename = "{}.json".format(fingerprint)
        return luigi.LocalTarget(path=self.path(filename=filename))

    def on_success(self):
        self.create_symlink(name="current")


class OpenCitationsOutboundStats(Task):
    """
    Outbound edge count distribution.
    """

    def requires(self):
        return OpenCitationsCitingCount()

    def run(self):
        output = shellout("zstdcat -T0 {input} | cut -f1 > {output}", input=self.input().path)
        df = pd.read_csv(output, header=None, names=["outbound_edges"], skip_blank_lines=True)
        percentiles = [0, 0.1, 0.25, 0.5, 0.75, 0.95, 0.99, 0.999, 1]
        with tempfile.NamedTemporaryFile(mode="wb", delete=False) as f:
            df.describe(percentiles=percentiles).to_json(f)
        luigi.LocalTarget(f.name).move(self.output().path)
        os.remove(output)

    def output(self):
        fingerprint = self.open_citations_url_hash()
        filename = "{}.json".format(fingerprint)
        return luigi.LocalTarget(path=self.path(filename=filename))

    def on_success(self):
        self.create_symlink(name="current")


class OpenCitationsUniqueDOI(Task):
    """
    List of unique DOI in citation dataset.
    """

    def requires(self):
        return {
            "s": OpenCitationsSourceDOI(),
            "t": OpenCitationsTargetDOI(),
        }

    def run(self):
        output = shellout("""
                          LC_ALL=C sort -u -S 50%
                            <(zstdcat -T0 {s} | LC_ALL=C uniq)
                            <(zstdcat -T0 {t} | LC_ALL=C uniq)
                          | zstd -c -T0 > {output}
                          """,
                          s=self.input().get("s").path,
                          t=self.input().get("t").path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        fingerprint = self.open_citations_url_hash()
        filename = "{}.json.zst".format(fingerprint)
        return luigi.LocalTarget(path=self.path(filename=filename))

    def on_success(self):
        self.create_symlink(name="current")


class IndexMappedDOI(Task):
    """
    A list of unique DOI which have a mapping to catalog identifier; sorted.
    """
    date = luigi.DateParameter(default=datetime.date.today())

    def requires(self):
        return IdMappingTable(date=self.date)

    def run(self):
        output = shellout("""
                          zstdcat -T0 {input} |
                          LC_ALL=C cut -f 2 |
                          LC_ALL=C tr [:upper:] [:lower:] |
                          LC_ALL=C sort -u -S 40% |
                          zstd -c -T0 > {output}
                          """,
                          input=self.input().path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"))

    def on_success(self):
        self.create_symlink(name="current")


class StatsCommonDOI(Task):
    """
    Run `comm` against open citations and index doi list.

    Example data point: 2022-02-02, 44311206 / 72736981; ratio: .6092; about
    60% of the documents in the index that have a DOI also have an entry in the
    citation graph.
    """
    date = luigi.DateParameter(default=datetime.date.today())

    def requires(self):
        return {
            "index": IndexMappedDOI(date=self.date),
            "oci": OpenCitationsUniqueDOI(),
        }

    def run(self):
        output = shellout("""
                          LC_ALL=C comm -12
                            <(zstdcat -T0 {index})
                            <(zstdcat -T0 {oci}) |
                          zstd -c -T0 > {output}
                          """,
                          index=self.input().get("index").path,
                          oci=self.input().get("oci").path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"))

    def on_success(self):
        self.create_symlink(name="current")


class StatsReportData(Task):
    """
    A daily overview (data).

    We want some json overview.
    """
    date = luigi.DateParameter(default=datetime.date.today())

    def requires(self):
        return {
            "common": StatsCommonDOI(date=self.date),
            "index_unique": IndexMappedDOI(date=self.date),
            "oci_inbound": OpenCitationsInboundStats(),
            "oci_outbound": OpenCitationsOutboundStats(),
            "oci_unique": OpenCitationsUniqueDOI(),
        }

    def run(self):
        data = {
            "date": str(datetime.date.today()),
            "index": {
                "num_mapped_doi": T(self.input().get("index_unique")).linecount(),
            },
            "oci": {
                "num_doi": T(self.input().get("oci_unique")).linecount(),
                "stats_inbound": T(self.input().get("oci_inbound")).json(),
                "stats_outbound": T(self.input().get("oci_outbound")).json(),
            },
            "num_common_doi": T(self.input().get("common").linecount()),
            "ratio_corpus": (T(self.input().get("common")).linecount() / T(self.input().get("oci_unique")).linecount()),
        }
        with self.output().open("w") as output:
            json.dump(data, output)

    def output(self):
        # TODO: exclude outputs from this task from cleanup.
        return luigi.LocalTarget(path=self.path(ext="json"))
