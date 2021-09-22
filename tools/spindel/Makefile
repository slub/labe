SHELL := /bin/bash
TARGETS := spindel
VERSION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GOLDFLAGS += -X main.Version=$(VERSION)
GOLDFLAGS += -X main.Buildtime=$(BUILDTIME)
GOLDFLAGS += -w -s
GOFLAGS = -ldflags "$(GOLDFLAGS)"
GOFILES = $(shell find . -name *.go)

.PHONY: build
spindel: $(GOFILES)
	go build -ldflags "$(GOLDFLAGS)" .

.PHONY: clean
clean:
	rm -f $(TARGETS)

