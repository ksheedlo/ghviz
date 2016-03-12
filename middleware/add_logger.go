package middleware

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/context"
)

func AddLogger(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseId := context.Get(r, CtxResponseId).(string)
		logger := log.New(os.Stdout, responseId+" ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
		context.Set(r, CtxLog, logger)
		handler(w, r)
	}
}
