//go:build darwin

package network

import "os/exec"

const (
	pingBin       = "/sbin/ping"
	tracerouteBin = "/usr/sbin/traceroute"
	digBin        = "/usr/bin/dig"
	nslookupBin   = "/usr/bin/nslookup"
	whoisBin      = "/usr/bin/whois"
)

func getDefaultGateway() string {
	out, err := exec.Command("sh", "-c", "route get default | awk '/gateway/{print $2}'").Output()
	if err != nil {
		return ""
	}
	return string(out[:len(out)-1]) // trim newline
}

func checkNetworkQualityAvailable() bool {
	_, err := exec.LookPath("networkQuality")
	return err == nil
}
