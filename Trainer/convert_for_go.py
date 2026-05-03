"""
Convert trained XGBoost model to native format for Go inference.
Exports: model binary, scaler params, feature selector, label mapping.
"""
import os
import json
import joblib
import numpy as np

MODEL_DIR = r"C:\New folder (2)\Redo_ATTT\Model\XGBoost"
EXPORT_DIR = r"C:\New folder (2)\Redo_ATTT\Model\XGBoost\go_export"


def export_model():
    """Save XGBoost model in JSON format for Go leaves library."""
    os.makedirs(EXPORT_DIR, exist_ok=True)

    model = joblib.load(os.path.join(MODEL_DIR, "trained_model.joblib"))
    # Save as JSON format (leaves library compatible)
    export_path = os.path.join(EXPORT_DIR, "xgboost_model.json")
    model.save_model(export_path)
    print(f"  Model exported: xgboost_model.json")


def export_scaler():
    """Export RobustScaler parameters (center_, scale_) as JSON."""
    scaler = joblib.load(os.path.join(MODEL_DIR, "scaler.joblib"))

    scaler_params = {
        "center": scaler.center_.tolist(),
        "scale": scaler.scale_.tolist(),
    }
    path = os.path.join(EXPORT_DIR, "scaler_params.json")
    with open(path, "w") as f:
        json.dump(scaler_params, f, indent=2)
    print(f"  Scaler exported: scaler_params.json ({len(scaler.center_)} features)")


def export_feature_selector():
    """Export VarianceThreshold selected column indices as JSON."""
    selector = joblib.load(os.path.join(MODEL_DIR, "feature_selector.joblib"))

    indices = np.where(selector.get_support())[0].tolist()
    path = os.path.join(EXPORT_DIR, "feature_indices.json")
    with open(path, "w") as f:
        json.dump({"selected_indices": indices}, f, indent=2)
    print(f"  Feature selector exported: {len(indices)} features selected")


def export_label_mapping():
    """Export label encoder mapping as JSON."""
    config = joblib.load(os.path.join(MODEL_DIR, "model_config.joblib"))
    encoder = joblib.load(os.path.join(MODEL_DIR, "label_encoder.joblib"))

    mapping = {str(i): label for i, label in enumerate(encoder.classes_)}
    path = os.path.join(EXPORT_DIR, "label_mapping.json")
    with open(path, "w") as f:
        json.dump(mapping, f, indent=2)
    print(f"  Label mapping exported: {len(mapping)} classes")


def main():
    """Run all export steps."""
    print("=" * 60)
    print("  EXPORTING MODEL FOR GO API")
    print("=" * 60)

    export_model()
    export_scaler()
    export_feature_selector()
    export_label_mapping()

    print(f"\n  All files saved to: {EXPORT_DIR}")
    print("  Ready for Go API consumption!")


if __name__ == "__main__":
    main()
