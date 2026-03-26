//go:build linux

package quickactions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/caioricciuti/dev-cockpit/internal/logger"
)

func (m *Model) initActions() {
	m.actions = []Action{
		// Performance
		{
			Name:        "Kill Heavy Processes",
			Description: "Terminate resource-intensive processes",
			Category:    "Performance",
			Command:     m.killHeavyProcesses,
		},
		{
			Name:         "Clear RAM Cache",
			Description:  "Drop kernel caches to free memory",
			Category:     "Performance",
			Command:      m.clearRAM,
			RequiresSudo: true,
		},

		// Network Fixes
		{
			Name:        "Fix WiFi",
			Description: "Reset WiFi via NetworkManager",
			Category:    "Network",
			Command:     m.fixWiFi,
		},
		{
			Name:        "Flush DNS",
			Description: "Clear DNS resolver cache",
			Category:    "Network",
			Command:     m.flushDNS,
		},
		{
			Name:        "Reset Network",
			Description: "Complete network reset",
			Category:    "Network",
			Command:     m.resetNetwork,
		},

		// System Fixes
		{
			Name:         "Fix Bluetooth",
			Description:  "Restart Bluetooth service",
			Category:     "System",
			Command:      m.fixBluetooth,
			RequiresSudo: true,
		},
		{
			Name:         "Fix Audio",
			Description:  "Restart audio subsystem",
			Category:     "System",
			Command:      m.fixAudio,
			RequiresSudo: false,
		},
		{
			Name:        "Fix Permissions",
			Description: "Repair home directory permissions",
			Category:    "System",
			Command:     m.fixPermissions,
		},

		// Cleanup
		{
			Name:        "Empty Trash",
			Description: "Empty XDG trash",
			Category:    "Cleanup",
			Command:     m.emptyTrash,
		},
		{
			Name:        "Clean Downloads",
			Description: "Remove old downloads",
			Category:    "Cleanup",
			Command:     m.cleanDownloads,
		},
	}

	m.categories = []string{"All", "Performance", "Network", "System", "Cleanup"}
	m.rebuildGroups()
}

func (m *Model) commonFixes() []func() error {
	return []func() error{
		m.flushDNS,
		m.fixPermissions,
	}
}

// Action implementations

func (m *Model) killHeavyProcesses() error {
	cmd := "ps aux | awk '$3 > 80 && NR > 1 {print $2}' | head -5"
	output, err := runShellWithTimeoutOutput(shortCommandTimeout, cmd)
	if err != nil {
		return fmt.Errorf("failed to get heavy processes: %v", err)
	}

	pids := strings.Fields(strings.TrimSpace(string(output)))
	killed := 0

	for _, pid := range pids {
		if pid != "" {
			logger.Debug("Attempting to kill process %s", pid)
			if err := runCommandWithTimeout(shortCommandTimeout, "kill", "-9", pid); err == nil {
				killed++
				logger.Info("Killed process %s", pid)
			} else {
				logger.Warn("Failed to kill process %s: %v", pid, err)
			}
		}
	}

	if killed == 0 {
		return fmt.Errorf("no heavy processes found or killed")
	}

	logger.Info("Killed %d heavy processes", killed)
	return nil
}

func (m *Model) clearRAM() error {
	// Drop page cache, dentries and inodes
	return executeSudoShell("sync && echo 3 > /proc/sys/vm/drop_caches")
}

func (m *Model) fixWiFi() error {
	logger.Info("Starting WiFi reset")

	// Try NetworkManager first
	if err := runCommandWithTimeout(shortCommandTimeout, "nmcli", "radio", "wifi", "off"); err == nil {
		time.Sleep(2 * time.Second)
		if err := runCommandWithTimeout(shortCommandTimeout, "nmcli", "radio", "wifi", "on"); err == nil {
			logger.Info("WiFi reset via NetworkManager")
			return nil
		}
	}

	// Fallback: try rfkill
	if err := runCommandWithTimeout(shortCommandTimeout, "rfkill", "block", "wifi"); err == nil {
		time.Sleep(2 * time.Second)
		if err := runCommandWithTimeout(shortCommandTimeout, "rfkill", "unblock", "wifi"); err == nil {
			logger.Info("WiFi reset via rfkill")
			return nil
		}
	}

	return fmt.Errorf("could not reset WiFi - nmcli and rfkill not available")
}

