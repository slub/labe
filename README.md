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

# SLOC

```
$ tokei -C -t=Go,Python
===============================================================================
 Language            Files        Lines         Code     Comments       Blanks
===============================================================================
 Go                      9         1612         1136          358          118
 Python                 10          700          589           36           75
===============================================================================
 Total                  19         2312         1725          394          193
===============================================================================
```
