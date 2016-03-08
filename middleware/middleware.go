package middleware

import (
	"net/http"
)

const (
	CtxResponseId = iota
	CtxLog
)

type Handler func(http.ResponseWriter, *http.Request)

type Middleware func(Handler) Handler

func Compose(middlewares ...Middleware) Middleware {
	if len(middlewares) == 1 {
		return middlewares[0]
	}
	return func(handler Handler) Handler {
		return middlewares[0](Compose(middlewares[1:]...)(handler))
	}
}
