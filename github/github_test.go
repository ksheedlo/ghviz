package github

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/ksheedlo/ghviz/interfaces"
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

	gh := NewClient(&Options{
		BaseUrl: ts.URL,
		Token:   "deadbeef",
	})
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

	gh := NewClient(&Options{
		BaseUrl: ts.URL,
		Token:   "deadbeef",
	})
	allStargazers, err := gh.ListStargazers(dummyLogger(t), "angular", "angular")
	assert.NoError(t, err)
	assert.Equal(t, call, 2)
	assert.Equal(t, len(allStargazers), 6)
}

const issuesJson string = `[
{"created_at":"2016-03-07T03:26:14.739Z","closed_at":null,"number":1,
"events_url":"https://api.example.com/issues/1/events","user":{"login":"tester"}},
{"created_at":"2016-03-07T03:23:53.002Z","closed_at":"2016-03-07T03:25:41.469Z",
 "events_url":"https://api.example.com/issues/2/events","number":2,
 "user":{"login":"tester"}},
{"created_at":"2016-03-07T03:46:36.717Z","closed_at":"2016-03-07T03:46:55.993Z",
 "pull_request":{},"events_url":"https://api.example.com/issues/3/events",
 "number":3,"user":{"login":"tester"}},
 {"created_at":"2016-03-07T03:46:46.458Z","pull_request":{},"number":4,
 "events_url":"https://api.example.com/issues/4/events","user":{"login":"tester"}}]`

func TestListIssues(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))
		assert.Equal(t,
			"/repos/lodash/lodash/issues?per_page=100&state=all&sort=created&direction=asc",
			pathAndQueryOnly(t, r.URL.String()),
		)
		fmt.Fprintln(w, issuesJson)
	}))
	defer ts.Close()

	gh := NewClient(&Options{
		BaseUrl: ts.URL,
		Token:   "deadbeef",
	})
	allIssues, err := gh.ListIssues(dummyLogger(t), "lodash", "lodash")
	assert.NoError(t, err)
	assert.Equal(t, len(allIssues), 4)
	assert.Equal(t, allIssues[0].EventsUrl, "https://api.example.com/issues/1/events")
	assert.False(t, allIssues[0].IsPr)
	assert.True(t, allIssues[2].IsPr)
}

const issuesBadCreatedAtJson = `[
{"created_at":"fish","events_url":"https://api.example.com/issues/1/events",
 "number":1,"user":{"login":"tester"}}]`

func TestListIssuesBadCreatedAt(t *testing.T) {
	t.SkipNow()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, issuesBadCreatedAtJson)
	}))
	defer ts.Close()

	gh := NewClient(&Options{
		BaseUrl: ts.URL,
		Token:   "deadbeef",
	})
	_, err := gh.ListIssues(dummyLogger(t), "lodash", "lodash")
	assert.Error(t, err)
}

const issuesBadClosedAtJson = `[
{"created_at":"2016-03-07T03:26:14.739Z","closed_at":"fish","number":1,
 "events_url":"https://api.example.com/issues/1/events","user":{"login":"tester"}}]`

func TestListIssuesBadClosedAt(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, issuesBadClosedAtJson)
	}))
	defer ts.Close()

	gh := NewClient(&Options{
		BaseUrl: ts.URL,
		Token:   "deadbeef",
	})
	_, err := gh.ListIssues(dummyLogger(t), "lodash", "lodash")
	assert.Error(t, err)
}

func TestRedisCacheHit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.FailNow(t, "Test should hit Redis and not call the API!")
	}))
	defer ts.Close()

	redisMock := &interfaces.MockRediser{}
	gh := NewClient(&Options{
		BaseUrl:     ts.URL,
		RedisClient: redisMock,
		Token:       "deadbeef",
	})

	redisMock.On("Get", "github:repo:lodash:lodash:issues").Return(issuesJson, nil)

	allIssues, err := gh.ListIssues(dummyLogger(t), "lodash", "lodash")
	assert.NoError(t, err)
	assert.Equal(t, len(allIssues), 4)
	assert.Equal(t, allIssues[0].EventsUrl, "https://api.example.com/issues/1/events")
	redisMock.AssertExpectations(t)
}

func TestRedisCacheSet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, issuesJson)
	}))
	defer ts.Close()

	redisMock := &interfaces.MockRediser{}
	gh := NewClient(&Options{
		BaseUrl:     ts.URL,
		RedisClient: redisMock,
		Token:       "deadbeef",
	})

	cacheKey := "github:repo:lodash:lodash:issues"
	redisMock.On("Get", cacheKey).Return("", nil)
	redisMock.On("Set", cacheKey, "", time.Duration(10)*time.Minute).Return(nil)

	_, err := gh.ListIssues(dummyLogger(t), "lodash", "lodash")
	assert.NoError(t, err)
	redisMock.AssertExpectations(t)
}
