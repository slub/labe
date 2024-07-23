"""
A default task class for local workflows.
"""

import configparser
import datetime
import hashlib
import logging
import os
import re
import subprocess
import tempfile

import luigi
import xdg

__all__ = [
    "BaseTask",
    "ClosestDateParameter",
    "ensure_minimum_file_size",
    "shellout",
]

logger = logging.getLogger("labe")


class ClosestDateParameter(luigi.DateParameter):
    """
    A marker parameter to replace date parameter value with whatever
    self.closest() returns. Use in conjunction with `gluish.task.BaseTask`.
    """

    use_closest_date = True


def is_closest_date_parameter(task, param_name):
    """Return the parameter class of param_name on task."""
    for name, obj in task.get_params():
        if name == param_name:
            return hasattr(obj, "use_closest_date")
    return False


def delistify(x):
    """A basic slug version of a given parameter list."""
    if isinstance(x, list):
        x = [e.replace("'", "") for e in x]
        return "-".join(sorted(x))
    return x


class BaseTask(luigi.Task):
    """
    A base task with a `path` method. BASE should be set to the root
    directory of all tasks. TAG is a shard for a group of related tasks.
    """

    # TODO: allow BASE to be override in config file
    BASE = os.environ.get("LABE_DATA_DIR", tempfile.gettempdir())
    TAG = "data"

    @property
    def logger(self):
        """
        Return the logger. Module logging uses singleton internally, so no worries.
        """
        return logging.getLogger("labe")

    # TODO: supply example config.ini in repo
    @property
    def config(self):
        """
        Configuration is loaded from labe.cfg, from current directory, XDG
        config home (https://wiki.archlinux.org/title/XDG_Base_Directory) or
        /etc/labe/labe.cfg.
        """
        if not hasattr(self, "_config"):
            _config_paths = [
                "/etc/labe/labe.cfg",
                os.path.join(xdg.xdg_config_home(), "labe", "labe.cfg"),
                "labe.cfg",
            ]
            existing_config_paths = [
                path for path in _config_paths if os.path.exists(path)
            ]
            if len(existing_config_paths) == 0:
                raise RuntimeError(
                    "we need a config file in one of: {}".format(_config_paths)
                )
            parser = configparser.ConfigParser()
            for path in existing_config_paths:
                parser.read(path)
            self._config = parser
        return self._config

    def closest(self):
        """Return the closest date for a given date.
        Defaults to the same date."""
        if not hasattr(self, "date"):
            raise AttributeError("Task has no date attribute.")
        return self.date

    def effective_task_id(self):
        """Replace date in task id with closest date."""
        params = self.param_kwargs
        if "date" in params and is_closest_date_parameter(self, "date"):
            params["date"] = self.closest()
            task_id_parts = sorted(["%s=%s" % (k, str(v)) for k, v in params.items()])
            return "%s(%s)" % (self.task_family, ", ".join(task_id_parts))
        else:
            return self.task_id

    def taskdir(self):
        """Return the directory under which all artefacts are stored."""
        return os.path.join(self.BASE, self.TAG, self.task_family)

    def create_symlink(self, name="current", suffix=""):
        """
        Allows to create a symlink pointing to the task output, optionally
        containing a suffix. Overwrites existing links.
        """
        dirname = self.taskdir()
        name = "{}-{}".format(name, suffix) if suffix else name
        current = os.path.join(dirname, name)
        if os.path.exists(current):
            os.remove(current)
        os.symlink(self.output().path, current)

    def path(
        self, filename=None, ext="tsv", digest=False, shard=False, encoding="utf-8"
    ):
        """
        Return the path for this class with a certain set of parameters.
        `ext` sets the extension of the file.
        If `digest` is true, the filename (w/o extenstion) will be hashed.
        If `shard` is true, the files are placed in shards, based on the first
        two chars of the filename (hashed).
        """
        if self.BASE is NotImplemented:
            raise RuntimeError("BASE directory must be set.")

        params = dict(self.get_params())

        if filename is None:
            parts = []

            for name, param in self.get_params():
                if not param.significant:
                    continue
                if name == "date" and is_closest_date_parameter(self, "date"):
                    parts.append("date-%s" % self.closest())
                    continue
                if hasattr(param, "is_list") and param.is_list:
                    es = "-".join([str(v) for v in getattr(self, name)])
                    parts.append("%s-%s" % (name, es))
                    continue

                val = getattr(self, name)

                if isinstance(val, datetime.datetime):
                    val = val.strftime("%Y-%m-%dT%H%M%S")
                elif isinstance(val, datetime.date):
                    val = val.strftime("%Y-%m-%d")
                elif isinstance(val, bool):
                    if val is True:
                        parts.append(name)
                    continue

                parts.append("%s-%s" % (name, val))

            name = "-".join(sorted(parts))
            if len(name) == 0:
                name = "output"
            if digest:
                name = hashlib.sha1(name.encode(encoding)).hexdigest()
            if not ext:
                filename = "{fn}".format(fn=name)
            else:
                filename = "{fn}.{ext}".format(ext=ext, fn=name)
            if shard:
                prefix = hashlib.sha1(filename.encode(encoding)).hexdigest()[:2]
                return os.path.join(
                    self.BASE, self.TAG, self.task_family, prefix, filename
                )

        return os.path.join(self.BASE, self.TAG, self.task_family, filename)


