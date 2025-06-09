package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func Tracing(operation string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		h := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oteltrace.
				SpanFromContext(r.Context()).
				SetAttributes(
					semconv.URLPath(r.URL.Path),
					semconv.URLQuery(r.URL.RawQuery),
				)
			next.ServeHTTP(w, r)
		}))

		return otelhttp.NewHandler(h, operation)
	}
}
