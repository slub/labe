# Makefile for building executables for labe.

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
GOFLAGS = -ldflags "$(GOLDFLAGS)"
GOFILES = $(shell find . -name \*.go -print)

.PHONY: all
all: $(TARGETS)

.PHONY: build
%: cmd/%/main.go server.go set/set.go
	go build -o $@ -ldflags "$(GOLDFLAGS)" $<

.PHONY: clean
clean:
	rm -f $(TARGETS)
	rm -f *.deb
	rm -rf packaging/deb/$(PKGNAME)/usr

.PHONY: deb
deb: all
	# executables
	mkdir -p packaging/deb/$(PKGNAME)/usr/local/bin
	cp $(TARGETS) packaging/deb/$(PKGNAME)/usr/local/bin
	# systemd unit file
	mkdir -p packaging/deb/$(PKGNAME)/usr/lib/systemd/system
	cp packaging/labed.service packaging/deb/$(PKGNAME)/usr/lib/systemd/system/
	# build package
	cd packaging/deb && fakeroot dpkg-deb --build $(PKGNAME) .
	mv packaging/deb/$(PKGNAME)_*.deb .

