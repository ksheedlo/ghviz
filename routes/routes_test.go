package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"text/template"
	"time"

	"github.com/ksheedlo/ghviz/errors"
	"github.com/ksheedlo/ghviz/github"
	"github.com/ksheedlo/ghviz/interfaces"
	"github.com/ksheedlo/ghviz/middleware"
	"github.com/ksheedlo/ghviz/mocks"
	"github.com/ksheedlo/ghviz/simulate"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockListStarEventser struct {
	mock.Mock
}

func (m *MockListStarEventser) ListStarEvents(
	logger *log.Logger,
	owner, repo string,
) ([]github.StarEvent, *errors.HttpError) {
	args := m.Called(logger, owner, repo)
	var starEvents []github.StarEvent = nil
	var err *errors.HttpError = nil
	eventsArg := args.Get(0)
	if eventsArg != nil {
		starEvents = eventsArg.([]github.StarEvent)
	}
	errArg := args.Get(1)
	if errArg != nil {
		err = errArg.(*errors.HttpError)
	}
	return starEvents, err
}

func TestListStarCounts(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	ghMock := &MockListStarEventser{}
	logger := mocks.DummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", ListStarCounts(ghMock))
	req := mocks.NewHttpRequest(t, "GET", "http://example.com/tester1/coolrepo", nil)
	context.Set(req, middleware.CtxLog, logger)

	ghMock.
		On("ListStarEvents", logger, "tester1", "coolrepo").
		Return([]github.StarEvent{
			github.StarEvent{StarredAt: time.Unix(1, 0)},
			github.StarEvent{StarredAt: time.Unix(2, 0)},
			github.StarEvent{StarredAt: time.Unix(3, 0)},
		}, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	ghMock.AssertExpectations(t)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var bodyContents []map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &bodyContents))
	assert.Len(t, bodyContents, 3)
	assert.Equal(t, 3.0, bodyContents[2]["stars"].(float64))
}

func TestListStarCountsError(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	ghMock := &MockListStarEventser{}
	logger := mocks.DummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", ListStarCounts(ghMock))
	req := mocks.NewHttpRequest(t, "GET", "http://example.com/tester1/coolrepo", nil)
	context.Set(req, middleware.CtxLog, logger)

	ghMock.
		On("ListStarEvents", logger, "tester1", "coolrepo").
		Return(nil, &errors.HttpError{
			Message: "Github API Error",
			Status:  http.StatusInternalServerError,
		})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	ghMock.AssertExpectations(t)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Github API Error\n", w.Body.String())
}

type MockListIssueser struct {
	mock.Mock
}

func (m *MockListIssueser) ListIssues(
	logger *log.Logger,
	owner, repo string,
) ([]github.Issue, *errors.HttpError) {
	args := m.Called(logger, owner, repo)
	var issues []github.Issue = nil
	var err *errors.HttpError = nil
	issuesArg := args.Get(0)
	if issuesArg != nil {
		issues = issuesArg.([]github.Issue)
	}
	errArg := args.Get(1)
	if errArg != nil {
		err = errArg.(*errors.HttpError)
	}
	return issues, err
}

func TestListOpenIssuesAndPrs(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	ghMock := &MockListIssueser{}
	logger := mocks.DummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", ListOpenIssuesAndPrs(ghMock))
	req := mocks.NewHttpRequest(t, "GET", "http://example.com/tester1/coolrepo", nil)
	context.Set(req, middleware.CtxLog, logger)

	ghMock.
		On("ListIssues", logger, "tester1", "coolrepo").
		Return([]github.Issue{
			github.Issue{CreatedAt: time.Unix(3, 0), IsPr: false, IsClosed: false},
			github.Issue{
				CreatedAt: time.Unix(1, 0),
				IsPr:      false,
				IsClosed:  true,
				ClosedAt:  time.Unix(2, 0),
			},
			github.Issue{
				CreatedAt: time.Unix(4, 0),
				IsPr:      true,
				IsClosed:  true,
				ClosedAt:  time.Unix(6, 0),
			},
			github.Issue{CreatedAt: time.Unix(5, 0), IsPr: true, IsClosed: false},
		}, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	ghMock.AssertExpectations(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var bodyContents []map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &bodyContents))
	assert.Len(t, bodyContents, 6)
	assert.Equal(t, 1.0, bodyContents[len(bodyContents)-1]["open_prs"].(float64))
	assert.Equal(t, 1.0, bodyContents[len(bodyContents)-1]["open_issues"].(float64))
}

