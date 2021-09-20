// mkocidb takes two columns and turns it into an indexed sqlite3 database.
// Example: Help to query an OCI data dump offline.
//
// TODO: sqlite pipe is gone, before stdin is done; maybe we need to import in
// smaller chunks
//
// 2021/09/20 10:36:01 [2301] wrote 1048576 (2412773376)
// 2021/09/20 10:36:01 [2302] wrote 1048576 (2413821952)
// 2021/09/20 10:36:01 copy failed: write |1: broken pipe
//
package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	dry        = flag.Bool("d", false, "dry run")
	outputFile = flag.String("o", "data.db", "output filename")
	initSQL    = `
PRAGMA journal_mode = OFF;
PRAGMA synchronous = 0;
PRAGMA cache_size = 1000000;
PRAGMA locking_mode = EXCLUSIVE;
PRAGMA temp_store = MEMORY;

CREATE TABLE IF NOT EXISTS map
(
	k TEXT,
	v TEXT
);

CREATE INDEX IF NOT EXISTS idx_k ON map(k);
CREATE INDEX IF NOT EXISTS idx_v ON map(v);
`
	runSQL = `
.mode tabs
.import /dev/stdin map
`
)

// TempFileReader returns path to temporary file with contents from reader.
func TempFileReader(r io.Reader) (string, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return f.Name(), nil
}

type LoggingWriter struct {
	i        int
	numBytes int
	w        io.Writer
}

func (w *LoggingWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	if err != nil {
		return
	}
	w.i++
	w.numBytes += n
	log.Printf("[%d] wrote %d (%d)", w.i, n, w.numBytes)
	return
}

// setupDatabase initializes the database, if necessary
func setupDatabase(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cmd := exec.Command("sqlite3", path)
		cmd.Stdin = strings.NewReader(initSQL)
		err := cmd.Run()
		if err == nil {
			log.Printf("[ok] initialized database at %s", path)
		}
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	if err := setupDatabase(*outputFile); err != nil {
		log.Fatal(err)
	}
	runFile, err := TempFileReader(strings.NewReader(runSQL))
	if err != nil {
		log.Fatal(err)
	}
	// TODO: start one process every N lines to save memory (broke with "m" dataset, 100M rows)
	cmd := exec.Command("sqlite3", "--init", runFile, *outputFile)
	cmdStdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer cmdStdin.Close()
		br := buf.NewReader(os.Stdin)
		for {
			line, err := br.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
		}
		if _, err := io.CopyBuffer(cmdStdin, os.Stdin); err != nil {
			log.Fatalf("copy failed: %v", err)
		}
	}()
	if *dry {
		log.Println(cmd)
		os.Exit(0)
	}
	if _, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("exec failed: %v", err)
	}
}
