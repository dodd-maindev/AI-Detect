# NIDS v2.0 — Network Intrusion Detection System

Hệ thống phát hiện xâm nhập mạng sử dụng **XGBoost** trên bộ dữ liệu **CICIoT2023** (34 loại tấn công), kết hợp **Go API Server** cho inference tốc độ cao.

---

## Kiến trúc tổng quan

```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│   CICIOT23/      │     │   Trainer/       │     │   Model/         │
│   Dataset CSV    │────▶│   Python Scripts │────▶│   XGBoost .joblib│
│   (2.2 GB)       │     │   (Train + Eval) │     │   + go_export/   │
└──────────────────┘     └──────────────────┘     └────────┬─────────┘
                                                           │
                                                           ▼
                                                  ┌──────────────────┐
                                                  │  API_Service/    │
                                                  │  Go HTTP Server  │
                                                  │  (Pure Go, 0    │
                                                  │   dependencies)  │
                                                  └──────────────────┘
```

## Cấu trúc thư mục

```
Redo_ATTT/
├── CICIOT23/                     # Dataset CICIoT2023
│   ├── train/train.csv           # 5.49M rows (1.5 GB)
│   ├── test/test.csv             # 1.17M rows (332 MB)
│   └── validation/validation.csv # 1.17M rows (332 MB)
│
├── Trainer/                      # Pipeline training (Python)
│   ├── config.py                 # Cấu hình paths, hyperparams
│   ├── data_loader.py            # Load CSV theo chunks (memory-safe)
│   ├── data_preprocessor.py      # Clean, encode, scale
│   ├── resampler.py              # SMOTE + Undersampling
│   ├── evaluator.py              # Metrics, classification report
│   ├── model_saver.py            # Serialize model artifacts
│   ├── train_xgboost.py          # Train XGBoost (GPU: cuda)
│   ├── train_random_forest.py    # Train Random Forest
│   ├── train_all.py              # Entry point — chạy toàn bộ pipeline
│   ├── convert_for_go.py         # Export model → JSON cho Go API
│   └── test_api.py               # Test Go API bằng dữ liệu thật
│
├── Model/                        # Model artifacts sau khi train
│   ├── XGBoost/
│   │   ├── trained_model.joblib  # Model (Python inference)
│   │   ├── scaler.joblib
│   │   ├── feature_selector.joblib
│   │   ├── label_encoder.joblib
│   │   ├── model_config.joblib
│   │   └── go_export/            # Artifacts cho Go API
│   │       ├── model.json        # XGBoost tree dump (82 MB)
│   │       ├── scaler_params.json
│   │       ├── feature_indices.json
│   │       └── label_mapping.json
│   └── RandomForest/
│
├── API_Service/                  # Go API Server
│   ├── main.go                   # Entry point
│   ├── config/config.go          # Paths, feature names
│   ├── preprocessor/
│   │   └── preprocessor.go       # RobustScaler + FeatureSelector (Go)
│   ├── predictor/
│   │   ├── predictor.go          # Predict interface + softmax
│   │   └── xgb_model.go          # XGBoost JSON tree parser (Go)
│   ├── handler/
│   │   ├── predict_handler.go    # POST /predict, /predict/batch
│   │   └── health_handler.go     # GET /health
│   ├── go.mod
│   └── nids-api.exe              # Binary (đã build sẵn)
│
└── README.md                     # File này
```

---

## Yêu cầu hệ thống

### Training (Python)
| Thành phần | Yêu cầu |
|-----------|---------|
| Python | 3.10+ |
| RAM | ≥ 8 GB (khuyến nghị 16 GB) |
| GPU | NVIDIA (hỗ trợ CUDA) — tùy chọn, XGBoost tự fallback CPU |
| Disk | ≥ 5 GB trống |

**Python packages:**
```
pandas >= 2.0
numpy >= 1.24
scikit-learn >= 1.3
xgboost >= 2.0
imbalanced-learn >= 0.11
joblib >= 1.3
```

