package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"regexp"
	"time"

	"github.com/araddon/dateparse"
	"github.com/miku/labe/go/ckit/xflag"
)

var (
	dir                   = flag.String("d", ".", "directory to search")
	afterDate  xflag.Date = xflag.Date{}
	beforeDate xflag.Date = xflag.Date{}

	patDate = regexp.MustCompile(`20[0-9][0-9]-[01][0-9]-[0-3][0-9]`)
)

// MustParse will panic on an unparsable date string.
func MustParse(value string) time.Time {
	t, err := dateparse.ParseStrict(value)
	if err != nil {
		panic(err)
	}
	return t
}

func main() {
	flag.Var(&afterDate, "a", "list files newer than date")
	flag.Var(&beforeDate, "b", "list files older than date")
	flag.Parse()
	err := filepath.Walk(*dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		v := patDate.FindString(info.Name())
		if v == "" {
			return nil
		}
		fileDate, err := time.Parse("2006-01-02", v)
		if err != nil {
			return err
		}
		switch {
		case !afterDate.Time.IsZero():
			if fileDate.After(afterDate.Time) {
				fmt.Println(path)
			}
		case !beforeDate.Time.IsZero():
			if fileDate.Before(beforeDate.Time) {
				fmt.Println(path)
			}
		default:
			fmt.Println(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
