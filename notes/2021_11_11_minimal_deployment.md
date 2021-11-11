# Minimal deployment

* use [XDG_DATA_HOME](https://wiki.archlinux.org/title/XDG_Base_Directory) for files

```
$ mkdir -p ~/.local/share/labe
```

Trying `rsync` first, not installed; will try scp.

```
$ scp -C -o port=21 ../../data/id_to_doi.db user@server.slub-dresden.de:/home/x/.local/share/labe
```
