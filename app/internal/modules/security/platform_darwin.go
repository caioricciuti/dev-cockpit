//go:build darwin

package security

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const cmdTimeout = 5 * time.Second

func (m *Model) refresh() tea.Cmd {
	return func() tea.Msg {
		fw := readFirewall()
		fv := readCmd("fdesetup", "status")
		sip := readCmd("csrutil", "status")
		gk := readCmd("spctl", "--status")
		return secMsg{firewall: fw, filevault: fv, sip: sip, gatekeeper: gk, note: "Security status refreshed"}
	}
}

func readFirewall() string {
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate").CombinedOutput()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	out, err = exec.CommandContext(ctx, "defaults", "read", "/Library/Preferences/com.apple.alf", "globalstate").CombinedOutput()
	if err != nil {
		return "Unknown"
	}
	s := strings.TrimSpace(string(out))
	switch s {
	case "0":
		return "Firewall is disabled (0)"
	case "1":
		return "Firewall is enabled (1)"
	case "2":
		return "Firewall is enabled for essential services (2)"
	default:
		return "Firewall state: " + s
	}
}

func readCmd(name string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return fmt.Sprintf("%s error", name)
	}
	return strings.TrimSpace(string(out))
}
