SHELL := /bin/bash
VERSION := 1.0.0

SOLRDUMP := $(HOME)/code/miku/solrdump/solrdump

labe-$(VERSION).zip:
	mkdir -p .build
	cp $(SOLRDUMP) .build
	cd go/ckit && make && cd -
	cp go/ckit/makta .build
	cp go/ckit/tabbedjson .build
	cp go/ckit/labed .build
	zip -j -r $@ .build/*

.PHONY: clean
clean:
	rm -rf .build
	rm -f labe-$(VERSION).zip