func (m *Model) flushDNS() error {
	// Try systemd-resolved
	if err := runCommandWithTimeout(shortCommandTimeout, "resolvectl", "flush-caches"); err == nil {
		logger.Info("DNS cache flushed via resolvectl")
		return nil
	}

	// Fallback: systemd-resolve (older systems)
	if err := runCommandWithTimeout(shortCommandTimeout, "systemd-resolve", "--flush-caches"); err == nil {
		logger.Info("DNS cache flushed via systemd-resolve")
		return nil
	}

	// Try nscd
	if err := executeSudoCommand("nscd", "-i", "hosts"); err == nil {
		logger.Info("DNS cache flushed via nscd")
		return nil
	}

	return fmt.Errorf("DNS flush failed - no supported resolver found")
}

func (m *Model) resetNetwork() error {
	logger.Info("Starting complete network reset")
	success := 0

	// Flush DNS
	if err := m.flushDNS(); err == nil {
		success++
	}

	// Try NetworkManager
	if err := runCommandWithTimeout(shortCommandTimeout, "nmcli", "networking", "off"); err == nil {
		time.Sleep(2 * time.Second)
		if err := runCommandWithTimeout(shortCommandTimeout, "nmcli", "networking", "on"); err == nil {
			logger.Info("Network reset via NetworkManager")
			success++
		}
	}

	// Fallback: restart networking service
	if success < 2 {
		for _, svc := range []string{"NetworkManager", "networking", "systemd-networkd"} {
			if err := executeSudoCommand("systemctl", "restart", svc); err == nil {
				logger.Info("Restarted %s", svc)
				success++
				break
			}
		}
	}

	if success == 0 {
		return fmt.Errorf("network reset failed - check your network configuration manually")
	}

	logger.Info("Network reset completed with %d successful operations", success)
	return nil
}

func (m *Model) fixBluetooth() error {
	logger.Info("Starting Bluetooth reset")

	if err := executeSudoCommand("systemctl", "restart", "bluetooth"); err == nil {
		logger.Info("Bluetooth service restarted")
		return nil
	}

	// Fallback: rfkill
	if err := runCommandWithTimeout(shortCommandTimeout, "rfkill", "block", "bluetooth"); err == nil {
		time.Sleep(2 * time.Second)
		if err := runCommandWithTimeout(shortCommandTimeout, "rfkill", "unblock", "bluetooth"); err == nil {
			logger.Info("Bluetooth reset via rfkill")
			return nil
		}
	}

	return fmt.Errorf("bluetooth reset failed")
}

func (m *Model) fixAudio() error {
	logger.Info("Starting audio system reset")

	// Try PipeWire first (modern distros)
	if err := runCommandWithTimeout(shortCommandTimeout, "systemctl", "--user", "restart", "pipewire"); err == nil {
		logger.Info("PipeWire restarted")
		return nil
	}

	// Try PulseAudio
	if err := runCommandWithTimeout(shortCommandTimeout, "pulseaudio", "-k"); err == nil {
		time.Sleep(2 * time.Second)
		logger.Info("PulseAudio restarted")
		return nil
	}

	return fmt.Errorf("audio reset failed - neither PipeWire nor PulseAudio found")
}

func (m *Model) fixPermissions() error {
	logger.Info("Starting permissions repair")
	homeDir, _ := os.UserHomeDir()
	success := 0

	logger.Info("Fixing home directory permissions")
	if err := runCommandWithTimeout(shortCommandTimeout, "chmod", "755", homeDir); err == nil {
		logger.Info("Home directory permissions fixed")
		success++
	} else {
		logger.Warn("Failed to fix home permissions: %v", err)
	}

	commonDirs := []string{
		filepath.Join(homeDir, "Desktop"),
		filepath.Join(homeDir, "Documents"),
		filepath.Join(homeDir, "Downloads"),
	}

	for _, dir := range commonDirs {
		if _, err := os.Stat(dir); err == nil {
			if err := runCommandWithTimeout(shortCommandTimeout, "chmod", "755", dir); err == nil {
				logger.Debug("Fixed permissions for: %s", dir)
				success++
			}
		}
	}

	// Verify filesystem (read-only check)
	logger.Info("Verifying filesystem")
	if err := executeSudoCommand("fsck", "-n", "/"); err == nil {
		logger.Info("Filesystem verification completed")
		success++
	} else {
		logger.Warn("Filesystem verification failed: %v", err)
	}

	if success == 0 {
		return fmt.Errorf("permission repair failed")
	}

	logger.Info("Permissions repair completed (%d operations successful)", success)
	return nil
}

