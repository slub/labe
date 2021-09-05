SHELL := /bin/bash

.PHONY: all
all:
	python setup.py develop

.PHONY: clean
clean:
	rm -rf labe.egg-info
