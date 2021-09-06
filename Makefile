SHELL := /bin/bash
PY_FILES := $(shell find . -name \*.py -print)
PKGNAME := labe

# The "zipapp" we build, cf. PEP441, https://www.python.org/dev/peps/pep-0441/,
# https://shiv.readthedocs.io/
ZIPAPP := $(PKGNAME).pyz

# IMPORTANT: Python version on dev (e.g. use https://github.com/pyenv/pyenv)
# and target *must match* (up to minor version) e.g. for aitio (2021), you
# might want to use:
# make refcat.pyz PYTHON_INTERPRETER='"/usr/bin/env python3.8"'
PYTHON_INTERPRETER := "/usr/bin/env python"

$(ZIPAPP): $(PY_FILES)
	# https://shiv.readthedocs.io/en/latest/cli-reference.html
	# note: use SHIV_ROOT envvar to override expansion dir (e.g. if home is networked)
	shiv --reproducible --compressed --entry-point labe.cli:main --python $(PYTHON_INTERPRETER) --output-file $(ZIPAPP) .

.PHONY: all
all:
	python setup.py develop

.PHONY: clean
clean:
	rm -rf labe.egg-info
	rm -rf labe.pyz

