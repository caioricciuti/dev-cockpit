//go:build linux

package security

import (
	"context"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const cmdTimeout = 5 * time.Second

func (m *Model) refresh() tea.Cmd {
	return func() tea.Msg {
		fw := checkLinuxFirewall()
		enc := checkDiskEncryption()
		return secMsg{
			firewall:   fw,
			filevault:  enc,
			sip:        "N/A (macOS-only)",
			gatekeeper: "N/A (macOS-only)",
			note:       "Security status refreshed",
		}
	}
}

func checkLinuxFirewall() string {
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ufw", "status").CombinedOutput()
	if err == nil {
		s := strings.TrimSpace(string(out))
		if strings.Contains(strings.ToLower(s), "active") {
			return "UFW: Active"
		}
		return "UFW: Inactive"
	}

	out, err = exec.CommandContext(ctx, "iptables", "-L", "-n", "--line-numbers").CombinedOutput()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		ruleCount := 0
		for _, line := range lines {
			if strings.HasPrefix(line, "Chain") || strings.TrimSpace(line) == "" || strings.HasPrefix(line, "num") || strings.HasPrefix(line, "target") {
				continue
			}
			ruleCount++
		}
		if ruleCount > 0 {
			return "iptables: Active (custom rules)"
		}
		return "iptables: Default (no custom rules)"
	}

	return "Firewall: Unknown (ufw/iptables not found)"
}

func checkDiskEncryption() string {
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "lsblk", "-o", "NAME,TYPE,FSTYPE", "--noheadings").CombinedOutput()
	if err == nil && strings.Contains(string(out), "crypto_LUKS") {
		return "Disk Encryption: LUKS enabled"
	}

	out, err = exec.CommandContext(ctx, "cat", "/etc/crypttab").CombinedOutput()
	if err == nil && strings.TrimSpace(string(out)) != "" {
		return "Disk Encryption: Configured (crypttab)"
	}

	return "Disk Encryption: Not detected"
}
