import fileinput
import collections

id_to_doi = collections.defaultdict(set)

for line in fileinput.input():
    fields = line.strip().split("\t")
    id_to_doi[fields[0]].add(fields[2])

for k, v in id_to_doi.items():
    if len(v) > 1:
        print(k, v)
