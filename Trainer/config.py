"""
Centralized configuration for CICIoT2023 training pipeline.
All paths, hyperparameters, and constants in one place.
"""
import os

# === PATHS ===
DATASET_DIR = r"C:\New folder (2)\Redo_ATTT\CICIOT23"
TRAIN_CSV = os.path.join(DATASET_DIR, "train", "train.csv")
TEST_CSV = os.path.join(DATASET_DIR, "test", "test.csv")
VAL_CSV = os.path.join(DATASET_DIR, "validation", "validation.csv")
MODEL_DIR = r"C:\New folder (2)\Redo_ATTT\Model"

# === DATASET ===
LABEL_COLUMN = "label"
RANDOM_STATE = 42
FEATURE_VARIANCE_THRESHOLD = 0.01

# === SAMPLING (memory-safe for 16GB RAM) ===
MAX_TRAIN_ROWS = 2_000_000
MAX_TEST_ROWS = 500_000
MAX_VAL_ROWS = 500_000
CHUNK_SIZE = 500_000
RESAMPLE_TARGET_PER_CLASS = 30_000

# === XGBOOST HYPERPARAMETERS ===
XGBOOST_PARAMS = {
    "n_estimators": 300,
    "max_depth": 12,
    "learning_rate": 0.1,
    "subsample": 0.8,
    "colsample_bytree": 0.8,
    "min_child_weight": 3,
    "gamma": 0.1,
    "reg_alpha": 0.1,
    "reg_lambda": 1.0,
    "tree_method": "hist",
    "device": "cuda",
    "n_jobs": -1,
    "random_state": RANDOM_STATE,
    "eval_metric": "mlogloss",
    "verbosity": 1,
}

# === RANDOM FOREST HYPERPARAMETERS ===
RF_PARAMS = {
    "n_estimators": 200,
    "max_depth": 20,
    "min_samples_split": 5,
    "min_samples_leaf": 2,
    "class_weight": "balanced",
    "n_jobs": -1,
    "random_state": RANDOM_STATE,
    "verbose": 1,
}

# === SVM HYPERPARAMETERS ===
SVM_PARAMS = {
    "C": 10.0,
    "kernel": "rbf",
    "gamma": "scale",
    "decision_function_shape": "ovr",
    "cache_size": 2000,
    "random_state": RANDOM_STATE,
    "verbose": True,
}
SVM_MAX_TRAIN_ROWS = 200_000

# === MODEL SUBDIRECTORIES ===
XGBOOST_DIR = os.path.join(MODEL_DIR, "XGBoost")
RF_DIR = os.path.join(MODEL_DIR, "RandomForest")
SVM_DIR = os.path.join(MODEL_DIR, "SVM")
