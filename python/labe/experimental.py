"""
Experimental.
"""

import luigi

from labe.base import Zstd, shellout
from labe.tasks import Task


class ExpRefcatDownload(Task):
    """
    Download refcat v2.
    """
    url = luigi.Parameter(default="https://archive.org/download/refcat_2022-01-03/refcat-doi-table-2022-01-03.json.zst")

    def run(self):
        output = shellout("""
                          curl -sL --fail {url} | zstd -c -d -T0 -q | LC_ALL=C sort -S50% | zstd -c -T0 > {output}
                          """, url=self.url)
        luigi.LocalTarget(output).move(self.output().path)

    def output(self):
        return luigi.LocalTarget(path=self.path(ext="json.zst", digest=True), format=Zstd)
