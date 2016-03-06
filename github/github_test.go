package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListStargazers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.github.v3.star+json" {
			t.Errorf("Expected 'Accept: application/vnd.github.v3.star+json', but it was '%s'", r.Header.Get("Accept"))
		}
		fmt.Fprintln(w, "[{}, {}, {}]")
	}))
	defer ts.Close()

	gh := NewClientWithBaseUrl("tester", "secretlel", ts.URL)
	allStargazers, err := gh.ListStargazers("angular", "angular")
	if err != nil {
		t.Error(err)
	}
	if len(allStargazers) != 3 {
		t.Errorf("Expected len(allStargazers) to be 3, but it was %d\n", len(allStargazers))
	}
}
