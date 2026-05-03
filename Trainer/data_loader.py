"""
Memory-safe data loader for CICIoT2023 dataset.
Loads CSV in chunks, converts dtypes, and samples rows.
"""
import pandas as pd
import numpy as np
from config import CHUNK_SIZE, LABEL_COLUMN


def load_csv_chunked(file_path, max_rows=None):
    """
    Load a large CSV file in chunks with memory optimization.

    Args:
        file_path: Path to CSV file.
        max_rows: Maximum number of rows to load. None = all.

    Returns:
        pd.DataFrame with optimized dtypes.
    """
    chunks = []
    loaded = 0

    for chunk in pd.read_csv(file_path, chunksize=CHUNK_SIZE, low_memory=False):
        chunk = _optimize_dtypes(chunk)
        chunks.append(chunk)
        loaded += len(chunk)

        if max_rows and loaded >= max_rows:
            break

    dataframe = pd.concat(chunks, ignore_index=True)

    if max_rows and len(dataframe) > max_rows:
        dataframe = dataframe.sample(n=max_rows, random_state=42)
        dataframe = dataframe.reset_index(drop=True)

    return dataframe


def _optimize_dtypes(dataframe):
    """Convert float64 columns to float32 to halve memory usage."""
    for col in dataframe.columns:
        if col == LABEL_COLUMN:
            continue
        if dataframe[col].dtype == np.float64:
            dataframe[col] = dataframe[col].astype(np.float32)

    return dataframe


def print_dataset_info(dataframe, name="Dataset"):
    """Print summary information about a loaded dataset."""
    memory_mb = dataframe.memory_usage(deep=True).sum() / (1024 ** 2)
    n_features = len(dataframe.columns) - 1
    n_labels = dataframe[LABEL_COLUMN].nunique()

    print(f"\n{'=' * 60}")
    print(f"  {name}")
    print(f"{'=' * 60}")
    print(f"  Rows:     {len(dataframe):,}")
    print(f"  Features: {n_features}")
    print(f"  Labels:   {n_labels} unique classes")
    print(f"  Memory:   {memory_mb:.1f} MB")
    print(f"{'=' * 60}")
