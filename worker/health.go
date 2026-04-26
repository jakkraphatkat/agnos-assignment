package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func newMetricsRouter() *gin.Engine {
	initWorkerMetrics()

	router := gin.New()
	router.Use(workerHTTPMetricsMiddleware())
	router.Use(gin.Recovery())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return router
}
