package middleware

import (
	"context"
	"net/http"
	"reflect"
	"runtime/debug"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"

	logutil "github.com/qase-tms/qase-services/utils/logger"
)

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
					err = errors.Errorf("error happend (%v)", v)
				}

				ctx := context.WithValue(r.Context(), logutil.CtxErrInfoKey, logutil.ErrInfo{
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
