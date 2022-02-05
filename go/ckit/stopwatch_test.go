package ckit

import "testing"

func TestRandString(t *testing.T) {
	// Only test for correct length.
	var cases = []struct {
		n int
	}{
		{0}, {1}, {2}, {16},
	}
	for _, c := range cases {
		v := randString(c.n)
		if len(v) != c.n {
			t.Fatalf("got %v, want %v", len(v), c.n)
		}
	}
}

func TestStopWatch(t *testing.T) {
	// Test a single, limited interaction only.
	var (
		sw      StopWatch
		entries []*Entry
	)

	sw.Record("a")
	sw.Record("b")
	entries = sw.Entries()
	if len(entries) != 2 {
		t.Fatalf("got %v, want %v", len(entries), 2)
	}
	if entries[0].Message != "a" {
		t.Fatalf("got %v, want a", entries[0].Message)
	}
	if entries[1].Message != "b" {
		t.Fatalf("got %v, want b", entries[1].Message)
	}
	sw.Reset()
	entries = sw.Entries()
	if len(entries) != 0 {
		t.Fatalf("got %v, want ", len(entries))
	}
}
