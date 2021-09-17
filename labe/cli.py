"""
Command line entry points for data processing, from acquisition to a prepared
index.

    $ labectl status
    $ labectl gc
    $ labectl help
    $ labectl run [TARGET]

Example:

    $ labectl run UpdateIndex

Put this into cron, to automate:

    0 8 * * * labectl run UpdateIndex && labectl gc
"""

import argparse
import logging
import os
import tempfile

from labe.coci import OpenCitationsDataset
from labe.settings import LOGGING_CONF_FILE, settings


def main():
    parser = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter
    )
    parser.add_argument(
        "-L",
        "--print-most-recent-download-url",
        action="store_true",
        help="show most recent OCI download URL",
    )
    parser.add_argument(
        "--logging-conf-file",
        default=LOGGING_CONF_FILE,
        help="path to logging configuration file",
    )
    parser.add_argument(
        "--tmp-dir", default=tempfile.gettempdir(), help="temporary directory to use"
    )
    args = parser.parse_args()
    if os.path.exists(LOGGING_CONF_FILE):
        logging.config.fileConfig(LOGGING_CONF_FILE)
    if settings.get("TMPDIR"):
        tempfile.tempdir = settings.TMPDIR

    # TODO: subcommands, e.g. "status", "run", "gc"

    if args.print_most_recent_download_url:
        ds = OpenCitationsDataset()
        print(ds.most_recent_download_url())