func TestListIssuesError(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	ghMock := &MockListIssueser{}
	logger := mocks.DummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", ListOpenIssuesAndPrs(ghMock))
	req := mocks.NewHttpRequest(t, "GET", "http://example.com/tester1/coolrepo", nil)
	context.Set(req, middleware.CtxLog, logger)

	ghMock.
		On("ListIssues", logger, "tester1", "coolrepo").
		Return(nil, &errors.HttpError{
			Message: "Github API Error",
			Status:  http.StatusInternalServerError,
		})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	ghMock.AssertExpectations(t)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Github API Error\n", w.Body.String())
}

type MockListTopIssueser struct {
	mock.Mock
}

func (m *MockListTopIssueser) ListTopIssues(
	logger *log.Logger,
	owner, repo string,
	limit int,
) ([]github.Issue, *errors.HttpError) {
	args := m.Called(logger, owner, repo, limit)
	var issues []github.Issue = nil
	var err *errors.HttpError = nil
	issuesArg := args.Get(0)
	if issuesArg != nil {
		issues = issuesArg.([]github.Issue)
	}
	errArg := args.Get(1)
	if errArg != nil {
		err = errArg.(*errors.HttpError)
	}
	return issues, err
}

func TestTopIssues(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	ghMock := &MockListTopIssueser{}
	logger := mocks.DummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", TopIssues(ghMock))
	req := mocks.NewHttpRequest(t, "GET", "http://example.com/tester1/coolrepo", nil)
	context.Set(req, middleware.CtxLog, logger)

	ghMock.
		On("ListTopIssues", logger, "tester1", "coolrepo", 5).
		Return([]github.Issue{
			github.Issue{Title: "Test Issue 1"},
			github.Issue{Title: "Test Issue 2"},
		}, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	ghMock.AssertExpectations(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var bodyContents []map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &bodyContents))
	assert.Len(t, bodyContents, 2)
	assert.Equal(t, "Test Issue 1", bodyContents[0]["title"].(string))
	assert.Equal(t, "Test Issue 2", bodyContents[1]["title"].(string))
}

func TestTopIssuesError(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	ghMock := &MockListTopIssueser{}
	logger := mocks.DummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", TopIssues(ghMock))
	req := mocks.NewHttpRequest(t, "GET", "http://example.com/tester1/coolrepo", nil)
	context.Set(req, middleware.CtxLog, logger)

	ghMock.
		On("ListTopIssues", logger, "tester1", "coolrepo", 5).
		Return(nil, &errors.HttpError{
			Message: "Github API Error",
			Status:  http.StatusInternalServerError,
		})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	ghMock.AssertExpectations(t)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Github API Error\n", w.Body.String())
}

type MockListTopPrser struct {
	mock.Mock
}

func (m *MockListTopPrser) ListTopPrs(
	logger *log.Logger,
	owner, repo string,
	limit int,
) ([]github.Issue, *errors.HttpError) {
	args := m.Called(logger, owner, repo, limit)
	var issues []github.Issue = nil
	var err *errors.HttpError = nil
	issuesArg := args.Get(0)
	if issuesArg != nil {
		issues = issuesArg.([]github.Issue)
	}
	errArg := args.Get(1)
	if errArg != nil {
		err = errArg.(*errors.HttpError)
	}
	return issues, err
}

func TestTopPrs(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	ghMock := &MockListTopPrser{}
	logger := mocks.DummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", TopPrs(ghMock))
	req := mocks.NewHttpRequest(t, "GET", "http://example.com/tester1/coolrepo", nil)
	context.Set(req, middleware.CtxLog, logger)

	ghMock.
		On("ListTopPrs", logger, "tester1", "coolrepo", 5).
		Return([]github.Issue{
			github.Issue{Title: "Test PR 1", IsPr: true},
			github.Issue{Title: "Test PR 2", IsPr: true},
		}, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	ghMock.AssertExpectations(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var bodyContents []map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &bodyContents))
	assert.Len(t, bodyContents, 2)
	assert.Equal(t, "Test PR 1", bodyContents[0]["title"].(string))
	assert.Equal(t, "Test PR 2", bodyContents[1]["title"].(string))
}

func TestTopPrsError(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	ghMock := &MockListTopPrser{}
	logger := mocks.DummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", TopPrs(ghMock))
	req := mocks.NewHttpRequest(t, "GET", "http://example.com/tester1/coolrepo", nil)
	context.Set(req, middleware.CtxLog, logger)

	ghMock.
		On("ListTopPrs", logger, "tester1", "coolrepo", 5).
		Return(nil, &errors.HttpError{
			Message: "Github API Error",
			Status:  http.StatusInternalServerError,
		})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	ghMock.AssertExpectations(t)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Github API Error\n", w.Body.String())
}

