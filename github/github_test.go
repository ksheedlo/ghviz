package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ksheedlo/ghviz/mocks"

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

const starsJson string = `[
{"starred_at":"2016-03-07T03:25:41.469Z"},
{"starred_at":"2016-03-07T03:23:53.002Z"},
{"starred_at":"2016-03-07T03:26:14.739Z"}]`

func TestListStarEvents(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/vnd.github.v3.star+json", r.Header.Get("Accept"))
		assert.Equal(t,
			"/repos/angular/angular/stargazers?per_page=100",
			pathAndQueryOnly(t, r.URL.String()),
		)
		fmt.Fprintln(w, starsJson)
	}))
	defer ts.Close()

	gh := NewClient(&Options{
		BaseUrl: ts.URL,
		Token:   "deadbeef",
	})
	starEvents, err := gh.ListStarEvents(mocks.DummyLogger(t), "angular", "angular")
	assert.NoError(t, err)
	assert.Equal(t, len(starEvents), 3)
	assert.True(t,
		starEvents[0].StarredAt.Before(starEvents[1].StarredAt),
		"Expected starEvents[0] to be before starEvents[1] !",
	)
	assert.True(t,
		starEvents[1].StarredAt.Before(starEvents[2].StarredAt),
		"Expected starEvents[1] to be before starEvents[2] !",
	)
}

const starsBadJson string = `["starred_at":"201`

func TestListStarEventsBadJson(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, starsBadJson)
	}))
	defer ts.Close()

	gh := NewClient(&Options{
		BaseUrl: ts.URL,
		Token:   "deadbeef",
	})
	_, err := gh.ListStarEvents(mocks.DummyLogger(t), "angular", "angular")
	assert.Error(t, err)
}

const starsBadStarredAtJson string = `[{"starred_at":"fish"}]`

func TestListStarEventsBadStarredAt(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, starsBadStarredAtJson)
	}))
	defer ts.Close()

	gh := NewClient(&Options{
		BaseUrl: ts.URL,
		Token:   "deadbeef",
	})
	_, err := gh.ListStarEvents(mocks.DummyLogger(t), "angular", "angular")
	assert.Error(t, err)
}

const starsJsonPage2 string = `[
{"starred_at":"2016-03-26T22:08:30.679Z"},
{"starred_at":"2016-03-26T22:08:35.319Z"},
{"starred_at":"2016-03-26T22:08:39.351Z"}]`

func TestPagination(t *testing.T) {
	var nextPage string

	t.Parallel()
	call := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		if call < 2 {
			w.Header().Add("Link", fmt.Sprintf("<%s>; rel=\"next\"", nextPage))
			fmt.Fprintln(w, starsJson)
		} else {
			assert.Equal(t, pathAndQueryOnly(t, nextPage), r.URL.String())
			fmt.Fprintln(w, starsJsonPage2)
		}
	}))
	defer ts.Close()
	nextPage = fmt.Sprintf("%s/repos/angular/angular/stargazers?per_page=100&page=2", ts.URL)

	gh := NewClient(&Options{
		BaseUrl: ts.URL,
		Token:   "deadbeef",
	})
	starEvents, err := gh.ListStarEvents(mocks.DummyLogger(t), "angular", "angular")
	assert.NoError(t, err)
	assert.Equal(t, call, 2)
	assert.Equal(t, len(starEvents), 6)
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
	allIssues, err := gh.ListIssues(mocks.DummyLogger(t), "lodash", "lodash")
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
	_, err := gh.ListIssues(mocks.DummyLogger(t), "lodash", "lodash")
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
	_, err := gh.ListIssues(mocks.DummyLogger(t), "lodash", "lodash")
	assert.Error(t, err)
}

func TestRedisCacheHit(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.FailNow(t, "Test should hit Redis and not call the API!")
	}))
	defer ts.Close()

	redisMock := &mocks.MockRediser{}
	gh := NewClient(&Options{
		BaseUrl:      ts.URL,
		MaxStaleness: 5,
		RedisClient:  redisMock,
		Token:        "deadbeef",
	})

	redisMock.On("Get", "github:repo:lodash:lodash:issues").Return(
		fmt.Sprintf("%d|%s", time.Now().Unix(), issuesJson),
		nil,
	)

	allIssues, err := gh.ListIssues(mocks.DummyLogger(t), "lodash", "lodash")
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

	redisMock := &mocks.MockRediser{}
	gh := NewClient(&Options{
		BaseUrl:      ts.URL,
		MaxStaleness: 5,
		RedisClient:  redisMock,
		Token:        "deadbeef",
	})

	cacheKey := "github:repo:lodash:lodash:issues"
	redisMock.On("Get", cacheKey).Return("", nil)
	redisMock.On("Set", cacheKey, "", time.Duration(0)).Return(nil)

	_, err := gh.ListIssues(mocks.DummyLogger(t), "lodash", "lodash")
	assert.NoError(t, err)
	redisMock.AssertExpectations(t)
}

