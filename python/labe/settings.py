"""
Setting up settings, living under XDG paths.

More information: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html

To access, import this module:

    from labe.settings import settings

"""

from pathlib import Path

from dynaconf import Dynaconf
from xdg import xdg_config_home

# logging and application settings files
LOGGING_CONF_FILE = str(xdg_config_home().joinpath("labe/logging.ini"))
CONFIG_FILE = str(xdg_config_home().joinpath("labe/settings.ini"))
ENVVAR_PREFIX_FOR_DYNACONF = "LABE"
ENV_FOR_DYNACONF = "default"

settings = Dynaconf(
    ENVIRONMENTS=True,
    ENV_FOR_DYNACONF=ENV_FOR_DYNACONF,
    ENVVAR_PREFIX_FOR_DYNACONF=ENVVAR_PREFIX_FOR_DYNACONF,
    SETTINGS_FILE_FOR_DYNACONF=CONFIG_FILE,
)
