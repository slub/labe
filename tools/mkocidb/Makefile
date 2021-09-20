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
	rm -f data.db
