SHELL := /bin/bash
TARGETS := spindel

GOLDFLAGS += -w -s
GOFLAGS = -ldflags "$(GOLDFLAGS)"

.PHONY: build
build:
	go build -ldflags "$(GOLDFLAGS)" .

.PHONY: clean
clean:
	rm -f $(TARGETS)

