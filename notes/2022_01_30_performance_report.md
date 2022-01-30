# Performance Report

> 2022-01-22

```
$ neofetch

       _,met$$$$$gg.          hey@sdvlabe
    ,g$$$$$$$$$$$$$$$P.       --------------
  ,g$$P"     """Y$$.".        OS: Debian GNU/Linux 10 (buster) x86_64
 ,$$P'              `$$$.     Host: VMware Virtual Platform None
',$$P       ,ggs.     `$$b:   Kernel: 4.19.0-18-amd64
`d$$'     ,$P"'   .    $$$    Uptime: 87 days, 55 mins
 $$P      d$'     ,    $$P    Packages: 469 (dpkg)
 $$:      $$.   -    ,d$$'    Shell: bash 5.0.3
 $$;      Y$b._   _,d$P'      Terminal: /dev/pts/0
 Y$$.    `.`"Y$$$$P"'         CPU: Intel Xeon Gold 5218 (8) @ 2.294GHz
 `$$b      "-.__              GPU: VMware SVGA II Adapter
  `Y$$                        Memory: 2523MiB / 16019MiB
   `Y$$.
     `$$b.
       `Y$$b.
          `"Y$b._
              `"""

$ labed -version
labed 88397c6 2022-01-30T01:03:12Z
```

Random ids are chosen from ~72M possible ids that have a map to a DOI (this
does not mean, that it has a relation in the citations database).

```
$ zstdcat -T0 /usr/share/labe/data/IdMappingTable/current | wc -l
72731297
```

The number of cached items in subsequent tests may be slightly bigger (as
additional items have been cached).

Format:

* number of local ids
* number of cached items
* parallel requests
* filter (y/n)

## 10K / 150K / 32 / n

```
$ time zstdcat -T0 /usr/share/labe/data/IdMappingTable/current | \
    awk '{print $1}' | shuf -n 10000 | \
    parallel -j 32 -I {} "curl -sL 'http://localhost:8000/id/{}'" | \
    jq -rc .extra.took > 10_150_32_n.tsv

real    0m57.044s
user    1m51.982s
sys     1m10.747s
```

> Results

```
count   6026.0
mean    0.069525608822768
std     0.13572305001969948
min     0.001951155
25%     0.01611085775
50%     0.0373924795
75%     0.06952258375
95%     0.2035256885
99%     0.73334005
99.5%   1.0769603348749999
99.9%   1.4530845264750059
99.99%  1.5616499532224175
100%    1.618536359
max     1.618536359
```

## 10K / 150K / 32 / y

```
$ time zstdcat -T0 /usr/share/labe/data/IdMappingTable/current | \
    awk '{print $1}' | shuf -n 10000 | \
    parallel -j 32 -I {} "curl -sL 'http://localhost:8000/id/{}?i=DE-14'" | \
    jq -rc .extra.took > 10_150_32_y.tsv

real    0m58.236s
user    1m54.580s
sys     1m12.651s
```

> Results

```
count   5999.0
mean    0.052833624121353555
std     0.0853497560623813
min     0.000430559
25%     0.013456665
50%     0.030074125
75%     0.0572592625
95%     0.16796195829999996
99%     0.4916874135799958
99.5%   0.6395566931400012
99.9%   0.7722342863040145
99.99%  1.1777203690489761
100%    1.194194023
max     1.194194023
```

## 100K / 150K / 16 / n

```
$ time zstdcat -T0 /usr/share/labe/data/IdMappingTable/current | \
    awk '{print $1}' | shuf -n 100000 | \
    parallel -j 16 -I {} "curl -sL 'http://localhost:8000/id/{}'" | \
    jq -rc .extra.took > 100_150_16_n.tsv

real    6m14.384s
user    13m12.131s
sys     9m42.016s
```

> Results

```
count   59746.0
mean    0.049754040905165206
std     0.09418759928968966
min     0.000474105
25%     0.0125099175
50%     0.029008653000000002
75%     0.055278471
95%     0.148788049
99%     0.39075252255000054
99.5%   0.5369267160250015
99.9%   1.2946787526800094
99.99%  2.522602559548393
100%    4.5396280860000005
max     4.5396280860000005
```

## 100K / 150K / 48 / n

```
$ time zstdcat -T0 /usr/share/labe/data/IdMappingTable/current | \
    awk '{print $1}' | shuf -n 100000 | \
    parallel -j 48 -I {} "curl -sL 'http://localhost:8000/id/{}'" | \
    jq -rc .extra.took > 100_150_48_n.tsv

real    6m13.058s
user    13m25.169s
sys     10m2.076s
```

![](../static/100_150_48_screenie.png)

> Results

```
count   59400.0
mean    0.059010946614276104
std     0.1174613430670048
min     0.000107
25%     0.013895456
50%     0.031852294
75%     0.0616479725
95%     0.18499460749999985
99%     0.5130498874900006
99.5%   0.780213243339988
99.9%   1.414341842967231
99.99%  2.0529509573230866
100%    7.522301854
max     7.522301854
```

