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
			desc:        "empty with institution",
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
		if err := resp.applyInstitutionFilter(c.institution); err != nil {
			t.Fatalf("[%s] unexpected error: %v", c.desc, err)
		}
		if string(mustMarshal(resp)) != string(mustMarshal(expected)) {
			log.Println(string(mustMarshal(resp)))
			log.Println(string(mustMarshal(expected)))
			t.Fatalf("[%s] %v", c.desc, cmp.Diff(resp, expected))
		}
	}
}

func TestApplyInstitutionFilterMalformedJSON(t *testing.T) {
	resp := Response{
		Citing: []json.RawMessage{[]byte(`not valid json`)},
	}
	err := resp.applyInstitutionFilter("DE-14")
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestValidISIL(t *testing.T) {
	valid := []string{"DE-14", "DE-Gla1", "DE-Ch1", "DE-D161", "DE-Brt1", "DE-L229", "DE-Bn3", "US-NNC"}
	for _, v := range valid {
		if !validISIL(v) {
			t.Errorf("expected %q to be valid ISIL", v)
		}
	}
	invalid := []string{"", "DE", "14", "DE-", "-14", `DE-14"; DROP TABLE`, "<script>alert(1)</script>", "DE 14", "../../etc/passwd"}
	for _, v := range invalid {
		if validISIL(v) {
			t.Errorf("expected %q to be invalid ISIL", v)
		}
	}
}

func TestSanitizeHost(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"localhost:8000", "localhost:8000"},
		{"example.com", "example.com"},
		{"10.0.0.1:8080", "10.0.0.1:8080"},
		{"evil.com\r\nX-Injected: true", "localhost"},
		{`"><script>alert(1)</script>`, "localhost"},
		{"", "localhost"},
	}
	for _, c := range cases {
		got := sanitizeHost(c.input)
		if got != c.expected {
			t.Errorf("sanitizeHost(%q) = %q, want %q", c.input, got, c.expected)
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
