# Deployment

* user/group is "labe"

## Cron

* https://crontab.guru/#10_0_*_*_*

```cron
SHELL=/bin/bash
PATH=...

10 0 * * * rm -f $(/usr/local/bin/labe.pyz --list-deletable) && /usr/local/bin/labe.pyz -r CombinedUpdate --workers 4    
```