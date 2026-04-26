package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerWorkerMetricsOnce sync.Once

	workerHTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agnos_http_requests_total",
			Help: "Total number of HTTP requests handled by the worker service.",
		},
		[]string{"service", "method", "path", "status"},
	)
	workerHTTPDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "agnos_http_request_duration_seconds",
			Help:    "HTTP request duration for the worker service.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path", "status"},
	)

	workerRunsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agnos_worker_runs_total",
			Help: "Total number of scheduled worker runs grouped by status.",
		},
		[]string{"status"},
	)
	workerRecordsUpdatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "agnos_worker_records_updated_total",
			Help: "Total number of records updated by the worker.",
		},
	)
	workerLastSuccessTimestampSeconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "agnos_worker_last_success_timestamp_seconds",
			Help: "Unix timestamp of the last successful worker run.",
		},
	)
)

func initWorkerMetrics() {
	registerWorkerMetricsOnce.Do(func() {
		prometheus.MustRegister(
			workerHTTPRequestsTotal,
			workerHTTPDurationSeconds,
			workerRunsTotal,
			workerRecordsUpdatedTotal,
			workerLastSuccessTimestampSeconds,
		)
	})
}

func workerHTTPMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		labels := []string{"worker", c.Request.Method, path, status}
		workerHTTPRequestsTotal.WithLabelValues(labels...).Inc()
		workerHTTPDurationSeconds.WithLabelValues(labels...).Observe(time.Since(startedAt).Seconds())
	}
}
