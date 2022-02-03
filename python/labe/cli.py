"""
Command line tool to run luigi tasks for project ≋ labe

  List tasks:

      $ labe.pyz -l
      CombinedUpdate
      IdMappingDatabase
      IdMappingTable
      OpenCitationsDatabase
      OpenCitationsDownload
      OpenCitationsSingleFile
      SolrDatabase
      SolrFetchDocs

  Run task:

      $ labe.pyz -r SolrDatabase --name main

  Show task output location:

      $ labe.pyz -O OpenCitationsDatabase
      /usr/share/labe/data/OpenCitationsDatabase/c90e82e35c9d02c00f81bee6d1f34b132953398c.db

Symlinks point to the current version of a task output. They will only be
updated, if the task ran successfully. This way we can identify outdated files:

  $ labe.pyz --list-deletable

Use cron job to schedule (a daily) task:

  0 5 * * * rm -rf $(labe.pyz --list-deletable) && labe.pyz -r CombinedUpdate

Relevant configuration files:

  /etc/labe/labe.cfg
  /etc/luigi/logging.ini
  /etc/luigi/luigi.cfg

Project homepage: https://github.com/slub/labe
"""

import argparse
import configparser
import datetime
import json
import logging
import os
import sys
import tempfile

import luigi
import pandas as pd
from luigi.cmdline_parser import CmdlineParser
from luigi.parameter import MissingParameterException
from luigi.task_register import Register, TaskClassNotFoundException
from xdg import xdg_config_home, xdg_data_home

from labe.deps import dump_deps
from labe.diff import stats_diff
from labe.oci import OpenCitationsDataset
from labe.stats import *
# We need a star import to import all tasks.
from labe.tasks import *
from labe.tasks import Task


def effective_task_names(suppress=None):
    """
    Runnable, relevant task names. Optionally pass a list of task names to
    suppress.
    """
    if suppress is None:
        # These are internal luigi names. TODO: may filter out by module name.
        suppress = [
            'BaseTask',
            'Config',
            'ExternalTask',
            'RangeBase',
            'RangeByMinutes',
            'RangeByMinutesBase',
            'RangeDaily',
            'RangeDailyBase',
            'RangeHourly',
            'RangeHourlyBase',
            'RangeMonthly',
            'Task',
            'TestNotificationsTask',
            'WrapperTask',
        ]
    names = (name for name in sorted(Register.task_names()))
    names = (name for name in names if name not in suppress and not name.islower())
    return names


