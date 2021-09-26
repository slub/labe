package spindel

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"
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
