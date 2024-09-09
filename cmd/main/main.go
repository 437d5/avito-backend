package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"zadanie-6105/internal/config"
	"zadanie-6105/internal/server/handlers"
	"zadanie-6105/internal/server/middleware/logger"
	"zadanie-6105/internal/storage/postgres"

	"github.com/gorilla/mux"
)

func main() {
	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Error(fmt.Errorf("cannot init config: %w", err).Error())
		os.Exit(1)
	}
	log.Info("config was initialized succesfully")
	log.Debug("debug messages are enabled")

	storage, err := postgres.New(ctx, *cfg)
	if err != nil {
		log.Error(fmt.Errorf("failed to init storage: %s", err).Error())
		os.Exit(1)
	}

	handler := handlers.New(storage)

	r := mux.NewRouter()
	apiRouter := r.PathPrefix("/api").Subrouter()

	apiRouter.Use(logger.New(log))

	apiRouter.HandleFunc("/ping", handler.PingHandler)

	server := &http.Server{
		Addr: cfg.SERVER_ADDRESS,
		Handler: apiRouter,
	}

	ch := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err = <-ch:
		log.Error(err.Error())
		os.Exit(1)
	case <- ctx.Done():
		timeout, cancel := context.WithTimeout(ctx, time.Second * 10)
		defer cancel()
		log.Warn(server.Shutdown(timeout).Error())
		os.Exit(0)
	}
}