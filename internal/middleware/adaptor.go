package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
)

// AdaptHandlerFuncToMux converts a http.HandlerFunc middleware to mux.MiddlewareFunc
func AdaptHandlerFuncToMux(middlewareFunc func(http.HandlerFunc) http.HandlerFunc) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareFunc(next.ServeHTTP)(w, r)
		})
	}
}
