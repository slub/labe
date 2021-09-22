package spindel

import (
	"encoding/json"
)

// Map is a generic lookup table. We use it together with sqlite3.
type Map struct {
	Key   string `db:"k"`
	Value string `db:"v"`
}

// Response contains a subset of index data fused with citation data.
type Response struct {
	// The local identifier.
	Identifier string `json:"id"`
	// The DOI for the local identifier.
	DOI string `json:"doi"`
	// We want to safe time not doing any serialization, if we do not need it.
	Citing []json.RawMessage `json:"citing"`
	Cited  []json.RawMessage `json:"cited"`
	// Some extra information for now, may not need these.
	Extra struct {
		Took        float64 `json:"took"`
		CitingCount int     `json:"citing_count"`
		CitedCount  int     `json:"cited_count"`
	} `json:"extra"`
}