func TestServeIndex(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	logger := mocks.DummyLogger(t)
	tpl, err := template.New("index").Parse("<Test>{{.Owner}}|{{.Repo}}</Test>")
	assert.NoError(t, err)
	r.HandleFunc("/", ServeIndex(&IndexParams{
		Owner: "tester1",
		Repo:  "coolrepo",
	}, tpl))
	req := mocks.NewHttpRequest(t, "GET", "http://example.com/", nil)
	context.Set(req, middleware.CtxLog, logger)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "<Test>tester1|coolrepo</Test>", w.Body.String())
}

func TestHighScoresBadYear(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	logger := mocks.DummyLogger(t)
	r.HandleFunc("/{owner}/{repo}/{year}/{month}", HighScores(nil))
	req := mocks.NewHttpRequest(t,
		"GET",
		"http://example.com/tester1/coolrepo/foof/03",
		nil,
	)
	context.Set(req, middleware.CtxLog, logger)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var bodyContents map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &bodyContents))
	assert.Equal(t, "error", bodyContents["type"].(string))
	assert.Equal(t, 400, int(bodyContents["code"].(float64)))
	assert.Equal(t, "foof is not a valid year", bodyContents["message"].(string))
}

func TestHighScoresBadMonth(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	logger := mocks.DummyLogger(t)
	r.HandleFunc("/{owner}/{repo}/{year}/{month}", HighScores(nil))
	req := mocks.NewHttpRequest(t,
		"GET",
		"http://example.com/tester1/coolrepo/2016/barf",
		nil,
	)
	context.Set(req, middleware.CtxLog, logger)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var bodyContents map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &bodyContents))
	assert.Equal(t, "error", bodyContents["type"].(string))
	assert.Equal(t, 400, int(bodyContents["code"].(float64)))
	assert.Equal(t, "barf is not a valid month between 01-12", bodyContents["message"].(string))
}

func TestHighScoresNotFound(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	logger := mocks.DummyLogger(t)
	redis := &mocks.MockRediser{}
	r.HandleFunc("/{owner}/{repo}/{year}/{month}", HighScores(redis))
	req := mocks.NewHttpRequest(t,
		"GET",
		"http://example.com/tester1/coolrepo/2016/03",
		nil,
	)
	context.Set(req, middleware.CtxLog, logger)

	redis.
		On("Get", "gh:repos:tester1:coolrepo:issue_event_setid").
		Return("", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	redis.AssertExpectations(t)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var bodyContents map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &bodyContents))
	assert.Equal(t, "error", bodyContents["type"].(string))
	assert.Equal(t, 404, int(bodyContents["code"].(float64)))
	assert.Equal(t, "Scores for tester1/coolrepo were not found.", bodyContents["message"].(string))
}

func TestHighScoresRedisError(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	logger := mocks.DummyLogger(t)
	redis := &mocks.MockRediser{}
	r.HandleFunc("/{owner}/{repo}/{year}/{month}", HighScores(redis))
	req := mocks.NewHttpRequest(t,
		"GET",
		"http://example.com/tester1/coolrepo/2016/03",
		nil,
	)
	context.Set(req, middleware.CtxLog, logger)

	redis.
		On("Get", "gh:repos:tester1:coolrepo:issue_event_setid").
		Return("deadbeef", nil)

	redis.
		On(
			"ZRangeByScore",
			"gh:repos:tester1:coolrepo:issue_events:deadbeef",
			&interfaces.ZRangeByScoreOpts{
				Min: strconv.FormatInt(
					time.Date(2016, time.Month(3), 1, 0, 0, 0, 0, time.UTC).Unix(),
					10,
				),
				Max: strconv.FormatInt(
					time.Date(2016, time.Month(4), 1, 0, 0, 0, 0, time.UTC).Unix(),
					10,
				),
			}).
		Return(nil, mocks.ConstantError("Redis Error"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	redis.AssertExpectations(t)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var bodyContents map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &bodyContents))
	assert.Equal(t, "error", bodyContents["type"].(string))
	assert.Equal(t, 500, int(bodyContents["code"].(float64)))
	assert.Equal(t, "Internal Server Error", bodyContents["message"].(string))
}

