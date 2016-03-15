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

const issuesJson string = `[{
	"created_at":"2016-03-07T03:26:14.739Z",
	"closed_at":null,
  "events_url":"https://api.example.com/issues/1/events",
  "html_url":"https://api.example.com/issues/1",
	"number":1,
	"title":"Test 1",
	"user":{"login":"tester1"}
}, {
	"created_at":"2016-03-07T03:23:53.002Z",
	"closed_at":"2016-03-07T03:25:41.469Z",
  "events_url":"https://api.example.com/issues/2/events",
  "html_url":"https://api.example.com/issues/2",
	"number":2,
	"title":"Test 2",
	"user":{"login":"tester1"}
}, {
	"created_at":"2016-03-07T03:46:36.717Z",
	"closed_at":"2016-03-07T03:46:55.993Z",
	"events_url":"https://api.example.com/issues/3/events",
  "html_url":"https://api.example.com/pull/3",
	"number":3,
  "pull_request":{},
	"title":"Test 3",
	"user":{"login":"tester1"}
}, {
	"created_at":"2016-03-07T03:46:46.458Z",
  "events_url":"https://api.example.com/issues/4/events",
  "html_url":"https://api.example.com/pull/4",
	"number":4,
	"pull_request":{},
	"title":"Test 4",
	"user":{"login":"tester1"}
}]`

func TestListIssues(t *testing.T) {
	t.Parallel()

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

const issuesBadCreatedAtJson = `[{
	"created_at":"fish",
	"events_url":"https://api.example.com/issues/1/events",
  "html_url":"https://api.example.com/issues/1",
	"number":1,
	"title":"Test 1",
	"user":{"login":"tester1"}
}]`

func TestListIssuesBadCreatedAt(t *testing.T) {
	t.Parallel()

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

const issuesBadClosedAtJson = `[{
	"created_at":"2016-03-07T03:26:14.739Z",
	"closed_at":"fish","title":"Test 1",
  "events_url":"https://api.example.com/issues/1/events",
  "html_url":"https://api.example.com/issues/1",
	"number":1,
	"title":"Test 1",
	"user":{"login":"tester1"}
}]`

func TestListIssuesBadClosedAt(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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

const topIssuesJsonPage1 string = `[{
	"created_at":"2016-03-07T03:26:14.739Z",
	"closed_at":null,
  "events_url":"https://api.example.com/issues/1/events",
  "html_url":"https://api.example.com/issues/1",
	"number":1,
  "title":"Test 1",
	"user":{"login":"tester1"}
}, {
	"created_at":"2016-03-07T03:23:53.002Z",
	"closed_at":"2016-03-07T03:25:41.469Z",
  "events_url":"https://api.example.com/issues/2/events",
  "html_url":"https://api.example.com/issues/2",
	"number":2,
	"title":"Test 2",
	"user":{"login":"tester1"}
}, {
	"created_at":"2016-03-07T03:46:36.717Z",
	"closed_at":"2016-03-07T03:46:55.993Z",
  "pull_request":{},
	"events_url":"https://api.example.com/issues/3/events",
  "html_url":"https://api.example.com/pull/3",
	"number":3,
  "title":"Test 3",
	"user":{"login":"tester1"}
}, {
	"created_at":"2016-03-07T03:46:46.458Z",
  "events_url":"https://api.example.com/issues/4/events",
  "html_url":"https://api.example.com/pull/4",
	"number":4,
	"pull_request":{},
	"title":"Test 4",
	"user":{"login":"tester1"}
}, {
	"created_at":"2017-03-07T03:23:53.002Z",
	"closed_at":null,
  "events_url":"https://api.example.com/issues/5/events",
  "html_url":"https://api.example.com/issues/5",
	"number":5,
	"title":"Test 5",
	"user":{"login":"tester1"}
}]`

const topIssuesJsonPage2 string = `[{
	"created_at":"2016-06-07T03:26:14.739Z",
	"closed_at":null,
  "events_url":"https://api.example.com/issues/6/events",
  "html_url":"https://api.example.com/issues/6",
	"number":6,
  "title":"Test 6",
	"user":{"login":"tester1"}
}, {
	"created_at":"2016-06-07T03:23:53.002Z",
	"closed_at":"2016-06-07T03:25:41.469Z",
  "events_url":"https://api.example.com/issues/7/events",
  "html_url":"https://api.example.com/issues/7",
	"number":7,
	"title":"Test 7",
	"user":{"login":"tester1"}
}, {
	"created_at":"2016-07-07T03:23:53.002Z",
	"closed_at":"2016-07-07T03:25:41.469Z",
  "events_url":"https://api.example.com/issues/8/events",
  "html_url":"https://api.example.com/issues/8",
	"number":8,
	"title":"Test 8",
	"user":{"login":"tester1"}
}]`

func TestTopIssues(t *testing.T) {
	t.Parallel()

	var nextPage string
	call := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		if call < 2 {
			w.Header().Add("Link", fmt.Sprintf("<%s>; rel=\"next\"", nextPage))
			fmt.Fprintln(w, topIssuesJsonPage1)
		} else {
			assert.Equal(t, pathAndQueryOnly(t, nextPage), r.URL.String())
			fmt.Fprintln(w, topIssuesJsonPage2)
		}
	}))
	defer ts.Close()
	nextPage = fmt.Sprintf("%s/repos/lodash/lodash/issues?page=2", ts.URL)

	gh := NewClient(&Options{
		BaseUrl: ts.URL,
		Token:   "deadbeef",
	})
	allIssues, err := gh.ListTopIssues(dummyLogger(t), "lodash", "lodash", 5)
	assert.NoError(t, err)
	assert.Equal(t, call, 2)
	assert.Equal(t, len(allIssues), 5)
	assert.Equal(t, allIssues[len(allIssues)-1].Title, "Test 7")

	for _, issue := range allIssues {
		assert.False(t, issue.IsPr)
	}
}

