#!/bin/bash
#
# Installation script for labe tools.

set -eu -o pipefail

BIN=${HOME}/.local/bin

mkdir -p "$BIN"
# TODO: generate uninstall script on the fly from list of binaries in this file
for prog in solrdump tabbedjson makta labed labe-uninstall.sh; do
    cp "$prog" "$BIN"
done

