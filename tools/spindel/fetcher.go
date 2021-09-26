package spindel

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	// ErrBlobNotFound can be used for unfetchable blobs.
	ErrBlobNotFound = errors.New("blob not found")
	client          = http.Client{
		Timeout: 5 * time.Second,
	}
)

type Pinger interface {
	Ping() error
}

// Fetcher fetches a blob of data for a given identifier.
type Fetcher interface {
	Fetch(id string) ([]byte, error)
}

// BlobServer implements access to a running microblob instance.
type BlobServer struct {
	BaseURL string
}

// Ping is a healthcheck.
func (bs *BlobServer) Ping() error {
	resp, err := client.Get(bs.BaseURL)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("blobserver: expected 200 OK, got: %v", resp.Status)
	}
	return nil
}

// Fetch constructs a URL from a template and retrieves the blob.
func (bs *BlobServer) Fetch(id string) ([]byte, error) {
	u, err := url.Parse(bs.BaseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, id)
	return fetchURL(u.String())
}

func fetchURL(u string) ([]byte, error) {
	resp, err := client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, ErrBlobNotFound
	}
	return ioutil.ReadAll(resp.Body)
}

// SqliteBlob serves index documents from sqlite.
type SqliteBlob struct {
	Path string
	mu   sync.Mutex
	db   *sqlx.DB
}

func (b *SqliteBlob) init() (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.db, err = sqlx.Open("sqlite3", b.Path)
	return
}

// Fetch document.
func (b *SqliteBlob) Fetch(id string) (p []byte, err error) {
	if b.db == nil {
		if err := b.init(); err != nil {
			return nil, err
		}
	}
	var s string
	if err := b.db.Get(&s, "SELECT v FROM map WHERE k = ?", id); err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func (b *SqliteBlob) FetchSet(ids []string) (p [][]byte, err error) {
	query, args, err := sqlx.In("SELECT * FROM map WHERE v IN (?)", ids)
	if err != nil {
		return nil, err
	}
	query = b.db.Rebind(query)
	if err := b.db.Select(&p, query, args...); err != nil {
		return nil, err
	}
	return p, nil
}

// Ping pings the database.
func (b *SqliteBlob) Ping() error {
	if b.db == nil {
		if err := b.init(); err != nil {
			return err
		}
	}
	return b.db.Ping()
}
