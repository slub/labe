# Makefile for labe project
# Last-Modified: 2021-11-18
SHELL := /bin/bash
VERSION := 1.0.0
INSTALLER_FILENAME := labe-$(VERSION).run
TARGETS := $(INSTALLER_FILENAME)

# Other static binaries we may need on the target machine (due to lack of root
# access in 11/2021).
#
# https://github.com/ubleipzig/solrdump
EXTRA_SOLRDUMP := $(HOME)/code/miku/solrdump/solrdump
# https://www.sqlite.org/download.html
# E.g. in 11/2021 on Linux:
# https://www.sqlite.org/2021/sqlite-tools-linux-x86-3360000.zip, we only want
# the "sqlite3" command.
EXTRA_SQLITE3 := $(HOME)/opt/sqlite-tools-linux-x86-3360000/sqlite3

.PHONY: all
all: $(TARGETS)

$(TARGETS):
	mkdir -p .build

	# extra tools
	cp $(EXTRA_SQLITE3) .build
	cp $(EXTRA_SOLRDUMP) .build

	# build and copy our tools
	cd go/ckit && make && cd -
	cp go/ckit/makta .build
	cp go/ckit/tabbedjson .build
	cp go/ckit/labed .build

	# copy installer script
	cp extra/install.sh .build

	makeself .build $(INSTALLER_FILENAME) "installing labe $(VERSION)..." ./install.sh

.PHONY: clean
clean:
	rm -rf .build
	rm -f $(TARGETS)
