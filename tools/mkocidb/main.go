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
// A chucked insert seems to mitigate this issue.
//
// $ time ./mkocidb < sample-m.tsv
// 2021/09/20 13:13:55 [ok] initialized database at data.db
// written 5G -- 6.6M/s
//
// real    12m46.585s
// user    11m17.193s
// sys     3m2.369s
//
// sqlite> select count(*) from map;
// 99057347
//
// sqlite> select v from map where k = '10.1182/blood.v91.5.1533.1533_1533_1541';
// 10.1126/science.142.3595.1069
// 10.1016/0006-291x(68)90266-0
// 10.1001/jama.1966.03100120106029
// 10.1002/1097-0142(19821101)50:9<1683::aid-cncr2820500904>3.0.co;2-x
// 10.1002/1097-0142(19860215)57:4<718::aid-cncr2820570406>3.0.co;2-p
// 10.1056/nejm199004123221504
// 10.1093/jnci/80.17.1412
// 10.1016/0163-7258(85)90021-x
// 10.1016/0006-2952(85)90543-x
// 10.1200/jco.1993.11.8.1523
// 10.1126/science.7973634
// 10.1073/pnas.91.14.6674
// 10.1111/j.1749-6632.1985.tb17191.x
// 10.1182/blood.v70.6.1824.1824
// 10.1172/jci113437
// 10.1002/ajh.2830320206
// 10.1182/blood.v69.1.109.109
// 10.1182/blood.v79.10.2555.2555
// 10.1002/1097-0142(19800801)46:3<455::aid-cncr2820460306>3.0.co;2-n
// 10.1007/bf00262739
// 10.1200/jco.1995.13.5.1089
// 10.1007/s002800050256
// 10.1007/bf01117450
// 10.1002/jps.2600830726
// 10.1023/a:1018953810705
// 10.1007/s002800050569
// 10.1016/0014-2964(78)90253-0
// 10.1016/s0305-7372(79)80008-0
// 10.2165/00003088-198916040-00002
// 10.1200/jco.1994.12.9.1754
// 10.1200/jco.1990.8.3.527
// 10.1182/blood.v80.9.2425.2425
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
`
	indexSQL = `
PRAGMA journal_mode = OFF;
PRAGMA synchronous = 0;
PRAGMA cache_size = 1000000;
PRAGMA locking_mode = EXCLUSIVE;
PRAGMA temp_store = MEMORY;

-- https://stackoverflow.com/q/1983979/89391
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

// runScript initializes the database, if necessary
func runScript(path, script, message string) error {
	cmd := exec.Command("sqlite3", path)
	cmd.Stdin = strings.NewReader(script)
	err := cmd.Run()
	if err == nil {
		log.Printf("[ok] %s -- %s", message, path)
	}
	return err
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
	if _, err := os.Stat(*outputFile); os.IsNotExist(err) {
		if err := runScript(*outputFile, initSQL, "initialized database"); err != nil {
			log.Fatal(err)
		}
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
	log.Printf("creating index")
	if err := runScript(*outputFile, indexSQL, "created index"); err != nil {
		log.Fatal(err)
	}
}