const topPrsJsonPage2 string = `[{
	"created_at":"2016-06-07T03:26:14.739Z",
	"closed_at":null,
  "events_url":"https://api.example.com/issues/6/events",
  "html_url":"https://api.example.com/issues/6",
	"number":6,
	"pull_request":{},
  "title":"PR 6",
	"user":{"login":"tester1"}
}, {
	"created_at":"2016-06-07T03:23:53.002Z",
	"closed_at":"2016-06-07T03:25:41.469Z",
  "events_url":"https://api.example.com/issues/7/events",
  "html_url":"https://api.example.com/issues/7",
	"number":7,
	"pull_request":{},
	"title":"PR 7",
	"user":{"login":"tester1"}
}, {
	"created_at":"2016-07-07T03:23:53.002Z",
	"closed_at":"2016-07-07T03:25:41.469Z",
  "events_url":"https://api.example.com/issues/8/events",
  "html_url":"https://api.example.com/issues/8",
	"number":8,
	"pull_request":{},
	"title":"PR 8",
	"user":{"login":"tester1"}
}, {
	"created_at":"2016-08-07T03:23:53.002Z",
	"closed_at":"2016-08-07T03:25:41.469Z",
  "events_url":"https://api.example.com/issues/8/events",
  "html_url":"https://api.example.com/issues/8",
	"number":9,
	"pull_request":{},
	"title":"PR 9",
	"user":{"login":"tester1"}
}]`

func TestTopPrs(t *testing.T) {
	t.Parallel()

	var nextPage string
	call := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		if call < 2 {
			w.Header().Add("Link", fmt.Sprintf("<%s>; rel=\"next\"", nextPage))
			fmt.Fprintln(w, topIssuesJsonPage1)
		} else {
			assert.Equal(t, pathAndQueryOnly(t, nextPage), r.URL.String())
			fmt.Fprintln(w, topPrsJsonPage2)
		}
	}))
	defer ts.Close()
	nextPage = fmt.Sprintf("%s/repos/lodash/lodash/issues?page=2", ts.URL)

	gh := NewClient(&Options{
		BaseUrl: ts.URL,
		Token:   "deadbeef",
	})
	allIssues, err := gh.ListTopPrs(dummyLogger(t), "lodash", "lodash", 5)
	assert.NoError(t, err)
	assert.Equal(t, call, 2)
	assert.Equal(t, len(allIssues), 5)
	assert.Equal(t, allIssues[len(allIssues)-1].Title, "PR 8")

	for _, issue := range allIssues {
		assert.True(t, issue.IsPr)
	}
}
