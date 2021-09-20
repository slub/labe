# mkocidb

Create a sqlite3 version of a tab separated file.

```sh
$ zstdcat -T0 ../../data/6741422v11s.zst | head
10.1001/10-v4n2-hsf10003        10.1177/003335490912400218
10.1001/10-v4n2-hsf10003        10.1097/01.bcr.0000155527.76205.a2
10.1001/amaguidesnewsletters.1996.novdec01      10.1056/nejm199312303292707
10.1001/amaguidesnewsletters.1996.novdec01      10.1016/s0363-5023(05)80265-5
10.1001/amaguidesnewsletters.1996.novdec01      10.1001/jama.1994.03510440069036
10.1001/amaguidesnewsletters.1997.julaug01      10.1097/00007632-199612150-00003
10.1001/amaguidesnewsletters.1997.mayjun01      10.1164/ajrccm/147.4.1056
10.1001/amaguidesnewsletters.1997.mayjun01      10.1136/thx.38.10.760
10.1001/amaguidesnewsletters.1997.mayjun01      10.1056/nejm199507133330207
10.1001/amaguidesnewsletters.1997.mayjun01      10.1378/chest.88.3.376

$ time zstdcat -T0 ../../data/6741422v11s1.zst | ./mkocidb
2021/09/20 17:34:55 [ok] initialized database -- data.db
written 57.6G -- 41.4M/s
2021/09/20 17:58:40 db setup done
2021/09/20 17:58:40 creating index
2021/09/20 19:07:53 [ok] created index -- data.db

real    92m58.885s
user    79m56.305s
sys     10m0.852s
```
