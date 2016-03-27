package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestCompose(t *testing.T) {
	t.Parallel()

	var calls []string
	outer := func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "outer:before")
			hf(w, r)
			calls = append(calls, "outer:after")
		}
	}
	inner := func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "inner:before")
			hf(w, r)
			calls = append(calls, "inner:after")
		}
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, "handler")
		fmt.Fprintln(w, "Goodbye, cruel world!")
	}

	r := mux.NewRouter()
	withMiddleware := Compose(outer, inner)
	r.HandleFunc("/", withMiddleware(handler))

	req, err := http.NewRequest("GET", "http://example.com/", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t,
		[]string{"outer:before", "inner:before", "handler", "inner:after", "outer:after"},
		calls,
	)
	assert.Equal(t, "Goodbye, cruel world!\n", w.Body.String())
}
