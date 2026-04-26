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
	defer store.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	healthServer := &http.Server{
		Addr:              ":" + cfg.HealthPort,
		Handler:           newMetricsRouter(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := healthServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logJSON("error", "worker metrics server shutdown failed", map[string]any{
				"error": err.Error(),
			})
		}
	}()

	go func() {
		logJSON("info", "worker metrics server starting", map[string]any{
			"addr": ":" + cfg.HealthPort,
			"env":  cfg.AppEnv,
		})

		if err := healthServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logJSON("error", "worker metrics server failed", map[string]any{
				"error": err.Error(),
			})
		}
	}()

	if err := runWorker(ctx, cfg, store); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}
