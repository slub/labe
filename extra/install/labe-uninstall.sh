#!/bin/sh
#
# Uninstaller for labe tools.

set -eu -o pipefail
BIN=${HOME}/.local/bin

for prog in "solrdump tabbedjson makta labed"; do
    rm -f "$BIN/$prog"
done
