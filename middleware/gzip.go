package middleware

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
)

func Gzip(handler http.HandlerFunc) http.HandlerFunc {
	gzipHandler := gziphandler.GzipHandler(handler)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gzipHandler.ServeHTTP(w, r)
	})
}
