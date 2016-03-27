package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ksheedlo/ghviz/errors"
	"github.com/ksheedlo/ghviz/github"
	"github.com/ksheedlo/ghviz/middleware"
)

func dummyLogger(t *testing.T) *log.Logger {
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0777)
	assert.NoError(t, err)
	return log.New(devnull, "", 0)
}

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
	logger := dummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", ListStarCounts(ghMock))
	req, err := http.NewRequest("GET", "http://example.com/tester1/coolrepo", nil)
	assert.NoError(t, err)
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
	logger := dummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", ListStarCounts(ghMock))
	req, err := http.NewRequest("GET", "http://example.com/tester1/coolrepo", nil)
	assert.NoError(t, err)
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
	logger := dummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", ListOpenIssuesAndPrs(ghMock))
	req, err := http.NewRequest("GET", "http://example.com/tester1/coolrepo", nil)
	assert.NoError(t, err)
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
	logger := dummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", ListOpenIssuesAndPrs(ghMock))
	req, err := http.NewRequest("GET", "http://example.com/tester1/coolrepo", nil)
	assert.NoError(t, err)
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
	logger := dummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", TopIssues(ghMock))
	req, err := http.NewRequest("GET", "http://example.com/tester1/coolrepo", nil)
	assert.NoError(t, err)
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
	logger := dummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", TopIssues(ghMock))
	req, err := http.NewRequest("GET", "http://example.com/tester1/coolrepo", nil)
	assert.NoError(t, err)
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
	logger := dummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", TopPrs(ghMock))
	req, err := http.NewRequest("GET", "http://example.com/tester1/coolrepo", nil)
	assert.NoError(t, err)
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
	logger := dummyLogger(t)
	r.HandleFunc("/{owner}/{repo}", TopPrs(ghMock))
	req, err := http.NewRequest("GET", "http://example.com/tester1/coolrepo", nil)
	assert.NoError(t, err)
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
