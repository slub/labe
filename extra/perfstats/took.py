#!/bin/bash

# Take a single column file with durations (e.g. in seconds) and print out same stats.
#
# 0.028951295
# 0.019983235
# 0.014776105
# 0.007000522
# 0.006672236
# 0.006075279
# 0.007077829
# 0.007280961
# 0.007379425
# 0.120662236

import pandas as pd
import sys

if len(sys.argv) < 2:
    print("usage: %s FILE" % sys.argv[0], file=sys.stderr)
    sys.exit(1)

df = pd.read_csv(sys.argv[1], skip_blank_lines=True, header=None, names=["s"])
df.describe(percentiles=[0.25, 0.5, 0.75, 0.95, 0.99, 0.995, 0.999, 0.9999, 1]).to_csv(
    sys.stdout, sep="\t", header=None
)
