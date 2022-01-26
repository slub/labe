# Makefile for building executables for labe.
#
# Release build:
#
#   $ make
#
# Debug build:
#
#   $ make all EXTRA_BUILD_FLAGS="-tags debug"
#
# Create debian package:
#
#   $ make deb

SHELL := /bin/bash
PKGNAME := ckit
TARGETS := \
	doisniffer \
	labed \
	makta \
    tabjson

# This is an automatic version string using the git commit id. The debian
# package version is currently separately defined in (and only in):
# packaging/deb/ckit/DEBIAN/control
VERSION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GOLDFLAGS += -X main.Version=$(VERSION)
GOLDFLAGS += -X main.Buildtime=$(BUILDTIME)
GOLDFLAGS += -w -s
EXTRA_BUILD_FLAGS = # e.g. add "-tags debug" here for debug builds
GOFILES = $(shell find . -name \*.go -print)

.PHONY: help
help: ## print info about all commands
	@echo "Commands:"
	@grep -E '^[/.a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[01;32m%-40s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: $(TARGETS) ## build all targets

.PHONY: build
%: cmd/%/main.go $(GOFILES)
	go build -o $@ -ldflags "$(GOLDFLAGS)" $(EXTRA_BUILD_FLAGS) $<

.PHONY: clean
clean: ## clean artifacts
	rm -f $(TARGETS)
	rm -f *.deb
	rm -rf packaging/deb/$(PKGNAME)/usr

.PHONY: test
test: ## run tests
	go test -v -cover ./...

.PHONY: deb
deb: all ## build debian package
	# executables
	mkdir -p packaging/deb/$(PKGNAME)/usr/local/bin
	cp $(TARGETS) packaging/deb/$(PKGNAME)/usr/local/bin
	# systemd unit file
	mkdir -p packaging/deb/$(PKGNAME)/usr/lib/systemd/system
	cp packaging/labed.service packaging/deb/$(PKGNAME)/usr/lib/systemd/system/
	# build package
	cd packaging/deb && fakeroot dpkg-deb --build $(PKGNAME) .
	mv packaging/deb/$(PKGNAME)_*.deb .

