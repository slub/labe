# Deployment

* default user/group is "labe"
* you can set the version of ckit and labe packages to be installed in
  [roles/labe/defaults/main.yml](https://github.com/slub/labe/blob/main/ansible/roles/labe/defaults/main.yml)

After deployment the `labed.service` is not started (or restarted); the server
requires the sqlite3 database files anyway, which are not present directly
after deployment; also: the installed cron jobs will restart the server, once
the database files have been successfully created.

The deployment has been tested with Debian 10
([buster](https://www.debian.org/releases/buster/)) and Debian 11
([bullseye](https://www.debian.org/releases/bullseye/)).

## Cron

The labe role includes crontab setup.

## Lint

* via [ansible-lint](https://ansible-lint.readthedocs.io)

```
$ ansible-lint
```