### API Server (Go)
| Thành phần | Yêu cầu |
|-----------|---------|
| Go | 1.21+ |
| RAM | ≥ 1 GB (model ~82 MB JSON) |
| Disk | ~100 MB |

---

## Hướng dẫn chạy

### Bước 1: Train model

> **Bỏ qua bước này nếu đã có model trong `Model/XGBoost/`**

```powershell
# Train tất cả models (XGBoost + Random Forest)
cd Redo_ATTT/Trainer
python train_all.py

# Hoặc train riêng từng model
python train_all.py --model=xgboost
python train_all.py --model=rf
```

**Output:**
- `Model/XGBoost/` — XGBoost model + preprocessing artifacts
- `Model/RandomForest/` — Random Forest model

**Thời gian ước tính:**
| Model | Thời gian | Ghi chú |
|-------|-----------|---------|
| XGBoost (GPU) | ~2 phút | Cần CUDA |
| XGBoost (CPU) | ~10 phút | Tự động fallback |
| Random Forest | ~5 phút | CPU only |

### Bước 2: Export model cho Go API

```powershell
cd Redo_ATTT/Trainer
python convert_for_go.py
```

**Output:** 4 file JSON trong `Model/XGBoost/go_export/`

### Bước 3: Build và chạy Go API

```powershell
cd Redo_ATTT/API_Service

# Build
go build -o nids-api.exe .

# Chạy server
./nids-api.exe
```
cd ~/API_Service
sudo fuser -k 5000/tcp
go build -o nids-engine .
sudo ./nids-engine



**Server khởi động:**
```
============================================================
  NIDS API Server v2.0 (Go + XGBoost)
  Dataset: CICIoT2023 | 34 Attack Classes
============================================================
  Loading preprocessor...
  ✓ Preprocessor loaded
  Loading XGBoost model...
  Loaded: 10200 trees, 34 classes
  ✓ Model loaded

  Server listening on :5000
============================================================
```

### Bước 4: Test API

```powershell
# Health check
curl http://localhost:5000/health

# Dự đoán 1 sample (46 features)
curl -X POST http://localhost:5000/predict ^
  -H "Content-Type: application/json" ^
  -d "{\"features\": [0.1, 2.0, 6.0, ...]}"

# Dự đoán batch
curl -X POST http://localhost:5000/predict/batch ^
  -H "Content-Type: application/json" ^
  -d "{\"samples\": [[...], [...], [...]]}"

# Hoặc dùng script test tự động
cd Redo_ATTT/Trainer
python test_api.py
```

---

## API Endpoints

### `GET /health`
Kiểm tra trạng thái server.

**Response:**
```json
{
  "status": "healthy",
  "model": "XGBoost (CICIoT2023)",
  "dataset": "CICIoT2023 - 34 classes",
  "go_version": "go1.23.6",
  "uptime": "1m30s"
}
```

### `POST /predict`
Dự đoán 1 sample.

**Request:**
```json
{
  "features": [0.0, 386.0, 6.0, 0.003, 333.33, ...]
}
```
> `features` là mảng **46 giá trị float64** theo thứ tự cột trong dataset CICIoT2023.

**Response:**
```json
{
  "label": "DDoS-TCP_Flood",
  "class_index": 13,
  "confidence": 0.9999,
  "latency_ms": 1.726
}
```

### `POST /predict/batch`
Dự đoán nhiều samples cùng lúc.

**Request:**
```json
{
  "samples": [
    [0.0, 386.0, 6.0, ...],
    [0.0, 220.0, 17.0, ...],
    [0.0, 160.0, 6.0, ...]
  ]
}
```

**Response:**
```json
{
  "results": [
    {"label": "DDoS-TCP_Flood", "class_index": 13, "confidence": 0.9999},
    {"label": "DoS-UDP_Flood", "class_index": 21, "confidence": 1.0},
    {"label": "BenignTraffic", "class_index": 1, "confidence": 0.9823}
  ],
  "count": 3,
  "total_ms": 5.214,
  "avg_ms": 1.738
}
```

---

## Thứ tự 46 Features (CICIoT2023)

