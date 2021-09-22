package spindel

// Map is a generic lookup table. We use it together with sqlite3.
type Map struct {
	Key   string `db:"k"`
	Value string `db:"v"`
}
