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
