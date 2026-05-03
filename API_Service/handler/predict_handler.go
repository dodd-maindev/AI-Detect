// Package handler implements HTTP endpoint handlers for the NIDS API.
package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"nids-api/predictor"
	"nids-api/preprocessor"
)

// PredictHandler handles realtime prediction requests.
type PredictHandler struct {
	preprocessor *preprocessor.Preprocessor
	predictor    *predictor.Predictor
}

// NewPredictHandler creates a handler with loaded model and preprocessor.
func NewPredictHandler(p *preprocessor.Preprocessor, m *predictor.Predictor) *PredictHandler {
	return &PredictHandler{preprocessor: p, predictor: m}
}

// PredictRequest is the JSON body for single-sample prediction.
type PredictRequest struct {
	Features []float64 `json:"features"`
}

// PredictResponse is the JSON response with classification result.
type PredictResponse struct {
	Label      string  `json:"label"`
	ClassIndex int     `json:"class_index"`
	Confidence float64 `json:"confidence"`
	LatencyMs  float64 `json:"latency_ms"`
}

// BatchRequest holds multiple samples for batch prediction.
type BatchRequest struct {
	Samples [][]float64 `json:"samples"`
}

// BatchResponse holds results for all samples in a batch.
type BatchResponse struct {
	Results   []PredictResponse `json:"results"`
	Count     int               `json:"count"`
	TotalMs   float64           `json:"total_ms"`
	AvgMs     float64           `json:"avg_ms"`
}

// HandlePredict handles POST /predict for a single sample.
func (h *PredictHandler) HandlePredict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PredictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	start := time.Now()
	scaled := h.preprocessor.Transform(req.Features)
	result := h.predictor.Predict(scaled)
	elapsed := time.Since(start)

	resp := PredictResponse{
		Label:      result.Label,
		ClassIndex: result.ClassIndex,
		Confidence: result.Confidence,
		LatencyMs:  float64(elapsed.Microseconds()) / 1000.0,
	}
	writeJSON(w, http.StatusOK, resp)
}

// HandleBatch handles POST /predict/batch for multiple samples.
func (h *PredictHandler) HandleBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	start := time.Now()
	results := make([]PredictResponse, len(req.Samples))
	for i, sample := range req.Samples {
		scaled := h.preprocessor.Transform(sample)
		result := h.predictor.Predict(scaled)
		results[i] = PredictResponse{
			Label:      result.Label,
			ClassIndex: result.ClassIndex,
			Confidence: result.Confidence,
		}
	}
	totalMs := float64(time.Since(start).Microseconds()) / 1000.0

	writeJSON(w, http.StatusOK, BatchResponse{
		Results: results,
		Count:   len(results),
		TotalMs: totalMs,
		AvgMs:   totalMs / float64(len(results)),
	})
}
