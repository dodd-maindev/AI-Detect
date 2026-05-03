"""
Random Forest trainer for CICIoT2023.
Uses balanced class weights to handle imbalance natively.
"""
import gc
from sklearn.ensemble import RandomForestClassifier
from config import RF_PARAMS, RF_DIR
from evaluator import evaluate_model
from model_saver import save_model_artifacts


def train_random_forest(x_train, y_train, x_test, y_test, label_mapping,
                         scaler, selector, encoder, selected_features):
    """
    Train Random Forest classifier.

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
        Trained RandomForestClassifier model.
    """
    print("\n" + "=" * 60)
    print("  TRAINING: Random Forest")
    print("=" * 60)

    print(f"  n_estimators: {RF_PARAMS['n_estimators']}")
    print(f"  max_depth:    {RF_PARAMS['max_depth']}")
    print(f"  class_weight: {RF_PARAMS['class_weight']}")

    model = RandomForestClassifier(**RF_PARAMS)
    model.fit(x_train, y_train)

    metrics = evaluate_model(model, x_test, y_test, label_mapping, "Random Forest")

    save_model_artifacts(
        model=model, scaler=scaler, selector=selector,
        encoder=encoder, label_mapping=label_mapping,
        selected_features=selected_features,
        model_name="RandomForest", save_dir=RF_DIR,
        metrics=metrics,
    )

    gc.collect()
    return model
