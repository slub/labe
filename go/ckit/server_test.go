package ckit

import (
	"log"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/segmentio/encoding/json"
)

func TestBatchedStrings(t *testing.T) {
	var cases = []struct {
		desc     string
		s        []string
		n        int
		expected [][]string
	}{
		{
			"empty slice", []string{}, 100, [][]string{
				[]string{},
			},
		},
		{
			"single batch", []string{"a", "b", "c"}, 100, [][]string{
				[]string{"a", "b", "c"},
			},
		},
		{
			"two batches", []string{"a", "b", "c"}, 2, [][]string{
				[]string{"a", "b"}, []string{"c"},
			},
		},
		{
			"full batches", []string{"a", "b", "c", "d"}, 2, [][]string{
				[]string{"a", "b"}, []string{"c", "d"},
			},
		},
	}
	for _, c := range cases {
		result := batchedStrings(c.s, c.n)
		if !reflect.DeepEqual(result, c.expected) {
			t.Fatalf("{%s] got %v, want %v", c.desc, result, c.expected)
		}
	}
}

func TestApplyInstitutionFilter(t *testing.T) {
	var cases = []struct {
		desc        string
		institution string
		resp        []byte // use serialized form for simplicity
		expected    []byte
	}{
		{
			desc:        "empty",
			institution: "",
			resp:        []byte("{}"),
			expected:    []byte("{}"),
		},
		{
			desc:        "empty",
			institution: "any",
			resp:        []byte("{}"),
			expected:    []byte(`{"extra": {"institution": "any"}}`),
		},
		{
			desc:        "one cited doc matches",
			institution: "a",
			resp: []byte(`
            {
              "cited": [
                {
                  "institution": ["a"]
                },
                {
                  "institution": ["b"]
                }
              ]
            }
			`),
			expected: []byte(`
			{
			  "cited": [
				{
				  "institution": ["a"]
				}
			  ],
			  "unmatched": {
				"cited": [
				  {
					"institution": ["b"]
				  }
				]
			  },
			  "extra": {
				"took": 0,
				"unmatched_citing_count": 0,
				"unmatched_cited_count": 1,
				"citing_count": 0,
				"cited_count": 1,
				"cached": false,
				"institution": "a"
			  }
			}
			`),
		},
	}
	for _, c := range cases {
		var (
			resp     Response
			expected Response
		)
		if err := json.Unmarshal(c.resp, &resp); err != nil {
			t.Fatalf("could not unmarshal test response: %v", err)
		}
		if err := json.Unmarshal(c.expected, &expected); err != nil {
			t.Fatalf("could not unmarshal test response: %v", err)
		}
		resp.applyInstitutionFilter(c.institution)
		if string(mustMarshal(resp)) != string(mustMarshal(expected)) {
			log.Printf(string(mustMarshal(resp)))
			log.Printf(string(mustMarshal(expected)))
			t.Fatalf("[%s] %v", c.desc, cmp.Diff(resp, expected))
		}
	}
}

func TestServerBasic(t *testing.T) {
	a, err := OpenDatabase("testdata/id_doi.db")
	if err != nil {
		t.Fatalf("test data: %v", err)
	}
	b, err := OpenDatabase("testdata/doi_doi.db")
	if err != nil {
		t.Fatalf("test data: %v", err)
	}
	g := &FetchGroup{}
	if err := g.FromFiles("testdata/id_metadata.db"); err != nil {
		t.Fatalf("test data: %v", err)
	}
	srv := &Server{
		IdentifierDatabase: a,
		OciDatabase:        b,
		IndexData:          g,
		Router:             mux.NewRouter(),
	}
	rr := httptest.NewRecorder()
	t.Logf("setup %v [%v]", srv, rr)
	// TODO: execute handlers
}

func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
