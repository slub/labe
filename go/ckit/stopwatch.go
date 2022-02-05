package ckit

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// Entry is a stopwatch entry.
type Entry struct {
	T       time.Time
	Message string
}

// StopWatch allows to record events over time and render them in a pretty
// table; thread-safe. Example log output (via stopwatch.LogTable()).
//
// 2021/09/29 17:22:40 timings for XVlB
//
// > XVlB    0    0s              0.00    started query for: ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTIxMC9qYy4yMDExLTAzODU
// > XVlB    1    134.532µs       0.00    found doi for id: 10.1210/jc.2011-0385
// > XVlB    2    67.918529ms     0.24    found 0 outbound and 4628 inbound edges
// > XVlB    3    32.293723ms     0.12    mapped 4628 dois back to ids
// > XVlB    4    3.358704ms      0.01    recorded unmatched ids
// > XVlB    5    68.636671ms     0.25    fetched 2567 blob from index data store
// > XVlB    6    105.771005ms    0.38    encoded JSON
// > XVlB    -    -               -       -
// > XVlB    S    278.113164ms    1.00    total
//
type StopWatch struct {
	sync.Mutex
	id       string
	entries  []*Entry
	disabled bool
}

// SetEnabled enables or disables the stopwatch. If disabled, any call will be
// a noop.
func (s *StopWatch) SetEnabled(enabled bool) {
	s.Lock()
	defer s.Unlock()
	s.disabled = !enabled
}

// Record records a message.
func (s *StopWatch) Record(msg string) {
	s.Recordf(msg)
}

// Recordf records a message.
func (s *StopWatch) Recordf(msg string, vs ...interface{}) {
	if s.disabled {
		return
	}
	s.Lock()
	defer s.Unlock()
	if s.id == "" {
		s.id = randString(8)
	}
	s.entries = append(s.entries, &Entry{
		T:       time.Now(),
		Message: fmt.Sprintf(msg, vs...),
	})
}

// Reset resets the stopwatch.
func (s *StopWatch) Reset() {
	if s.disabled {
		return
	}
	s.entries = nil
}

// Elapsed returns the total elapsed time.
func (s *StopWatch) Elapsed() time.Duration {
	if len(s.entries) == 0 {
		return 0
	}
	return time.Now().Sub(s.entries[0].T)
}

// Entries returns the accumulated messages for this stopwatch.
func (s *StopWatch) Entries() []*Entry {
	return s.entries
}

// LogTable write a table using standard library log facilities.
func (s *StopWatch) LogTable() {
	if s.disabled {
		return
	}
	log.Printf("timings for %s\n\n"+s.Table()+"\n", s.id)
}

// Table format the timings as table.
func (s *StopWatch) Table() string {
	s.Lock()
	defer s.Unlock()
	if s.disabled || len(s.entries) == 0 {
		return ""
	}
	var (
		buf   strings.Builder
		w     = tabwriter.NewWriter(&buf, 0, 0, 4, ' ', 0)
		total = s.entries[len(s.entries)-1].T.Sub(s.entries[0].T)
	)
	for i, entry := range s.entries {
		var (
			diff time.Duration
			pct  float64
		)
		if i > 0 {
			diff = s.entries[i].T.Sub(s.entries[i-1].T)
			pct = float64(diff) / float64(total)
		}
		fmt.Fprintf(w, "> %s\t%d\t%s\t%0.2f\t%s\n", s.id, i, diff, pct, entry.Message)
	}
	fmt.Fprintf(w, "> %s\t-\t-\t-\t-\n", s.id)
	fmt.Fprintf(w, "> %s\tS\t%s\t%0.2f\ttotal\n", s.id, total, 1.0)
	w.Flush()
	return buf.String()
}
