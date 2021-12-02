package ckit

import (
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

// WithReadOnly opens a sqlite database in read-only mode.
func WithReadOnly(path string) string {
	return fmt.Sprintf("file:%s?mode=ro", path)
}

// RunScript runs a script on an sqlite3 database.
func RunScript(path, script, message string) error {
	cmd := exec.Command("sqlite3", path)
	cmd.Stdin = strings.NewReader(script)
	b, err := cmd.CombinedOutput()
	if err == nil {
		log.Printf("[ok] %s · %s", message, path)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", string(b))
	}
	return err
}

// RunImport reads data to be imported (e.g. two column TSV) from reader into a
// given database. Before importing, read commands from a given init file.
func RunImport(r io.Reader, initFile, outputFile string) (int64, error) {
	// TODO: Unify parameter order, e.g. put outputFile first.
	cmd := exec.Command("sqlite3", "--init", initFile, outputFile)
	cmdStdin, err := cmd.StdinPipe()
	if err != nil {
		return 0, fmt.Errorf("stdin pipe: %w", err)
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
			return
		}
		written += n
	}()
	if b, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintln(os.Stderr, string(b))
		return written, fmt.Errorf("exec failed: %w", err)
	}
	wg.Wait()
	return written, copyErr
}

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

// Flushf for messages that should stay on a single line.
func Flushf(s string, vs ...interface{}) {
	t := time.Now().Format("2006/01/02 15:04:05")
	msg := fmt.Sprintf("\r"+t+" [io] "+s, vs...)
	fmt.Printf("\r" + strings.Repeat(" ", len(msg)+1))
	fmt.Printf(msg)
}

// HumanSpeed returns a human readable throughput number, e.g. 10MB/s.
func HumanSpeed(bytesWritten int64, elapsedSeconds float64) string {
	speed := float64(bytesWritten) / elapsedSeconds
	return fmt.Sprintf("%s/s", ByteSize(int(speed)))
}

// ByteSize returns a human-readable byte string of the form 10M, 12.5K, and so
// forth.  The following units are available: E: Exabyte, P: Petabyte, T:
// Terabyte, G: Gigabyte, M: Megabyte, K: Kilobyte, B: Byte, The unit that
// results in the smallest number greater than or equal to 1 is always chosen.
func ByteSize(bytes int) string {
	const (
		BYTE = 1 << (10 * iota)
		KB
		MB
		GB
		TB
		PB
		EB
	)
	var (
		u      = ""
		v      = float64(bytes)
		result string
	)
	switch {
	case bytes >= EB:
		u = "E"
		v = v / EB
	case bytes >= PB:
		u = "P"
		v = v / PB
	case bytes >= TB:
		u = "T"
		v = v / TB
	case bytes >= GB:
		u = "G"
		v = v / GB
	case bytes >= MB:
		u = "M"
		v = v / MB
	case bytes >= KB:
		u = "K"
		v = v / KB
	case bytes >= BYTE:
		u = "B"
	case bytes == 0:
		return "0B"
	}
	result = strconv.FormatFloat(v, 'f', 1, 64)
	result = strings.TrimSuffix(result, ".0")
	return result + u
}
