// labectl manages the path from raw data to databases for OCI and index data.
// This program manages a local path and will be able to answer certain
// questions regerding updates. This program should also be usable with cron.
//
// Keeps a internal cache and operations are atomic.
//
// $ labectl -A -o 1.db -S http://data.index/1
// $ labectl -A -o 2.db -S http://data.index/2
// $ labectl -A -o i.db -I -S http://data.index/1 -S http://data.index/2
// $ labectl -A -o o.db -O
package main

import (
	"flag"
	"fmt"
)

var (
	updateSolr               = flag.String("S", "", "create a new database from an index dump, given the solr url")
	updateIdentifierDatabase = flag.Bool("I", false, "create identifier database")
	updateOciDatabase        = flag.String("O", "", "create OCI database from default or supplied location")
	outputFile               = flag.String("o", "", "output file for operation")
	autoUpdate               = flag.Bool("A", false, "if updated version passes plausibility check, run update automatically, otherwise log issues")
	logFile                  = flag.String("l", "", "path to logfile, use stderr if empty")
)

func main() {
	fmt.Println("labectl")
}
