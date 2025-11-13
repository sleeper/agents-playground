package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/example/agents-playground/internal/config"
	"github.com/example/agents-playground/internal/http/transport"
	"github.com/example/agents-playground/internal/logging"
	"github.com/example/agents-playground/internal/storage/sqlite"
)

func main() {
	logging.Configure()
	cfg := config.Load()

	store, err := sqlite.Open(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}
	defer store.Close()

	router := transport.NewRouter(cfg, store)

	srv := &http.Server{
		Addr:         cfg.HTTPAddress,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("address", cfg.HTTPAddress).Msg("starting http server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	waitForShutdown(srv)
}

func waitForShutdown(srv *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Info().Msg("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
}
