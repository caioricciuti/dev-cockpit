package sudo

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

var (
	cachedPassword string
	cacheMutex     sync.Mutex
)

// ErrCancelled is returned when the user dismisses the password prompt.
var ErrCancelled = errors.New("sudo authorization cancelled")

// Run executes a command with sudo privileges, prompting the user for
// their password via a secure macOS dialog when necessary. Output from the
// command is returned as a string (stdout + stderr combined).
func Run(command string, args ...string) (string, error) {
	// First, try to run using any existing sudo session timestamp.
	output, err := exec.Command("sudo", append([]string{"-n", command}, args...)...).CombinedOutput()
	if err == nil {
		return strings.TrimRight(string(output), "\n"), nil
	}

	if !requiresPassword(output, err) {
		return strings.TrimRight(string(output), "\n"), fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}

	password, err := ensurePassword()
	if err != nil {
		return "", err
	}

	sudoArgs := append([]string{"-S", "-p", "", command}, args...)
	cmd := exec.Command("sudo", sudoArgs...)
	cmd.Stdin = strings.NewReader(password + "\n")

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if runErr := cmd.Run(); runErr != nil {
		cacheMutex.Lock()
		if requiresPassword(buf.Bytes(), runErr) {
			cachedPassword = ""
		}
		cacheMutex.Unlock()
		return strings.TrimRight(buf.String(), "\n"), fmt.Errorf("%s", strings.TrimSpace(buf.String()))
	}

	return strings.TrimRight(buf.String(), "\n"), nil
}

// RunShell executes a shell command (`sh -c`) with sudo privileges.
func RunShell(command string) (string, error) {
	return Run("sh", "-c", command)
}

func ensurePassword() (string, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if cachedPassword != "" {
		return cachedPassword, nil
	}

	pwd, err := promptPassword()
	if err != nil {
		return "", err
	}

	cachedPassword = pwd
	return cachedPassword, nil
}

func requiresPassword(output []byte, err error) bool {
	if err == nil {
		return false
	}

	text := strings.ToLower(string(output))
	if strings.Contains(text, "a password is required") || strings.Contains(text, "sorry, try again") {
		return true
	}

	// When sudo exits with status 1 without writing an explicit message it may
	// still indicate that credentials are required.
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		if len(text) == 0 {
			return true
		}
	}

	return false
}

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
