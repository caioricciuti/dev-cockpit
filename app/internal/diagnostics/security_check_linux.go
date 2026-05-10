//go:build linux

package diagnostics

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const securityCmdTimeout = 5 * time.Second

func CheckSecurity() CheckResult {
	r := CheckResult{Category: "Security"}

	fw := checkLinuxFirewall()
	enc := checkLinuxDiskEncryption()

	r.Details = append(r.Details,
		fmt.Sprintf("Firewall: %s", fw.label),
		fmt.Sprintf("Disk Encryption: %s", enc.label),
		"SIP: N/A (macOS-only)",
		"Gatekeeper: N/A (macOS-only)",
	)

	disabled := 0
	for _, s := range []secStatus{fw, enc} {
		if !s.enabled {
			disabled++
			r.Suggestions = append(r.Suggestions, s.fix)
		}
	}

	switch {
	case disabled >= 2:
		r.Status = Critical
		r.Summary = fmt.Sprintf("%d security features disabled", disabled)
	case disabled == 1:
		r.Status = Warning
		r.Summary = "1 security feature disabled"
	default:
		r.Status = OK
		r.Summary = "All checked security features enabled"
	}

	return r
}

func checkLinuxFirewall() secStatus {
	ctx, cancel := context.WithTimeout(context.Background(), securityCmdTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ufw", "status").CombinedOutput()
	if err == nil {
		s := strings.ToLower(strings.TrimSpace(string(out)))
		if strings.Contains(s, "active") && !strings.Contains(s, "inactive") {
			return secStatus{true, "UFW Active", ""}
		}
		return secStatus{false, "UFW Inactive", "Enable firewall: sudo ufw enable"}
	}

	out, err = exec.CommandContext(ctx, "iptables", "-L", "-n").CombinedOutput()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		rules := 0
		for _, line := range lines {
			l := strings.TrimSpace(line)
			if l == "" || strings.HasPrefix(l, "Chain") || strings.HasPrefix(l, "target") {
				continue
			}
			rules++
		}
		if rules > 0 {
			return secStatus{true, "iptables Active", ""}
		}
		return secStatus{false, "iptables (no rules)", "Configure iptables or install ufw"}
	}

	return secStatus{false, "Not found", "Install a firewall: sudo apt install ufw"}
}

func checkLinuxDiskEncryption() secStatus {
	ctx, cancel := context.WithTimeout(context.Background(), securityCmdTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "lsblk", "-o", "NAME,TYPE,FSTYPE", "--noheadings").CombinedOutput()
	if err == nil && strings.Contains(string(out), "crypto_LUKS") {
		return secStatus{true, "LUKS Enabled", ""}
	}

	out, _ = exec.CommandContext(ctx, "cat", "/etc/crypttab").CombinedOutput()
	content := strings.TrimSpace(string(out))
	if content != "" && !strings.HasPrefix(content, "#") {
		return secStatus{true, "Configured (crypttab)", ""}
	}

	return secStatus{false, "Not detected", "Consider enabling LUKS disk encryption"}
}
