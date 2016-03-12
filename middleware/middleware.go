package middleware

import (
	"net/http"
)

const (
	CtxResponseId = iota
	CtxLog
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func Compose(middlewares ...Middleware) Middleware {
	if len(middlewares) == 1 {
		return middlewares[0]
	}
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return middlewares[0](Compose(middlewares[1:]...)(handler))
	}
}
