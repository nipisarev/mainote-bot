package middleware

import (
	"net/http"
)

type MiddlewareFunc func(http.Handler) http.Handler

// ApplyMiddlewares wraps a http.Handler with the provided middlewares.
// The middlewares are applied in the reverse order to preserve the order of execution.
func ApplyMiddlewares(h http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	if len(middlewares) == 0 {
		return h
	}

	wrapped := h
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}

	return wrapped
}
