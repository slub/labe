"""
Experimental.
"""

import luigi

from labe.base import Zstd, shellout
from labe.tasks import OpenCitationsSingleFile, Task

__all__ = [
    'ExpRefcatDownload',
    'OpenCitationsDOITable',
    'ExpRefcatDownload',
    'ExpOpenCitationsOnly',
    'ExpCombinedCitationsTable',
]


class ExpRefcatDownload(Task):
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


class OpenCitationsDOITable(Task):
    """
    DOI to DOI table, sorted.
    """

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
        fingerprint = self.open_citations_url_hash()
        filename = "{}.tsv.zst".format(fingerprint)
        return luigi.LocalTarget(path=self.path(filename=filename))

    def on_success(self):
        self.create_symlink(name="current")


class ExpRefcatOnly(Task):
    """
    Refcat only DOI-DOI pairs.
    """

    def requires(self):
        return {
            "refcat": ExpRefcatDownload(),
            "oci": OpenCitationsDOITable(),
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
        return luigi.LocalTarget(path=self.path(ext="tsv.zst", digest=True), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class ExpOpenCitationsOnly(Task):
    """
    OCI only DOI-DOI pairs.
    """

    def requires(self):
        return {
            "oci": OpenCitationsDOITable(),
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
        return luigi.LocalTarget(path=self.path(ext="tsv.zst", digest=True), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")


class ExpCombinedCitationsTable(Task):
    """
    OCI and refcat as a single file with duplicates removed.
    """

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
        return luigi.LocalTarget(path=self.path(ext="tsv.zst", digest=True), format=Zstd)

    def on_success(self):
        self.create_symlink(name="current")
