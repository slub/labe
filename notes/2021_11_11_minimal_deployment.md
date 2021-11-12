# Minimal deployment

* use [XDG_DATA_HOME](https://wiki.archlinux.org/title/XDG_Base_Directory) for files

```
$ mkdir -p ~/.local/share/labe
```

Trying `rsync` first, not installed; will try scp.

```
$ scp -C -o port=21 ../../data/id_to_doi.db user@server.slub-dresden.de:/home/x/.local/share/labe
```

## First try, w/o root

We cannot use apt, so we try to setup a minimal dev environment.

* [x] download packages into `~/opt`
* [x] install Go into `~/.local/go`

```
$ tar -C ~/.local/ -xzf go1.17.3.linux-amd64.tar.gz
```

* [x] in .bashrc add `~/.local/go/bin` to path

It works.

```
$ go version
go version go1.17.3 linux/amd64
```

Install git from source:

* [https://mirrors.edge.kernel.org/pub/software/scm/git/](https://mirrors.edge.kernel.org/pub/software/scm/git/)

Ok, we do not even have a C compiler. Let's stop here.

## Truly

All we need are executables: solrdump, labed, etc. and we can package Python as
a single file runnable as well. Build zip file locally. Or self executable
files, that extract the binaries into e.g. ~/.local/bin. [unzipsfx](https://www.unix.com/man-page/linux/1/unzipsfx/).

Could write installed files into ~/.local/share/ckit/installed and then all to
uninstall or upgrade as well.
