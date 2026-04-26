package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBackendHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newRouter(config{
		AppEnv:      "dev",
		MinIOBucket: "agnos-dev-records",
	}, nil)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", response["status"])
	}

	if response["service"] != "backend" {
		t.Fatalf("expected service backend, got %q", response["service"])
	}

	if response["environment"] != "dev" {
		t.Fatalf("expected environment dev, got %q", response["environment"])
	}

	if response["bucket"] != "agnos-dev-records" {
		t.Fatalf("expected bucket agnos-dev-records, got %q", response["bucket"])
	}
}

func TestBackendMock500Endpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newRouter(config{
		AppEnv:      "dev",
		MinIOBucket: "agnos-dev-records",
	}, nil)

	request := httptest.NewRequest(http.MethodGet, "/mock/500", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["error"] != "mock internal server error" {
		t.Fatalf("expected mock error message, got %q", response["error"])
	}

	if response["service"] != "backend" {
		t.Fatalf("expected service backend, got %q", response["service"])
	}
}
