"""
Resampling utilities for handling severe class imbalance.
CICIoT2023 has 5600:1 imbalance ratio.
"""
import numpy as np
from collections import Counter
from imblearn.over_sampling import SMOTE
from imblearn.under_sampling import RandomUnderSampler
from config import RANDOM_STATE, RESAMPLE_TARGET_PER_CLASS


def resample_training_data(x_train, y_train):
    """
    Balance training data using SMOTE + RandomUnderSampler.

    Strategy:
    1. SMOTE minority classes up to target count
    2. Undersample majority classes down to target count

    Args:
        x_train: Training features (numpy array).
        y_train: Training labels (numpy array).

    Returns:
        Tuple of (x_resampled, y_resampled).
    """
    target = RESAMPLE_TARGET_PER_CLASS
    print(f"\n  Resampling target: {target:,} per class")
    print(f"  Before: {len(y_train):,} samples, {len(np.unique(y_train))} classes")

    _print_distribution(y_train, "Before")

    # Step 1: SMOTE - oversample minority classes
    smote_strategy = _build_smote_strategy(y_train, target)
    if smote_strategy:
        smote = SMOTE(
            sampling_strategy=smote_strategy,
            random_state=RANDOM_STATE,
            k_neighbors=min(5, _min_class_count(y_train) - 1),
        )
        x_train, y_train = smote.fit_resample(x_train, y_train)
        print(f"  After SMOTE: {len(y_train):,} samples")

    # Step 2: Undersample - reduce majority classes
    under_strategy = {label: target for label in np.unique(y_train)}
    under = RandomUnderSampler(
        sampling_strategy=under_strategy,
        random_state=RANDOM_STATE,
    )
    x_train, y_train = under.fit_resample(x_train, y_train)

    print(f"  After Undersampling: {len(y_train):,} samples")
    _print_distribution(y_train, "After")

    return x_train, y_train


def _build_smote_strategy(y_train, target):
    """Build SMOTE strategy dict: only oversample classes below target."""
    counts = Counter(y_train)
    return {label: target for label, count in counts.items() if count < target}


def _min_class_count(y_train):
    """Return the count of the rarest class."""
    return min(Counter(y_train).values())


def _print_distribution(y_data, phase=""):
    """Print class distribution summary."""
    counts = Counter(y_data)
    sorted_counts = sorted(counts.items(), key=lambda x: x[1], reverse=True)
    top3 = sorted_counts[:3]
    bot3 = sorted_counts[-3:]
    print(f"    {phase} top-3: {top3}")
    print(f"    {phase} bot-3: {bot3}")
