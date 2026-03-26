//go:build darwin

package sudo

import (
	"fmt"
	"os/exec"
	"strings"
)

func promptPassword() (string, error) {
	script := `tell application "System Events"
activate
with timeout of 120 seconds
    display dialog "Dev Cockpit requires your administrator password to continue." default answer "" with hidden answer buttons {"Cancel", "Allow"} default button "Allow" with icon caution
end timeout
end tell`

	cmd := exec.Command("osascript",
		"-e", script,
		"-e", "text returned of result",
	)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", ErrCancelled
		}
		return "", fmt.Errorf("failed to request administrator approval: %w", err)
	}

	password := strings.TrimSpace(string(output))
	if password == "" {
		return "", ErrCancelled
	}

	return password, nil
}
