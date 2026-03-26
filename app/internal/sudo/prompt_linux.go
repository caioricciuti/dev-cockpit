//go:build linux

package sudo

import (
	"os/exec"
	"strings"
)

func promptPassword() (string, error) {
	// Try systemd-ask-password for a secure terminal prompt
	if path, err := exec.LookPath("systemd-ask-password"); err == nil {
		cmd := exec.Command(path, "Dev Cockpit requires your administrator password:")
		output, err := cmd.Output()
		if err != nil {
			return "", ErrCancelled
		}
		password := strings.TrimSpace(string(output))
		if password == "" {
			return "", ErrCancelled
		}
		return password, nil
	}

	// Fallback: let sudo handle its own terminal prompt via sudo -n in Run()
	return "", ErrCancelled
}
