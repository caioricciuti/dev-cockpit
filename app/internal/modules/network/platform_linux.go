//go:build linux

package network

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

const (
	pingBin       = "ping"
	tracerouteBin = "traceroute"
	digBin        = "dig"
	nslookupBin   = "nslookup"
	whoisBin      = "whois"
)

func getDefaultGateway() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "sh", "-c", "ip route show default | awk '{print $3}'").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func checkNetworkQualityAvailable() bool {
	// networkQuality is macOS-only; not available on Linux
	return false
}
