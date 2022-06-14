package fly

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/3vilive/fly/pkg/flylog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	viper.SetDefault("http-server.graceful-shutdown-timeout", 5*time.Second)
}

func RunHttpServer(server *http.Server) error {
	// catch terminate signals
	termSignal := make(chan os.Signal, 1)
	signal.Notify(termSignal, syscall.SIGINT, syscall.SIGTERM)

	// graceful shutdown
	waitShutdown := make(chan bool, 1)
	go func() {
		<-termSignal

		shutdownTimeout := viper.GetDuration("http-server.graceful-shutdown-timeout")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			flylog.Error("graceful shutdown server error", zap.Error(err))
		}

		waitShutdown <- true
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return errors.Wrap(err, "listen and serve error")
	}

	<-waitShutdown

	return nil
}
