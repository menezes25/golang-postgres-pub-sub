package http

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

func NewServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    address,
		Handler: handler,
	}
}

func StartServer(srv *http.Server, l *zap.SugaredLogger, errchan chan error) {
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		l.Error("rest server failed with: ", err.Error())
		errchan <- err
	}
}

func ShutdownServer(ctx context.Context, srv *http.Server, l *zap.SugaredLogger) {
	if err := srv.Shutdown(ctx); err != nil {
		l.Error("rest server failed to shutdown with: ", err.Error())
	}

	l.Debug("server exited with success")
}
