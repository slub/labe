package spindel

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	// ErrBlobNotFound can be used for unfetchable blobs.
	ErrBlobNotFound   = errors.New("blob not found")
	ErrBackendsFailed = errors.New("all backends failed")
	client            = http.Client{
		// We use the client to fetch data from backends. Often, we request one
		// item after another and there will be a 5 second timeout per request,
		// not for the whole operation.
		Timeout: 5 * time.Second,
	}
)

// Pinger allows to perform a simple health check.
type Pinger interface {
	Ping() error
}

// Fetcher fetches one or more blobs given their identifiers.
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

// fetchURL is a helper to read the response body for given link.
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

// SqliteBlob serves index documents from sqlite database.
type SqliteBlob struct {
	DB *sqlx.DB
}

// Fetch document.
func (b *SqliteBlob) Fetch(id string) (p []byte, err error) {
	var s string
	if err := b.DB.Get(&s, "SELECT v FROM map WHERE k = ?", id); err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// Ping pings the database.
func (b *SqliteBlob) Ping() error {
	return b.DB.Ping()
}

// SolrBlob implements access to a running microblob instance. The base url
// would be something like http://localhost/solr/biblio (e.g. without the
// select part of the path).
type SolrBlob struct {
	BaseURL string
}

// Ping is a healthcheck. Solr typically responds with 404 on the URL without
// any handler; http://localhost:8085/solr/biblio/admin/ping
func (b *SolrBlob) Ping() error {
	baseURL := strings.TrimRight(b.BaseURL, "/")
	link := fmt.Sprintf("%s/admin/ping", baseURL)
	resp, err := client.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("solrblob: expected 200 OK, got: %v", resp.Status)
	}
	return nil
}

// Fetch constructs a URL from a template and retrieves the blob.
func (b *SolrBlob) Fetch(id string) ([]byte, error) {
	baseURL := strings.TrimRight(b.BaseURL, "/")
	link := fmt.Sprintf("%s/select?q=id:%q&wt=json", baseURL, id)
	return fetchURL(link)
}

// FetchGroup allows to run a index data fetch operation in a cascade over a
// couple of backends.
type FetchGroup struct {
	Backends []Fetcher
}

// Ping is a healthcheck. Solr typically responds with 404 on the URL without
// any handler; http://localhost:8085/solr/biblio/admin/ping
func (g *FetchGroup) Ping() error {
	for _, v := range g.Backends {
		w, ok := v.(Pinger)
		if !ok {
			continue
		}
		if err := w.Ping(); err != nil {
			return err
		}
	}
	return nil
}

// Fetch constructs a URL from a template and retrieves the blob.
func (g *FetchGroup) Fetch(id string) ([]byte, error) {
	for _, v := range g.Backends {
		if p, err := v.Fetch(id); err != nil {
			return nil, err
		} else {
			return p, nil
		}
	}
	return nil, ErrBackendsFailed
}
