//go:build darwin

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

	fw := checkFirewall()
	fv := checkFileVault()
	sip := checkSIP()
	gk := checkGatekeeper()

	r.Details = append(r.Details,
		fmt.Sprintf("Firewall: %s", fw.label),
		fmt.Sprintf("FileVault: %s", fv.label),
		fmt.Sprintf("SIP: %s", sip.label),
		fmt.Sprintf("Gatekeeper: %s", gk.label),
	)

	disabled := 0
	for _, s := range []secStatus{fw, fv, sip, gk} {
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
		r.Summary = "All security features enabled"
	}

	return r
}

func checkFirewall() secStatus {
	ctx, cancel := context.WithTimeout(context.Background(), securityCmdTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate").CombinedOutput()
	if err != nil {
		out, err = exec.CommandContext(ctx, "defaults", "read", "/Library/Preferences/com.apple.alf", "globalstate").CombinedOutput()
		if err != nil {
			return secStatus{false, "Unknown", "Enable Firewall in System Settings > Network > Firewall"}
		}
		val := strings.TrimSpace(string(out))
		enabled := val != "0"
		label := "Disabled"
		if enabled {
			label = "Enabled"
		}
		return secStatus{enabled, label, "Enable Firewall in System Settings > Network > Firewall"}
	}

	s := strings.ToLower(string(out))
	enabled := strings.Contains(s, "enabled")
	label := "Disabled"
	if enabled {
		label = "Enabled"
	}
	return secStatus{enabled, label, "Enable Firewall in System Settings > Network > Firewall"}
}

func checkFileVault() secStatus {
	ctx, cancel := context.WithTimeout(context.Background(), securityCmdTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "fdesetup", "status").CombinedOutput()
	if err != nil {
		return secStatus{false, "Unknown", "Enable FileVault in System Settings > Privacy & Security"}
	}
	s := strings.ToLower(string(out))
	enabled := strings.Contains(s, "on")
	label := "Off"
	if enabled {
		label = "On"
	}
	return secStatus{enabled, label, "Enable FileVault in System Settings > Privacy & Security"}
}

func checkSIP() secStatus {
	ctx, cancel := context.WithTimeout(context.Background(), securityCmdTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "csrutil", "status").CombinedOutput()
	if err != nil {
		return secStatus{false, "Unknown", "Re-enable SIP via Recovery Mode: csrutil enable"}
	}
	s := strings.ToLower(string(out))
	enabled := strings.Contains(s, "enabled")
	label := "Disabled"
	if enabled {
		label = "Enabled"
	}
	return secStatus{enabled, label, "Re-enable SIP via Recovery Mode: csrutil enable"}
}

func checkGatekeeper() secStatus {
	ctx, cancel := context.WithTimeout(context.Background(), securityCmdTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "spctl", "--status").CombinedOutput()
	if err != nil {
		return secStatus{false, "Unknown", "Enable Gatekeeper: sudo spctl --master-enable"}
	}
	s := strings.ToLower(string(out))
	enabled := strings.Contains(s, "enabled")
	label := "Disabled"
	if enabled {
		label = "Enabled"
	}
	return secStatus{enabled, label, "Enable Gatekeeper: sudo spctl --master-enable"}
}
