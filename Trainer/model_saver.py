"""
Model serialization utilities.
Saves trained model + all preprocessing artifacts.
"""
import os
import joblib
import pandas as pd


def save_model_artifacts(
    model, scaler, selector, encoder, label_mapping,
    selected_features, model_name, save_dir, metrics=None
):
    """
    Save model and all preprocessing artifacts to disk.

    Args:
        model: Trained model object.
        scaler: Fitted RobustScaler.
        selector: Fitted VarianceThreshold selector.
        encoder: Fitted LabelEncoder.
        label_mapping: Dict int -> label string.
        selected_features: List of feature column names.
        model_name: Name of the model (e.g., 'XGBoost').
        save_dir: Directory to save all files.
        metrics: Optional evaluation metrics dict.
    """
    os.makedirs(save_dir, exist_ok=True)

    joblib.dump(model, os.path.join(save_dir, "trained_model.joblib"))
    joblib.dump(scaler, os.path.join(save_dir, "scaler.joblib"))
    joblib.dump(selector, os.path.join(save_dir, "feature_selector.joblib"))
    joblib.dump(encoder, os.path.join(save_dir, "label_encoder.joblib"))

    config = {
        "model_type": model_name,
        "selected_features": selected_features,
        "label_mapping": label_mapping,
        "n_features": len(selected_features),
        "n_classes": len(label_mapping),
        "dataset": "CICIoT2023",
        "training_date": pd.Timestamp.now().strftime("%Y-%m-%d %H:%M:%S"),
    }
    if metrics:
        config["metrics"] = {
            k: v for k, v in metrics.items()
            if k != "confusion_matrix"
        }

    joblib.dump(config, os.path.join(save_dir, "model_config.joblib"))

    print(f"\n  Saved to: {save_dir}")
    print(f"    - trained_model.joblib")
    print(f"    - scaler.joblib")
    print(f"    - feature_selector.joblib")
    print(f"    - label_encoder.joblib")
    print(f"    - model_config.joblib")

    model_size = os.path.getsize(os.path.join(save_dir, "trained_model.joblib"))
    print(f"  Model size: {model_size / (1024 * 1024):.2f} MB")
