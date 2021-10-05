# Data conversions, for documentation.

6741422v11: 6741422v11.zip
	unzip -d 6741422v11 6741422v11.zip

6741422v11.zst: 6741422v11
	for f in $$(find 6741422v11 -name "*zip"); do unzip -p $$f; done | LC_ALL=C grep -vF 'oci,citing' | zstd -c -T0 > 6741422v11.zst

6741422v11.db: 6741422v11s1.zst
	zstdcat -T0 6741422v11s1.zst | LC_ALL=C grep -v '^"' | slikv -o 6741422v11.db

