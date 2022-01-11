# Makefile for labe python package.

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
	rm -rf $(PKGNAME)_*.deb

.PHONY: deploy
deploy: $(ZIPAPP)
	# temporary deploy option for development, you can put:
	# DEPLOY_DST=user@srv.slub-dresden.de:/home/user/bin
	# into ".env" file (read by makefile)
	rsync -avzP -e 'ssh -p 21' labe.pyz $(DEPLOY_DST)

.PHONY: deb
deb: $(ZIPAPP)
	# executables
	mkdir -p packaging/deb/$(PKGNAME)/usr/local/bin
	cp $(ZIPAPP) packaging/deb/$(PKGNAME)/usr/local/bin
	# config files
	mkdir -p packaging/deb/$(PKGNAME)/etc/luigi
	mkdir -p packaging/deb/$(PKGNAME)/etc/labe
	cp luigi.cfg logging.ini packaging/deb/$(PKGNAME)/etc/luigi
	cp labe.cfg packaging/deb/$(PKGNAME)/etc/labe
	# build package
	cd packaging/deb && fakeroot dpkg-deb --build $(PKGNAME) .
	mv packaging/deb/$(PKGNAME)_*.deb .

