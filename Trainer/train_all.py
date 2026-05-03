"""
CICIoT2023 Training Pipeline - Main Entry Point.
Trains XGBoost (GPU), Random Forest, and SVM on CICIoT2023 dataset.

Usage:
    python train_all.py
    python train_all.py --model xgboost
    python train_all.py --model rf
    python train_all.py --model svm
"""
import sys
import time
import gc
import warnings
import numpy as np
from data_loader import load_csv_chunked, print_dataset_info
from data_preprocessor import (
    clean_dataframe, encode_labels, select_features,
    fit_scaler, split_features_labels,
)
from resampler import resample_training_data
from config import (
    TRAIN_CSV, TEST_CSV, VAL_CSV,
    MAX_TRAIN_ROWS, MAX_TEST_ROWS, LABEL_COLUMN,
)

warnings.filterwarnings("ignore")


def load_and_preprocess():
    """Load, clean, encode, select features, scale, and resample."""
    total_start = time.time()

    # Step 1: Load data
    print("\n" + "=" * 60)
    print("  STEP 1: LOADING DATA")
    print("=" * 60)
    train_df = load_csv_chunked(TRAIN_CSV, max_rows=MAX_TRAIN_ROWS)
    print_dataset_info(train_df, "Training Set")

    test_df = load_csv_chunked(TEST_CSV, max_rows=MAX_TEST_ROWS)
    print_dataset_info(test_df, "Test Set")

    # Step 2: Clean
    print("\n  STEP 2: CLEANING")
    train_df = clean_dataframe(train_df)
    test_df = clean_dataframe(test_df)

    # Step 3: Encode labels
    print("\n  STEP 3: ENCODING LABELS")
    train_df, encoder, label_mapping = encode_labels(train_df)
    test_df["label_encoded"] = encoder.transform(test_df[LABEL_COLUMN])

    # Step 4: Feature selection (fit on train only)
    print("\n  STEP 4: FEATURE SELECTION")
    feature_cols = [c for c in train_df.columns if c not in [LABEL_COLUMN, "label_encoded"]]
    train_features = train_df[feature_cols].astype(np.float32)
    train_features, selector, selected_cols = select_features(train_features)

    # Step 5: Split features/labels
    x_train = train_features.values
    y_train = train_df["label_encoded"].values
    x_test = selector.transform(test_df[feature_cols].astype(np.float32))
    y_test = test_df["label_encoded"].values
    del train_df, test_df, train_features
    gc.collect()

    # Step 6: Resample training data
    print("\n  STEP 6: RESAMPLING")
    x_train, y_train = resample_training_data(x_train, y_train)

    # Step 7: Scale
    print("\n  STEP 7: SCALING")
    scaler, x_train = fit_scaler(x_train)
    x_test = scaler.transform(x_test)

    elapsed = time.time() - total_start
    print(f"\n  Preprocessing completed in {elapsed:.1f}s")
    print(f"  Train: {x_train.shape}, Test: {x_test.shape}")

    return (x_train, y_train, x_test, y_test,
            scaler, selector, encoder, label_mapping, selected_cols)


def main():
    """Main training pipeline."""
    model_filter = sys.argv[1].replace("--model=", "") if len(sys.argv) > 1 and "--model" in sys.argv[1] else "all"

    print("=" * 60)
    print("  CICIoT2023 TRAINING PIPELINE")
    print(f"  Models: {model_filter}")
    print("=" * 60)

    data = load_and_preprocess()
    x_train, y_train, x_test, y_test = data[:4]
    scaler, selector, encoder, label_mapping, selected_cols = data[4:]
    shared = (x_test, y_test, label_mapping, scaler, selector, encoder, selected_cols)

    if model_filter in ("all", "xgboost"):
        from train_xgboost import train_xgboost
        train_xgboost(x_train, y_train, *shared)
        gc.collect()

    if model_filter in ("all", "rf"):
        from train_random_forest import train_random_forest
        train_random_forest(x_train, y_train, *shared)
        gc.collect()

    if model_filter in ("all", "svm"):
        from train_svm import train_svm
        train_svm(x_train, y_train, *shared)
        gc.collect()

    print("\n" + "=" * 60)
    print("  ALL TRAINING COMPLETED!")
    print("=" * 60)


if __name__ == "__main__":
    main()
