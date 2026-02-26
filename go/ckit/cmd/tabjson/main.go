// TODO: value could be encoded, compressed
//
// Does compression help in this text domain? We get 0.4 of the original size,
// so should reduce db size in half, probably. Uses gzip and base64.
//
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA2NzU1Mw                2124  977   0.459981
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA1MjExMA                2049  961   0.469009
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8wMDAzLTA2NnguMzEuMTIuODQz  2228  1001  0.449282
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA1MDEzMA                2101  977   0.465017
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8wNzM1LTcwMjguNy4yLjIxNA    2216  993   0.448105
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDAzNTk3Mg                2468  1081  0.438006
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzOC9zai9sZXUvMjQwMTI5MA        1841  969   0.526344
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8wNzM1LTcwMjguOC4xLjgx      2223  981   0.441296
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDAzODE1NA                2226  981   0.440701
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA1NzM1Mw                2141  961   0.448856
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8vMDg4Mi03OTc0LjIuMi4xMzg   2397  1037  0.432624
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDAzODE1NQ                2386  1009  0.422883
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8wNzM1LTcwMjguOC4xLjcy      2064  917   0.444283
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA1MDM5Mg                2419  1057  0.436957
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8wNzM1LTcwMjguOC4xLjg4      1967  913   0.464159
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA0Njk4Mg                2352  1017  0.432398
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzOC9zai9sZXUvMjQwMTI5MQ        1841  973   0.528517
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDAzNzA0MA                2311  1025  0.443531
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8vMDg5NC00MTA1LjE1LjQuNDYy  2434  1053  0.432621
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA4OC8wMDMxLTkxNTUvMjQvNC80MTk   2371  1093  0.460987
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAxNy9zMDk1MDAxNzA5OTAwMDA5NA    2460  1177  0.478455
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA2NTI1OA                2184  1005  0.460165
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDAyMDU1NQ                1937  909   0.469282
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA3MzU4OQ                1979  921   0.465387
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA4Ni8yNjUwMjk                   2047  1065  0.520274
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA2ODU5OA                2244  989   0.440731
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy8vMDA5Ny03NDAzLjIxLjIuMTE2  2830  1161  0.410247
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA1MDQxNg                2065  981   0.475061
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA2NzAwMg                2357  1045  0.443360
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA1NzY5NQ                1972  917   0.465010
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA3NDY1Nw                2307  1017  0.440832
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA3NjU3Ng                2351  1061  0.451297
// ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTAzNy9oMDA2NzA5NA                2113  977   0.462376

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/miku/parallel"
	"github.com/segmentio/encoding/json"
)

var (
	Version string

	compressValue = flag.Bool("C", false, "compress value; gz+b64")
	compressTable = flag.Bool("T", false, "emit table showing possible savings through compression")
	showVersion   = flag.Bool("version", false, "show version and exit")
)

// Doc is the part of the document we are interested in. If this tool should be
// more generic, this would need to be generated on the fly.
type Doc struct {
	ID string `json:"id"`
}

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("tabjson %s\n", Version)
		os.Exit(0)
	}
	pp := parallel.NewProcessor(os.Stdin, os.Stdout, func(p []byte) ([]byte, error) {
		var doc Doc
		if err := json.Unmarshal(p, &doc); err != nil {
			return nil, err
		}
		if !*compressValue {
			return append([]byte(doc.ID+"\t"), p...), nil
		}
		var (
			buf = bytes.NewReader(p)
			dst bytes.Buffer
			enc = base64.NewEncoder(base64.StdEncoding, &dst)
			w   = gzip.NewWriter(enc)
		)
		if _, err := io.Copy(w, buf); err != nil {
			return nil, err
		}
		if err := w.Flush(); err != nil {
			return nil, err
		}
		if _, err := io.WriteString(&dst, "\n"); err != nil {
			return nil, err
		}
		if *compressTable {
			ratio := float64(len(dst.Bytes())) / float64(len(p))
			line := fmt.Sprintf("%v\t%d\t%d\t%f\n", doc.ID, len(p), len(dst.Bytes()), ratio)
			return []byte(line), nil
		} else {
			return append([]byte(doc.ID+"\t"), dst.Bytes()...), nil
		}
	})
	if err := pp.Run(); err != nil {
		log.Fatal(err)
	}
}
