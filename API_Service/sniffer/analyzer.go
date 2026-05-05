// Package sniffer provides the analysis loop and IPS blocking logic.
package sniffer

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket"
)

// analysisLoop runs every 2 seconds to aggregate flows and run AI inference.
func (ips *IPSCore) analysisLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Window Tracking: Use PopActiveFlows to reset state every 2s
		flows := ips.flowManager.PopActiveFlows()
		for _, flow := range flows {
			ips.analyzeFlow(flow)
		}
	}
}

// analyzeFlow performs AI inference and hybrid heuristic checks on a single flow.
func (ips *IPSCore) analyzeFlow(flow *FlowTracker) {
	features := flow.ExtractFeatures()

	aiStart := time.Now()
	scaled := ips.preprocessor.Transform(features)
	result := ips.predictor.Predict(scaled)
	aiLatency := time.Since(aiStart)

	// LỚP 1: NGƯỠNG ĐỘNG (Dynamic Thresholds) - Mỗi loại tấn công cần ngưỡng khác nhau
	threshold := 0.85
	threshold = ips.getDynamicThreshold(result.Label, threshold)

	isAttack := result.Label != "BenignTraffic" && result.Confidence >= threshold

	// LỚP 2: LUẬT TĨNH (Heuristic Rules) - Chống False Positive và False Negative
	rate := flow.PktCount / flow.Duration()
	isAttack, result.Label = ips.applyHeuristics(isAttack, flow.PktCount, rate, result.Label)


	// Hiển thị log với màu sắc tương ứng
	latencyStr := fmt.Sprintf("%v", aiLatency)
	if isAttack {
		LogAttack(flow.SrcIP, flow.DstIP, flow.PktCount, result.Label, result.Confidence, latencyStr)
		ips.handleIntrusion(flow, result.Label, result.Confidence)
	} else if result.Confidence >= 0.50 && result.Label != "BenignTraffic" {
		LogSuspicious(flow.SrcIP, flow.DstIP, flow.PktCount, result.Label, result.Confidence, latencyStr)
	} else {
		LogBenign(flow.SrcIP, flow.DstIP, flow.PktCount, result.Label, result.Confidence, latencyStr)
	}
}

// getDynamicThreshold returns a confidence threshold tailored to the attack category.
func (ips *IPSCore) getDynamicThreshold(label string, defaultVal float64) float64 {
	switch label {
	case "Recon-PortScan", "Recon-OSScan", "Recon-HostDiscovery", "Recon-PingSweep":
		return 0.60
	case "MITM-ArpSpoofing", "DNS_Spoofing":
		return 0.70
	case "DoS-Slowloris", "DictionaryBruteForce":
		return 0.75
	default:
		return defaultVal
	}
}

// applyHeuristics adjusts the attack decision using rule-based logic.
func (ips *IPSCore) applyHeuristics(isAttack bool, pktCount float64, rate float64, label string) (bool, string) {
	floodLabels := map[string]bool{
		"DoS-HTTP_Flood": true, "DDoS-HTTP_Flood": true,
		"DDoS-TCP_Flood": true, "DoS-TCP_Flood": true,
		"DDoS-RSTFINFlood": true, "DoS-SYN_Flood": true,
		"DDoS-UDP_Flood": true, "DoS-UDP_Flood": true,
		"DDoS-SynonymousIP_Flood": true, "VulnerabilityScan": true,
		"DDoS-SYN_Flood": true, "DDoS-ICMP_Flood": true, 
		"DoS-ICMP_Flood": true, "DDoS-ACK_Flood": true,
		"DDoS-PSHACK_Flood": true, "DDoS-UDP_Fragmentation": true,
	}

	// Chống False Positive Tối Hậu (Ép AI phải học cách phân biệt Web):
	// Trí tuệ nhân tạo (Dataset IoT) rất "ngu ngơ" với luồng tải Web. Nó thấy tốc độ tải tài nguyên nhanh (vài trăm pkt) là nó quy chụp thành Flood.
	// Luật thép cho Web Server:
	// 1. Dưới 500 packets: Quá ít để gọi là DDoS -> Chắc chắn là Web Load.
	// 2. Tốc độ < 500 pkts/sec: Quá chậm để làm sập NGINX -> Chắc chắn là Web Load.
	// -> Ép tất cả các chẩn đoán Flood thỏa mãn điều kiện này thành BenignTraffic (Web Load).
	if (pktCount < 500 || rate < 500) && floodLabels[label] {
		return false, "BenignTraffic"
	}
	
	// Chống False Negative: Dội bom cực lớn với Rate > 500 nhưng AI mất tự tin
	if !isAttack && pktCount > 1000 && rate > 500 && floodLabels[label] {
		return true, label
	}
	return isAttack, label
}

