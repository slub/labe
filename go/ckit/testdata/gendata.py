#!/usr/bin/env python

# id-doi, "id_doi.tsv"
# doi-doi, "doi_doi.tsv"
# id-metadata, "id_metadata.tsv"

import random
random.seed(0)

num_docs = 100

with open("id_doi.tsv", "w") as f:
    for i in range(num_docs):
        print("i{:04}\td{:04}".format(i, i), file=f)

with open("doi_doi.tsv", "w") as f:
    num_dois = num_docs + num_docs
    for i in range(num_dois):
        a, b = random.sample(range(num_dois), 2)
        print("d{:04}\td{:04}".format(a, b), file=f)

with open("id_metadata.tsv", "w") as f:
    vs = ["DE-1", "DE-2", "DE-3"]
    for i in range(num_docs):
        institution = vs[i%len(vs)]
        data = """{"a": "%s", "b": "ok", "institution": ["%s"]}""" % (i, institution)
        print("i{:04}\t{}".format(i, data), file=f)

# $ head *tsv
# ==> doi_doi.tsv <==
# d0098   d0194
# d0107   d0010
# d0066   d0130
# d0124   d0103
# d0077   d0122
# d0091   d0149
# d0055   d0129
# d0035   d0072
# d0035   d0193
# d0024   d0158
#
# ==> id_doi.tsv <==
# i0000   d0000
# i0001   d0001
# i0002   d0002
# i0003   d0003
# i0004   d0004
# i0005   d0005
# i0006   d0006
# i0007   d0007
# i0008   d0008
# i0009   d0009
#
# ==> id_metadata.tsv <==
# i0000   {"a": "0", "b": "ok", "institution": ["DE-1"]}
# i0001   {"a": "1", "b": "ok", "institution": ["DE-2"]}
# i0002   {"a": "2", "b": "ok", "institution": ["DE-3"]}
# i0003   {"a": "3", "b": "ok", "institution": ["DE-1"]}
# i0004   {"a": "4", "b": "ok", "institution": ["DE-2"]}
# i0005   {"a": "5", "b": "ok", "institution": ["DE-3"]}
# i0006   {"a": "6", "b": "ok", "institution": ["DE-1"]}
# i0007   {"a": "7", "b": "ok", "institution": ["DE-2"]}
# i0008   {"a": "8", "b": "ok", "institution": ["DE-3"]}
# i0009   {"a": "9", "b": "ok", "institution": ["DE-1"]}