func TestRedisStaleCacheHit(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, issuesJson)
	}))
	defer ts.Close()

	redisMock := &mocks.MockRediser{}
	gh := NewClient(&Options{
		BaseUrl:      ts.URL,
		MaxStaleness: 5,
		RedisClient:  redisMock,
		Token:        "deadbeef",
	})

	cacheKey := "github:repo:lodash:lodash:issues"
	redisMock.On("Get", cacheKey).Return(
		fmt.Sprintf("%d|meh", time.Now().Add(time.Duration(-6)*time.Minute).Unix()),
		nil,
	)
	redisMock.On("Set", cacheKey, "", time.Duration(0)).Return(nil)

	allIssues, err := gh.ListIssues(mocks.DummyLogger(t), "lodash", "lodash")
	assert.NoError(t, err)
	assert.Equal(t, len(allIssues), 4)
	assert.Equal(t, allIssues[0].EventsUrl, "https://api.example.com/issues/1/events")
	redisMock.AssertExpectations(t)
}

func TestBadRedisValues(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, issuesJson)
	}))
	defer ts.Close()

	redisMock := &mocks.MockRediser{}
	gh := NewClient(&Options{
		BaseUrl:      ts.URL,
		MaxStaleness: 5,
		RedisClient:  redisMock,
		Token:        "deadbeef",
	})

	cacheKey := "github:repo:lodash:lodash:issues"
	redisMock.On("Get", cacheKey).Return("chicken", nil)
	redisMock.On("Set", cacheKey, "", time.Duration(0)).Return(nil)

	allIssues, err := gh.ListIssues(mocks.DummyLogger(t), "lodash", "lodash")
	assert.NoError(t, err)
	assert.Equal(t, len(allIssues), 4)
	assert.Equal(t, allIssues[0].EventsUrl, "https://api.example.com/issues/1/events")
	redisMock.AssertExpectations(t)
}

func TestBadRedisTimestamp(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, issuesJson)
	}))
	defer ts.Close()

	redisMock := &mocks.MockRediser{}
	gh := NewClient(&Options{
		BaseUrl:      ts.URL,
		MaxStaleness: 5,
		RedisClient:  redisMock,
		Token:        "deadbeef",
	})

	cacheKey := "github:repo:lodash:lodash:issues"
	redisMock.On("Get", cacheKey).Return("fish|chicken", nil)
	redisMock.On("Set", cacheKey, "", time.Duration(0)).Return(nil)

	allIssues, err := gh.ListIssues(mocks.DummyLogger(t), "lodash", "lodash")
	assert.NoError(t, err)
	assert.Equal(t, len(allIssues), 4)
	assert.Equal(t, allIssues[0].EventsUrl, "https://api.example.com/issues/1/events")
	redisMock.AssertExpectations(t)
}

func TestBadRedisJSON(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, issuesJson)
	}))
	defer ts.Close()

	redisMock := &mocks.MockRediser{}
	gh := NewClient(&Options{
		BaseUrl:      ts.URL,
		MaxStaleness: 5,
		RedisClient:  redisMock,
		Token:        "deadbeef",
	})

	brokenIssueJson := `{"title":"Test Issue`
	cacheKey := "github:repo:lodash:lodash:issues"
	redisMock.On("Get", cacheKey).Return(
		fmt.Sprintf("%d|%s", time.Now().Unix(), brokenIssueJson),
		nil,
	)
	redisMock.On("Set", cacheKey, "", time.Duration(0)).Return(nil)

	allIssues, err := gh.ListIssues(mocks.DummyLogger(t), "lodash", "lodash")
	assert.NoError(t, err)
	assert.Equal(t, len(allIssues), 4)
	assert.Equal(t, allIssues[0].EventsUrl, "https://api.example.com/issues/1/events")
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
	allIssues, err := gh.ListTopIssues(mocks.DummyLogger(t), "lodash", "lodash", 5)
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
	allIssues, err := gh.ListTopPrs(mocks.DummyLogger(t), "lodash", "lodash", 5)
	assert.NoError(t, err)
	assert.Equal(t, call, 2)
	assert.Equal(t, len(allIssues), 5)
	assert.Equal(t, allIssues[len(allIssues)-1].Title, "PR 8")

	for _, issue := range allIssues {
		assert.True(t, issue.IsPr)
	}
}

