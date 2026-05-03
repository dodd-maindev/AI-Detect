// Package predictor implements XGBoost inference from JSON model dump.
// Pure Go, zero external dependencies - parses XGBoost JSON tree format.
package predictor

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
)

// XGBModel holds the parsed XGBoost ensemble.
type XGBModel struct {
	Trees      []tree
	NumClasses int
	BaseScore  float64
}

type tree struct {
	SplitIndices    []int     `json:"split_indices"`
	SplitConditions []float64 `json:"split_conditions"`
	LeftChildren    []int     `json:"left_children"`
	RightChildren   []int     `json:"right_children"`
	DefaultLeft     []int     `json:"default_left"`
	BaseWeights     []float64 `json:"base_weights"`
}

type xgbJSON struct {
	Learner struct {
		GradientBooster struct {
			Model struct {
				Trees []tree `json:"trees"`
				Param struct {
					NumTrees string `json:"num_trees"`
				} `json:"gbtree_model_param"`
			} `json:"model"`
		} `json:"gradient_booster"`
		Param struct {
			NumClass  string `json:"num_class"`
			BaseScore string `json:"base_score"`
		} `json:"learner_model_param"`
	} `json:"learner"`
}

// LoadXGBModel loads an XGBoost model from JSON dump file.
func LoadXGBModel(path string) (*XGBModel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading model file: %w", err)
	}

	var raw xgbJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing model JSON: %w", err)
	}

	numClasses := 0
	fmt.Sscanf(raw.Learner.Param.NumClass, "%d", &numClasses)

	baseScore := 0.5
	fmt.Sscanf(raw.Learner.Param.BaseScore, "%f", &baseScore)

	model := &XGBModel{
		Trees:      raw.Learner.GradientBooster.Model.Trees,
		NumClasses: numClasses,
		BaseScore:  baseScore,
	}

	fmt.Printf("  Loaded: %d trees, %d classes\n", len(model.Trees), numClasses)
	return model, nil
}

// Predict runs inference on a single feature vector. Returns raw scores.
func (m *XGBModel) Predict(features []float64) []float64 {
	scores := make([]float64, m.NumClasses)

	for i, t := range m.Trees {
		classIdx := i % m.NumClasses
		leafValue := traverseTree(&t, features)
		scores[classIdx] += leafValue
	}

	return scores
}

// traverseTree walks a single decision tree and returns the leaf value.
func traverseTree(t *tree, features []float64) float64 {
	nodeIdx := 0
	for {
		left := t.LeftChildren[nodeIdx]
		if left == -1 {
			return t.BaseWeights[nodeIdx]
		}
		splitFeature := t.SplitIndices[nodeIdx]
		threshold := t.SplitConditions[nodeIdx]
		featureVal := safeGet(features, splitFeature)

		if math.IsNaN(featureVal) {
			if t.DefaultLeft[nodeIdx] == 1 {
				nodeIdx = left
			} else {
				nodeIdx = t.RightChildren[nodeIdx]
			}
		} else if featureVal < threshold {
			nodeIdx = left
		} else {
			nodeIdx = t.RightChildren[nodeIdx]
		}
	}
}

func safeGet(features []float64, idx int) float64 {
	if idx < len(features) {
		return features[idx]
	}
	return math.NaN()
}
