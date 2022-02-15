// makta takes a two column TSV file and turns it into an indexed sqlite3 database.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/andrew-d/go-termutil"
	"github.com/slub/labe/go/ckit"
	"github.com/slub/labe/go/ckit/tabutils"
)

var (
	Version    string
	Buildtime  string
	validTypes = []string{"INTEGER", "READ", "TEXT", "BLOB"} // sqlite3

	showVersion  = flag.Bool("version", false, "show version and exit")
	outputFile   = flag.String("o", "data.db", "output filename")
	bufferSize   = flag.Int("B", 64*1<<20, "buffer size")
	indexMode    = flag.Int("I", 3, "index mode: 0=none, 1=k, 2=v, 3=kv")
	cacheSize    = flag.Int("C", 1000000, "sqlite3 cache size, needs memory = C x page size")
	initDatabase = flag.Bool("init", false, "on start, initialize database, even when the file already exists")
	valueType    = flag.String("T", "TEXT", "sqlite3 type for value column")
	verbose      = flag.Bool("verbose", false, "be verbose")
)

func main() {
	flag.Parse()
	if !ckit.SliceContains(validTypes, *valueType) {
		log.Fatalf("invalid type for value column: %v %v", *valueType, validTypes)
	}
	var (
		err      error
		initFile string
		pragma   = fmt.Sprintf(`
PRAGMA journal_mode = OFF;
PRAGMA synchronous = 0;
PRAGMA cache_size = %d;
PRAGMA locking_mode = EXCLUSIVE;`, *cacheSize)
		initSQL = `
CREATE TABLE IF NOT EXISTS map (k TEXT, v %s);`
		keyIndexSQL = fmt.Sprintf(`
%s
CREATE INDEX IF NOT EXISTS idx_k ON map(k);`, pragma)
		valueIndexSQL = fmt.Sprintf(`
%s
CREATE INDEX IF NOT EXISTS idx_v ON map(v);`, pragma)
		importSQL = fmt.Sprintf(`
%s
PRAGMA temp_store = MEMORY;

.mode tabs
.import /dev/stdin map`, pragma)
	)
	if *showVersion {
		fmt.Printf("makta %s %s\n", Version, Buildtime)
		os.Exit(0)
	}
	if termutil.Isatty(os.Stdin.Fd()) {
		log.Println("stdin: no data")
		os.Exit(1)
	}
	_, err = os.Stat(*outputFile)
	if err != nil || *initDatabase {
		if os.IsNotExist(err) || *initDatabase {
			if err := tabutils.RunScript(*outputFile, fmt.Sprintf(initSQL, *valueType), "initialized database"); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}
	if initFile, err = tabutils.TempFileReader(strings.NewReader(importSQL)); err != nil {
		log.Fatal(err)
	}
	var (
		br          = bufio.NewReader(os.Stdin)
		buf         bytes.Buffer
		written     int64
		started     = time.Now()
		elapsed     float64
		numBatches  int
		importBatch = func() error {
			n, err := tabutils.RunImport(&buf, initFile, *outputFile)
			if err != nil {
				return fmt.Errorf("import: %w", err)
			}
			written += n
			numBatches++
			elapsed = time.Since(started).Seconds()
			if *verbose {
				log.Printf("imported batch #%d, %s at %s",
					numBatches,
					tabutils.ByteSize(int(written)),
					tabutils.HumanSpeed(written, elapsed))
			} else {
				tabutils.Flushf("written %s · %s",
					tabutils.ByteSize(int(written)),
					tabutils.HumanSpeed(written, elapsed))
			}
			return nil
		}
		indexScripts []string
	)
	for {
		b, err := br.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("read: %v", err)
		}
		if _, err := buf.Write(b); err != nil {
			log.Fatalf("write: %v", err)
		}
		if buf.Len() >= *bufferSize {
			if err := importBatch(); err != nil {
				log.Fatalf("batch: %v", err)
			}
		}
	}
	if err := importBatch(); err != nil {
		log.Fatalf("batch: %v", err)
	}
	fmt.Println()
	switch *indexMode {
	case 1:
		indexScripts = []string{
			keyIndexSQL,
		}
	case 2:
		indexScripts = []string{
			valueIndexSQL,
		}
	case 3:
		indexScripts = []string{
			keyIndexSQL,
			valueIndexSQL,
		}
	default:
		log.Printf("no index requested")
	}
	log.Printf("[io] building %d indices ...", len(indexScripts))
	for i, script := range indexScripts {
		msg := fmt.Sprintf("%d/%d created index", i+1, len(indexScripts))
		if err := tabutils.RunScript(*outputFile, script, msg); err != nil {
			log.Fatalf("run script: %v", err)
		}
	}
}
