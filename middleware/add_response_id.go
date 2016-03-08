package middleware

import (
	"crypto/rand"
	"math/big"
	"net/http"

	"github.com/gorilla/context"
)

func AddResponseId(handler Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		responseIdInt, err := rand.Int(rand.Reader, big.NewInt(1<<62))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error\n"))
			return
		}
		responseId := responseIdInt.Text(36)
		context.Set(r, CtxResponseId, responseId)
		w.Header().Add("X-Request-Id", responseId)
		handler(w, r)
	}
}
