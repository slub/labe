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

* [2021-09-23](notes/2021_09_23_meeting_minutes.md)
* [2021-09-16](notes/2021_09_16_meeting_minutes.md)
* [2021-09-14](notes/2021_09_14_meeting_minutes.md)

# SLOC

```
===============================================================================
 Language            Files        Lines         Code     Comments       Blanks
===============================================================================
 Go                      7         1396          952          343          101
 Makefile                5           97           64           11           22
 Python                  9          670          585           14           71
 Plain Text              1            6            0            6            0
-------------------------------------------------------------------------------
 HTML                    2          648          578           13           57
 |- JavaScript           2           66           64            0            2
 (Total)                            714          642           13           59
-------------------------------------------------------------------------------
 Markdown               12         1016            0          775          241
 |- Go                   1            5            4            1            0
 |- JSON                 4           39           37            0            2
 (Total)                           1060           41          776          243
===============================================================================
 Total                  36         3833         2179         1162          492
===============================================================================
```
