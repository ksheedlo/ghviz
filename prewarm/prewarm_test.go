package prewarm

import (
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ksheedlo/ghviz/errors"
	"github.com/ksheedlo/ghviz/github"
	"github.com/ksheedlo/ghviz/interfaces"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func dummyLogger(t *testing.T) *log.Logger {
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0777)
	assert.NoError(t, err)
	return log.New(devnull, "", 0)
}

type MockListAllPrEventser struct {
	mock.Mock
}

func (m *MockListAllPrEventser) ListAllPrEvents(
	logger *log.Logger,
	owner, repo string,
) ([]github.DetailedIssueEvent, *errors.HttpError) {
	args := m.Called(logger, owner, repo)
	errArg := args.Get(1)
	if errArg == nil {
		return args.Get(0).([]github.DetailedIssueEvent), nil
	}
	return args.Get(0).([]github.DetailedIssueEvent), errArg.(*errors.HttpError)
}

type MockRandomTagger struct {
	mock.Mock
}

func (m *MockRandomTagger) RandomTag() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

type ErrorFunc func() string

func (f ErrorFunc) Error() string {
	return f()
}

func ConstantError(msg string) ErrorFunc {
	return ErrorFunc(func() string {
		return msg
	})
}

func TestPrewarmHighScores(t *testing.T) {
	t.Parallel()

	redisMock := &interfaces.MockRediser{}
	ghMock := &MockListAllPrEventser{}
	clock := clockwork.NewFakeClock()
	randomTagger := &MockRandomTagger{}
	logger := dummyLogger(t)

	ghMock.
		On("ListAllPrEvents", logger, "tester1", "coolrepo").
		Return([]github.DetailedIssueEvent{
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(1, 0),
				Detail:      nil,
				EventType:   github.IssueCreated,
				IssueNumber: 1,
			},
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(2, 0),
				Detail:      map[string]interface{}{"name": "ready label"},
				EventType:   github.IssueLabeled,
				IssueNumber: 1,
			},
			github.DetailedIssueEvent{
				ActorId:     "tester2",
				CreatedAt:   time.Unix(3, 0),
				Detail:      map[string]interface{}{"name": "ready label"},
				EventType:   github.IssueUnlabeled,
				IssueNumber: 1,
			},
		}, nil)

	randomTagger.On("RandomTag").Return("deadbeef", nil)

	redisMock.
		On("ZAdd", "gh:repos:tester1:coolrepo:issue_events:deadbeef").
		Return(int64(0), nil)

	redisMock.
		On("Get", "gh:repos:tester1:coolrepo:issue_event_setid").
		Return("dogetest", nil)

	redisMock.
		On("Set", "gh:repos:tester1:coolrepo:issue_event_setid", "", time.Duration(0)).
		Return(nil)

	redisMock.
		On("Del", "gh:repos:tester1:coolrepo:issue_events:dogetest").
		Return(int64(0), nil)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := PrewarmHighScores(
			logger,
			ghMock,
			redisMock,
			clock,
			randomTagger,
			"tester1",
			"coolrepo",
		)
		assert.NoError(t, err)
		wg.Done()
	}()

	clock.BlockUntil(1)
	clock.Advance(6 * time.Second)

	wg.Wait()

	ghMock.AssertExpectations(t)
	randomTagger.AssertExpectations(t)
	redisMock.AssertExpectations(t)
}

func TestPrewarmPropagatesGithubError(t *testing.T) {
	t.Parallel()

	ghMock := &MockListAllPrEventser{}
	logger := dummyLogger(t)

	ghMock.
		On("ListAllPrEvents", logger, "tester1", "coolrepo").
		Return(
			[]github.DetailedIssueEvent{},
			&errors.HttpError{Message: "Server Error", Status: 500},
		)
	err := PrewarmHighScores(logger, ghMock, nil, nil, nil, "tester1", "coolrepo")

	assert.Error(t, err)
	ghMock.AssertExpectations(t)
}

func TestPrewarmPropagatesRandomIOError(t *testing.T) {
	t.Parallel()

	ghMock := &MockListAllPrEventser{}
	randomTagger := &MockRandomTagger{}
	logger := dummyLogger(t)

	ghMock.
		On("ListAllPrEvents", logger, "tester1", "coolrepo").
		Return([]github.DetailedIssueEvent{
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(1, 0),
				Detail:      nil,
				EventType:   github.IssueCreated,
				IssueNumber: 1,
			},
		}, nil)

	randomTagger.On("RandomTag").Return("", ConstantError("I/O Error"))

	err := PrewarmHighScores(
		logger,
		ghMock,
		nil,
		nil,
		randomTagger,
		"tester1",
		"coolrepo",
	)

	assert.Error(t, err)
	ghMock.AssertExpectations(t)
	randomTagger.AssertExpectations(t)
}