func emptyTrashInternal() error {
	logger.Info("=== Starting Empty Trash operation ===")

	homeDir, _ := os.UserHomeDir()
	// XDG trash path
	trashPath := filepath.Join(homeDir, ".local/share/Trash/files")
	trashInfoPath := filepath.Join(homeDir, ".local/share/Trash/info")
	logger.Debug("Trash path: %s", trashPath)

	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		logger.Info("Trash directory not found or already empty")
		return nil
	}

	itemsBefore := countDirEntries(trashPath)
	logger.Info("Items in trash before cleanup: %d", itemsBefore)

	if itemsBefore == 0 {
		logger.Info("Trash is already empty")
		return nil
	}

	// Remove files
	if err := removeDirectoryContents(trashPath, defaultCommandTimeout); err != nil {
		logger.Warn("Direct removal failed: %v", err)
		// Fallback with find
		if err := runCommandWithTimeout(defaultCommandTimeout, "find", trashPath, "-mindepth", "1", "-delete"); err != nil {
			return fmt.Errorf("failed to empty trash: %v", err)
		}
	}

	// Also clean trash info files
	if _, err := os.Stat(trashInfoPath); err == nil {
		_ = removeDirectoryContents(trashInfoPath, defaultCommandTimeout)
	}

	itemsAfter := countDirEntries(trashPath)
	logger.Info("Items in trash after cleanup: %d", itemsAfter)

	if itemsAfter == 0 {
		logger.Info("=== Empty Trash SUCCEEDED ===")
		return nil
	}

	if itemsAfter < itemsBefore {
		return fmt.Errorf("partially cleaned: %d items remain (some may be in use)", itemsAfter)
	}

	return fmt.Errorf("failed to empty trash: %d items remain", itemsAfter)
}

// EmptyTrash exposes the operation for CLI usage.
func EmptyTrash() error { return emptyTrashInternal() }

func (m *Model) emptyTrash() error { return emptyTrashInternal() }

func (m *Model) cleanDownloads() error {
	logger.Info("Starting Downloads cleanup")
	homeDir, _ := os.UserHomeDir()
	downloadsPath := filepath.Join(homeDir, "Downloads")

	if _, err := os.Stat(downloadsPath); os.IsNotExist(err) {
		logger.Error("Downloads directory not found: %s", downloadsPath)
		return fmt.Errorf("downloads directory not found")
	}

	beforeOutput, _ := runCommandWithTimeoutOutput(shortCommandTimeout, "find", downloadsPath, "-type", "f", "-mtime", "+30")
	filesBefore := countLines(string(beforeOutput))
	logger.Info("Files older than 30 days in Downloads: %d", filesBefore)

	if filesBefore == 0 {
		logger.Info("No old files to clean in Downloads")
		return nil
	}

	logger.Info("Removing files older than 30 days from Downloads")
	err := runCommandWithTimeout(defaultCommandTimeout, "find", downloadsPath, "-type", "f", "-mtime", "+30", "-delete")

	if err != nil {
		logger.Error("Failed to clean Downloads: %v", err)
		return fmt.Errorf("failed to clean downloads: %v", err)
	}

	afterOutput, _ := runCommandWithTimeoutOutput(shortCommandTimeout, "find", downloadsPath, "-type", "f", "-mtime", "+30")
	filesAfter := countLines(string(afterOutput))

	removed := filesBefore - filesAfter
	logger.Info("Removed %d old files from Downloads", removed)

	if removed == 0 && filesBefore > 0 {
		return fmt.Errorf("could not remove old files (they may be in use)")
	}

	return nil
}
