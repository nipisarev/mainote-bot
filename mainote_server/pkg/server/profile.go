package server

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/cockroachdb/errors"
)

func Profile(port uint16) error {
	handler := http.NewServeMux()
	handler.HandleFunc("/debug/pprof/", pprof.Index)
	handler.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	handler.HandleFunc("/debug/pprof/profile", pprof.Profile)
	handler.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	handler.HandleFunc("/debug/pprof/trace", pprof.Trace)

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	return errors.WithStack(srv.ListenAndServe())
}
