package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestLogRequest(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	r.HandleFunc("/foof", LogRequest(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Foo", "foofbarf")
			w.Write([]byte("Test Response\n"))
		},
	))

	req, err := http.NewRequest("GET", "/foof", nil)
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	context.Set(req, CtxLog, log.New(buf, "", 0))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	logLineRegex, err := regexp.Compile(`hndl GET /foof (\d{3}) [^\s]+`)
	assert.NoError(t, err)

	match := logLineRegex.FindStringSubmatch(buf.String())
	assert.NotNil(t, match)
	assert.Equal(t, "200", match[1])
	assert.Equal(t, "foofbarf", w.Header().Get("X-Foo"))
}

func TestLogRequestPassesHeader(t *testing.T) {
	t.Parallel()

	r := mux.NewRouter()
	r.HandleFunc("/foof", LogRequest(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Nope\n"))
		},
	))

	req, err := http.NewRequest("GET", "/foof", nil)
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	context.Set(req, CtxLog, log.New(buf, "", 0))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	logLineRegex, err := regexp.Compile(`hndl GET /foof (\d{3}) [^\s]+`)
	assert.NoError(t, err)

	match := logLineRegex.FindStringSubmatch(buf.String())
	assert.NotNil(t, match)
	assert.Equal(t, "400", match[1])
}
