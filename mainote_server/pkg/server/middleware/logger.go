package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/qase-tms/qase-services/utils/logger"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	// Default the status code to 200, as this is what net/http defaults to
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logger(operation string) MiddlewareFunc {
	return func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), logger.CtxLoggerKey, logger.ContextInfo{
				TraceParent: r.Header.Get(logger.TraceParentKey),
				Endpoint:    operation,
			})

			start := time.Now()

			wrappedWriter := newResponseWriter(w)

			inner.ServeHTTP(wrappedWriter, r.WithContext(ctx))

			log.Info().
				Ctx(ctx).
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Str("user_agent", r.UserAgent()).
				Int("response_code", wrappedWriter.statusCode).
				Dur("duration", time.Since(start)).
				Send()
		})
	}
}
