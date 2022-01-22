package cache

import (
	"io/ioutil"
	"testing"
)

func TestCache(t *testing.T) {
	f, err := ioutil.TempFile(t.TempDir(), "cache")
	if err != nil {
		t.Fatalf("failed to create temporary test file: %v", err)
	}
	defer f.Close()
	t.Log(f.Name())
	cache, err := New(f.Name())
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	if size, err := cache.Len(); err != nil {
		t.Fatalf("failed to get number of entries: %v", err)
	} else {
		if size != 0 {
			t.Fatalf("want 0, got %v", size)
		}
	}
	if err := cache.Set("a", []byte("abc")); err != nil {
		t.Fatalf("failed to set value: %v", err)
	}
	if size, err := cache.Len(); err != nil {
		t.Fatalf("failed to get number of entries: %v", err)
	} else {
		if size != 1 {
			t.Fatalf("want 1, got %v", size)
		}
	}
	if v, err := cache.Get("a"); err != nil {
		t.Fatalf("failed to set value: %v", err)
	} else {
		if string(v) != "abc" {
			t.Fatalf("want abc, got %v", v)
		}
	}
	if err := cache.Close(); err != nil {
		t.Fatalf("failed to close db: %v", err)
	}
}
