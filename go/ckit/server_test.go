package ckit

import (
	"reflect"
	"testing"
)

func TestBatchedStrings(t *testing.T) {
	var cases = []struct {
		s        []string
		n        int
		expected [][]string
	}{
		{
			[]string{}, 100, [][]string{
				[]string{},
			},
		},
		{
			[]string{"a", "b", "c"}, 100, [][]string{
				[]string{"a", "b", "c"},
			},
		},
		{
			[]string{"a", "b", "c"}, 2, [][]string{
				[]string{"a", "b"}, []string{"c"},
			},
		},
	}
	for _, c := range cases {
		result := batchedStrings(c.s, c.n)
		if !reflect.DeepEqual(result, c.expected) {
			t.Fatalf("got %v, want %v", result, c.expected)
		}
	}
}
