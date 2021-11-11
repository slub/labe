#!/usr/bin/env python
#
# Test data generator for makta; generates a TSV file that we can feed into
# makta/sqlite3.
#
# $ python make_2col_tsv.py 10
# K0      V0
# K1      V1
# K2      V2
# K3      V3
# K4      V4
# K5      V5
# K6      V6
# K7      V7
# K8      V8
# K9      V9
#
# Generating 1B rows takes about 15min.
#
# real    14m36.929s
# user    14m14.175s
# sys     0m20.676s

import sys

N = 1_000_000_000 if len(sys.argv) == 1 else int(sys.argv[1])

for i in range(N):
    print("K%s\tV%s" % (i, hash(i)))

