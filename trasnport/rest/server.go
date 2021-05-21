package rest

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

/* Usage example with graceful shutdown
zapLogger, err := zap.NewDevelopment()
if err != nil {
	log.Fatal("failed to get logger with", err.Error())
}
l := zapLogger.Sugar()

srv := rest.NewServer("8080", http.NewServerMux())

errchan := make(chan error)
go rest.StartServer(srv, l, errchan)

doneChan := make(chan os.Signal, 1)
signal.Notify(doneChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
defer cancel()

select {
case oscall := <-doneChan:
	l.Infow("shuting servers down with", "sys_signal", oscall.String())
	rest.ShutdownServer(ctx, srv, l.With("server", "rest"))
case err := <-errchan:
	l.Info("rest server crashed with ", err.Error())
}
*/

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
