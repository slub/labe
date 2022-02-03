# LABE

![](static/canvas.png)

Merging citations and catalog data at [SLUB Dresden](https://www.slub-dresden.de/).

![](static/Overview.png)

> Status: testing

## Project

The project is composed of a couple of command line tools, both written in
Python and Go.

* [ckit](go/ckit), citation toolkit contains an API server, plus a few command line tools
* [python](python), orchestration helper to assemble data files regularly (based on luigi)

## Meeting Minutes

* [2022-02-03](notes/2022_02_03_meeting_minutes.md) (wrap-up)
* [2022-01-20](notes/2022_01_20_meeting_minutes.md)
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
* [2021-09-14](notes/2021_09_14_meeting_minutes.md) (kick-off)

## Project structure

```shell
$ tree -d
.
├── ansible
│   └── roles
│       ├── common
│       │   └── tasks
│       └── labe
│           ├── defaults
│           └── tasks
├── data
├── go
│   └── ckit
│       ├── cmd
│       │   ├── doisniffer
│       │   ├── labed
│       │   ├── makta
│       │   └── tabjson
│       ├── doi
│       ├── fixtures
│       ├── packaging
│       │   └── deb
│       │       └── ckit
│       │           └── DEBIAN
│       ├── set
│       ├── static
│       └── xflag
├── notes
├── python
│   ├── labe
│   ├── packaging
│   │   └── deb
│   │       └── labe
│   │           └── DEBIAN
│   └── tests
└── static

33 directories
```

## SLOC

```
$ tokei -C -t=Go,Python,yaml
===============================================================================
 Language            Files        Lines         Code     Comments       Blanks
===============================================================================
 Go                     18         2550         2029          338          183
 Python                 14         2006         1647           88          271
 YAML                    4          194          162           17           15
===============================================================================
 Total                  36         4750         3838          443          469
===============================================================================
```

