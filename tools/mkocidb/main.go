// mkocidb takes two columns and turns it into an indexed sqlite3 database.
// Example: Help to query an OCI data dump offline.
//
// $ time mkocidb < sample-m.tsv
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
// ...
//
// Note that the machine this runs on probably needs at least 4K * cache_size
// free memory; not sure how much performance varies for these ops, if we
// reduce.
//
// Indexing dstat on nvme.
//
// --total-cpu-usage-- -dsk/total- -net/total- ---paging-- ---system--
// usr sys idl wai stl| read  writ| recv  send|  in   out | int   csw
//  14   2  84   0   0|  23M   63M|   0     0 |  92k  241k|7859    44k
//  14   3  81   2   0| 561M  888k| 252k 5546B|   0     0 |5711    10k
//  17   1  83   0   0|8192B    0 |2200B   94B|   0     0 |2831  8652
//  17   6  75   1   0| 115M  793M|   0   359B|  28k    0 |4611    14k
//  16   4  77   2   0| 807M    0 | 313k 7332B|  28k    0 |8005    15k
//  18   3  79   1   0| 337M    0 |   0     0 |   0     0 |5327    12k
//  18   2  79   0   0|  16k 1000k|  42B  359B|   0     0 |3920    10k
//  15   6  76   3   0| 490M  793M| 257k 5640B|   0    48k|7172    15k
//  17   6  75   2   0| 768M  152k|2242B  188B|4096B   76k|8282    17k
//  20   1  79   0   0|  24k  528k|   0   359B|  12k    0 |3379    11k
//  16   3  81   0   0|8192B    0 | 264k 6392B|4096B    0 |2995  9128
//  12   5  81   3   0| 768M  800k|2313B  188B|   0     0 |6880    10k
//  15   2  81   1   0| 481M    0 |  42B  359B|   0     0 |5621  9342
//  18   1  81   0   0|8192B    0 | 321k 6674B|   0     0 |2988  9067
//  14   6  79   2   0| 259M  793M|   0     0 |   0     0 |5069  9360
//  11   4  82   3   0| 882M 8192B|   0   359B|   0  4096B|7338    10k
//  17   1  81   0   0| 113M  816k| 270k 5686B|   0     0 |4259  8830
//  16   4  78   2   0|  56k 1550M|2373B   94B|   0     0 |4868  9010
//  11   4  81   3   0| 867M  253M|  80k 2615B|   0    36k|7978    10k
// ...
// More on sqlite3 pragmas:
//
// https://www.sqlite.org/pragma.html
// https://stackoverflow.com/q/1983979/89391
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
	outputFile = flag.String("o", "data.db", "output filename")
	bufferSize = flag.Int("B", 64*1<<20, "buffer size")
	indexMode  = flag.Int("I", 3, "index mode: 0=none, 1=k, 2=v, 3=kv")

	initSQL = `
CREATE TABLE IF NOT EXISTS map
(
	k TEXT,
	v TEXT
);
`
	keyIndexSQL = `
PRAGMA journal_mode = OFF;
PRAGMA synchronous = 0;
PRAGMA cache_size = 1000000;
PRAGMA locking_mode = EXCLUSIVE;
CREATE INDEX IF NOT EXISTS idx_k ON map(k);
`
	valueIndexSQL = `
PRAGMA journal_mode = OFF;
PRAGMA synchronous = 0;
PRAGMA cache_size = 1000000;
PRAGMA locking_mode = EXCLUSIVE;

CREATE INDEX IF NOT EXISTS idx_v ON map(v);
`
	importSQL = `
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
	var (
		unit  = ""
		value = float64(bytes)
	)
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

// Flushf for messages that should stay on a single line.
func Flushf(s string, vs ...interface{}) {
	msg := fmt.Sprintf("\r"+s, vs...)
	fmt.Printf("\r" + strings.Repeat(" ", len(msg)+1))
	fmt.Printf(msg)
}

// HumanSpeed returns a human readable throughput number, e.g. 10MB/s,
// 12.3kB/s, etc.
func HumanSpeed(bytesWritten int64, elapsedSeconds float64) string {
	speed := float64(bytesWritten) / elapsedSeconds
	return fmt.Sprintf("%s/s", ByteSize(int64(speed)))
}

func main() {
	flag.Parse()
	if _, err := os.Stat(*outputFile); os.IsNotExist(err) {
		if err := runScript(*outputFile, initSQL, "initialized database"); err != nil {
			log.Fatal(err)
		}
	}
	runFile, err := TempFileReader(strings.NewReader(importSQL))
	if err != nil {
		log.Fatal(err)
	}
	var (
		br      = bufio.NewReader(os.Stdin)
		buf     bytes.Buffer
		written int64
		started = time.Now()
		elapsed float64
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
		if buf.Len() >= *bufferSize {
			n, err := runImport(&buf, runFile, *outputFile)
			if err != nil {
				log.Fatal(err)
			}
			written += n
			elapsed = time.Since(started).Seconds()
			Flushf("written %s -- %s", ByteSize(written), HumanSpeed(written, elapsed))
		}
	}
	n, err := runImport(&buf, runFile, *outputFile)
	if err != nil {
		log.Fatal(err)
	}
	written += n
	elapsed = time.Since(started).Seconds()
	Flushf("written %s -- %s", ByteSize(written), HumanSpeed(written, elapsed))
	fmt.Println()
	log.Printf("import done")
	log.Printf("creating index")
	var indexScripts []string
	switch *indexMode {
	case 0:
	case 1:
		indexScripts = append(indexScripts, keyIndexSQL)
	case 2:
		indexScripts = append(indexScripts, valueIndexSQL)
	case 3:
		indexScripts = append(indexScripts, keyIndexSQL)
		indexScripts = append(indexScripts, valueIndexSQL)
	}
	for i, script := range indexScripts {
		msg := fmt.Sprintf("%d/%d created index", i+1, len(indexScripts))
		if err := runScript(*outputFile, script, msg); err != nil {
			log.Fatal(err)
		}
	}
}
