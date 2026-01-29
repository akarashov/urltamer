package handler

import "net/http"

// Adapt converts middleware
func Adapt(mw func(http.HandlerFunc) http.HandlerFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mw(http.HandlerFunc(next.ServeHTTP))(w, r)
		})
	}
}
