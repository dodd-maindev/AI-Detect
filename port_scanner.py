import socket
import threading
import time
import sys
import random

# === CẤU HÌNH MỤC TIÊU ===
TARGET_IP = "10.203.152.105" # Đổi IP này thành IP máy Ubuntu của bạn nếu nó khác
THREADS = 50

# Danh sách các cổng phổ biến để quét mồi
COMMON_PORTS = [21, 22, 23, 25, 53, 80, 110, 111, 135, 139, 143, 443, 445, 993, 995, 1723, 3306, 3389, 5900, 8080]

def stealth_syn_scan():
    """Mô phỏng SYN Stealth Scan (Quét diện rộng nmap -sS)
    Thay vì quét từng cổng tuần tự dễ bị NIDS bắt,
    chúng ta sẽ quét ngẫu nhiên toàn bộ 65535 cổng kết hợp với các cổng phổ biến
    để tạo ra mẫu lưu lượng "UniqueDstPorts" rất lớn.
    """
    while True:
        # Chọn ngẫu nhiên giữa việc quét cổng phổ biến hoặc quét bừa bãi toàn bộ mạng
        if random.random() > 0.3:
            port = random.choice(COMMON_PORTS)
        else:
            port = random.randint(1, 65535)
            
        try:
            s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            s.settimeout(0.05) # Timeout cực thấp để quét siêu nhanh
            # Tạo kết nối giả mạo (chỉ gửi SYN rồi tự vứt, không cần hoàn thành 3-way handshake)
            s.connect((TARGET_IP, port))
            s.close()
        except:
            pass

def aggressive_connect_scan():
    """Mô phỏng TCP Connect Scan (nmap -sT)
    Tốc độ quét cực nhanh, nhắm vào dải port liên tiếp để làm nhiễu loạn bộ phân tích IAT.
    """
    while True:
        start_port = random.randint(1, 60000)
        for port in range(start_port, start_port + 100):
            try:
                s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                s.settimeout(0.1)
                s.connect((TARGET_IP, port))
                s.close()
            except:
                pass

print("==================================================")
print(f"[*] RECON TOOL PRO (Nmap Simulator) - Target: {TARGET_IP}")
print("[1] Stealth SYN Scan (-sS): Quét tàng hình ngẫu nhiên toàn bộ 65k cổng")
print("[2] Aggressive Connect Scan (-sT): Quét càn quét tốc độ cao")
print("==================================================")

choice = input("Select scanning strategy (1-2): ")

print(f"\n[!] Bắt đầu quét thăm dò {TARGET_IP} với {THREADS} threads...")
print("[*] Qua Ubuntu xem AI nhận diện Port Scan có nhạy không nhé!")

attack_dict = {
    '1': stealth_syn_scan,
    '2': aggressive_connect_scan,
}

scan_func = attack_dict.get(choice, stealth_syn_scan)

for i in range(THREADS):
    thread = threading.Thread(target=scan_func)
    thread.daemon = True
    thread.start()

try:
    while True:
        time.sleep(1)
except KeyboardInterrupt:
    print("\n[!] Dừng quét.")
    sys.exit()
