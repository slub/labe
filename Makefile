SHELL := /bin/bash
ANSIBLE_OPTS = ANSIBLE_RETRY_FILES_ENABLED=false ANSIBLE_NOCOWS=true ANSIBLE_HOST_KEY_CHECKING=false

.PHONY: deploy
deploy:
	$(ANSIBLE_OPTS) ansible-playbook -b -v -i ansible/hosts ansible/site.yml
