# Mock data generator
#
# Run `make` to regenerate the data. Change random.seed if you need other data.
#
# No `make clean` since even with random.seed, sqlite3 seems to encode the
# creation time in the database (running `make` again should yields same data,
# but databases differ).
SHELL := /bin/bash
TARGETS := id_doi.db doi_doi.db id_metadata.db

.PHONY: all
all: $(TARGETS)

.PHONY: fakes
fakes:
	python gendata.py

id_doi.db: fakes
	makta -I 3 -o $@ < id_doi.tsv

doi_doi.db: fakes
	makta -I 3 -o $@ < doi_doi.tsv

id_metadata.db: fakes
	makta -I 1 -o $@ < id_metadata.tsv

