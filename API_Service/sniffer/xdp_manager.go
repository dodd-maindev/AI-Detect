// Package sniffer provides XDP/eBPF hardware-level packet blocking.
package sniffer

import (
	"fmt"
	"net"
	"os/exec"
)

// XDPManager controls the XDP firewall attached to a network interface.
type XDPManager struct {
	iface   string
	bpfPath string
	pinPath string
}

// NewXDPManager loads the XDP program and pins the BPF map for easy access.
func NewXDPManager(iface string, bpfObjPath string) (*XDPManager, error) {
	mgr := &XDPManager{
		iface:   iface,
		bpfPath: bpfObjPath,
		pinPath: "/sys/fs/bpf/blocked_ips",
	}
	if err := mgr.loadXDP(); err != nil {
		return nil, err
	}
	return mgr, nil
}

// loadXDP attaches the XDP program and pins the blocked_ips map.
func (x *XDPManager) loadXDP() error {
	// Detach any existing XDP program first
	exec.Command("ip", "link", "set", "dev", x.iface, "xdp", "off").Run()
	// Clean stale pin
	exec.Command("rm", "-f", x.pinPath).Run()

	// Load XDP program onto NIC
	cmd := exec.Command("ip", "link", "set", "dev", x.iface, "xdp", "obj", x.bpfPath, "sec", "xdp")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("XDP load: %s: %w", string(out), err)
	}

	// Pin the map so we can access it by path (no ID parsing needed)
	cmd = exec.Command("bpftool", "map", "pin", "name", "blocked_ips", x.pinPath)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("map pin: %s: %w", string(out), err)
	}
	return nil
}

// BlockIP adds an IP to the XDP hardware blacklist.
func (x *XDPManager) BlockIP(ipStr string) error {
	hexBytes := ipToHexSlice(ipStr)
	if hexBytes == nil {
		return fmt.Errorf("invalid IP: %s", ipStr)
	}
	// bpftool requires each hex byte as a separate argument
	args := []string{"map", "update", "pinned", x.pinPath, "key", "hex"}
	args = append(args, hexBytes...)
	args = append(args, "value", "hex", "01", "00", "00", "00")
	cmd := exec.Command("bpftool", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("block failed: %s: %w", string(out), err)
	}
	return nil
}

// UnblockIP removes an IP from the XDP hardware blacklist.
func (x *XDPManager) UnblockIP(ipStr string) error {
	hexBytes := ipToHexSlice(ipStr)
	if hexBytes == nil {
		return fmt.Errorf("invalid IP: %s", ipStr)
	}
	args := []string{"map", "delete", "pinned", x.pinPath, "key", "hex"}
	args = append(args, hexBytes...)
	exec.Command("bpftool", args...).Run()
	return nil
}

// Cleanup detaches XDP and removes the pinned map.
func (x *XDPManager) Cleanup() {
	exec.Command("ip", "link", "set", "dev", x.iface, "xdp", "off").Run()
	exec.Command("rm", "-f", x.pinPath).Run()
}

// ipToHexSlice converts "10.203.152.91" to ["0a","cb","98","5b"] for bpftool.
func ipToHexSlice(ipStr string) []string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return nil
	}
	return []string{
		fmt.Sprintf("%02x", ip4[0]),
		fmt.Sprintf("%02x", ip4[1]),
		fmt.Sprintf("%02x", ip4[2]),
		fmt.Sprintf("%02x", ip4[3]),
	}
}
