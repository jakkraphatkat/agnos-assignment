package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type config struct {
	AppEnv            string
	WorkerInterval    time.Duration
	HealthPort        string
	MinIOEndpoint     string
	MinIOAccessKey    string
	MinIOSecretKey    string
	MinIOBucket       string
	MinIORegion       string
	MinIORecordPrefix string
	MinIOUseSSL       bool
}

type envDefaults struct {
	WorkerInterval    time.Duration
	HealthPort        string
	MinIOEndpoint     string
	MinIOAccessKey    string
	MinIOSecretKey    string
	MinIOBucket       string
	MinIORegion       string
	MinIORecordPrefix string
	MinIOUseSSL       bool
}

func loadConfig() (config, error) {
	appEnv, err := normalizeAppEnv(getEnv("APP_ENV", "dev"))
	if err != nil {
		return config{}, err
	}

	defaults := defaultsForEnv(appEnv)

	workerInterval, err := getDurationEnv("WORKER_INTERVAL", defaults.WorkerInterval)
	if err != nil {
		return config{}, err
	}

	useSSL, err := getBoolEnv("MINIO_USE_SSL", defaults.MinIOUseSSL)
	if err != nil {
		return config{}, err
	}

	return config{
		AppEnv:            appEnv,
		WorkerInterval:    workerInterval,
		HealthPort:        getEnv("HEALTH_PORT", defaults.HealthPort),
		MinIOEndpoint:     getEnv("MINIO_ENDPOINT", defaults.MinIOEndpoint),
		MinIOAccessKey:    getEnv("MINIO_ACCESS_KEY", defaults.MinIOAccessKey),
		MinIOSecretKey:    getEnv("MINIO_SECRET_KEY", defaults.MinIOSecretKey),
		MinIOBucket:       getEnv("MINIO_BUCKET", defaults.MinIOBucket),
		MinIORegion:       getEnv("MINIO_REGION", defaults.MinIORegion),
		MinIORecordPrefix: getEnv("MINIO_RECORD_PREFIX", defaults.MinIORecordPrefix),
		MinIOUseSSL:       useSSL,
	}, nil
}

func normalizeAppEnv(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "", "dev":
		return "dev", nil
	case "uat":
		return "uat", nil
	case "prod":
		return "prod", nil
	default:
		return "", fmt.Errorf("APP_ENV must be one of dev, uat, prod, got %q", value)
	}
}

func defaultsForEnv(appEnv string) envDefaults {
	switch appEnv {
	case "uat":
		return envDefaults{
			WorkerInterval:    30 * time.Second,
			HealthPort:        "8081",
			MinIOEndpoint:     "minio:9000",
			MinIOAccessKey:    "uat-minio-user",
			MinIOSecretKey:    "uat-minio-password",
			MinIOBucket:       "agnos-uat-records",
			MinIORegion:       "us-east-1",
			MinIORecordPrefix: "uat/records",
			MinIOUseSSL:       false,
		}
	case "prod":
		return envDefaults{
			WorkerInterval:    time.Minute,
			HealthPort:        "8081",
			MinIOEndpoint:     "minio:9000",
			MinIOAccessKey:    "prod-minio-user",
			MinIOSecretKey:    "prod-minio-password",
			MinIOBucket:       "agnos-prod-records",
			MinIORegion:       "us-east-1",
			MinIORecordPrefix: "prod/records",
			MinIOUseSSL:       false,
		}
	default:
		return envDefaults{
			WorkerInterval:    15 * time.Second,
			HealthPort:        "8081",
			MinIOEndpoint:     "localhost:9000",
			MinIOAccessKey:    "minioadmin",
			MinIOSecretKey:    "minioadmin",
			MinIOBucket:       "agnos-dev-records",
			MinIORegion:       "us-east-1",
			MinIORecordPrefix: "dev/records",
			MinIOUseSSL:       false,
		}
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func getDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}

	return parsed, nil
}

func getBoolEnv(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a valid boolean: %w", key, err)
	}

	return parsed, nil
}
