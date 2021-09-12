"""
Command line entry points.
"""

import argparse
import logging
import os
import tempfile

from labe.settings import LOGGING_CONF_FILE, settings


def main():
    parser = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter
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

    print("labectl")