| # | Feature | Mô tả |
|---|---------|-------|
| 0 | `flow_duration` | Thời gian flow |
| 1 | `Header_Length` | Tổng chiều dài header |
| 2 | `Protocol Type` | Loại protocol (TCP=6, UDP=17, ICMP=1) |
| 3 | `Duration` | Thời gian kết nối |
| 4 | `Rate` | Tốc độ gói tin |
| 5 | `Srate` | Source rate |
| 6 | `Drate` | Destination rate |
| 7–14 | `fin/syn/rst/psh/ack/ece/cwr_flag_number` | Số cờ TCP |
| 14–18 | `ack/syn/fin/urg/rst_count` | Tổng đếm cờ |
| 19–32 | `HTTP, HTTPS, DNS, ..., LLC` | Protocol indicators (0/1) |
| 33 | `Tot sum` | Tổng kích thước packets |
| 34 | `Min` | Packet nhỏ nhất |
| 35 | `Max` | Packet lớn nhất |
| 36 | `AVG` | Kích thước trung bình |
| 37 | `Std` | Độ lệch chuẩn |
| 38 | `Tot size` | Tổng kích thước flow |
| 39 | `IAT` | Inter-Arrival Time |
| 40 | `Number` | Số lượng packets |
| 41 | `Magnitue` | Magnitude |
| 42 | `Radius` | Radius phân bố |
| 43 | `Covariance` | Hiệp phương sai |
| 44 | `Variance` | Phương sai |
| 45 | `Weight` | Trọng số flow |

---

## 34 Attack Classes

| Nhóm | Các loại |
|------|----------|
| **DDoS** (11) | ACK_Fragmentation, HTTP_Flood, ICMP_Flood, ICMP_Fragmentation, PSHACK_Flood, RSTFINFlood, SYN_Flood, SlowLoris, SynonymousIP_Flood, TCP_Flood, UDP_Flood, UDP_Fragmentation |
| **DoS** (4) | HTTP_Flood, SYN_Flood, TCP_Flood, UDP_Flood |
| **Mirai** (3) | greeth_flood, greip_flood, udpplain |
| **Recon** (4) | HostDiscovery, OSScan, PingSweep, PortScan |
| **Spoofing** (2) | MITM-ArpSpoofing, DNS_Spoofing |
| **Web** (4) | SqlInjection, XSS, CommandInjection, BrowserHijacking |
| **Other** (4) | Backdoor_Malware, DictionaryBruteForce, Uploading_Attack, VulnerabilityScan |
| **Benign** (1) | BenignTraffic |

---

## Hiệu năng

| Metric | XGBoost |
|--------|---------|
| **Accuracy** | 99.44% |
| **F1-Score (weighted)** | 99.44% |
| **API Latency** | ~1.7 ms/sample |
| **Batch Throughput** | ~570 samples/sec |
| **Model Load Time** | ~15 giây (parse 82MB JSON) |
| **Binary Size** | 7.7 MB (Go executable) |

---

## Troubleshooting

| Vấn đề | Nguyên nhân | Giải pháp |
|--------|-------------|-----------|
| `CUDA error` khi train | Không có GPU / driver cũ | XGBoost tự fallback CPU, hoặc đổi `device: "cpu"` trong `config.py` |
| `MemoryError` khi train | RAM < 8GB | Giảm `MAX_TRAIN_ROWS` trong `config.py` |
| API trả `404` | Sai endpoint | Kiểm tra URL: `/predict` (không phải `/api/predict`) |
| Model load chậm | File 82MB JSON | Bình thường ~15s lần đầu, sau đó cache OS |
| Prediction sai class | Thứ tự features sai | Đảm bảo 46 features đúng thứ tự bảng trên |



Tại sao ngay trong lúc cấm IP mà nó vẫn nhận diện đang bị tấn công bởi IP đó  ??



Một câu hỏi cực kỳ tinh tế và thể hiện bạn hiểu rất sâu về luồng đi của hệ thống! Đây chính là một điểm cộng cực lớn nếu bạn biết cách giải thích cho giảng viên trong lúc bảo vệ đồ án.

