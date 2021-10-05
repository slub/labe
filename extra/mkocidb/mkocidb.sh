#!/bin/bash

# Generate an OCI database version from a given URL.

set -e
set -o errexit -o pipefail -o nounset

# Adjust this URL for updates.
DATASET_URL=https://figshare.com/ndownloader/articles/6741422/versions/11
OUTPUT=oci-$(date +"%Y-%m-%d")-$(echo -n $DATASET_URL | sha1sum | awk '{print $1}').db
FILE=""

while [[ "$#" -gt 0 ]]; do
	case $1 in
	-h | --help)
		cat <<EOF
usage: mkocidb [OPTIONS]

  -u, --url URL

    The direct URL to an OCI data dump. This is expected to be a zip of zips.
    Example data point, version 11 of the dataset was 30GB in size.
    Default: $DATASET_URL

  -f, --file FILE

    Instead of downloading a dump from a URL, use FILE instead. If this flag is
    set it is preferred over URL.

  -o, --output FILE

    Filename for sqlite3 database. This file can be used for the labesvc
    µ-service. Default: $OUTPUT

EOF
		exit 0
		;;
	-u | --url)
		DATASET_URL="$2"
		shift
		;;
	-f | --file)
		FILE="$2"
		shift
		;;
	-o | --output)
		OUTPUT="$2"
		shift
		;;
	-u | --uglify) uglify=1 ;;
	*)
		echo "unknown parameter passed: $1"
		exit 1
		;;
	esac
	shift
done

for req in curl unzip zstd hck slikv; do
	command -v $req >/dev/null 2>&1 || {
		echo >&2 "$req required"
		exit 1
	}
done

T=$(mktemp)
D=$(mktemp -d)
Z=$(mktemp)
F=$(mktemp)

function cleanup() {
	rm -f "$T" "$Z" "$F"
	rm -rf "$D"
}

if [ -z "$FILE" ]; then
	curl -vL "$DATASET_URL" >"$T" && mv "$T" "$FILE"
fi

unzip -d "$D" "$T"
for f in $(find "$D" -name "*zip"); do
	unzip -p "$f"
done | LC_ALL=C grep -vF 'oci,citing' | zstd -c -T0 >"$Z"
zstdcat -T0 "$Z" | hck -d, -f2,3 | slikv -o "$F" && mv "$F" "$OUTPUT"

