// Package preprocessor implements RobustScaler and feature selection in pure Go.
package preprocessor

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
)

// ScalerParams holds RobustScaler center (median) and scale (IQR) values.
type ScalerParams struct {
	Center []float64 `json:"center"`
	Scale  []float64 `json:"scale"`
}

// FeatureSelector holds the indices of selected features.
type FeatureSelector struct {
	SelectedIndices []int `json:"selected_indices"`
}

// Preprocessor applies feature selection and scaling to raw input.
type Preprocessor struct {
	scaler   ScalerParams
	selector FeatureSelector
}

// NewPreprocessor loads scaler and selector parameters from JSON files.
func NewPreprocessor(scalerPath, selectorPath string) (*Preprocessor, error) {
	scaler, err := loadScaler(scalerPath)
	if err != nil {
		return nil, fmt.Errorf("loading scaler: %w", err)
	}

	selector, err := loadSelector(selectorPath)
	if err != nil {
		return nil, fmt.Errorf("loading selector: %w", err)
	}

	return &Preprocessor{scaler: scaler, selector: selector}, nil
}

// Transform applies feature selection then scaling to a single sample.
func (p *Preprocessor) Transform(rawFeatures []float64) []float64 {
	selected := p.selectFeatures(rawFeatures)
	return p.scaleFeatures(selected)
}

// selectFeatures picks only the selected feature indices.
func (p *Preprocessor) selectFeatures(raw []float64) []float64 {
	result := make([]float64, len(p.selector.SelectedIndices))
	for i, idx := range p.selector.SelectedIndices {
		if idx < len(raw) {
			result[i] = sanitize(raw[idx])
		}
	}
	return result
}

// scaleFeatures applies RobustScaler: (x - center) / scale.
func (p *Preprocessor) scaleFeatures(features []float64) []float64 {
	result := make([]float64, len(features))
	for i, val := range features {
		if i < len(p.scaler.Scale) && p.scaler.Scale[i] != 0 {
			result[i] = (val - p.scaler.Center[i]) / p.scaler.Scale[i]
		}
	}
	return result
}

// sanitize replaces NaN and Inf with 0.
func sanitize(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0.0
	}
	return v
}

func loadScaler(path string) (ScalerParams, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ScalerParams{}, err
	}
	var params ScalerParams
	return params, json.Unmarshal(data, &params)
}

func loadSelector(path string) (FeatureSelector, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FeatureSelector{}, err
	}
	var selector FeatureSelector
	return selector, json.Unmarshal(data, &selector)
}
