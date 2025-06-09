package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const serverShutdownTimeout = 5 * time.Second

func ServeHTTP(
	ctx context.Context,
	handler http.Handler,
	port uint16,
	group *errgroup.Group,
	groupCtx context.Context,
) *http.Server {
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	group.Go(func() error {
		log.Info().Msgf("http server is listening at %s", httpServer.Addr)
		return errors.WithStack(httpServer.ListenAndServe())
	})

	group.Go(func() error {
		<-groupCtx.Done()
		tCtx, cancel := context.WithTimeout(groupCtx, serverShutdownTimeout)
		defer cancel()
		return errors.WithStack(httpServer.Shutdown(tCtx))
	})

	return httpServer
}

func ServeGRPC(
	registerFunc func(server *grpc.Server),
	port uint16,
	group *errgroup.Group,
	groupCtx context.Context,
) *grpc.Server {
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	registerFunc(grpcServer)

	group.Go(func() error {
		listener, err := net.Listen(
			"tcp",
			fmt.Sprintf(":%d", port),
		)
		if err != nil {
			return errors.WithStack(err)
		}
		log.Info().Msgf("grpc server is listening at %s", listener.Addr())

		return errors.WithStack(grpcServer.Serve(listener))
	})

	group.Go(func() error {
		<-groupCtx.Done()
		tCtx, cancel := context.WithTimeout(groupCtx, serverShutdownTimeout)
		defer cancel()

		stopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(stopped)
		}()

		select {
		case <-tCtx.Done():
			return tCtx.Err()
		case <-stopped:
		}

		return nil
	})

	return grpcServer
}
