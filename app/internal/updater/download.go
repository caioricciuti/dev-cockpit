package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DownloadAndVerify downloads the binary and checksum, then verifies integrity
func DownloadAndVerify(release *Release) (string, error) {
	// Create temporary directory
	tempDir := filepath.Join("/tmp", fmt.Sprintf("devcockpit-update-%d", os.Getpid()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	binaryPath := filepath.Join(tempDir, binaryName)
	checksumPath := filepath.Join(tempDir, checksumName)

	// Find assets
	binaryAsset := release.FindAsset(binaryName)
	checksumAsset := release.FindAsset(checksumName)

	if binaryAsset == nil || checksumAsset == nil {
		return "", fmt.Errorf("required assets not found in release")
	}

	// Download checksum first (it's small)
	if err := downloadFile(checksumAsset.BrowserDownloadURL, checksumPath); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to download checksum: %w", err)
	}

	// Download binary
	if err := downloadFile(binaryAsset.BrowserDownloadURL, binaryPath); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to download binary: %w", err)
	}

	// Verify checksum
	if err := verifyChecksum(binaryPath, checksumPath); err != nil {
		os.RemoveAll(tempDir)
		return "", err // Already formatted with security warning
	}

	return binaryPath, nil
}

// downloadFile downloads a file from URL to destination
func downloadFile(url, destPath string) error {
	client := &http.Client{Timeout: 5 * time.Minute}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// verifyChecksum verifies the SHA256 checksum of a file
func verifyChecksum(filePath, checksumPath string) error {
	// Read expected checksum from file
	checksumData, err := os.ReadFile(checksumPath)
	if err != nil {
		return fmt.Errorf("failed to read checksum file: %w", err)
	}

	// Parse checksum file (format: "abc123  filename" or just "abc123")
	parts := strings.Fields(string(checksumData))
	if len(parts) == 0 {
		return fmt.Errorf("invalid checksum file: empty")
	}
	expectedChecksum := strings.ToLower(strings.TrimSpace(parts[0]))

	// Calculate actual checksum
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open binary for verification: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}
	actualChecksum := hex.EncodeToString(h.Sum(nil))

	// Compare
	if actualChecksum != expectedChecksum {
		return fmt.Errorf(`⚠️  SECURITY ALERT: Checksum verification failed!

Expected: %s
Actual:   %s

The downloaded file may be corrupted or tampered with.
Update aborted for your safety.`, expectedChecksum, actualChecksum)
	}

	return nil
}
