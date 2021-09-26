package main

import (
	"log"
	"os"

	"github.com/miku/parallel"
	"github.com/segmentio/encoding/json"
)

type Doc struct {
	ID string `json:"id"`
}

func main() {
	pp := parallel.NewProcessor(os.Stdin, os.Stdout, func(p []byte) ([]byte, error) {
		var doc Doc
		if err := json.Unmarshal(p, &doc); err != nil {
			return nil, err
		}
		return append([]byte(doc.ID+"\t"), p...), nil
	})
	if err := pp.Run(); err != nil {
		log.Fatal(err)
	}
}
