package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"confluence-mcp/internal/auth"
	"confluence-mcp/internal/config"
	"confluence-mcp/internal/confluence"
	"confluence-mcp/internal/handlers"
	"confluence-mcp/internal/policy"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to YAML config")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	policies := make(map[string][]string, len(cfg.AccessControl.Policies))
	for _, p := range cfg.AccessControl.Policies {
		policies[p.Subject] = p.AllowedSpaces
	}

	api := &handlers.API{
		Client:      confluence.NewClient(cfg.Confluence.BaseURL, cfg.Confluence.AuthToken),
		Auth:        &auth.Authenticator{SharedSecret: []byte(cfg.AccessControl.SharedSecret)},
		PolicyStore: policy.NewStore(cfg.AccessControl.DefaultSpaces, policies),
	}

	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      api.Router(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("starting MCP Confluence gateway on %s", cfg.Server.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received, draining")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}

	log.Println("server stopped")
}
