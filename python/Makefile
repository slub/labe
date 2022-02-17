# Makefile for labe python package.

SHELL := /bin/bash
PY_FILES := $(shell find labe -name \*.py -print)
PKGNAME := labe

# The "zipapp" we build, cf. PEP441, https://www.python.org/dev/peps/pep-0441/,
# https://shiv.readthedocs.io/
ZIPAPP := $(PKGNAME).pyz

# IMPORTANT: Python version on dev (e.g. use https://github.com/pyenv/pyenv)
# and target *must match* (up to minor version) e.g. for debian 10 (2021), you
# might want to use:
# make labe.pyz PYTHON_INTERPRETER='"/usr/bin/env python3.7"'
PYTHON_INTERPRETER := "/usr/bin/env python3.7"

.PHONY: help
help: ## print info about all commands
	@echo "Commands:"
	@grep -E '^[/.a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sed -e 's@Makefile:@@' | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[01;32m%-40s\033[0m %s\n", $$1, $$2}'

$(ZIPAPP): $(PY_FILES) ## build single python executable with shiv
	# https://shiv.readthedocs.io/en/latest/cli-reference.html note: use
	# SHIV_ROOT envvar to override expansion dir (e.g. if home is networked)
	# You can use `shiv-info -j labe.pyz` to show info about the built file.
	shiv --no-modify --reproducible --compressed --entry-point labe.cli:main --python $(PYTHON_INTERPRETER) --output-file $(ZIPAPP) .

.PHONY: fmt
fmt: ## apply code formatting
	yapf -p -i -r $(PKGNAME) tests
	isort $(PKGNAME)

.PHONY: clean
clean: ## clean artifacts
	find . -name "__pycache__" -exec rm -rf "{}" +
	rm -rf $(PKGNAME).egg-info/
	rm -rf $(PKGNAME)_*.deb
	rm -rf $(ZIPAPP)
	rm -rf .pytest_cache/ build/ dist/
	rm -rf packaging/deb/labe/etc/ packaging/deb/labe/usr/

.PHONY: deb
deb: $(ZIPAPP) ## build a debian package
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
