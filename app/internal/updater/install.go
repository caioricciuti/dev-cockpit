package updater

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/caioricciuti/dev-cockpit/internal/sudo"
)

const defaultInstallPath = "/usr/local/bin/devcockpit"

// InstallUpdate replaces the current binary with the new one
// Uses atomic operations with backup and rollback on failure
func InstallUpdate(newBinaryPath, targetPath string) error {
	if targetPath == "" {
		targetPath = defaultInstallPath
	}

	// Create backup path
	backupPath := filepath.Join("/tmp", fmt.Sprintf("devcockpit-backup-%d", os.Getpid()))

	// Check if target exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("current binary not found at %s", targetPath)
	}

	// Step 1: Backup current binary
	if err := copyFile(targetPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	defer os.Remove(backupPath) // Cleanup backup on success

	// Step 2: Replace binary (requires sudo)
	if _, err := sudo.Run("cp", newBinaryPath, targetPath); err != nil {
		// No need to rollback - original is still in place
		if err == sudo.ErrCancelled {
			return fmt.Errorf("update cancelled")
		}
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Step 3: Set permissions
	if _, err := sudo.Run("chmod", "+x", targetPath); err != nil {
		// Rollback: restore from backup
		sudo.Run("cp", backupPath, targetPath)
		return fmt.Errorf("failed to set permissions, rolled back: %w", err)
	}

	// Step 4: Test new binary
	cmd := exec.Command(targetPath, "version")
	if err := cmd.Run(); err != nil {
		// Rollback: restore from backup
		sudo.Run("cp", backupPath, targetPath)
		sudo.Run("chmod", "+x", targetPath)
		return fmt.Errorf("new binary test failed, rolled back: %w", err)
	}

	// Success - cleanup temp directory with new binary
	os.RemoveAll(filepath.Dir(newBinaryPath))

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
