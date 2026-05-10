//go:build darwin

package network

import (
	"context"
	"os/exec"
	"time"
)

const (
	pingBin       = "/sbin/ping"
	tracerouteBin = "/usr/sbin/traceroute"
	digBin        = "/usr/bin/dig"
	nslookupBin   = "/usr/bin/nslookup"
	whoisBin      = "/usr/bin/whois"
)

func getDefaultGateway() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "sh", "-c", "route get default | awk '/gateway/{print $2}'").Output()
	if err != nil {
		return ""
	}
	return string(out[:len(out)-1]) // trim newline
}

func checkNetworkQualityAvailable() bool {
	_, err := exec.LookPath("networkQuality")
	return err == nil
}
