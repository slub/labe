# Labe

![](extra/banner/static/canvas.png)

Prototyping citation data flows.

```sh
$ du -hs .
839G
```

# Notes

Install dev dependencies.

```
$ pip install -e .[dev]
```

# TODO

* [ ] python distribution setup
* [ ] sync graph data
* [ ] define log format, structured logging
* [ ] calculate subset, data reduction (via solr)
* [ ] generate snapshot
* [ ] indexing pipeline
* [ ] http/rest api
* [ ] delta report
* [ ] ...

# Meeting Minutes

* [2021-10-14](notes/2021_10_14_meeting_minutes.md)
* [2021-10-07](notes/2021_10_07_meeting_minutes.md)
* [2021-09-30](notes/2021_09_30_meeting_minutes.md)
* [2021-09-23](notes/2021_09_23_meeting_minutes.md)
* [2021-09-16](notes/2021_09_16_meeting_minutes.md)
* [2021-09-14](notes/2021_09_14_meeting_minutes.md)

# SLOC

```
$ tokei -C -t=Go,Python
===============================================================================
 Language            Files        Lines         Code     Comments       Blanks
===============================================================================
 Go                      7         1396          952          343          101
 Python                  9          670          585           14           71
===============================================================================
 Total                  16         2066         1537          357          172
===============================================================================
```
