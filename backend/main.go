package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	store, err := openStore(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if closeErr := store.Close(); closeErr != nil {
			logJSON("error", "backend store close failed", map[string]any{
				"error": closeErr.Error(),
			})
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	router := newRouter(cfg, store)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logJSON("error", "server shutdown failed", map[string]any{
				"error": err.Error(),
			})
		}
	}()

	logJSON("info", "backend server starting", map[string]any{
		"addr": ":" + cfg.Port,
		"env":  cfg.AppEnv,
	})

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
