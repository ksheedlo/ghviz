package github

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func pathAndQueryOnly(t *testing.T, rawurl string) string {
	urlObj, err := url.Parse(rawurl)
	assert.NoError(t, err)
	if urlObj.RawQuery != "" {
		return fmt.Sprintf("%s?%s", urlObj.Path, urlObj.RawQuery)
	}
	return urlObj.Path
}

func dummyLogger(t *testing.T) *log.Logger {
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0777)
	assert.NoError(t, err)
	return log.New(devnull, "", 0)
}

func TestListStargazers(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/vnd.github.v3.star+json", r.Header.Get("Accept"))
		assert.Equal(t,
			"/repos/angular/angular/stargazers?per_page=100",
			pathAndQueryOnly(t, r.URL.String()),
		)
		fmt.Fprintln(w, "[{}, {}, {}]")
	}))
	defer ts.Close()

	gh := NewClientWithBaseUrl("tester", "secretlel", ts.URL)
	allStargazers, err := gh.ListStargazers(dummyLogger(t), "angular", "angular")
	assert.NoError(t, err)
	assert.Equal(t, len(allStargazers), 3)
}

func TestPagination(t *testing.T) {
	var nextPage string

	t.Parallel()
	call := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		if call < 2 {
			w.Header().Add("Link", fmt.Sprintf("<%s>; rel=\"next\"", nextPage))
		} else {
			assert.Equal(t, pathAndQueryOnly(t, nextPage), r.URL.String())
		}
		fmt.Fprintln(w, "[{}, {}, {}]")
	}))
	defer ts.Close()
	nextPage = fmt.Sprintf("%s/repos/angular/angular/stargazers?per_page=100&page=2", ts.URL)

	gh := NewClientWithBaseUrl("tester", "secretlel", ts.URL)
	allStargazers, err := gh.ListStargazers(dummyLogger(t), "angular", "angular")
	assert.NoError(t, err)
	assert.Equal(t, call, 2)
	assert.Equal(t, len(allStargazers), 6)
}

func TestListIssues(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/vnd.github.v3.star+json", r.Header.Get("Accept"))
		assert.Equal(t,
			"/repos/lodash/lodash/stargazers?per_page=100",
			pathAndQueryOnly(t, r.URL.String()),
		)
		fmt.Fprintln(w, "[{}, {}, {}, {}]")
	}))
	defer ts.Close()

	gh := NewClientWithBaseUrl("tester", "secretlel", ts.URL)
	allStargazers, err := gh.ListStargazers(dummyLogger(t), "lodash", "lodash")
	assert.NoError(t, err)
	assert.Equal(t, len(allStargazers), 4)
}
