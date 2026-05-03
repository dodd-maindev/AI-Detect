// Package predictor wraps the XGBoost model for inference.
package predictor

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
)

// Predictor holds the loaded XGBoost model and label mapping.
type Predictor struct {
	model        *XGBModel
	labelMapping map[int]string
}

// NewPredictor loads the XGBoost model and label mapping.
func NewPredictor(modelPath, labelMappingPath string) (*Predictor, error) {
	model, err := LoadXGBModel(modelPath)
	if err != nil {
		return nil, fmt.Errorf("loading XGBoost model: %w", err)
	}

	mapping, err := loadLabelMapping(labelMappingPath)
	if err != nil {
		return nil, fmt.Errorf("loading label mapping: %w", err)
	}

	return &Predictor{model: model, labelMapping: mapping}, nil
}

// PredictionResult contains the classification output.
type PredictionResult struct {
	Label      string  `json:"label"`
	ClassIndex int     `json:"class_index"`
	Confidence float64 `json:"confidence"`
}

// Predict classifies a single preprocessed feature vector.
func (p *Predictor) Predict(features []float64) PredictionResult {
	scores := p.model.Predict(features)
	bestIdx := argmax(scores)
	confidence := softmaxConfidence(scores, bestIdx)

	return PredictionResult{
		Label:      p.labelMapping[bestIdx],
		ClassIndex: bestIdx,
		Confidence: confidence,
	}
}

func argmax(scores []float64) int {
	bestIdx := 0
	for i := 1; i < len(scores); i++ {
		if scores[i] > scores[bestIdx] {
			bestIdx = i
		}
	}
	return bestIdx
}

func softmaxConfidence(scores []float64, idx int) float64 {
	maxVal := scores[0]
	for _, v := range scores[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	sumExp := 0.0
	for _, v := range scores {
		sumExp += math.Exp(v - maxVal)
	}
	return math.Exp(scores[idx]-maxVal) / sumExp
}

func loadLabelMapping(path string) (map[int]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	mapping := make(map[int]string, len(raw))
	keys := make([]int, 0, len(raw))
	for k := range raw {
		var idx int
		fmt.Sscanf(k, "%d", &idx)
		keys = append(keys, idx)
	}
	sort.Ints(keys)
	for _, idx := range keys {
		mapping[idx] = raw[fmt.Sprintf("%d", idx)]
	}
	return mapping, nil
}
