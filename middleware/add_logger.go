package middleware

import (
	"io"
	"log"
	"net/http"

	"github.com/gorilla/context"
)

func AddLogger(writer io.Writer) Middleware {
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			responseId := context.Get(r, CtxResponseId).(string)
			logger := log.New(writer, responseId+" ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
			context.Set(r, CtxLog, logger)
			handler(w, r)
		}
	}
}
