"""Export XGBoost model in multiple formats to find leaves-compatible one."""
import joblib
import os

MODEL_DIR = r"C:\New folder (2)\Redo_ATTT\Model\XGBoost"
OUT_DIR = r"C:\New folder (2)\Redo_ATTT\Model\XGBoost\go_export"

model = joblib.load(os.path.join(MODEL_DIR, "trained_model.joblib"))
booster = model.get_booster()

# Save as .ubj (UBJSON)
ubj_path = os.path.join(OUT_DIR, "model.ubj")
booster.save_model(ubj_path)
print("Saved as .ubj:", os.path.getsize(ubj_path) / (1024 * 1024), "MB")

# Save as .json
json_path = os.path.join(OUT_DIR, "model.json")
booster.save_model(json_path)
print("Saved as .json:", os.path.getsize(json_path) / (1024 * 1024), "MB")

# Check headers
with open(ubj_path, "rb") as f:
    header = f.read(16)
    print("UBJ header bytes:", list(header[:8]))

with open(json_path, "rb") as f:
    header = f.read(100).decode("utf-8", errors="ignore")
    print("JSON header:", header[:80])

# Also try downgrading - save config and use old format
config = booster.save_config()
print("Config type:", type(config))
print("Config preview:", config[:200])
