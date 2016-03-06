package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListStargazers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("Accept"), "application/vnd.github.v3.star+json")
		fmt.Fprintln(w, "[{}, {}, {}]")
	}))
	defer ts.Close()

	gh := NewClientWithBaseUrl("tester", "secretlel", ts.URL)
	allStargazers, err := gh.ListStargazers("angular", "angular")
	assert.NoError(t, err)
	assert.Equal(t, len(allStargazers), 3)
}
