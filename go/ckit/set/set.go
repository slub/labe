package set

import (
	"sort"
	"strings"
)

// Set implements basic string set operations, not thread-safe.
type Set map[string]struct{}

// New creates a new set.
func New() Set {
	var s = make(Set)
	return s
}

// FromSlice initializes a set from a slice.
func FromSlice(vs []string) Set {
	s := New()
	for _, v := range vs {
		s.Add(v)
	}
	return s
}

// Clear removes all elements.
func (s Set) Clear() {
	for k := range s {
		delete(s, k)
	}
}

// Add adds an element.
func (s Set) Add(v string) Set {
	s[v] = struct{}{}
	return s
}

// Len returns number of elements in set.
func (s Set) Len() int {
	return len(s)
}

// IsEmpty returns if set has zero elements.
func (s Set) IsEmpty() bool {
	return s.Len() == 0
}

// Equals returns true, if sets contain the same elements.
func (s Set) Equals(t Set) bool {
	for k := range s {
		if !t.Contains(k) {
			return false
		}
	}
	return s.Len() == t.Len()
}

// Contains returns membership status.
func (s Set) Contains(v string) bool {
	_, ok := (s)[v]
	return ok
}

// Intersection returns a new set containing all elements found in both sets.
func (s Set) Intersection(t Set) Set {
	u := New()
	for k := range s {
		if t.Contains(k) {
			u.Add(k)
		}
	}
	return u
}

// Union returns the union of two sets.
func (s Set) Union(t Set) Set {
	u := New()
	for k := range s {
		u.Add(k)
	}
	for k := range t {
		u.Add(k)
	}
	return u
}

func (s Set) Difference(t Set) Set {
	u := New()
	for k := range s {
		if !t.Contains(k) {
			u.Add(k)
		}
	}
	return u
}

// Slice returns all elements as a slice.
func (s Set) Slice() (result []string) {
	for k := range s {
		result = append(result, k)
	}
	return
}

// Sorted returns all elements as a slice, sorted.
func (s Set) Sorted() (result []string) {
	for k := range s {
		result = append(result, k)
	}
	sort.Strings(result)
	return
}

// TopK returns at most k sorted elements.
func (s Set) TopK(k int) Set {
	var top []string
	for i, v := range s.Sorted() {
		if i < k {
			top = append(top, v)
		}
	}
	return FromSlice(top)
}

// Product returns a slice of pairs, representing the cartesian product of two sets.
func (s Set) Product(t Set) (result [][]string) {
	for k := range s {
		for l := range t {
			result = append(result, []string{k, l})
		}
	}
	return
}

// Jaccard returns the jaccard index of sets s and t, between 0 and 1, where 1
// means equality.
func (s Set) Jaccard(t Set) float64 {
	if s.IsEmpty() && t.IsEmpty() {
		return 1
	}
	if u := s.Union(t); u.IsEmpty() {
		return 0
	} else {
		return float64(s.Intersection(t).Len()) / float64(u.Len())
	}
}

// Join joins elements from a set with given separator.
func (s Set) Join(sep string) string {
	return strings.Join(s.Slice(), sep)
}

// Max returns the size of the largest set.
func Max(ss ...Set) (max int) {
	for _, s := range ss {
		if s.Len() > max {
			max = s.Len()
		}
	}
	return
}

// Min returns the size of the smallest set.
func Min(ss ...Set) (min int) {
	min = 2 << 30
	for _, s := range ss {
		if s.Len() < min {
			min = s.Len()
		}
	}
	return
}

// Filter returns a set containing all elements, which satisfy a given predicate.
func Filter(s Set, f func(string) bool) Set {
	t := New()
	for v := range s {
		if f(v) {
			t.Add(v)
		}
	}
	return t
}
