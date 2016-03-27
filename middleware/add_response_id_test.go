package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ksheedlo/ghviz/interfaces"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func ConstantTag(tag string) interfaces.RandomTaggerFunc {
	return func() (string, error) {
		return tag, nil
	}
}

func ErrorTag(err error) interfaces.RandomTaggerFunc {
	return func() (string, error) {
		return "", err
	}
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

func TestAddResponseId(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	var responseId string
	r.HandleFunc("/", AddResponseId(ConstantTag("foof1234"))(
		func(w http.ResponseWriter, r *http.Request) {
			responseId = context.Get(r, CtxResponseId).(string)
		},
	))

	req, err := http.NewRequest("GET", "http://example.com/", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, "foof1234", responseId)
	assert.Equal(t, "foof1234", w.Header().Get("X-Response-Id"))
}

func TestAddResponseIdError(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	r.HandleFunc("/", AddResponseId(ErrorTag(ConstantError("Oops")))(
		func(w http.ResponseWriter, r *http.Request) {},
	))

	req, err := http.NewRequest("GET", "http://example.com/", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Server Error\n", w.Body.String())
}
