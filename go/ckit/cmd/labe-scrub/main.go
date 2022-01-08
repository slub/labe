// Script to cleanup space.
//
// Some ideas:
//
// * check if labed is running and get all open files
//
// $ labe-scrub -s /var/lib/labe -d /var/data
//
// ....
//
// $ tree -sh
// .
// ├── [4.0K]  IdMappingDatabase
// │   ├── [  29]  date-2021-11-25.db -> /home/czygan/tmp/id_to_doi.db
// │   └── [ 13G]  date-2022-01-05.db
// ├── [4.0K]  IdMappingTable
// │   └── [438M]  date-2022-01-05.tsv.zst
// ├── [4.0K]  OpenCitationsDatabase
// │   ├── [144G]  a6d28e3e04bc206d3c4021a3381deb5c4443104b.db
// │   └── [150G]  c90e82e35c9d02c00f81bee6d1f34b132953398c.db
// ├── [4.0K]  OpenCitationsDownload
// │   ├── [  28]  a6d28e3e04bc206d3c4021a3381deb5c4443104b.zip -> /home/czygan/tmp/6741422.zip
// │   └── [ 31G]  c90e82e35c9d02c00f81bee6d1f34b132953398c.zip
// ├── [4.0K]  OpenCitationsSingleFile
// │   ├── [ 29G]  a6d28e3e04bc206d3c4021a3381deb5c4443104b.zst
// │   └── [ 31G]  c90e82e35c9d02c00f81bee6d1f34b132953398c.zst
// ├── [4.0K]  SolrDatabase
// │   ├── [ 40G]  date-2021-11-25-name-ai.db
// │   ├── [5.5G]  date-2021-11-25-name-main.db
// │   ├── [217M]  date-2021-11-25-name-slub-production.db
// │   ├── [ 42G]  date-2022-01-05-name-ai.db
// │   ├── [5.5G]  date-2022-01-05-name-main.db
// │   └── [217M]  date-2022-01-05-name-slub-production.db
// └── [4.0K]  SolrFetchDocs
//     ├── [5.5G]  date-2021-11-25-name-ai.zst
//     ├── [968M]  date-2021-11-25-name-main.zst
//     ├── [ 23M]  date-2021-11-25-name-slub-production.zst
//     ├── [5.7G]  date-2022-01-05-name-ai.zst
//     ├── [966M]  date-2022-01-05-name-main.zst
//     ├── [ 23M]  date-2022-01-05-name-slub-production.zst
//     ├── [ 20G]  date-2022-01-07-name-main-short-False.zst
//     └── [231M]  date-2022-01-07-name-slub-production-short-False.zst
//
// 7 directories, 23 files
//
package main

import (
	"flag"
	"log"
)

var (
	dataDir    = flag.String("-d", "", "were all the data lives")
	currentDir = flag.String("-s", "", "where the symlinks to the current databases live")
)

func main() {
	flag.Parse()
	log.Println("scrub")
	// (1) find the targets of all the symlinks
	// (2) find all files in data dir
	// (3) all files, which we do not link to we can could get rid of
	// (4) maybe keep only the newest file per task
}
