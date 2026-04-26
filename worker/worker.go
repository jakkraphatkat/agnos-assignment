package main

import (
	"context"
	"time"
)

func runWorker(ctx context.Context, cfg config, store *Store) error {
	initWorkerMetrics()

	run := func() error {
		now := time.Now().UTC()
		updated, err := store.TouchTodayRecords(now)
		if err != nil {
			workerRunsTotal.WithLabelValues("error").Inc()
			logJSON("error", "worker run failed", map[string]any{
				"error": err.Error(),
			})
			return err
		}

		workerRunsTotal.WithLabelValues("success").Inc()
		workerRecordsUpdatedTotal.Add(float64(updated))
		workerLastSuccessTimestampSeconds.Set(float64(now.Unix()))

		logJSON("info", "worker run completed", map[string]any{
			"updated":   updated,
			"timestamp": now.Format(time.RFC3339),
			"bucket":    cfg.MinIOBucket,
		})
		return nil
	}

	if err := run(); err != nil {
		return err
	}

	ticker := time.NewTicker(cfg.WorkerInterval)
	defer ticker.Stop()

	logJSON("info", "worker started", map[string]any{
		"interval": cfg.WorkerInterval.String(),
		"env":      cfg.AppEnv,
		"bucket":   cfg.MinIOBucket,
	})

	for {
		select {
		case <-ctx.Done():
			logJSON("info", "worker stopped", nil)
			return ctx.Err()
		case <-ticker.C:
			if err := run(); err != nil {
				return err
			}
		}
	}
}
