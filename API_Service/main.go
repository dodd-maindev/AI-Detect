// NIDS API Server - XGBoost inference in pure Go.
// Uses leaves library for model loading, no CGO required.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"nids-api/config"
	"nids-api/handler"
	"nids-api/predictor"
	"nids-api/preprocessor"
	"nids-api/sniffer"
)

func main() {
	// Khai báo flag cấu hình chạy bằng dòng lệnh
	modeFlag := flag.String("mode", "active", "Operation mode: 'active' (block IPs) or 'passive' (detect only)")
	flag.Parse()

	fmt.Println("============================================================")
	fmt.Println("  NIDS API Server v2.0 (Go + XGBoost)")
	fmt.Println("  Dataset: CICIoT2023 | 34 Attack Classes")
	fmt.Println("============================================================")

	paths := config.GetModelPaths()

	fmt.Println("\n  Loading preprocessor...")
	preproc, err := preprocessor.NewPreprocessor(paths.ScalerFile, paths.FeatureFile)
	if err != nil {
		log.Fatalf("  Failed to load preprocessor: %v", err)
	}
	fmt.Println("  ✓ Preprocessor loaded")

	fmt.Println("  Loading XGBoost model...")
	pred, err := predictor.NewPredictor(paths.ModelFile, paths.LabelMappingFile)
	if err != nil {
		log.Fatalf("  Failed to load model: %v", err)
	}
	fmt.Println("  ✓ Model loaded")

	// Đọc cấu hình từ cờ dòng lệnh (flags)
	activeIPS := false
	if *modeFlag == "active" {
		activeIPS = true
	}

	predictHandler := handler.NewPredictHandler(preproc, pred)

	snifferConfig := sniffer.Config{
		Interface: "ens33",
		TargetIP:  "10.203.152.105",
		ActiveIPS: activeIPS,
	}
	ipsAgent := sniffer.NewIPSCore(snifferConfig, preproc, pred)
	go ipsAgent.Start()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.HandleHealth)
	mux.HandleFunc("/predict", predictHandler.HandlePredict)
	mux.HandleFunc("/predict/batch", predictHandler.HandleBatch)

	fmt.Printf("\n  Server listening on %s\n", config.ServerPort)
	fmt.Println("  Endpoints:")
	fmt.Println("    GET  /health         - Health check")
	fmt.Println("    POST /predict        - Single prediction")
	fmt.Println("    POST /predict/batch  - Batch predictions")
	fmt.Println("============================================================")

	if err := http.ListenAndServe(config.ServerPort, mux); err != nil {
		log.Fatalf("  Server failed: %v", err)
	}
}
