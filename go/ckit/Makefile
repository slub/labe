# Makefile for building executables for labe.
#
# Release build:
#
#   $ make
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
VERSION := 0.1.46

GOLDFLAGS += -X main.Version=$(VERSION)
GOLDFLAGS += -w -s
GOFILES = $(shell find . -name \*.go -print)

.PHONY: help
help: ## print info about all commands
	@echo "Commands:"
	@grep -E '^[/.a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[01;32m%-40s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: $(TARGETS) ## build all targets

.PHONY: build
%: cmd/%/main.go $(GOFILES)
	go build -o $@ -ldflags "$(GOLDFLAGS)" $<

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

