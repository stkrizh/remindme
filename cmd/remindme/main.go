package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"remindme/internal/app"
	"remindme/internal/app/consumers"
	"remindme/internal/app/deps"
	"remindme/internal/app/services"
	"syscall"
	"time"

	dl "remindme/internal/core/domain/logging"
)

func main() {
	deps, shutdownDeps := deps.InitDeps()
	defer shutdownDeps()

	services := services.InitServices(deps)

	shutdownConsumers := consumers.InitConsumers(deps, services)
	defer shutdownConsumers()

	httpServer := app.InitHttpServer(deps, services)
	go start(httpServer, deps)

	stopCh, closeCh := createChannel()
	defer closeCh()

	<-stopCh
	shutdown(context.Background(), httpServer)

	deps.Logger.Info(context.Background(), "Shutdown completed.")
}

func createChannel() (chan os.Signal, func()) {
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	return stopCh, func() {
		close(stopCh)
	}
}

func start(server *http.Server, deps *deps.Deps) {
	deps.Logger.Info(
		context.Background(),
		"HTTP server has started.",
		dl.Entry("address", server.Addr),
		dl.Entry("isTestMode", deps.Config.IsTestMode),
		dl.Entry("telegramBots", deps.Config.TelegramBots),
	)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	} else {
		deps.Logger.Info(context.Background(), "HTTP server is stopping gracefully.")
	}
}

func shutdown(ctx context.Context, server *http.Server) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}
}
