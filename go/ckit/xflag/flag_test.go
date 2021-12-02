package xflag

import "testing"

func TestArray(t *testing.T) {
	var a Array
	a.Set("a")
	a.Set("b")
	if len(a) != 2 {
		t.Fatalf("got %d, want 2", len(a))
	}
	if a[0] != "a" || a[1] != "b" {
		t.Fatalf("got %s, want [a b]", a)
	}
}
