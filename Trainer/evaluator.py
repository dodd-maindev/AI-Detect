"""
Model evaluation utilities.
Computes metrics, classification report, and confusion matrix.
"""
import time
import numpy as np
from sklearn.metrics import (
    accuracy_score, precision_score, recall_score,
    f1_score, classification_report, confusion_matrix,
)


def evaluate_model(model, x_test, y_test, label_mapping, model_name="Model"):
    """
    Evaluate a trained model on test data.

    Args:
        model: Trained sklearn/xgboost model.
        x_test: Test features (scaled numpy array).
        y_test: Test labels (encoded integers).
        label_mapping: Dict mapping int -> label string.
        model_name: Display name for logging.

    Returns:
        Dict of evaluation metrics.
    """
    print(f"\n{'=' * 60}")
    print(f"  EVALUATING: {model_name}")
    print(f"{'=' * 60}")

    start = time.time()
    y_pred = model.predict(x_test)
    pred_time = time.time() - start

    accuracy = accuracy_score(y_test, y_pred)
    precision = precision_score(y_test, y_pred, average="weighted", zero_division=0)
    recall = recall_score(y_test, y_pred, average="weighted", zero_division=0)
    f1 = f1_score(y_test, y_pred, average="weighted", zero_division=0)

    print(f"  Accuracy:   {accuracy:.4f}")
    print(f"  Precision:  {precision:.4f}")
    print(f"  Recall:     {recall:.4f}")
    print(f"  F1-Score:   {f1:.4f}")
    print(f"  Pred Time:  {pred_time:.2f}s ({len(x_test):,} samples)")
    print(f"  Per Sample: {pred_time / len(x_test) * 1000:.4f} ms")

    target_names = [label_mapping[i] for i in sorted(label_mapping.keys())]
    print(f"\n  Classification Report:")
    report = classification_report(
        y_test, y_pred,
        target_names=target_names,
        zero_division=0,
    )
    print(report)

    return {
        "accuracy": accuracy,
        "precision": precision,
        "recall": recall,
        "f1_score": f1,
        "prediction_time": pred_time,
        "per_sample_ms": pred_time / len(x_test) * 1000,
        "confusion_matrix": confusion_matrix(y_test, y_pred),
    }
