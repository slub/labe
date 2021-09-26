SHELL := /bin/bash
TARGETS := tabbedjson

GOLDFLAGS += -w -s
GOFLAGS = -ldflags "$(GOLDFLAGS)"

.PHONY: build
build:
	go build -ldflags "$(GOLDFLAGS)" .

.PHONY: clean
clean:
	rm -f $(TARGETS)

