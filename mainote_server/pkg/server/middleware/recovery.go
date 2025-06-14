package middleware

import (
	"context"
	"net/http"
	"reflect"
	"runtime/debug"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
)

// Custom context keys and types to replace logutil dependency
type contextKey string

const (
	CtxErrInfoKey contextKey = "errInfo"
)

// ErrInfo holds error information for logging
type ErrInfo struct {
	IsPanic bool
	ErrType string
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rc := recover(); rc != nil {
				var err error

				switch v := rc.(type) {
				case error:
					err = v
				case string:
					err = errors.New(v)
				default:
					err = errors.Errorf("error happened (%v)", v)
				}

				ctx := context.WithValue(r.Context(), CtxErrInfoKey, ErrInfo{
					IsPanic: true,
					ErrType: reflect.TypeOf(err).String(),
				})

				log.Error().
					Ctx(ctx).
					Err(err).
					Str("stacktrace", string(debug.Stack())).
					Send()

				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
