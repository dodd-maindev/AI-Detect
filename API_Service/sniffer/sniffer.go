// Package sniffer implements real-time packet capture, flow feature extraction,
// and automated IPS response using iptables.
package sniffer

import (
	"fmt"
	"log"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"

	"nids-api/predictor"
	"nids-api/preprocessor"
)

// Config defines sniffer settings.
type Config struct {
	Interface string
	TargetIP  string // IP of the NGINX server to protect
	ActiveIPS bool   // true = Block IP, false = Detect Only (Passive NIDS)
}

// IPSCore represents the sniffer and intrusion prevention engine.
type IPSCore struct {
	cfg          Config
	preprocessor *preprocessor.Preprocessor
	predictor    *predictor.Predictor
	blockedIPs   map[string]bool
	flowManager  *FlowManager
	mu           sync.Mutex
}

// NewIPSCore initializes a new IPS engine.
func NewIPSCore(cfg Config, prep *preprocessor.Preprocessor, pred *predictor.Predictor) *IPSCore {
	return &IPSCore{
		cfg:          cfg,
		preprocessor: prep,
		predictor:    pred,
		blockedIPs:   make(map[string]bool),
		flowManager:  NewFlowManager(),
	}
}

// Start begins capturing packets and predicting.
func (ips *IPSCore) Start() {
	ips.printBanner()

	handle, err := pcap.OpenLive(ips.cfg.Interface, 1600, true, pcap.BlockForever)
	if err != nil {
		log.Fatalf("Error opening interface %s: %v\n(Run with sudo?)", ips.cfg.Interface, err)
	}
	defer handle.Close()

	go ips.analysisLoop()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for pkt := range packetSource.Packets() {
		ips.processPacket(pkt)
	}
}

// printBanner displays the startup banner with mode information.
func (ips *IPSCore) printBanner() {
	fmt.Println("============================================================")
	fmt.Printf("  %s[*]%s IPS Agent (GoPacket) Started on interface: %s%s%s\n",
		BrightCyan, Reset, Bold+BrightWhite, ips.cfg.Interface, Reset)
	if ips.cfg.ActiveIPS {
		fmt.Printf("  %s[*]%s Mode: %s ACTIVE IPS (Block + Alert) %s\n",
			BrightCyan, Reset, BgRed+Bold+White, Reset)
	} else {
		fmt.Printf("  %s[*]%s Mode: %s PASSIVE NIDS (Monitor Only) %s\n",
			BrightCyan, Reset, BgBlue+Bold+White, Reset)
	}
	fmt.Printf("  %s[*]%s Monitoring traffic and protecting NGINX...\n", BrightCyan, Reset)
	fmt.Println("============================================================")
}
