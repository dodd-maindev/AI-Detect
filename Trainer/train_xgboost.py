"""
XGBoost trainer for CICIoT2023.
Uses GPU acceleration (3050ti) with cuda tree_method.
"""
import gc
from xgboost import XGBClassifier
from config import XGBOOST_PARAMS, XGBOOST_DIR
from evaluator import evaluate_model
from model_saver import save_model_artifacts


def train_xgboost(x_train, y_train, x_test, y_test, label_mapping,
                   scaler, selector, encoder, selected_features):
    """
    Train XGBoost classifier with GPU acceleration.

    Args:
        x_train: Scaled training features.
        y_train: Encoded training labels.
        x_test: Scaled test features.
        y_test: Encoded test labels.
        label_mapping: Dict int -> label string.
        scaler: Fitted scaler for serialization.
        selector: Fitted feature selector.
        encoder: Fitted label encoder.
        selected_features: List of feature names.

    Returns:
        Trained XGBClassifier model.
    """
    print("\n" + "=" * 60)
    print("  TRAINING: XGBoost (GPU: cuda)")
    print("=" * 60)

    n_classes = len(label_mapping)
    params = {**XGBOOST_PARAMS, "num_class": n_classes}

    print(f"  n_estimators: {params['n_estimators']}")
    print(f"  max_depth:    {params['max_depth']}")
    print(f"  device:       {params['device']}")
    print(f"  Classes:      {n_classes}")

    model = XGBClassifier(**params)
    model.fit(x_train, y_train)

    metrics = evaluate_model(model, x_test, y_test, label_mapping, "XGBoost")

    save_model_artifacts(
        model=model, scaler=scaler, selector=selector,
        encoder=encoder, label_mapping=label_mapping,
        selected_features=selected_features,
        model_name="XGBoost", save_dir=XGBOOST_DIR,
        metrics=metrics,
    )

    gc.collect()
    return model