func TestHighScoresBadJsonError(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	logger := mocks.DummyLogger(t)
	redis := &mocks.MockRediser{}
	r.HandleFunc("/{owner}/{repo}/{year}/{month}", HighScores(redis))
	req := mocks.NewHttpRequest(t,
		"GET",
		"http://example.com/tester1/coolrepo/2016/03",
		nil,
	)
	context.Set(req, middleware.CtxLog, logger)

	redis.
		On("Get", "gh:repos:tester1:coolrepo:issue_event_setid").
		Return("deadbeef", nil)

	redis.
		On(
			"ZRangeByScore",
			"gh:repos:tester1:coolrepo:issue_events:deadbeef",
			&interfaces.ZRangeByScoreOpts{
				Min: strconv.FormatInt(
					time.Date(2016, time.Month(3), 1, 0, 0, 0, 0, time.UTC).Unix(),
					10,
				),
				Max: strconv.FormatInt(
					time.Date(2016, time.Month(4), 1, 0, 0, 0, 0, time.UTC).Unix(),
					10,
				),
			}).
		Return([]string{
			`{"actor_id":"foof`,
		}, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	redis.AssertExpectations(t)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var bodyContents map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &bodyContents))
	assert.Equal(t, "error", bodyContents["type"].(string))
	assert.Equal(t, 500, int(bodyContents["code"].(float64)))
	assert.Equal(t, "Internal Server Error", bodyContents["message"].(string))
}

func marshalEachScoringEvent(t *testing.T, sevs ...simulate.ScoringEvent) []string {
	var result []string
	for _, sev := range sevs {
		jsonBytes := mocks.MarshalJSON(t, &sev)
		result = append(result, string(jsonBytes))
	}
	return result
}

func TestHighScores(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	logger := mocks.DummyLogger(t)
	redis := &mocks.MockRediser{}
	r.HandleFunc("/{owner}/{repo}/{year}/{month}", HighScores(redis))
	req := mocks.NewHttpRequest(t,
		"GET",
		"http://example.com/tester1/coolrepo/2016/03",
		nil,
	)
	context.Set(req, middleware.CtxLog, logger)

	redis.
		On("Get", "gh:repos:tester1:coolrepo:issue_event_setid").
		Return("deadbeef", nil)

	redis.
		On(
			"ZRangeByScore",
			"gh:repos:tester1:coolrepo:issue_events:deadbeef",
			&interfaces.ZRangeByScoreOpts{
				Min: strconv.FormatInt(
					time.Date(2016, time.Month(3), 1, 0, 0, 0, 0, time.UTC).Unix(),
					10,
				),
				Max: strconv.FormatInt(
					time.Date(2016, time.Month(4), 1, 0, 0, 0, 0, time.UTC).Unix(),
					10,
				),
			}).
		Return(marshalEachScoringEvent(t,
			simulate.ScoringEvent{ActorId: "tester1", EventType: simulate.IssueOpened},
			simulate.ScoringEvent{ActorId: "tester2", EventType: simulate.IssueReviewed},
		), nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	redis.AssertExpectations(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var bodyContents []map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &bodyContents))
	assert.Len(t, bodyContents, 2)
	assert.Equal(t, "tester2", bodyContents[0]["actor_id"].(string))
	assert.Equal(t, 1000, int(bodyContents[0]["score"].(float64)))
	assert.Equal(t, "tester1", bodyContents[1]["actor_id"].(string))
	assert.Equal(t, 200, int(bodyContents[1]["score"].(float64)))
}

func TestHighScoresYearWraparound(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	logger := mocks.DummyLogger(t)
	redis := &mocks.MockRediser{}
	r.HandleFunc("/{owner}/{repo}/{year}/{month}", HighScores(redis))
	req := mocks.NewHttpRequest(t,
		"GET",
		"http://example.com/tester1/coolrepo/2015/12",
		nil,
	)
	context.Set(req, middleware.CtxLog, logger)

	redis.
		On("Get", "gh:repos:tester1:coolrepo:issue_event_setid").
		Return("deadbeef", nil)

	redis.
		On(
			"ZRangeByScore",
			"gh:repos:tester1:coolrepo:issue_events:deadbeef",
			&interfaces.ZRangeByScoreOpts{
				Min: strconv.FormatInt(
					time.Date(2015, time.Month(12), 1, 0, 0, 0, 0, time.UTC).Unix(),
					10,
				),
				Max: strconv.FormatInt(
					time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC).Unix(),
					10,
				),
			}).
		Return(marshalEachScoringEvent(t,
			simulate.ScoringEvent{ActorId: "tester2", EventType: simulate.IssueReviewed},
			simulate.ScoringEvent{ActorId: "tester3", EventType: simulate.IssueReviewed},
			simulate.ScoringEvent{ActorId: "tester3", EventType: simulate.IssueReviewed},
		), nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	redis.AssertExpectations(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var bodyContents []map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &bodyContents))
	assert.Len(t, bodyContents, 2)
	assert.Equal(t, "tester3", bodyContents[0]["actor_id"].(string))
	assert.Equal(t, 2000, int(bodyContents[0]["score"].(float64)))
	assert.Equal(t, "tester2", bodyContents[1]["actor_id"].(string))
	assert.Equal(t, 1000, int(bodyContents[1]["score"].(float64)))
}