const prEventsJson string = `[{
	"event": "whatever",
	"issue": {
		"created_at":"2016-06-07T03:26:14.739Z",
		"closed_at":null,
	  "events_url":"https://api.example.com/issues/6/events",
	  "html_url":"https://api.example.com/issues/6",
		"number":6,
		"pull_request":{},
	  "title":"PR 6",
		"user":{"login":"tester1"}
	}
}, {
	"actor": {"login": "tester2"},
	"commit_id": "deadbeef",
	"created_at": "2016-03-16T22:21:39.799Z",
	"event": "closed",
	"id": 87930,
	"issue": {
		"created_at":"2016-06-07T03:23:53.002Z",
		"closed_at":"2016-06-07T03:25:41.469Z",
	  "events_url":"https://api.example.com/issues/7/events",
	  "html_url":"https://api.example.com/issues/7",
		"number":7,
		"pull_request":{},
		"title":"PR 7",
		"user":{"login":"tester1"}
	}
}, {
  "actor": {"login": "tester3"},
	"created_at": "2016-03-16T22:24:01.888Z",
	"event": "labeled",
	"id": 87931,
	"issue": {
		"created_at":"2016-08-07T03:23:53.002Z",
		"closed_at":"2016-08-07T03:25:41.469Z",
	  "events_url":"https://api.example.com/issues/8/events",
	  "html_url":"https://api.example.com/issues/8",
		"number":9,
		"pull_request":{},
		"title":"PR 9",
		"user":{"login":"tester1"}
	},
	"label": "ready for review"
}, {
	"actor": {"login": "tester2"},
	"commit_id": "deadbeef",
	"created_at": "2016-03-16T22:21:39.799Z",
	"event": "closed",
	"id": 87932,
	"issue": {
		"created_at":"2016-06-07T03:26:14.739Z",
		"closed_at":null,
	  "events_url":"https://api.example.com/issues/10/events",
	  "html_url":"https://api.example.com/issues/10",
		"number":10,
	  "title":"Test 10",
		"user":{"login":"tester1"}
	}
}]`

func assertIssueEventContents(
	t *testing.T,
	issueEvent DetailedIssueEvent,
	actorId string,
	eventType DetailedIssueEventType,
	id string,
	issueNumber int,
) {
	assert.Equal(t, issueEvent.ActorId, actorId)
	assert.Equal(t, issueEvent.EventType, eventType)
	assert.Equal(t, issueEvent.Id, id)
	assert.Equal(t, issueEvent.IssueNumber, issueNumber)
}

func TestListAllPrEvents(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, prEventsJson)
	}))
	defer ts.Close()

	gh := NewClient(&Options{
		BaseUrl: ts.URL,
		Token:   "deadbeef",
	})

	prEvents, err := gh.ListAllPrEvents(mocks.DummyLogger(t), "lodash", "lodash")
	assert.NoError(t, err)
	assert.Len(t, prEvents, 5)

	assertIssueEventContents(t, prEvents[0], "tester1", IssueCreated, "cr6", 6)
	assertIssueEventContents(t, prEvents[1], "tester1", IssueCreated, "cr7", 7)
	assertIssueEventContents(t, prEvents[2], "tester2", IssueClosed, "87930", 7)
	assertIssueEventContents(t, prEvents[3], "tester1", IssueCreated, "cr9", 9)
	assertIssueEventContents(t, prEvents[4], "tester3", IssueLabeled, "87931", 9)
	assert.Equal(t, prEvents[4].Detail.(string), "ready for review")
}

func TestGithubError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Trigger a low-level HTTP error by causing a redirect loop.
		http.Redirect(w, r, r.URL.String(), http.StatusFound)
	}))
	defer ts.Close()

	gh := NewClient(&Options{
		BaseUrl: ts.URL,
		Token:   "deadbeef",
	})

	_, err := gh.ListIssues(mocks.DummyLogger(t), "lodash", "lodash")
	assert.Error(t, err)
}

func TestMarshalIssue(t *testing.T) {
	t.Parallel()

	jsonBytes := mocks.MarshalJSON(t, &Issue{
		CreatedAt: time.Unix(1458966366, 892000000).UTC(),
		ClosedAt:  time.Unix(1458969687, 787000000).UTC(),
		EventsUrl: "https://api.github.com/repos/88/issues/99/events",
		HtmlUrl:   "https://github.com/lodash/lodash/issues/99",
		IsClosed:  true,
		IsPr:      false,
		Number:    99,
		Submitter: "tester1",
		Title:     "Test Issue",
	})
	var issue map[string]interface{}
	assert.NoError(t, json.Unmarshal(jsonBytes, &issue))
	assert.Equal(t, "2016-03-26T04:26:06.892Z", issue["created_at"].(string))
	assert.Equal(t, "2016-03-26T05:21:27.787Z", issue["closed_at"].(string))
	assert.Equal(t,
		"https://api.github.com/repos/88/issues/99/events",
		issue["events_url"].(string),
	)
	assert.Equal(t,
		"https://github.com/lodash/lodash/issues/99",
		issue["html_url"].(string),
	)
	assert.Equal(t, true, issue["is_closed"].(bool))
	assert.Equal(t, false, issue["is_pr"].(bool))
	assert.Equal(t, 99.0, issue["number"].(float64))
	assert.Equal(t, "tester1", issue["submitter"].(string))
	assert.Equal(t, "Test Issue", issue["title"].(string))
}
