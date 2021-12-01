# Makefile for labe.

# use .env if present, https://lithic.tech/blog/2020-05/makefile-dot-env
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

SHELL := /bin/bash
PY_FILES := $(shell find . -name \*.py -print)
PKGNAME := labe

# The "zipapp" we build, cf. PEP441, https://www.python.org/dev/peps/pep-0441/,
# https://shiv.readthedocs.io/
ZIPAPP := $(PKGNAME).pyz

# IMPORTANT: Python version on dev (e.g. use https://github.com/pyenv/pyenv)
# and target *must match* (up to minor version) e.g. for debian 10 (2021), you
# might want to use:
# make labe.pyz PYTHON_INTERPRETER='"/usr/bin/env python3.7"'
PYTHON_INTERPRETER := "/usr/bin/env python3.7"

$(ZIPAPP): $(PY_FILES)
	# https://shiv.readthedocs.io/en/latest/cli-reference.html
	# note: use SHIV_ROOT envvar to override expansion dir (e.g. if home is networked)
	shiv --reproducible --compressed --entry-point labe.cli:main --python $(PYTHON_INTERPRETER) --output-file $(ZIPAPP) .

.PHONY: fmt
fmt:
	yapf -p -i -r $(PKGNAME) tests
	isort $(PKGNAME)

.PHONY: clean
clean:
	rm -rf .pytest_cache/
	rm -rf build/
	rm -rf dist/
	rm -rf $(PKGNAME).egg-info/
	rm -rf $(ZIPAPP)
	find . -name "__pycache__" -exec rm -rf "{}" +

.PHONY: deploy
deploy: $(ZIPAPP)
	# temporary deploy option
	rsync -avzP -e 'ssh -p 21' labe.pyz $(DEPLOY_DST)