// labectl manages data for labe, e.g. download and transformations; using
// external tools. Data lives under a single directory, by default under
// $XDG_DATA_HOME.
package main

import (
	"flag"
	"log"
	"path"

	"github.com/adrg/xdg"
)

var (
	dataHome = flag.String("b", path.Join(xdg.DataHome, "labe"), "dir where all artifacts will be stored")
)

func main() {
	log.Printf("labectl (%s)", *dataHome)
	// (1) Turn remote OCI zip file into a local database in one go. We do not
	// need to download the zip file, if figshare supports range requests; cf.
	// "seekable http",
	// https://blog.gopheracademy.com/advent-2017/seekable-http/ Update:
	// unfortunately, figshare/s3 does not support http range requests.
}