def main():
    parser = argparse.ArgumentParser(prog="labe.pyz", formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    parser.add_argument(
        "-L",
        "--print-most-recent-download-url",
        action="store_true",
        help="show most recent OCI download URL",
    )
    parser.add_argument("-l", "--list", action="store_true", help="list task names")
    parser.add_argument("-O", "--show-output-path", metavar="TASK", type=str, help="show output path of task")
    parser.add_argument("-r", "--run", metavar="TASK", type=str, help="task to run")
    parser.add_argument(
        "-c",
        "--config-file",
        default=os.path.join(xdg_config_home(), "labe", "labe.cfg"),
        help="path to configuration file",
    )
    parser.add_argument(
        "--logging-conf-file",
        default=os.path.join(xdg_config_home(), "labe", "logging.ini"),
        help="path to logging configuration file",
    )
    parser.add_argument("--data-dir",
                        "-D",
                        default=os.path.join(xdg_data_home(), "labe"),
                        help="root directory for all tasks, we follow XDG")
    parser.add_argument("--tmp-dir", "-T", default=tempfile.gettempdir(), help="temporary directory to use")
    parser.add_argument("--list-deletable", action="store_true", help="list task outputs, which could be deleted")
    parser.add_argument("--deps", metavar="TASK", type=str, help="show task dependencies")
    parser.add_argument("--deps-dot", metavar="TASK", type=str, help="print task dependencies in dot format")
    parser.add_argument("--diff", action="store_true", help="generate a diff json from two stats json docs")
    parser.add_argument("--diff-institution",
                        metavar="ISIL",
                        default="DE-14",
                        help="which isil to compute the stats for (only relevant with --diff)")

    # Task may have their own arguments, which we ignore.
    args, unparsed = parser.parse_known_args()

    # Hack to set the base directory of all tasks.
    Task.BASE = args.data_dir
    # Hack to override autodetection of config file, if the given file exists.
    if os.path.exists(args.config_file):
        parser = configparser.ConfigParser()
        parser.read(args.config_file)
        Task._config = parser

    # Setup, configure.
    tempfile.tempdir = args.tmp_dir
    if os.path.exists(args.logging_conf_file):
        # TODO: This won't work with [2:] ...
        logging.config.fileConfig(args.logging_conf_file)

    # Wrapper around OCI.
    if args.print_most_recent_download_url:
        ds = OpenCitationsDataset()
        print(ds.most_recent_download_url())
        sys.exit(0)

    # List available tasks.
    if args.list:
        for name in effective_task_names():
            print(name)
        sys.exit(0)

    elif args.diff:
        # TODO: this can live in a separate task in stats as well
        # in pandas 1.4.0 date_range will get an inclusive kwarg:
        # https://pandas.pydata.org/docs/reference/api/pandas.date_range.html#pandas-date-range
        rng = pd.date_range(start="2022-01-01", end=datetime.date.today(), freq="D")
        found = []
        for date in rng.to_pydatetime():
            task = StatsReportData(date=date, institution=args.institution)
            path = task.output().path
            if os.path.exists(path):
                found.append(path)

        # take first and last for now
        if len(found) < 2:
            print("not enough stats (only found: {})".format(found))
            sys.exit(1)

        found = sorted(found)
        # TODO: just to have some output for now
        first, last = found[0], found[-1]
        with open(first) as f:
            with open(last) as g:
                a = json.load(f)
                b = json.load(g)
                doc = {
                    "a": a,
                    "b": b,
                    "diff": stats_diff(a, b),
                }
                print(json.dumps(doc))

        sys.exit(0)

    elif args.list_deletable:
        # TODO: whitelist tasks or directories, e.g. for reports
        # rm -f $(labe.pyz --list-deletable)
        filenames = set()
        symlinked = set()
        for root, dirs, files in os.walk(args.data_dir):
            keepfile = os.path.join(root, ".labekeep")
            if os.path.exists(keepfile):
                continue
            for name in files:
                full = os.path.join(root, name)
                if os.path.isfile(full) and not os.path.islink(full):
                    filenames.add(full)
                if os.path.islink(full):
                    symlinked.add(os.readlink(full))
        for path in sorted(filenames - symlinked):
            print(path)

    # Show output filename for task.
    elif args.show_output_path:
        try:
            parser = CmdlineParser([args.show_output_path] + unparsed)
            output = parser.get_task_obj().output()
            try:
                print(output.path)
            except AttributeError as err:
                print('output of task has no path', file=sys.stderr)
                sys.exit(1)
        except MissingParameterException as err:
            print('missing parameter: %s' % err, file=sys.stderr)
            sys.exit(1)
        except TaskClassNotFoundException as err:
            print(err, file=sys.stderr)
            sys.exit(1)

    # Show dependency graph.
    elif args.deps:
        if len(sys.argv) < 2:
            raise ValueError("task name required")
        try:
            parser = CmdlineParser([args.deps] + unparsed)
            obj = parser.get_task_obj()
            dump_deps(obj)
            sys.exit(0)
        except TaskClassNotFoundException as exc:
            print("no such task")
            sys.exit(1)

    # Render graphviz dot.
    elif args.deps_dot:
        # labe.pyz --deps-dot CombinedUpdate | dot -Tpng > CombinedUpdate.png
        if len(sys.argv) < 2:
            raise ValueError("task name required")
        try:
            parser = CmdlineParser([args.deps_dot] + unparsed)
            obj = parser.get_task_obj()
            dump_deps(obj, dot=True)
            sys.exit(0)
        except TaskClassNotFoundException as exc:
            print("no such task")
            sys.exit(1)

    # Run luigi task, tweak args so we can use luigi.run, again.
    elif args.run:
        try:
            sys.argv = [sys.argv[0], args.run] + unparsed
            luigi.run(local_scheduler=True)
        except MissingParameterException as err:
            print('missing parameter: %s' % err, file=sys.stderr)
            sys.exit(1)
        except TaskClassNotFoundException as err:
            print(err, file=sys.stderr)
            sys.exit(1)

    else:
        print(__doc__)
