SHELL := /bin/bash
ANSIBLE_OPTS = ANSIBLE_RETRY_FILES_ENABLED=false ANSIBLE_NOCOWS=true ANSIBLE_HOST_KEY_CHECKING=false

.PHONY: help
help: ## print info about all commands
	@echo "Commands:"
	@grep -E '^[/.a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[01;32m%-40s\033[0m %s\n", $$1, $$2}'

.PHONY: deploy
deploy: ## deploy ansible/set.yml
	$(ANSIBLE_OPTS) ansible-playbook --ask-become-pass -b -v -i ansible/hosts ansible/site.yml

.PHONY: deb
deb: ## shortcut to build both ckit and labe debian packages
	(cd python && make clean && make deb)
	(cd go/ckit && make clean && make -j deb)

.PHONY: clean
clean:
	(cd python && make clean)
	(cd go/ckit && make clean)
