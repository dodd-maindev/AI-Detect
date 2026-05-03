// Package sniffer provides ANSI color utilities for terminal output.
package sniffer

import (
	"fmt"
	"time"
)

// ANSI escape code constants for terminal coloring.
const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"

	// Foreground colors
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright foreground colors
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Background colors
	BgRed    = "\033[41m"
	BgGreen  = "\033[42m"
	BgYellow = "\033[43m"
	BgBlue   = "\033[44m"
	BgCyan   = "\033[46m"
)

// LogBenign prints a green-toned log line for benign traffic.
func LogBenign(src, dst string, pkt float64, label string, conf float64, latency string) {
	loc := time.FixedZone("UTC+7", 7*3600)
	ts := time.Now().In(loc).Format("15:04:05.000")
	fmt.Printf("%s[%s]%s %s%s%s -> %s%s%s | Pkt: %s%.0f%s | %s%s (%.1f%%)%s | %sBenign%s | %sAI Processing Time: %s%s\n",
		Dim, ts, Reset,
		Cyan, src, Reset,
		Cyan, dst, Reset,
		White, pkt, Reset,
		Green, label, conf*100, Reset,
		BrightGreen, Reset,
		Dim, latency, Reset,
	)
}

// LogSuspicious prints a yellow-toned log line for suspicious but not confirmed traffic.
func LogSuspicious(src, dst string, pkt float64, label string, conf float64, latency string) {
	loc := time.FixedZone("UTC+7", 7*3600)
	ts := time.Now().In(loc).Format("15:04:05.000")
	fmt.Printf("%s[%s]%s %s%s%s -> %s%s%s | Pkt: %s%.0f%s | %s%s (%.1f%%)%s | %sSuspect%s | %sAI Processing Time: %s%s\n",
		Yellow, ts, Reset,
		BrightYellow, src, Reset,
		BrightYellow, dst, Reset,
		Bold+White, pkt, Reset,
		Bold+Yellow, label, conf*100, Reset,
		Yellow, Reset,
		Dim, latency, Reset,
	)
}

// LogAttack prints a red alert log line for confirmed attacks.
func LogAttack(src, dst string, pkt float64, label string, conf float64, latency string) {
	loc := time.FixedZone("UTC+7", 7*3600)
	ts := time.Now().In(loc).Format("15:04:05.000")
	fmt.Printf("%s[%s]%s %s%s%s -> %s%s%s | Pkt: %s%.0f%s | %s%s (%.1f%%)%s | %sisAttack: True%s | %sAI Processing Time: %s%s\n",
		BrightRed, ts, Reset,
		Bold+BrightRed, src, Reset,
		Bold+BrightRed, dst, Reset,
		Bold+White, pkt, Reset,
		Bold+BrightRed, label, conf*100, Reset,
		BgRed+Bold+White, Reset,
		Dim, latency, Reset,
	)
}

// LogIntrusion prints the full intrusion alert banner.
func LogIntrusion(label string, conf float64, ip string, mode string, detTime string) {
	fmt.Println()
	fmt.Printf("  %s ⚠  INTRUSION DETECTED: %s (Conf: %.2f%%) %s\n", BgRed+Bold+White, label, conf*100, Reset)
	fmt.Printf("  %s[*]%s Attacker IP: %s%s%s %s[%s]%s\n", BrightRed, Reset, Bold+BrightWhite, ip, Reset, BrightYellow, mode, Reset)
	fmt.Printf("  %s[⏱]%s Detection Time: %s%s%s\n", BrightCyan, Reset, Bold+BrightGreen, detTime, Reset)
	fmt.Println()
}

// LogUnban prints the unban notification.
func LogUnban(ip string) {
	fmt.Printf("\n  %s[✓]%s IP %s%s%s đã được %sgỡ block%s sau 1 phút.\n\n",
		BrightGreen, Reset,
		Bold+White, ip, Reset,
		BrightGreen, Reset,
	)
}
