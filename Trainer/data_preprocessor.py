"""
Data preprocessing pipeline for CICIoT2023.
Handles cleaning, encoding, scaling, and class imbalance.
"""
import numpy as np
import pandas as pd
from sklearn.preprocessing import LabelEncoder, RobustScaler
from sklearn.feature_selection import VarianceThreshold
from config import LABEL_COLUMN, FEATURE_VARIANCE_THRESHOLD, RANDOM_STATE


def clean_dataframe(dataframe):
    """Remove NaN, Inf, duplicates from the dataset."""
    initial = len(dataframe)
    dataframe = dataframe.drop_duplicates()
    dataframe = dataframe.replace([np.inf, -np.inf], np.nan)
    dataframe = dataframe.dropna()
    removed = initial - len(dataframe)

    print(f"  Cleaned: removed {removed:,} rows ({removed/initial*100:.1f}%)")
    print(f"  Remaining: {len(dataframe):,} rows")
    return dataframe


def encode_labels(dataframe):
    """Encode string labels to integers. Returns encoder + encoded df."""
    encoder = LabelEncoder()
    dataframe = dataframe.copy()
    dataframe["label_encoded"] = encoder.fit_transform(dataframe[LABEL_COLUMN])

    mapping = dict(zip(
        encoder.transform(encoder.classes_),
        encoder.classes_,
    ))
    print(f"  Encoded {len(mapping)} classes:")
    for idx, name in sorted(mapping.items()):
        count = (dataframe["label_encoded"] == idx).sum()
        print(f"    {idx:2d}: {name:<35s} ({count:>8,})")

    return dataframe, encoder, mapping


def select_features(features_df):
    """Remove low-variance features using VarianceThreshold."""
    selector = VarianceThreshold(threshold=FEATURE_VARIANCE_THRESHOLD)
    selected_array = selector.fit_transform(features_df)
    selected_cols = features_df.columns[selector.get_support()].tolist()

    removed = features_df.shape[1] - len(selected_cols)
    print(f"  Features: {features_df.shape[1]} -> {len(selected_cols)} (removed {removed})")

    return pd.DataFrame(selected_array, columns=selected_cols), selector, selected_cols


def fit_scaler(train_features):
    """Fit RobustScaler on training data. Returns scaler + scaled array."""
    scaler = RobustScaler()
    scaled = scaler.fit_transform(train_features)
    print(f"  Fitted RobustScaler on {train_features.shape[0]:,} samples")
    return scaler, scaled


def split_features_labels(dataframe, selected_cols=None):
    """Split dataframe into X (features) and y (encoded labels)."""
    drop_cols = [LABEL_COLUMN, "label_encoded"]
    feature_cols = [c for c in dataframe.columns if c not in drop_cols]

    if selected_cols:
        feature_cols = [c for c in selected_cols if c in dataframe.columns]

    x_data = dataframe[feature_cols].values.astype(np.float32)
    y_data = dataframe["label_encoded"].values

    return x_data, y_data
