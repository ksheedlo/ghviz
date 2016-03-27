package middleware

import (
	"net/http"

	"github.com/ksheedlo/ghviz/interfaces"

	"github.com/gorilla/context"
)

func AddResponseId(tagger interfaces.RandomTagger) Middleware {
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			responseId, err := tagger.RandomTag()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Server Error\n"))
				return
			}
			context.Set(r, CtxResponseId, responseId)
			w.Header().Add("X-Response-Id", responseId)
			handler(w, r)
		}
	}
}
