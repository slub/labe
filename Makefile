# Makefile for labe project
# Last-Modified: 2021-11-19
#
# We want a single command build process that combines go tools, python
# packaging, systemd units and other configuration files.
SHELL := /bin/bash

VERSION := 1.0.0
PKGNAME := labe

GITCOMMIT := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

ANSIBLE_OPTS := ANSIBLE_RETRY_FILES_ENABLED=false ANSIBLE_NOCOWS=true ANSIBLE_HOST_KEY_CHECKING=false

.PHONY: all
all:
	cd go/ckit && make all && cd -

.PHONY: clean
clean:
	rm -f $(PKGNAME)_*.deb

.PHONY: deb
deb: all
	mkdir -p packaging/deb/$(PKGNAME)/usr/local/bin
	# copy files one by one
	cp go/ckit/makta packaging/deb/$(PKGNAME)/usr/local/bin
	cp go/ckit/tabbedjson packaging/deb/$(PKGNAME)/usr/local/bin
	cp go/ckit/labed packaging/deb/$(PKGNAME)/usr/local/bin
	# build package
	cd packaging/deb && fakeroot dpkg-deb --build $(PKGNAME) .
	mv packaging/deb/$(PKGNAME)_*.deb .

.PHONY: deploy
deploy: deb
	mkdir -p ansible/roles/app/files
	cp -v $(PKGNAME)_*deb ansible/roles/app/files/
	$(ANSIBLE_OPTS) ansible-playbook -b -i ansible/hosts ansible/site.yml

