"""
Command line entry points for labe commands.


    usage: labe.pyz [-h] [-L] [-l] [-O TASK] [-r TASK] [-c CONFIG_FILE]
                    [--logging-conf-file LOGGING_CONF_FILE] [--data-dir DATA_DIR]
                    [--tmp-dir TMP_DIR]

    optional arguments:
      -h, --help            show this help message and exit
      -L, --print-most-recent-download-url
                            show most recent OCI download URL (default: False)
      -l, --list            list task nam namess (default: False)
      -O TASK, --show-output-path TASK
                            show output path of task (default: None)
      -r TASK, --run TASK   task to run (default: None)
      -c CONFIG_FILE, --config-file CONFIG_FILE
                            path to configuration file (default:
                            /home/tir/.config/labe/labe.cfg)
      --logging-conf-file LOGGING_CONF_FILE
                            path to logging configuration file (default:
                            /home/tir/.config/labe/logging.ini)
      --data-dir DATA_DIR, -D DATA_DIR
                            root directory for all tasks, we follow XDG (override
                            in settings.ini) (default: /home/tir/.local/share)
      --tmp-dir TMP_DIR, -T TMP_DIR
                            temporary directory to use (default: /tmp)

Example:

    $ labe.pyz -r SolrDatabase --name main

Put this into cron, to automate:

    0 8 * * * labe.pyz -r SolrDatabase --name main

"""

import argparse
import logging
import configparser
import os
import sys
import tempfile

import luigi
from luigi.cmdline_parser import CmdlineParser
from luigi.parameter import MissingParameterException
from luigi.task_register import Register, TaskClassNotFoundException
from xdg import xdg_config_home, xdg_data_home

from labe.oci import OpenCitationsDataset
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
    names = (name for name in names
             if name not in suppress and not name.islower())
    return names


def main():
    parser = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    parser.add_argument(
        "-L",
        "--print-most-recent-download-url",
        action="store_true",
        help="show most recent OCI download URL",
    )
    parser.add_argument("-l",
                        "--list",
                        action="store_true",
                        help="list task nam namess")
    parser.add_argument("-O",
                        "--show-output-path",
                        metavar="TASK",
                        type=str,
                        help="show output path of task")
    parser.add_argument("-r",
                        "--run",
                        metavar="TASK",
                        type=str,
                        help="task to run")
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
    parser.add_argument(
        "--data-dir",
        "-D",
        default=xdg_data_home(),
        help=
        "root directory for all tasks, we follow XDG (override in settings.ini)"
    )
    parser.add_argument("--tmp-dir",
                        "-T",
                        default=tempfile.gettempdir(),
                        help="temporary directory to use")

    # Task may have their own arguments, which we ignore.
    args, _ = parser.parse_known_args()

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

    # Show output filename for task.
    elif args.show_output_path:
        try:
            parser = CmdlineParser(sys.argv[2:])
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

    # Run luigi task, tweak args so we can use luigi.run, again.
    if args.run:
        try:
            sys.argv = [sys.argv[0], *sys.argv[2:]]
            luigi.run(local_scheduler=True)
        except MissingParameterException as err:
            print('missing parameter: %s' % err, file=sys.stderr)
            sys.exit(1)
        except TaskClassNotFoundException as err:
            print(err, file=sys.stderr)
            sys.exit(1)
