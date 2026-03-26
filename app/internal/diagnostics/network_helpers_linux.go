//go:build linux

package diagnostics

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func checkDNS() (ok bool, ms int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	out, err := exec.CommandContext(ctx, "dig", "+short", "google.com").Output()
	elapsed := time.Since(start).Milliseconds()

	if err != nil || strings.TrimSpace(string(out)) == "" {
		return false, 0
	}
	return true, int(elapsed)
}

func checkGateway() (ok bool, ms int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gwOut, err := exec.CommandContext(ctx, "sh", "-c", "ip route show default | awk '{print $3}'").Output()
	if err != nil {
		return false, 0
	}

	gateway := strings.TrimSpace(string(gwOut))
	if gateway == "" {
		return false, 0
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()

	out, err := exec.CommandContext(ctx2, "ping", "-c", "1", "-W", "2", gateway).Output()
	if err != nil {
		return false, 0
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "avg") {
			parts := strings.Split(line, "/")
			if len(parts) >= 5 {
				avg, err := strconv.ParseFloat(parts[4], 64)
				if err == nil {
					return true, int(avg)
				}
			}
		}
	}
	return true, 0
}
