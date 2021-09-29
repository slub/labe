package spindel

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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
// table. Example log output (via stopwatch.LogTable()).
//
// 2021/09/29 17:22:40 timings for hTHc
//
// > hTHc    0    0s             0.00    started query for: ai-49-aHR0cDovL2R4LmRvaS5vcmcvMTAuMTA5OC9yc3BhLjE5OTguMDE2NA
// > hTHc    1    383.502µs      0.01    found doi for id: 10.1098/rspa.1998.0164
// > hTHc    2    478.925µs      0.01    found 8 citing items
// > hTHc    3    14.477959ms    0.32    found 456 cited items
// > hTHc    4    10.778333ms    0.24    mapped 464 dois back to ids
// > hTHc    5    396.743µs      0.01    recorded unmatched ids
// > hTHc    6    13.375994ms    0.29    fetched 302 blob from index data store
// > hTHc    7    5.782197ms     0.13    encoded JSON
// > hTHc    -    -              -       -
// > hTHc    S    45.673653ms    1.0     total
//
// By default a stopwatch is disabled, which means all functions will be noops,
// use SetEnabled to toggle mode.
type StopWatch struct {
	sync.Mutex
	id      string
	entries []*Entry
	enabled bool
}

func (s *StopWatch) SetEnabled(enabled bool) {
	s.Lock()
	defer s.Unlock()
	s.enabled = enabled
}

// Record records a message.
func (s *StopWatch) Record(msg string) {
	if !s.enabled {
		return
	}
	s.Recordf(msg)
}

// Recordf records a message.
func (s *StopWatch) Recordf(msg string, vs ...interface{}) {
	if !s.enabled {
		return
	}
	s.Lock()
	defer s.Unlock()
	if s.id == "" {
		s.id = randString(4)
	}
	s.entries = append(s.entries, &Entry{
		T:       time.Now(),
		Message: fmt.Sprintf(msg, vs...),
	})
}

// LogTable write a table using standard library log facilities.
func (s *StopWatch) LogTable() {
	if !s.enabled {
		return
	}
	log.Printf("timings for %s\n\n"+s.Table()+"\n", s.id)
}

// Table format the timings as table.
func (s *StopWatch) Table() string {
	if !s.enabled || len(s.entries) == 0 {
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
	fmt.Fprintf(w, "> %s\tS\t%s\t1.0\ttotal\n", s.id, total)
	w.Flush()
	return buf.String()
}

// Reset resets the stopwatch.
func (s *StopWatch) Reset() {
	if !s.enabled {
		return
	}
	s.entries = nil
}
