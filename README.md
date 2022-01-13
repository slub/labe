# LABE

![](static/canvas.png)

Merging citations and catalog data at [SLUB Dresden](https://www.slub-dresden.de/).

## Project

The project is composed of a couple of command line tools, both written in
Python and Go.

* [ckit](go/ckit), the citation toolkit contains the API server and a few command line utilities
* [python](python), orchestration helper to assemble data files regularly (based on luigi)

## Meeting Minutes

* [2022-01-13](notes/2022_01_13_meeting_minutes.md)
* [2022-01-06](notes/2022_01_06_meeting_minutes.md)
* [2021-12-16](notes/2021_12_16_meeting_minutes.md)
* [2021-11-25](notes/2021_11_25_meeting_minutes.md)
* [2021-11-18](notes/2021_11_18_meeting_minutes.md)
* [2021-11-04](notes/2021_11_04_meeting_minutes.md)
* [2021-10-21](notes/2021_10_21_meeting_minutes.md)
* [2021-10-14](notes/2021_10_14_meeting_minutes.md)
* [2021-10-07](notes/2021_10_07_meeting_minutes.md)
* [2021-09-30](notes/2021_09_30_meeting_minutes.md)
* [2021-09-23](notes/2021_09_23_meeting_minutes.md)
* [2021-09-16](notes/2021_09_16_meeting_minutes.md)
* [2021-09-14](notes/2021_09_14_meeting_minutes.md)

## SLOC

```
$ tokei -C -t=Go,Python
===============================================================================
 Language            Files        Lines         Code     Comments       Blanks
===============================================================================
 Go                     15         1926         1523          249          154
 Python                 11         1305         1085           62          158
===============================================================================
 Total                  26         3231         2608          311          312
===============================================================================
```

