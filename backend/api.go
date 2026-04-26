package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type createRecordRequest struct {
	Name string `json:"name" binding:"required,max=120"`
}

func newRouter(cfg config, store *Store) *gin.Engine {
	initBackendMetrics()

	router := gin.New()
	router.Use(backendMetricsMiddleware())
	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logMap := map[string]any{
			"timestamp": param.TimeStamp.Format(time.RFC3339),
			"status":    param.StatusCode,
			"latency":   param.Latency.String(),
			"client_ip": param.ClientIP,
			"method":    param.Method,
			"path":      param.Path,
			"error":     param.ErrorMessage,
		}

		jsonLog, _ := json.Marshal(logMap)
		return string(jsonLog) + "\n"
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"environment": cfg.AppEnv,
			"service":     "backend",
			"bucket":      cfg.MinIOBucket,
			"time":        time.Now().UTC().Format(time.RFC3339),
		})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/mock/500", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "mock internal server error",
			"service": "backend",
		})
	})

	router.GET("/records", func(c *gin.Context) {
		records, err := store.ListRecords()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load records"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"records": records,
			"count":   len(records),
		})
	})

	router.POST("/records", func(c *gin.Context) {
		var request createRecordRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		record, err := store.CreateRecord(strings.TrimSpace(request.Name))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create record"})
			return
		}

		c.JSON(http.StatusCreated, record)
	})

	return router
}
