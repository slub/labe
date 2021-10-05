// labectl manages data for labe, e.g. download and transformations; using
// external tools. Data lives under a single directory, by default under
// $XDG_DATA_HOME.
//
// (1) Download OCI dumps, check for updates; allow for manual override.
// (2) Turn OCI zip to db with slikv.
// (3) Download index copies, e.g. via solrdump.
// (4) Generate id-to-doi mapping, jq, slikv.
// (5) Generate db version of index, slikv.
package main

import (
	"flag"
	"fmt"
	"log"
	"path"

	"github.com/adrg/xdg"
)

var (
	dataHome            = flag.String("b", path.Join(xdg.DataHome, "labe"), "dir where all artifacts will be stored")
	generateOciDatabase = flag.String("C", "", "generate OCI database from zip URL")
	outputFile          = flag.String("o", "", "output filename")

	usage = `usage: labectl [OPTIONS]

  $ labectl -o oci.db -C https://figshare.com/ndownloader/articles/6741422/versions/11
`
)

func main() {
	flag.Usage = func() {
		fmt.Println(usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	log.Printf("labectl (%s)", *dataHome)
	// (1) Turn remote OCI zip file into a local database in one go. We do not
	// need to download the zip file, if figshare supports range requests; cf.
	// "seekable http",
	// https://blog.gopheracademy.com/advent-2017/seekable-http/ Update:
	// unfortunately, figshare/s3 does not support http range requests.
}