// handleIntrusion logs the intrusion and optionally blocks the IP.
func (ips *IPSCore) handleIntrusion(flow *FlowTracker, label string, conf float64) {
	attackerIP := flow.SrcIP
	if attackerIP == "10.203.152.105" || attackerIP == ips.cfg.TargetIP {
		attackerIP = flow.DstIP
	}
	detectionTime := time.Since(flow.StartTime)

	if ips.cfg.ActiveIPS {
		ips.blockIP(attackerIP, label, conf, detectionTime)
	} else {
		LogIntrusion(label, conf, attackerIP, "DETECT ONLY - NO BLOCK", fmt.Sprintf("%v", detectionTime))
	}
}

// processPacket filters and routes raw packets into the FlowManager.
func (ips *IPSCore) processPacket(pkt gopacket.Packet) {
	ipLayer := pkt.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return
	}

	ip, _ := ipLayer.(*layers.IPv4)
	// HIDS Filter: Chỉ theo dõi các gói tin ĐI ĐẾN hoặc ĐI TỪ máy chủ NGINX của chúng ta
	if ip.SrcIP.String() != ips.cfg.TargetIP && ip.DstIP.String() != ips.cfg.TargetIP {
		return
	}

	// Lọc bỏ các cổng phát sóng nền rác (mDNS, SSDP, LLMNR, NetBIOS)
	if udpLayer := pkt.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		if udp.DstPort == 5353 || udp.DstPort == 1900 || udp.DstPort == 5355 || udp.DstPort == 137 || udp.DstPort == 138 {
			return
		}
	}

	// Bỏ qua SSH (Port 22) để không bao giờ tự khoá mình
	if tcpLayer := pkt.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		if tcp.DstPort == 22 || tcp.SrcPort == 22 {
			return
		}
	}

	ips.flowManager.ProcessPacket(pkt)
}

// blockIP blocks an attacker IP using XDP (hardware) or iptables (kernel) fallback.
func (ips *IPSCore) blockIP(ip string, label string, conf float64, detectionTime time.Duration) {
	ips.mu.Lock()
	if ips.blockedIPs[ip] {
		ips.mu.Unlock()
		return
	}
	ips.blockedIPs[ip] = true
	ips.mu.Unlock()

	detTimeStr := fmt.Sprintf("%v", detectionTime)

	if ips.xdpManager != nil {
		if err := ips.xdpManager.BlockIP(ip); err == nil {
			LogIntrusion(label, conf, ip, "XDP HARDWARE BLOCK - 1 MIN", detTimeStr)
		} else {
			// XDP failed, fallback to iptables for this IP
			exec.Command("iptables", "-A", "INPUT", "-s", ip, "-j", "DROP").Run()
			LogIntrusion(label, conf, ip, "IPTABLES BLOCK - 1 MIN", detTimeStr)
		}
	} else {
		exec.Command("iptables", "-A", "INPUT", "-s", ip, "-j", "DROP").Run()
		LogIntrusion(label, conf, ip, "ACTIVE IPS - BLOCKED 1 MINUTE", detTimeStr)
	}

	go func() {
		time.Sleep(60 * time.Second)

		ips.mu.Lock()
		delete(ips.blockedIPs, ip)
		ips.mu.Unlock()

		LogUnban(ip)
		if ips.xdpManager != nil {
			ips.xdpManager.UnblockIP(ip)
		} else {
			exec.Command("iptables", "-D", "INPUT", "-s", ip, "-j", "DROP").Run()
		}
	}()
}