func TestPrewarmPropagatesRedisZAddError(t *testing.T) {
	t.Parallel()

	redisMock := &interfaces.MockRediser{}
	ghMock := &MockListAllPrEventser{}
	randomTagger := &MockRandomTagger{}
	logger := dummyLogger(t)

	ghMock.
		On("ListAllPrEvents", logger, "tester1", "coolrepo").
		Return([]github.DetailedIssueEvent{
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(1, 0),
				Detail:      nil,
				EventType:   github.IssueCreated,
				IssueNumber: 1,
			},
		}, nil)

	randomTagger.On("RandomTag").Return("deadbeef", nil)

	redisMock.
		On("ZAdd", "gh:repos:tester1:coolrepo:issue_events:deadbeef").
		Return(int64(0), ConstantError("ZAdd Error"))

	err := PrewarmHighScores(
		logger,
		ghMock,
		redisMock,
		nil,
		randomTagger,
		"tester1",
		"coolrepo",
	)

	assert.Error(t, err)
	ghMock.AssertExpectations(t)
	randomTagger.AssertExpectations(t)
	redisMock.AssertExpectations(t)
}

func TestPrewarmPropagatesRedisSetError(t *testing.T) {
	t.Parallel()

	redisMock := &interfaces.MockRediser{}
	ghMock := &MockListAllPrEventser{}
	randomTagger := &MockRandomTagger{}
	logger := dummyLogger(t)

	ghMock.
		On("ListAllPrEvents", logger, "tester1", "coolrepo").
		Return([]github.DetailedIssueEvent{
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(1, 0),
				Detail:      nil,
				EventType:   github.IssueCreated,
				IssueNumber: 1,
			},
		}, nil)

	randomTagger.On("RandomTag").Return("deadbeef", nil)

	redisMock.
		On("ZAdd", "gh:repos:tester1:coolrepo:issue_events:deadbeef").
		Return(int64(0), nil)

	redisMock.
		On("Get", "gh:repos:tester1:coolrepo:issue_event_setid").
		Return("dogetest", nil)

	redisMock.
		On("Set", "gh:repos:tester1:coolrepo:issue_event_setid", "", time.Duration(0)).
		Return(ConstantError("Redis Error"))

	err := PrewarmHighScores(
		logger,
		ghMock,
		redisMock,
		nil,
		randomTagger,
		"tester1",
		"coolrepo",
	)

	assert.Error(t, err)
	ghMock.AssertExpectations(t)
	randomTagger.AssertExpectations(t)
	redisMock.AssertExpectations(t)
}

func TestPrewarmRecoversDeleteOldIssueEventset(t *testing.T) {
	t.Parallel()

	redisMock := &interfaces.MockRediser{}
	ghMock := &MockListAllPrEventser{}
	clock := clockwork.NewFakeClock()
	randomTagger := &MockRandomTagger{}
	logger := dummyLogger(t)

	ghMock.
		On("ListAllPrEvents", logger, "tester1", "coolrepo").
		Return([]github.DetailedIssueEvent{
			github.DetailedIssueEvent{
				ActorId:     "tester1",
				CreatedAt:   time.Unix(1, 0),
				Detail:      nil,
				EventType:   github.IssueCreated,
				IssueNumber: 1,
			},
		}, nil)

	randomTagger.On("RandomTag").Return("deadbeef", nil)

	redisMock.
		On("ZAdd", "gh:repos:tester1:coolrepo:issue_events:deadbeef").
		Return(int64(0), nil)

	redisMock.
		On("Get", "gh:repos:tester1:coolrepo:issue_event_setid").
		Return("dogetest", nil)

	redisMock.
		On("Set", "gh:repos:tester1:coolrepo:issue_event_setid", "", time.Duration(0)).
		Return(nil)

	redisMock.
		On("Del", "gh:repos:tester1:coolrepo:issue_events:dogetest").
		Return(int64(0), ConstantError("Delete Event Set Failed"))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := PrewarmHighScores(
			logger,
			ghMock,
			redisMock,
			clock,
			randomTagger,
			"tester1",
			"coolrepo",
		)
		assert.NoError(t, err)
		wg.Done()
	}()

	clock.BlockUntil(1)
	clock.Advance(6 * time.Second)

	wg.Wait()

	ghMock.AssertExpectations(t)
	randomTagger.AssertExpectations(t)
	redisMock.AssertExpectations(t)
}
