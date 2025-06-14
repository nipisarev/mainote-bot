package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Logger returns a middleware that logs HTTP requests
func Logger(operationName string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("operation", operationName).
				Msg("Request started")

			next.ServeHTTP(w, r)

			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("operation", operationName).
				Dur("duration", time.Since(start)).
				Msg("Request completed")
		})
	}
}

// Tracing returns a middleware that adds tracing information
func Tracing(operationName string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add tracing headers
			w.Header().Set("X-Operation-Name", operationName)
			next.ServeHTTP(w, r)
		})
	}
}
