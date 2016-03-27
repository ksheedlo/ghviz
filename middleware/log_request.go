package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/context"
)

type statusTrackingResponseWriter struct {
	w      http.ResponseWriter
	status int
}

func newStatusTrackingResponseWriter(w http.ResponseWriter) *statusTrackingResponseWriter {
	writer := new(statusTrackingResponseWriter)
	writer.w = w
	writer.status = -1
	return writer
}

func (strw *statusTrackingResponseWriter) Header() http.Header {
	return strw.w.Header()
}

func (strw *statusTrackingResponseWriter) Write(b []byte) (int, error) {
	if strw.status == -1 {
		strw.status = http.StatusOK
	}
	return strw.w.Write(b)
}

func (strw *statusTrackingResponseWriter) WriteHeader(status int) {
	strw.status = status
	strw.w.WriteHeader(status)
}

func LogRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		tracker := newStatusTrackingResponseWriter(w)
		handler(tracker, r)
		logger := context.Get(r, CtxLog).(*log.Logger)
		logger.Printf("hndl %s %s %d %s", r.Method, r.URL.String(), tracker.status, time.Since(startTime).String())
	}
}
