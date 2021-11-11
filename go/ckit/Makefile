SHELL := /bin/bash
TARGETS := labed tabbedjson makta
VERSION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GOLDFLAGS += -X main.Version=$(VERSION)
GOLDFLAGS += -X main.Buildtime=$(BUILDTIME)
GOLDFLAGS += -w -s
GOFLAGS = -ldflags "$(GOLDFLAGS)"
GOFILES = $(shell find . -name *.go)

.PHONY: all
all: $(TARGETS)

.PHONY: build
%: cmd/%/main.go server.go set/set.go
	go build -o $@ -ldflags "$(GOLDFLAGS)" $<

.PHONY: clean
clean:
	rm -f $(TARGETS)

