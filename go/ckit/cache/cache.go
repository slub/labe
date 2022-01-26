// Package cache implements caching helpers, e.g. an sqlite3 based cache.
// Caching is important for the most cited items, these can take seconds to
// assemble; so we cache the serialized JSON in sqlite3 and serve subsequent
// requests from there.
//
// Without cache:
//
//   | Rank    | Links | T    |
//   |---------+-------+------|
//   | ~5000   |  2999 | 2.8s |
//   | ~10000  |  2108 | 3.5s |
//   | ~50000  |   937 | 1.2s |
//   | ~100000 |   659 | 0.8s |
//   | ~150000 |   538 | 0.6s |
//
// A data point: The hundert most expensive ids take 175s to request (in
// parallel). After caching, this time reduces to 2.78s. Individual requests
// from cache are in the 1-10ms range.
package cache

import (
	"errors"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/slub/labe/go/ckit/tabutils"

	_ "github.com/mattn/go-sqlite3"
)

var ErrCacheMiss = errors.New("cache miss")

// Cache is a minimalistic cache based on sqlite. In the future, values could
// be transparently compressed as well.
type Cache struct {
	Path string

	sync.Mutex
	db *sqlx.DB
}

func New(path string) (*Cache, error) {
	conn, err := sqlx.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	c := &Cache{Path: path, db: conn}
	if err := c.init(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Cache) init() error {
	s := `
PRAGMA journal_mode = OFF;
PRAGMA synchronous = 0;
PRAGMA locking_mode = EXCLUSIVE;
PRAGMA temp_store = MEMORY;
CREATE TABLE IF NOT EXISTS map (k TEXT, v TEXT);
CREATE INDEX IF NOT EXISTS idx_k ON map(k);
	`
	return tabutils.RunScript(c.Path, s, "initialized database")
}

// Close closes the underlying database.
func (c *Cache) Close() error {
	return c.db.Close()
}

// Flush empties the cache.
func (c *Cache) Flush() error {
	c.Lock()
	defer c.Unlock()
	_, err := c.db.Exec(`DELETE FROM map`)
	return err
}

// ItemCount returns the number of entries in the cache.
func (c *Cache) ItemCount() (int, error) {
	row := c.db.QueryRow(`SELECT count(k) FROM map`)
	var v int
	if err := row.Scan(&v); err != nil {
		return 0, err
	}
	return v, nil
}

// Set key value pair.
func (c *Cache) Set(key string, value []byte) error {
	c.Lock()
	defer c.Unlock()
	s := `INSERT into map (k, v) VALUES (?, ?)`
	_, err := c.db.Exec(s, key, value)
	return err
}

// Get value for a key.
func (c *Cache) Get(key string) ([]byte, error) {
	var (
		row = c.db.QueryRow(`SELECT v FROM map WHERE k = ?`, key)
		v   string
	)
	if err := row.Scan(&v); err != nil {
		return nil, err
	}
	if v == "" {
		return nil, ErrCacheMiss
	}
	return []byte(v), nil
}
