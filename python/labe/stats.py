"""
Stats related tasks. Mainly:

* [x] number of DOI in citation corpus
* [x] edge count distibution (across the whole graph)
* [x] size of overlap with the index

Other ideas:

* [x] per institution overlaps

"""

import datetime
import json
import os
import tempfile

import luigi
import pandas as pd

from labe.base import Zstd, shellout
from labe.tasks import (IdMappingTable, OpenCitationsSingleFile, SolrFetchDocs,
                        Task)

__all__ = [
    'IdMappingTableForInstitution',
    'IndexMappedDOI',
    'IndexMappedDOIForInstitution',
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
        return luigi.LocalTarget(path=self.path(filename=filename), format=Zstd)

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
        return luigi.LocalTarget(path=self.path(filename=filename), format=Zstd)

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
        return luigi.LocalTarget(path=self.path(filename=filename), format=Zstd)

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
        return luigi.LocalTarget(path=self.path(filename=filename), format=Zstd)

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
        return luigi.LocalTarget(path=self.path(filename=filename), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class IdMappingTableForInstitution(Task):
    """
    This is like IdMappingTable, but for a single institution.
    """
    date = luigi.DateParameter(default=datetime.date.today())
    institution = luigi.Parameter(default="DE-14")

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
                              jq -rc 'select((.institution | index("{institution}")) != null) | [.id, .doi_str_mv[0]] | @tsv' |
                              zstd -c -T0 >> {output} """,
                          input=self.input().get("main").path,
                          institution=self.institution)

        # In 01/2022, we use "doisniffer" for slub-production as well.
        shellout(""" zstd -q -d -c -T0 {input} |
                     doisniffer |
                     jq -rc 'select((.institution | index("{institution}")) != null) | [.id, .doi_str_mv[0]] | @tsv' |
                     zstd -c -T0 >> {output} """,
                 output=output,
                 input=self.input().get("slub-production").path,
                 institution=self.institution)

        # In 01/2022, the "doi_str_mv" field is included in "ai" - with 73881207 values.
        shellout(r""" zstd -q -d -c -T0 {input} |
                     parallel -j 8 --pipe --block 10M "jq -rc 'select(.doi_str_mv | length > 0) | select((.institution | index(\"{institution}\")) != null) | [.id, .doi_str_mv[0]] | @tsv'" |
                     zstd -T0 -c >> {output} """,
                 output=output,
                 input=self.input().get("ai").path,
                 institution=self.institution)

        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"))


class IndexMappedDOIForInstitution(Task):
    """
    A list of unique DOI which have a mapping to catalog identifier; sorted; for a single institution.
    """
    date = luigi.DateParameter(default=datetime.date.today())
    institution = luigi.Parameter(default="DE-14")

    def requires(self):
        return IdMappingTableForInstitution(date=self.date, institution=self.institution)

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
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"), format=Zstd)


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
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"), format=Zstd)


class StatsCommonDOIForInstitution(Task):
    """
    Run `comm` against open citations and index doi list for institution.
    """
    date = luigi.DateParameter(default=datetime.date.today())
    institution = luigi.Parameter(default="DE-14")

    def requires(self):
        return {
            "index": IndexMappedDOIForInstitution(date=self.date, institution=self.institution),
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
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"), format=Zstd)


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
        return luigi.LocalTarget(path=self.path(ext="tsv.zst"), format=Zstd)


class StatsReportData(Task):
    """
    A daily overview (data).
    """
    date = luigi.DateParameter(default=datetime.date.today())
    institution = luigi.Parameter(default="DE-14")

    def requires(self):
        return {
            "common": StatsCommonDOI(date=self.date),
            "common_institution": StatsCommonDOIForInstitution(date=self.date, institution=self.institution),
            "index_unique": IndexMappedDOI(date=self.date),
            "index_unique_institution": IndexMappedDOIForInstitution(date=self.date, institution=self.institution),
            "oci_inbound": OpenCitationsInboundStats(),
            "oci_outbound": OpenCitationsOutboundStats(),
            "oci_unique": OpenCitationsUniqueDOI(),
            "oci": OpenCitationsSingleFile(),
        }

    def run(self):
        """
        Ok to open files and not close them, as this is a sink task.
        """
        si = self.input()
        data = {
            "version": "2",
            "date": str(self.date),
            "institution": {
                self.institution: {
                    "num_mapped_doi": sum(1 for _ in si.get("index_unique_institution").open()),
                    "num_common_doi": sum(1 for _ in si.get("common_institution").open()),
                    "ratio_common_mapped": (sum(1 for _ in si.get("common_institution").open()) / sum(1 for _ in si.get("index_unique_institution").open())),
                    "ratio_common_exp": (sum(1 for _ in si.get("common_institution").open()) / sum(1 for _ in si.get("oci_unique").open())),
                },
            },
            "index": {
                "num_mapped_doi": sum(1 for _ in si.get("index_unique").open()),
                "num_common_doi": sum(1 for _ in si.get("common").open()),
                "ratio_common_mapped": (sum(1 for _ in si.get("common").open()) / sum(1 for _ in si.get("index_unique").open())),
                "ratio_common_exp": (sum(1 for _ in si.get("common").open()) / sum(1 for _ in si.get("oci_unique").open())),
            },
            "oci": {
                "num_edges": sum(1 for _ in si.get("oci").open()),
                "num_doi": sum(1 for _ in si.get("oci_unique").open()),
                "stats_inbound": json.load(si.get("oci_inbound").open()).get("inbound_edges"),
                "stats_outbound": json.load(si.get("oci_outbound").open()).get("outbound_edges"),
            },
        }
        with self.output().open("w") as output:
            json.dump(data, output)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="json"))
