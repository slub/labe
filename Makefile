SHELL := /bin/bash

.PHONY: all
all:
	python setup develop

.PHONY: clean
clean:
	rm -rf labe.egg-info