def ensure_minimum_file_size(filename, size=1024):
    """
    Raises an exception, when condition fails, otherwise returns None.
    """
    sz = os.path.getsize(filename)
    if sz >= size:
        return
    raise RuntimeError(
        "size underflow for {}, got {}, want at least {}".format(filename, sz, size)
    )


def shellout(
    template,
    preserve_whitespace=False,
    executable="/bin/bash",
    ignoremap=None,
    encoding=None,
    pipefail=True,
    temp_prefix="labe-",
    **kwargs,
):
    """
    Takes a shell command template and executes it. The template must use the
    new (2.6+) format mini language. `kwargs` must contain any defined
    placeholder, only `output` is optional and will be autofilled with a
    temporary file if it used, but not specified explicitly.
    If `pipefail` is `False` no subshell environment will be spawned, where a
    failed pipe will cause an error as well. If `preserve_whitespace` is `True`,
    no whitespace normalization is performed. A custom shell executable name can
    be passed in `executable` and defaults to `/bin/bash`.
    Raises RuntimeError on nonzero exit codes. To ignore certain errors, pass a
    dictionary in `ignoremap`, with the error code to ignore as key and a string
    message as value.
    Simple template:
        wc -l < {input} > {output}
    Quoted curly braces:
        ps ax|awk '{{print $1}}' > {output}
    Usage with luigi:
        ...
        tmp = shellout('wc -l < {input} > {output}', input=self.input().path)
        luigi.LocalTarget(tmp).move(self.output().path)
        ....
    """
    if "output" not in kwargs:
        kwargs.update({"output": tempfile.mkstemp(prefix=temp_prefix)[1]})
    if ignoremap is None:
        ignoremap = {}
    if encoding:
        command = template.decode(encoding).format(**kwargs)
    else:
        command = template.format(**kwargs)
    if not preserve_whitespace:
        command = re.sub("[ \t\n]+", " ", command)
    if pipefail:
        command = "(set -o pipefail && %s)" % command
    logger.debug(command)
    code = subprocess.call([command], shell=True, executable=executable)
    if not code == 0:
        if code in ignoremap:
            logger.info("Ignoring error via ignoremap: %s" % ignoremap.get(code))
        else:
            logger.error("%s: %s" % (command, code))
            error = RuntimeError("%s exitcode: %s" % (command, code))
            error.code = code
            raise error
    return kwargs.get("output")


class ZstdFormat(luigi.format.Format):
    """
    The zstandard format.
    """

    input = "bytes"
    output = "bytes"

    def __init__(self, compression_level=None):
        self.compression_level = compression_level
        self.zstd = ["zstd"]
        self.unzstd = ["unzstd"]

    def pipe_reader(self, input_pipe):
        return luigi.format.InputPipeProcessWrapper(self.unzstd, input_pipe)

    def pipe_writer(self, output_pipe):
        args = self.zstd
        if self.compression_level is not None:
            args.append("-" + str(int(self.compression_level)))
        return luigi.format.OutputPipeProcessWrapper(args, output_pipe)


Zstd = ZstdFormat()
