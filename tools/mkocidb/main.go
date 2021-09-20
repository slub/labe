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
// A chucked insert seems to work.
//
// $ time ./mkocidb < sample-m.tsv
// 2021/09/20 13:13:55 [ok] initialized database at data.db
// written 5G -- 6.6M/s
//
// real    12m46.585s
// user    11m17.193s
// sys     3m2.369s
//
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
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
PRAGMA journal_mode = OFF;
PRAGMA synchronous = 0;
PRAGMA cache_size = 1000000;
PRAGMA locking_mode = EXCLUSIVE;
PRAGMA temp_store = MEMORY;

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

// ByteSize returns a human-readable byte string of the form 10M, 12.5K, and so
// forth.  The following units are available: E: Exabyte, P: Petabyte, T:
// Terabyte, G: Gigabyte, M: Megabyte, K: Kilobyte, B: Byte, The unit that
// results in the smallest number greater than or equal to 1 is always chosen.
func ByteSize(bytes int64) string {
	const (
		BYTE = 1 << (10 * iota)
		KILOBYTE
		MEGABYTE
		GIGABYTE
		TERABYTE
		PETABYTE
		EXABYTE
	)
	unit := ""
	value := float64(bytes)
	switch {
	case bytes >= EXABYTE:
		unit = "E"
		value = value / EXABYTE
	case bytes >= PETABYTE:
		unit = "P"
		value = value / PETABYTE
	case bytes >= TERABYTE:
		unit = "T"
		value = value / TERABYTE
	case bytes >= GIGABYTE:
		unit = "G"
		value = value / GIGABYTE
	case bytes >= MEGABYTE:
		unit = "M"
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = "K"
		value = value / KILOBYTE
	case bytes >= BYTE:
		unit = "B"
	case bytes == 0:
		return "0B"
	}

	result := strconv.FormatFloat(value, 'f', 1, 64)
	result = strings.TrimSuffix(result, ".0")
	return result + unit
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

func runImport(r io.Reader, initFile, outputFile string) (int64, error) {
	cmd := exec.Command("sqlite3", "--init", initFile, outputFile)
	cmdStdin, err := cmd.StdinPipe()
	if err != nil {
		return 0, err
	}
	var (
		wg      sync.WaitGroup
		copyErr error
		written int64
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cmdStdin.Close()
		n, err := io.Copy(cmdStdin, r)
		if err != nil {
			copyErr = fmt.Errorf("copy failed: %w", err)
		}
		written += n
	}()
	if _, err := cmd.CombinedOutput(); err != nil {
		return written, fmt.Errorf("exec failed: %w", err)
	}
	wg.Wait()
	return written, copyErr
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
	var (
		buf     bytes.Buffer
		size    = 64 * 1 << 20
		written int64
		br      = bufio.NewReader(os.Stdin)
		started = time.Now()
	)
	for {
		b, err := br.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if _, err := buf.Write(b); err != nil {
			log.Fatal(err)
		}
		if buf.Len() > size {
			n, err := runImport(&buf, runFile, *outputFile)
			if err != nil {
				log.Fatal(err)
			}
			written += n
			fmt.Printf(strings.Repeat(" ", 40))
			fmt.Printf("\rwritten %s -- %s/s", ByteSize(written), ByteSize(int64(float64(written)/time.Since(started).Seconds())))
		}
	}
	n, err := runImport(&buf, runFile, *outputFile)
	if err != nil {
		log.Fatal(err)
	}
	written += n
	fmt.Printf(strings.Repeat(" ", 40))
	fmt.Printf("\rwritten %s -- %s/s", ByteSize(written), ByteSize(int64(float64(written)/time.Since(started).Seconds())))
	fmt.Println()
	log.Printf("db setup done")
}
