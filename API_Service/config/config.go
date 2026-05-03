// Package config provides centralized configuration for the NIDS API.
package config

import "path/filepath"

// ModelDir is the directory containing exported model artifacts.
var ModelDir = filepath.Join("..", "Model", "XGBoost", "go_export")

// ServerPort defines the port the API server listens on.
const ServerPort = ":5000"

// ModelPaths holds paths to all model artifacts.
type ModelPaths struct {
	ModelFile       string
	ScalerFile      string
	FeatureFile     string
	LabelMappingFile string
}

// GetModelPaths returns full paths to all model artifacts.
func GetModelPaths() ModelPaths {
	return ModelPaths{
		ModelFile:        filepath.Join(ModelDir, "model.json"),
		ScalerFile:       filepath.Join(ModelDir, "scaler_params.json"),
		FeatureFile:      filepath.Join(ModelDir, "feature_indices.json"),
		LabelMappingFile: filepath.Join(ModelDir, "label_mapping.json"),
	}
}

// AllFeatureNames lists all 46 CICIoT2023 feature columns in order.
var AllFeatureNames = []string{
	"flow_duration", "Header_Length", "Protocol Type", "Duration",
	"Rate", "Srate", "Drate", "fin_flag_number", "syn_flag_number",
	"rst_flag_number", "psh_flag_number", "ack_flag_number",
	"ece_flag_number", "cwr_flag_number", "ack_count", "syn_count",
	"fin_count", "urg_count", "rst_count",
	"HTTP", "HTTPS", "DNS", "Telnet", "SMTP", "SSH", "IRC",
	"TCP", "UDP", "DHCP", "ARP", "ICMP", "IPv", "LLC",
	"Tot sum", "Min", "Max", "AVG", "Std", "Tot size", "IAT",
	"Number", "Magnitue", "Radius", "Covariance", "Variance", "Weight",
}
