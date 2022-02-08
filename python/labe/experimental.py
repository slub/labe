"""
Experimental.
"""

import datetime
import json
import os
import tempfile

import luigi
import pandas as pd

from labe.base import Zstd, shellout
from labe.stats import IndexMappedDOI, IndexMappedDOIForInstitution
from labe.tasks import OpenCitationsSingleFile, Task

__all__ = [
    'ExpRefcatDownload',
    'OpenCitationsDOITable',
    'ExpRefcatDownload',
    'ExpOpenCitationsOnly',
    'ExpCombinedCitationsTable',
]


class ExpTask(Task):
    TAG = 'exp'


class ExpRefcatDownload(ExpTask):
    """
    Download refcat v2; over 21h.
    """
    url = luigi.Parameter(default="https://archive.org/download/refcat_2022-01-03/refcat-doi-table-2022-01-03.json.zst")

    def run(self):
        output = shellout("""
                          curl -sL --retry 3 --fail {url} | zstd -c -d -T0 -q | LC_ALL=C sort -S50% | zstd -c -T0 > {output}
                          """,
                          url=self.url)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="tsv.zst", digest=True), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class OpenCitationsDOITable(ExpTask):
    """
    DOI to DOI table, sorted.
    """
    exp = luigi.Parameter(default="1", description="experiment id")

    def requires(self):
        return OpenCitationsSingleFile()

    def run(self):
        output = shellout(r"""
                          zstdcat -T0 {input} |
                          cut -d, -f2,3 |
                          sed -e 's@,@\t@' |
                          LC_ALL=C sort -S 50% |
                          zstd -c -T0 > {output}
                          """,
                          input=self.input().path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path())

    def on_success(self):
        self.create_symlink(name="current")


