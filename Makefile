SHELL := /bin/bash
VERSION := 1.0.0
INSTALLER := labe-$(VERSION)-installer.run

# Other static binaries we may need.
SOLRDUMP := $(HOME)/code/miku/solrdump/solrdump

.PHONY: all
all: $(INSTALLER)

$(INSTALLER):
	mkdir -p .build
	cp $(SOLRDUMP) .build
	cd go/ckit && make && cd -
	cp go/ckit/makta .build
	cp go/ckit/tabbedjson .build
	cp go/ckit/labed .build
	cp extra/install/install.sh .build
	cp extra/install/labe-uninstall.sh .build

	makeself .build $(INSTALLER) "installing labe $(VERSION)..." ./install.sh

.PHONY: clean
clean:
	rm -rf .build
	rm -f $(INSTALLER)
