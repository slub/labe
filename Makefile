SHELL := /bin/bash
VERSION := 1.0.0
INSTALLER := labe-$(VERSION)-installer.run

# Other static binaries we may need.
#
# https://github.com/ubleipzig/solrdump
SOLRDUMP := $(HOME)/code/miku/solrdump/solrdump
# https://www.sqlite.org/download.html
SQLITE3 := $(HOME)/opt/sqlite-tools-linux-x86-3360000/sqlite3

.PHONY: all
all: $(INSTALLER)

$(INSTALLER):
	mkdir -p .build

	# extra tools
	cp $(SQLITE3) .build
	cp $(SOLRDUMP) .build

	# our tools
	cd go/ckit && make && cd -
	cp go/ckit/makta .build
	cp go/ckit/tabbedjson .build
	cp go/ckit/labed .build

	# installer and uninstaller
	cp extra/install/install.sh .build
	cp extra/install/labe-uninstall.sh .build

	makeself .build $(INSTALLER) "installing labe $(VERSION)..." ./install.sh

.PHONY: clean
clean:
	rm -rf .build
	rm -f $(INSTALLER)