class ExpRefcatOnly(ExpTask):
    """
    Refcat only DOI-DOI pairs.
    """
    exp = luigi.Parameter(default="1", description="experiment id")

    def requires(self):
        return {
            "refcat": ExpRefcatDownload(),
            "oci": OpenCitationsDOITable(exp=self.exp),
        }

    def run(self):
        output = shellout("""
                          LC_ALL=C comm -23 <(zstdcat -T0 {refcat}) <(zstdcat -T0 {oci}) |
                          zstd -c -T0 > {output}
                          """,
                          refcat=self.input().get("refcat").path,
                          oci=self.input().get("oci").path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class ExpOpenCitationsOnly(ExpTask):
    """
    OCI only DOI-DOI pairs.
    """
    exp = luigi.Parameter(default="1", description="experiment id")

    def requires(self):
        return {
            "oci": OpenCitationsDOITable(exp=self.exp),
            "refcat": ExpRefcatDownload(),
        }

    def run(self):
        output = shellout("""
                          LC_ALL=C comm -13 <(zstdcat -T0 {refcat}) <(zstdcat -T0 {oci}) |
                          zstd -c -T0 > {output}
                          """,
                          refcat=self.input().get("refcat").path,
                          oci=self.input().get("oci").path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class ExpCombinedCitationsTable(ExpTask):
    """
    OCI and refcat as a single file with duplicates removed; in 02/2022 we
    gather 1,565,087,463 edges (but many from refcat are dataset related;
    https://arxiv.org/pdf/2110.06595v2.pdf#page=2); 13GB compressed, 73GB
    uncompressed; deployment machine i/o throughput at 33MB/s for db generation
    (underlying i/o can go up to 400MB/s, so bottleneck seems to be the code);
    about 700k edges/s added; sqlite db was about 185GB (vs 155GB for OCI only).

    As is, stats generation keys on the URL of the oci dump, so it's a bit
    unwieldy to evaluate this dataset directly.
    """
    exp = luigi.Parameter(default="1", description="experiment id")

    def requires(self):
        return {
            "oci": OpenCitationsSingleFile(),
            "refcat": ExpRefcatDownload(),
        }

    def run(self):
        output = shellout(r"""
                          LC_ALL=C sort -u -S50%
                            <(zstdcat -T0 {oci} | cut -d, -f2,3 | sed -e 's@,@\t@' | tr [:upper:] [:lower:])
                            <(zstdcat -T0 {refcat} | tr [:upper:] [:lower:]) |
                          zstd -c -T0 > {output}
                          """,
                          oci=self.input().get("oci").path,
                          refcat=self.input().get("refcat").path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class ExpCitationsSourceDOI(ExpTask):
    """
    List of DOI that are source of a citation edge. Normalized and sorted.
    18m54.167s.

    Issues: We still have '"' in DOI.
    """
    exp = luigi.Parameter(default="1", description="experiment id")

    def requires(self):
        return ExpCombinedCitationsTable(exp=self.exp)

    def run(self):
        output = shellout("""
                          zstdcat -T0 {input} |
                          LC_ALL=C cut -f 1 |
                          LC_ALL=C tr [:upper:] [:lower:] |
                          LC_ALL=C sort -S50% |
                          zstd -c -T0 > {output}
                          """,
                          input=self.input().path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class ExpCitationsTargetDOI(ExpTask):
    """
    List of DOI that are target of a citation edge. Normalized and sorted.
    """
    exp = luigi.Parameter(default="1", description="experiment id")

    def requires(self):
        return ExpCombinedCitationsTable(exp=self.exp)

    def run(self):
        output = shellout("""
                          zstdcat -T0 {input} |
                          LC_ALL=C cut -f 2 |
                          LC_ALL=C tr [:upper:] [:lower:] |
                          LC_ALL=C sort -S50% |
                          zstd -c -T0 > {output}
                          """,
                          input=self.input().path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class ExpCitationsCitedCount(ExpTask):
    """
    Generate a table with two columns: inbound link count and DOI.
    """
    exp = luigi.Parameter(default="1", description="experiment id")

    def requires(self):
        return ExpCitationsTargetDOI(exp=self.exp)

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
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class ExpCitationsCitingCount(ExpTask):
    """
    Generate a table with two columns: outbound link count and DOI.
    """
    exp = luigi.Parameter(default="1", description="experiment id")

    def requires(self):
        return ExpCitationsSourceDOI(exp=self.exp)

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
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class ExpCitationsInboundStats(ExpTask):
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
    exp = luigi.Parameter(default="1", description="experiment id")

    def requires(self):
        return ExpCitationsCitedCount(exp=self.exp)

    def run(self):
        output = shellout("zstdcat -T0 {input} | cut -f1 > {output}", input=self.input().path)
        df = pd.read_csv(output, header=None, names=["inbound_edges"], skip_blank_lines=True)
        percentiles = [0, 0.1, 0.25, 0.5, 0.75, 0.95, 0.99, 0.999, 1]
        with tempfile.NamedTemporaryFile(mode="wb", delete=False) as f:
            df.describe(percentiles=percentiles).to_json(f)
        luigi.LocalTarget(f.name).move(self.output().path)
        os.remove(output)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="json"))

    def on_success(self):
        self.create_symlink(name="current")


class ExpCitationsOutboundStats(ExpTask):
    """
    Outbound edge count distribution.
    """
    exp = luigi.Parameter(default="1", description="experiment id")

    def requires(self):
        return ExpCitationsCitingCount(exp=self.exp)

    def run(self):
        output = shellout("zstdcat -T0 {input} | cut -f1 > {output}", input=self.input().path)
        df = pd.read_csv(output, header=None, names=["outbound_edges"], skip_blank_lines=True)
        percentiles = [0, 0.1, 0.25, 0.5, 0.75, 0.95, 0.99, 0.999, 1]
        with tempfile.NamedTemporaryFile(mode="wb", delete=False) as f:
            df.describe(percentiles=percentiles).to_json(f)
        luigi.LocalTarget(f.name).move(self.output().path)
        os.remove(output)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="json"))

    def on_success(self):
        self.create_symlink(name="current")


class ExpCitationsUniqueDOI(ExpTask):
    """
    List of unique DOI in citation dataset.
    """
    exp = luigi.Parameter(default="1", description="experiment id")

    def requires(self):
        return {
            "s": ExpCitationsSourceDOI(exp=self.exp),
            "t": ExpCitationsTargetDOI(exp=self.exp),
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
        return luigi.LocalTarget(path=self.path("tsv.zst"), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class ExpStatsCommonDOIForInstitution(ExpTask):
    """
    Run `comm` against open citations and index doi list for institution.
    """
    date = luigi.DateParameter(default=datetime.date.today())
    institution = luigi.Parameter(default="DE-14")
    exp = luigi.Parameter(default="1", description="experiment id")

    def requires(self):
        return {
            "index": IndexMappedDOIForInstitution(date=self.date, institution=self.institution),
            "oci": ExpCitationsUniqueDOI(exp=self.exp),
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
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class ExpStatsCommonDOI(ExpTask):
    """
    Run `comm` against open citations and index doi list.

    Example data point: 2022-02-02, 44311206 / 72736981; ratio: .6092; about
    60% of the documents in the index that have a DOI also have an entry in the
    citation graph.
    """
    date = luigi.DateParameter(default=datetime.date.today())
    exp = luigi.Parameter(default="1", description="experiment id")

    def requires(self):
        return {
            "index": IndexMappedDOI(date=self.date),
            "exp": ExpCitationsUniqueDOI(exp=self.exp),
        }

    def run(self):
        output = shellout("""
                          LC_ALL=C comm -12
                            <(zstdcat -T0 {index})
                            <(zstdcat -T0 {oci}) |
                          zstd -c -T0 > {output}
                          """,
                          index=self.input().get("index").path,
                          oci=self.input().get("exp").path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class ExpStatsReportData(ExpTask):
    """
    A daily overview (data).
    """
    date = luigi.DateParameter(default=datetime.date.today())
    institution = luigi.Parameter(default="DE-14")
    exp = luigi.Parameter(default="1", description="experiment id")

    def requires(self):
        return {
            "common": ExpStatsCommonDOI(date=self.date, exp=self.exp),
            "common_institution": ExpStatsCommonDOIForInstitution(date=self.date, institution=self.institution, exp=self.exp),
            "index_unique": IndexMappedDOI(date=self.date),
            "index_unique_institution": IndexMappedDOIForInstitution(date=self.date, institution=self.institution),
            "exp_inbound": ExpCitationsInboundStats(exp=self.exp),
            "exp_outbound": ExpCitationsOutboundStats(exp=self.exp),
            "exp_unique": ExpCitationsUniqueDOI(exp=self.exp),
            "exp": ExpCombinedCitationsTable(),
        }

    def run(self):
        """
        Ok to open files and not close them, as this is a sink task.
        """
        si = self.input()
        data = {
            "version": "1",
            "date": str(self.date),
            "institution": {
                self.institution: {
                    "num_mapped_doi": sum(1 for _ in si.get("index_unique_institution").open()),
                    "num_common_doi": sum(1 for _ in si.get("common_institution").open()),
                    "ratio": (sum(1 for _ in si.get("common_institution").open()) / sum(1 for _ in si.get("exp_unique").open())),
                },
            },
            "index": {
                "num_mapped_doi": sum(1 for _ in si.get("index_unique").open()),
                "num_common_doi": sum(1 for _ in si.get("common").open()),
                "ratio": (sum(1 for _ in si.get("common").open()) / sum(1 for _ in si.get("exp_unique").open())),
            },
            "exp": {
                "num_edges": sum(1 for _ in si.get("exp").open()),
                "num_doi": sum(1 for _ in si.get("exp_unique").open()),
                "stats_inbound": json.load(si.get("exp_inbound").open()).get("inbound_edges"),
                "stats_outbound": json.load(si.get("exp_outbound").open()).get("outbound_edges"),
            },
        }
        with self.output().open("w") as output:
            json.dump(data, output)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="json"))

    def on_success(self):
        self.create_symlink(name="current")
