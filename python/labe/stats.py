"""
Stats related tasks.
"""

import luigi

from labe.base import shellout
from labe.tasks import IdMappingTable, OpenCitationsSingleFile, Task

__all__ = [
    'OpenCitationsCitedCount',
    'OpenCitationsSourceDOI',
    'OpenCitationsTargetDOI',
    'OpenCitationsUniqueDOI',
    'IndexMappedDOI',
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
        shellout(r"""
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
                          > | zstd -c -T0 > {output}
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
