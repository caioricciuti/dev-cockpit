//go:build linux

package network

import (
	"os/exec"
	"strings"
)

const (
	pingBin       = "ping"
	tracerouteBin = "traceroute"
	digBin        = "dig"
	nslookupBin   = "nslookup"
	whoisBin      = "whois"
)

func getDefaultGateway() string {
	out, err := exec.Command("sh", "-c", "ip route show default | awk '{print $3}'").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func checkNetworkQualityAvailable() bool {
	// networkQuality is macOS-only; not available on Linux
	return false
}
