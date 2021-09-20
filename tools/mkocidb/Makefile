SHELL := /bin/bash
TARGETS := mkocidb

GOLDFLAGS += -w -s
GOFLAGS = -ldflags "$(GOLDFLAGS)"

.PHONY: build
build:
	go build -ldflags "$(GOLDFLAGS)" .

.PHONY: clean
clean:
	rm -f $(TARGETS)

.PHONY: purge
purge:
	rm -if data.db
