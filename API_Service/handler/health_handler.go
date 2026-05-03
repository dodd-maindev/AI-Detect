// Package handler - utility functions for HTTP responses.
package handler

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// HealthResponse is returned by the health check endpoint.
type HealthResponse struct {
	Status    string `json:"status"`
	Model     string `json:"model"`
	Dataset   string `json:"dataset"`
	GoVersion string `json:"go_version"`
	Uptime    string `json:"uptime"`
}

var startTime = time.Now()

// HandleHealth responds with system health information.
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:    "healthy",
		Model:     "XGBoost (CICIoT2023)",
		Dataset:   "CICIoT2023 - 34 classes",
		GoVersion: runtime.Version(),
		Uptime:    time.Since(startTime).Round(time.Second).String(),
	}
	writeJSON(w, http.StatusOK, resp)
}

// writeJSON encodes a value as JSON and writes it to the response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
