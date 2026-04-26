package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerBackendMetricsOnce sync.Once

	backendHTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agnos_http_requests_total",
			Help: "Total number of HTTP requests handled by the backend service.",
		},
		[]string{"service", "method", "path", "status"},
	)
	backendHTTPDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "agnos_http_request_duration_seconds",
			Help:    "HTTP request duration for the backend service.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path", "status"},
	)
)

func initBackendMetrics() {
	registerBackendMetricsOnce.Do(func() {
		prometheus.MustRegister(backendHTTPRequestsTotal, backendHTTPDurationSeconds)
	})
}

func backendMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		labels := []string{"backend", c.Request.Method, path, status}
		backendHTTPRequestsTotal.WithLabelValues(labels...).Inc()
		backendHTTPDurationSeconds.WithLabelValues(labels...).Observe(time.Since(startedAt).Seconds())
	}
}
