"""
Stats related tasks.
"""

import luigi

from labe.tasks import OpenCitationsSingleFile, Task

__all__ = [
    'OpenCitationsSourceDOI',
    'OpenCitationsTargetDOI',
    'OpenCitationsUniqueDOI',
]


class OpenCitationsSourceDOI(Task):
    """
    List of DOI that are source of a citation edge. Normalized and sorted.
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
        filename = "{}.tsv".format(fingerprint)
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
        filename = "{}.tsv".format(fingerprint)
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
                          > {output}
                          """,
                          s=self.input().get("s").path,
                          t=self.input().get("t").path)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        fingerprint = self.open_citations_url_hash()
        filename = "{}.json".format(fingerprint)
        return luigi.LocalTarget(path=self.path(filename=filename))

    def on_success(self):
        self.create_symlink(name="current")
