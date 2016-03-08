package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/context"
)

func LogRequest(handler Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		handler(w, r)
		logger := context.Get(r, CtxLog).(*log.Logger)
		logger.Printf("hndl %s %s %s", r.Method, r.URL.String(), time.Since(startTime).String())
	}
}
