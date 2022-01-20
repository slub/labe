SHELL := /bin/bash
ANSIBLE_OPTS = ANSIBLE_RETRY_FILES_ENABLED=false ANSIBLE_NOCOWS=true ANSIBLE_HOST_KEY_CHECKING=false

.PHONY: deploy
deploy:
	$(ANSIBLE_OPTS) ansible-playbook --ask-become-pass -b -v -i ansible/hosts ansible/site.yml

.PHONY: deb
deb:
	(cd python && make clean && make deb)
	(cd go/ckit && make clean && make -j deb)
