# LABE

![](static/canvas.png)

Prototyping citation data flows for [SLUB Dresden](https://www.slub-dresden.de/).

## Project

The project is composed of a couple of command line tools, both written in
Python and Go.

WIP: You can build a single debian package with a single `make` command. That
is you can make, test changes and create a deployment artifact locally. The
deployment artifacts includes executables, systemd unit files and
documentation.

* [ckit](go/ckit), the citation toolkit contains the API server and a few command line utilities
* [python](python), orchestration helper to assemble data files regularly (based on luigi)

## Meeting Minutes

* [2021-11-04](notes/2021_11_04_meeting_minutes.md)
* [2021-10-21](notes/2021_10_21_meeting_minutes.md)
* [2021-10-14](notes/2021_10_14_meeting_minutes.md)
* [2021-10-07](notes/2021_10_07_meeting_minutes.md)
* [2021-09-30](notes/2021_09_30_meeting_minutes.md)
* [2021-09-23](notes/2021_09_23_meeting_minutes.md)
* [2021-09-16](notes/2021_09_16_meeting_minutes.md)
* [2021-09-14](notes/2021_09_14_meeting_minutes.md)

## TODO

* [ ] python distribution setup
* [ ] sync graph data
* [ ] define log format, structured logging
* [ ] calculate subset, data reduction (via solr)
* [ ] generate snapshot
* [ ] indexing pipeline
* [ ] http/rest api
* [ ] delta report
* [ ] ...

# Metrics and Structure

## SLOC

```
$ tokei -C -t=Go,Python
===============================================================================
 Language            Files        Lines         Code     Comments       Blanks
===============================================================================
 Go                     14         2105         1549          400          156
 Python                 10         1135          957           50          128
===============================================================================
 Total                  24         3240         2506          450          284
===============================================================================
```

## Tree

```shell
.
├── [4.0K]  ansible
│   ├── [ 207]  hosts
│   ├── [4.0K]  roles
│   │   ├── [4.0K]  common
│   │   │   └── [4.0K]  tasks
│   │   │       └── [ 478]  main.yml
│   │   └── [4.0K]  labe
│   │       └── [4.0K]  tasks
│   │           └── [ 170]  main.yml
│   └── [  64]  site.yml
├── [4.0K]  data
│   ├── [9.8G]  6741422v11s1.zst
│   ├── [9.7G]  6741422v11s2.zst
│   ├── [ 11G]  6741422v11s.zst
│   ├── [ 29G]  6741422v11.zst
│   ├── [ 18G]  ai.json.zst
│   ├── [4.6G]  ai_to_doi.tsv
│   ├── [2.6G]  ai.url.tsv
│   ├── [ 259]  are_doi_unique_per_record.py
│   ├── [5.8K]  doisniffer.py
│   ├── [ 11G]  id_to_doi.db
│   ├── [4.7G]  id_to_doi.tsv
│   ├── [698M]  id_to_doi.tsv.zst
│   ├── [329G]  index.db
│   ├── [ 13M]  ma.doi.sniffed.tsv.zst
│   ├── [2.7M]  ma.doi.tsv.zst
│   ├── [448K]  ma.doi_unique_link_sample.jsonl.zst
│   ├── [1.5M]  ma.doi_unique.tsv.zst
│   ├── [9.4G]  ma.json.zst
│   ├── [ 374]  Makefile
│   ├── [ 42M]  ma_to_doi.tsv
│   ├── [9.4G]  ma_with_doi.json.zst
│   ├── [145G]  oci.db
│   └── [ 15K]  README.md
├── [4.0K]  go
│   ├── [4.0K]  ckit
│   │   ├── [4.0K]  cmd
│   │   │   ├── [4.0K]  doisniffer
│   │   │   │   └── [2.7K]  main.go
│   │   │   ├── [4.0K]  labed
│   │   │   │   └── [ 16K]  main.go
│   │   │   ├── [4.0K]  makta
│   │   │   │   └── [3.5K]  main.go
│   │   │   └── [4.0K]  tabjson
│   │   │       └── [4.8K]  main.go
│   │   ├── [4.0K]  doi
│   │   │   ├── [3.2K]  sniffer.go
│   │   │   └── [1.4K]  sniffer_test.go
│   │   ├── [4.3K]  fetcher.go
│   │   ├── [4.0K]  fixtures
│   │   │   ├── [7.6M]  id-1m.tsv.zst
│   │   │   ├── [ 519]  makta-2-col-gen.py
│   │   │   ├── [5.7M]  makta-sample-s.tsv.zst
│   │   │   ├── [ 679]  makta-sample-xs.tsv
│   │   │   └── [ 247]  README.md
│   │   ├── [ 647]  go.mod
│   │   ├── [3.2K]  go.sum
│   │   ├── [  23]  i.db -> ../../data/id_to_doi.db
│   │   ├── [  19]  index.db -> ../../data/index.db
│   │   ├── [1.0K]  Makefile
│   │   ├── [  17]  o.db -> ../../data/oci.db
│   │   ├── [4.0K]  packaging
│   │   │   └── [4.0K]  deb
│   │   │       └── [4.0K]  ckit
│   │   │           └── [4.0K]  DEBIAN
│   │   │               └── [ 244]  control
│   │   ├── [ 14K]  README.md
│   │   ├── [ 14K]  server.go
│   │   ├── [4.0K]  set
│   │   │   ├── [3.2K]  set.go
│   │   │   └── [ 915]  set_test.go
│   │   ├── [4.0K]  static
│   │   │   ├── [ 22K]  443238.cast
│   │   │   ├── [281K]  443238.gif
│   │   │   ├── [ 79K]  45582_reading_lg.gif
│   │   │   └── [ 39K]  45582_reading_md.gif
│   │   ├── [3.0K]  stopwatch.go
│   │   ├── [3.2K]  tabutils.go
│   │   └── [4.0K]  xflag
│   │       ├── [ 587]  flag.go
│   │       └── [ 233]  flag_test.go
│   └── [ 286]  README.md
├── [1.0K]  LICENSE
├── [4.0K]  notes
│   ├── [ 505]  2021_09_14_meeting_minutes.md
│   ├── [ 803]  2021_09_16_meeting_minutes.md
│   ├── [ 908]  2021_09_23_meeting_minutes.md
│   ├── [2.9K]  2021_09_30_meeting_minutes.md
│   ├── [ 13K]  2021_09_system_design.md
│   ├── [2.1K]  2021_10_07_meeting_minutes.md
│   ├── [2.3K]  2021_10_14_meeting_minutes.md
│   ├── [2.1K]  2021_10_21_meeting_minutes.md
│   ├── [1.3K]  2021_10_ops_notes.md
│   ├── [ 304]  2021_11_04_meeting_minutes.md
│   ├── [1.5K]  2021_11_11_minimal_deployment.md
│   ├── [ 490]  2021_11_18_meeting_minutes.md
│   ├── [ 645]  2021_11_19_deployment_ansible.md
│   ├── [ 507]  2021_11_25_meeting_minutes.md
│   ├── [ 401]  2021_12_16_meeting_minutes.md
│   ├── [ 936]  2022_01_06_meeting_minutes.md
│   ├── [ 37K]  Labe.drawio
│   ├── [ 37K]  Labe.png
│   └── [ 49K]  uJAWsAE.png
├── [4.0K]  packaging
│   └── [4.0K]  deb
│       └── [4.0K]  labe
│           └── [4.0K]  DEBIAN
│               └── [ 244]  control
├── [4.0K]  python
│   ├── [4.0K]  labe
│   │   ├── [7.0K]  base.py
│   │   ├── [6.2K]  cli.py
│   │   ├── [ 178]  __init__.py
│   │   ├── [5.1K]  oci.py
│   │   └── [9.3K]  tasks.py
│   ├── [ 205]  labe.cfg
│   ├── [ 498]  logging.ini
│   ├── [  31]  luigi.cfg
│   ├── [1.4K]  Makefile
│   ├── [ 167]  pytest.ini
│   ├── [1.4K]  README.md
│   ├── [1.2K]  setup.py
│   └── [4.0K]  tests
│       └── [1.6K]  test_oci.py
├── [7.5K]  README.md
├── [4.0K]  static
│   └── [6.6K]  canvas.png
├── [1.4K]  TODO.md
└── [3.0K]  Vagrantfile

32 directories, 97 files
```
