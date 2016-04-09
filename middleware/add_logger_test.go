package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ksheedlo/ghviz/mocks"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestAddLogger(t *testing.T) {
	buf := new(bytes.Buffer)

	r := mux.NewRouter()
	r.HandleFunc("/", AddLogger(buf)(func(w http.ResponseWriter, r *http.Request) {
		log := context.Get(r, CtxLog).(*log.Logger)
		log.Println("Test Message")
	}))

	req := mocks.NewHttpRequest(t, "GET", "http://example.com/", nil)
	context.Set(req, CtxResponseId, "deadbeef")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// The middle 3 columns are driven by the logger and can't be controlled
	// by configuring AddLogger. We use a regex to parse past them.
	logLineRegex := mocks.CompileRegex(t, `deadbeef(?:\s+[^\s]+){3}\s+(.*)`)

	assert.Equal(t, "Test Message", logLineRegex.FindStringSubmatch(buf.String())[1])
}