Việc NIDS vẫn nhìn thấy và báo cáo kẻ tấn công dù IP đó đã bị iptables cấm hoàn toàn là Bản chất của kiến trúc Linux Networking (Và nó là một tính năng tốt, không phải lỗi!).

Hãy hình dung quá trình một gói tin đi vào máy chủ Ubuntu của bạn như sau:

1. Luồng đi của gói tin:
Bước 1 (Cửa ngoài cùng): Gói tin từ Windows bay đến Card mạng vật lý/ảo (ens33) của Ubuntu.
Bước 2 (Trạm gác của NIDS): Thư viện GoPacket (dựa trên lõi libpcap) đứng ngay tại cửa của Card mạng (OSI Layer 2). Nó "copy" toàn bộ gói tin thô đi qua đây để mang về cho AI phân tích.
Bước 3 (Bức tường lửa iptables): Sau khi lọt qua cổng Card mạng, gói tin mới đi vào nhân hệ điều hành Linux (Linux Kernel). Tại đây, iptables (OSI Layer 3) đứng chờ sẵn. Nó dò thấy Rule cấm IP Windows, lập tức tiêu hủy (DROP) gói tin đó ném vào sọt rác.
Bước 4 (Đích đến): NGINX và Hệ thống Thư viện của bạn hoàn toàn bình yên, không nhận được bất kỳ gói tin nào, không bị tốn CPU.
2. Tại sao đây là một "TÍNH NĂNG" chứ không phải "LỖI"?
Vì GoPacket (Bước 2) đứng trước iptables (Bước 3), nên nó luôn "nhìn thấy" gói tin bay đến đập vào tường lửa. Nhờ vậy, NIDS vẫn bắt được và in log cho bạn xem.

Lợi ích của việc này trong thực tế (SOC - Security Operations Center):

Hiển thị tính liên tục của cuộc tấn công (Visibility): Giúp người quản trị mạng biết được rằng: "À, thằng Hacker này cay cú quá, dù bị khóa IP rồi nhưng nó vẫn đang ngoan cố cắm máy dội bom mình liên tục".
Theo dõi thời điểm kết thúc: Nếu tự nhiên log NIDS im bặt, bạn biết ngay là tên Hacker đã bỏ cuộc hoặc tắt máy.
Nếu iptables chặn mà GoPacket cũng mù luôn (không thấy gì), người quản trị mạng sẽ không có bất kỳ thông số nào để biết hệ thống có đang bị dội bom hay không.
Bạn có thể lưu lại lập luận này để chém gió lúc bảo vệ đồ án:

"Hệ thống của em đặt cảm biến (Sniffer) ở Tầng Data Link (Layer 2) bằng AF_PACKET, nằm trước Tầng Network (Layer 3) của Iptables. Nhờ kiến trúc này, tường lửa vẫn chặn đứng cuộc tấn công bảo vệ Web Server, nhưng NIDS vẫn thu thập được đầy đủ bằng chứng (Log) chứng minh Hacker vẫn đang cố tình dội bom."


1. Bật Active IPS (Phát hiện + Tự động Block IP)
Chế độ này NIDS sẽ trở thành một Vệ sĩ chủ động. Khi phát hiện Hacker, nó sẽ cảnh báo màn hình đỏ rực và tự động gõ lệnh iptables cấm cửa ngay lập tức.

bash
sudo ./nids-engine --mode=active


(Thực ra nếu bạn chỉ gõ sudo ./nids-engine không có cờ gì cả thì nó cũng tự hiểu là mặc định đang bật Active IPS).

2. Tắt IPS - Chuyển sang Passive NIDS (Chỉ phát hiện, KHÔNG Block)
Chế độ này NIDS sẽ trở thành một Camera quan sát. Nó vẫn phân tích gói tin, vẫn in ra dòng chữ Đỏ cảnh báo là đang bị tấn công, nhưng nó KHÔNG hề chặn IP đó. (Hữu ích khi bạn muốn test khả năng nhận diện liên tục mà không sợ IP bị khóa).

bash
sudo ./nids-engine --mode=passive