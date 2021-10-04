# Ops notes

Notes and ideas on operational issues, e.g. data updates, deployments, web
services, etc.

## Artifacts

The deployment artifact in 10/2021 can be a single
[deb](https://en.wikipedia.org/wiki/Deb_(file_format)) file. The deployment
target is [Debian 10](https://www.debian.org/releases/buster/releasenotes).

The deployment artifact should be builable in a single command on a development
machine or a CI system.

```
$ make deb
```

For the build to succeed, two tools are required: `fakeroot` and `dpkg-deb` -
available on the host, or within a build image (in which case, docker or podman
needs to be installed).

## Upgrade and downgrade path

To upgrade to a newer version, standard command like `dkpg` or `apt` need to suffice.

## Deploy units

* a server component ("REST-API")
* one or more command line tools required for data acquisition and preparation ("pipeline")
* servers and tools will use a single configuration file in a standard location; an example configuration in supplied on install

## System Landscape

The server opens a single, configurable port for HTTP. The backend server keeps
logs (rotated) and will use a reverse proxy (e.g. nginx). The nginx
configuration should be part of the documentation. The service may implement a
health check URL.
