package mocks

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/ksheedlo/ghviz/interfaces"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DummyLogger(t *testing.T) *log.Logger {
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0777)
	assert.NoError(t, err)
	return log.New(devnull, "", 0)
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

type MockRediser struct {
	mock.Mock
}

func (m *MockRediser) Del(key string) (int64, error) {
	args := m.Called(key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRediser) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockRediser) Set(key, value string, ttl time.Duration) error {
	// Set the value to "" because it's used to set JSON, which serializes
	// in an unpredictable order.
	args := m.Called(key, "", ttl)
	return args.Error(0)
}

func (m *MockRediser) ZAdd(key string, members ...interfaces.ZZ) (int64, error) {
	args := m.Called(key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRediser) ZRangeByScore(key string, opts *interfaces.ZRangeByScoreOpts) ([]string, error) {
	args := m.Called(key, opts)
	resultsArg := args.Get(0)
	if resultsArg != nil {
		return resultsArg.([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func MarshalJSON(t *testing.T, v interface{}) []byte {
	buf, err := json.Marshal(v)
	assert.NoError(t, err)
	return buf
}

func CompileRegex(t *testing.T, expr string) *regexp.Regexp {
	re, err := regexp.Compile(expr)
	assert.NoError(t, err)
	return re
}

func NewHttpRequest(t *testing.T, method, urlStr string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, urlStr, body)
	assert.NoError(t, err)
	return req
}
