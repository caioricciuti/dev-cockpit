//go:build darwin

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
			Name:         "Clear RAM",
			Description:  "Purge inactive memory",
			Category:     "Performance",
			Command:      m.clearRAM,
			RequiresSudo: true,
		},
		{
			Name:        "Disable Animations",
			Description: "Speed up UI by disabling animations",
			Category:    "Performance",
			Command:     m.disableAnimations,
		},
		{
			Name:        "Rebuild Launch Services",
			Description: "Fix app associations and duplicates",
			Category:    "Performance",
			Command:     m.rebuildLaunchServices,
		},

		// Network Fixes
		{
			Name:        "Fix WiFi",
			Description: "Reset WiFi configuration",
			Category:    "Network",
			Command:     m.fixWiFi,
		},
		{
			Name:         "Flush DNS",
			Description:  "Clear DNS cache",
			Category:     "Network",
			Command:      m.flushDNS,
			RequiresSudo: true,
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
			Description:  "Reset Bluetooth module",
			Category:     "System",
			Command:      m.fixBluetooth,
			RequiresSudo: true,
		},
		{
			Name:         "Fix Audio",
			Description:  "Reset Core Audio",
			Category:     "System",
			Command:      m.fixAudio,
			RequiresSudo: true,
		},
		{
			Name:        "Reset SMC",
			Description: "Reset System Management Controller",
			Category:    "System",
			Command:     m.resetSMC,
		},
		{
			Name:        "Reset NVRAM",
			Description: "Reset Non-Volatile RAM",
			Category:    "System",
			Command:     m.resetNVRAM,
		},
		{
			Name:         "Fix Spotlight",
			Description:  "Rebuild Spotlight index",
			Category:     "System",
			Command:      m.fixSpotlight,
			RequiresSudo: true,
		},
		{
			Name:         "Fix Time Machine",
			Description:  "Reset Time Machine and optimize",
			Category:     "System",
			Command:      m.fixTimeMachine,
			RequiresSudo: true,
		},
		{
			Name:        "Fix Permissions",
			Description: "Repair file permissions",
			Category:    "System",
			Command:     m.fixPermissions,
		},

		// Cleanup
		{
			Name:        "Empty Trash",
			Description: "Securely empty trash",
			Category:    "Cleanup",
			Command:     m.emptyTrash,
		},
		{
			Name:        "Clean Downloads",
			Description: "Remove old downloads",
			Category:    "Cleanup",
			Command:     m.cleanDownloads,
		},
		{
			Name:         "Purge Memory",
			Description:  "Free up inactive RAM",
			Category:     "Cleanup",
			Command:      m.purgeMemory,
			RequiresSudo: true,
		},
	}

	m.categories = []string{"All", "Performance", "Network", "System", "Cleanup"}
	m.rebuildGroups()
}

func (m *Model) commonFixes() []func() error {
	return []func() error{
		m.flushDNS,
		m.clearRAM,
		m.fixPermissions,
		m.rebuildLaunchServices,
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
	return executeSudoCommand("purge")
}

func (m *Model) disableAnimations() error {
	commands := [][]string{
		{"defaults", "write", "NSGlobalDomain", "NSAutomaticWindowAnimationsEnabled", "-bool", "false"},
		{"defaults", "write", "com.apple.dock", "expose-animation-duration", "-float", "0.1"},
		{"defaults", "write", "com.apple.dock", "autohide-time-modifier", "-float", "0"},
		{"defaults", "write", "NSGlobalDomain", "NSWindowResizeTime", "-float", "0.001"},
	}

	for _, cmd := range commands {
		if err := runCommandWithTimeout(shortCommandTimeout, cmd[0], cmd[1:]...); err != nil {
			logger.Warn("Failed to set animation preference: %v", err)
		}
	}

	logger.Info("Restarting Dock to apply animation changes")
	return runCommandWithTimeout(shortCommandTimeout, "killall", "Dock")
}

func (m *Model) rebuildLaunchServices() error {
	lsregisterPath := "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"

	logger.Info("Rebuilding Launch Services database")
	err := runCommandWithTimeout(longCommandTimeout, lsregisterPath,
		"-kill", "-r", "-domain", "local", "-domain", "system", "-domain", "user")

	if err != nil {
		logger.Error("Failed to rebuild launch services: %v", err)
		return fmt.Errorf("failed to rebuild launch services: %v", err)
	}

	logger.Info("Launch Services rebuild completed")
	return nil
}

func (m *Model) fixWiFi() error {
	logger.Info("Starting WiFi reset")

	wifiServices := []string{"Wi-Fi", "WiFi", "AirPort"}

	success := false
	for _, service := range wifiServices {
		logger.Debug("Trying WiFi service: %s", service)

		if err := runCommandWithTimeout(shortCommandTimeout, "networksetup", "-setairportpower", service, "off"); err == nil {
			logger.Info("WiFi turned off for service: %s", service)
			time.Sleep(2 * time.Second)

			if err := runCommandWithTimeout(shortCommandTimeout, "networksetup", "-setairportpower", service, "on"); err == nil {
				logger.Info("WiFi turned on for service: %s", service)
				success = true
				break
			} else {
				logger.Warn("Failed to turn WiFi back on for service %s: %v", service, err)
			}
		} else {
			logger.Debug("Service %s not found or failed: %v", service, err)
		}
	}

	if !success {
		logger.Info("Trying with interface names...")
		networkInterface := getActiveNetworkInterface()
		wifiInterfaces := []string{networkInterface, "en0", "en1"}

		for _, iface := range wifiInterfaces {
			logger.Debug("Trying WiFi interface: %s", iface)

			if err := runCommandWithTimeout(shortCommandTimeout, "networksetup", "-setairportpower", iface, "off"); err == nil {
				time.Sleep(2 * time.Second)
				if err := runCommandWithTimeout(shortCommandTimeout, "networksetup", "-setairportpower", iface, "on"); err == nil {
					logger.Info("WiFi reset successful via interface: %s", iface)
					success = true
					break
				}
			}
		}
	}

	if !success {
		return fmt.Errorf("could not reset WiFi - no working interface/service found")
	}

	return nil
}

func (m *Model) flushDNS() error {
	err1 := executeSudoCommand("dscacheutil", "-flushcache")
	err2 := executeSudoCommand("killall", "-HUP", "mDNSResponder")

	if err1 != nil && err2 != nil {
		return fmt.Errorf("DNS flush failed - admin privileges required")
	}

	return nil
}

func getActiveNetworkInterface() string {
	logger.Debug("Detecting active network interface")

	cmd := "route -n get default 2>/dev/null | grep 'interface:' | awk '{print $2}'"
	output, err := runShellWithTimeoutOutput(shortCommandTimeout, cmd)
	if err == nil && len(output) > 0 {
		iface := strings.TrimSpace(string(output))
		if iface != "" {
			logger.Debug("Found interface via route: %s", iface)
			return iface
		}
	}

	interfaces := []string{"en0", "en1", "en2"}
	for _, iface := range interfaces {
		checkCmd := fmt.Sprintf("ifconfig %s 2>/dev/null | grep -q 'status: active'", iface)
		if err := runShellWithTimeout(shortCommandTimeout, checkCmd); err == nil {
			logger.Debug("Found active interface: %s", iface)
			return iface
		}
	}

	services := []string{"Wi-Fi", "Ethernet", "en0", "en1"}
	for _, service := range services {
		if err := runCommandWithTimeout(shortCommandTimeout, "networksetup", "-getinfo", service); err == nil {
			logger.Debug("Found interface via networksetup: %s", service)
			return service
		}
	}

	logger.Warn("Could not determine network interface, using default: en0")
	return "en0"
}

func (m *Model) resetNetwork() error {
	logger.Info("Starting complete network reset")
	success := 0

	logger.Info("Flushing DNS cache")
	if err := m.flushDNS(); err == nil {
		success++
	}

	wifiServices := []string{"Wi-Fi", "WiFi"}
	for _, service := range wifiServices {
		logger.Debug("Trying to reset WiFi service: %s", service)
		if err := runCommandWithTimeout(shortCommandTimeout, "networksetup", "-setairportpower", service, "off"); err == nil {
			time.Sleep(2 * time.Second)
			if err := runCommandWithTimeout(shortCommandTimeout, "networksetup", "-setairportpower", service, "on"); err == nil {
				logger.Info("Successfully reset WiFi via service: %s", service)
				success++

				if err := runCommandWithTimeout(defaultCommandTimeout, "networksetup", "-setdhcp", service); err == nil {
					logger.Info("Renewed DHCP for: %s", service)
					success++
				}
				break
			}
		}
	}

	if success < 2 {
		networkInterface := getActiveNetworkInterface()
		logger.Debug("Trying to reset via interface: %s", networkInterface)

		if err := runCommandWithTimeout(shortCommandTimeout, "networksetup", "-setairportpower", networkInterface, "off"); err == nil {
			time.Sleep(2 * time.Second)
			if err := runCommandWithTimeout(shortCommandTimeout, "networksetup", "-setairportpower", networkInterface, "on"); err == nil {
				logger.Info("Successfully reset WiFi via interface: %s", networkInterface)
				success++
			}
		}
	}

	if success == 0 {
		return fmt.Errorf("network reset failed - try checking Network settings manually")
	}

	logger.Info("Network reset completed with %d successful operations", success)
	return nil
}

func (m *Model) fixBluetooth() error {
	logger.Info("Starting Bluetooth reset")

	logger.Info("Attempting to restart Bluetooth daemon")
	err1 := executeSudoCommand("pkill", "-9", "bluetoothd")

	if err1 == nil {
		logger.Info("Bluetooth daemon killed, waiting for automatic restart")
		time.Sleep(3 * time.Second)
		return nil
	}

	logger.Info("Trying alternative Bluetooth toggle method")
	toggleCmd := "blueutil -p 0 && sleep 2 && blueutil -p 1"
	if err := runShellWithTimeout(defaultCommandTimeout, toggleCmd); err == nil {
		logger.Info("Bluetooth toggled successfully via blueutil")
		return nil
	}

	homeDir, _ := os.UserHomeDir()
	btPlist := filepath.Join(homeDir, "Library/Preferences/com.apple.Bluetooth.plist")
	if _, err := os.Stat(btPlist); err == nil {
		logger.Info("Removing Bluetooth preferences file")
		os.Remove(btPlist)
	}

	if err1 != nil {
		return fmt.Errorf("bluetooth reset requires admin privileges")
	}

	return nil
}

func (m *Model) fixAudio() error {
	logger.Info("Starting audio system reset")

	err := executeSudoCommand("killall", "-9", "coreaudiod")
	if err != nil {
		logger.Error("Failed to restart audio daemon: %v", err)
		return fmt.Errorf("audio reset requires admin privileges")
	}

	logger.Info("Core Audio daemon restarted")
	time.Sleep(2 * time.Second)
	return nil
}

func (m *Model) resetSMC() error {
	return fmt.Errorf("SMC reset requires manual intervention: Shut down, press Shift-Control-Option on left side + power button")
}

func (m *Model) resetNVRAM() error {
	return fmt.Errorf("NVRAM reset requires manual steps: Shut down Mac, then hold Option+Command+P+R during startup")
}

func (m *Model) fixSpotlight() error {
	logger.Info("Starting Spotlight reindex")

	logger.Info("Disabling Spotlight indexing")
	err1 := executeSudoCommand("mdutil", "-i", "off", "/")
	if err1 != nil {
		logger.Error("Failed to disable Spotlight: %v", err1)
		return fmt.Errorf("spotlight reset requires admin privileges")
	}

	logger.Info("Erasing Spotlight index")
	err2 := executeSudoCommand("mdutil", "-E", "/")
	if err2 != nil {
		logger.Warn("Failed to erase index, continuing anyway: %v", err2)
	}

	time.Sleep(2 * time.Second)

	logger.Info("Re-enabling Spotlight indexing")
	err3 := executeSudoCommand("mdutil", "-i", "on", "/")
	if err3 != nil {
		logger.Error("Failed to re-enable Spotlight: %v", err3)
		return err3
	}

	logger.Info("Spotlight reindex started successfully")
	return nil
}

func (m *Model) fixTimeMachine() error {
	logger.Info("Starting Time Machine optimization")
	success := 0

	logger.Info("Checking Time Machine status")
	if err := runCommandWithTimeout(shortCommandTimeout, "tmutil", "status"); err == nil {
		logger.Info("Time Machine is accessible")
		success++
	}

	if err := runCommandWithTimeout(shortCommandTimeout, "tmutil", "destinationinfo"); err == nil {
		logger.Info("Time Machine destinations accessible")
		success++
	}

	homeDir, _ := os.UserHomeDir()
	tmPrefs := filepath.Join(homeDir, "Library/Preferences/com.apple.TimeMachine.plist")
	if _, err := os.Stat(tmPrefs); err == nil {
		logger.Info("Removing Time Machine preferences file")
		if err := os.Remove(tmPrefs); err == nil {
			success++
		}
	}

	logger.Info("Attempting to trigger Time Machine backup check")
	if err := runCommandWithTimeout(shortCommandTimeout, "tmutil", "startbackup", "-b"); err == nil {
		logger.Info("Time Machine backup check initiated")
		success++
	}

	if success == 0 {
		return fmt.Errorf("time Machine is not configured or requires admin privileges")
	}

	logger.Info("Time Machine optimization completed (%d operations successful)", success)
	return nil
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

	logger.Info("Verifying disk permissions")
	if err := runCommandWithTimeout(defaultCommandTimeout, "diskutil", "verifyVolume", "/"); err == nil {
		logger.Info("Disk verification completed")
		success++
	} else {
		logger.Warn("Disk verification failed: %v", err)
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
	trashPath := filepath.Join(homeDir, ".Trash")
	logger.Debug("Trash path: %s", trashPath)

	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		logger.Error("Trash directory not found: %s", trashPath)
		return fmt.Errorf("trash directory not found")
	}

	itemsBefore := countDirEntries(trashPath)
	logger.Info("Items in trash before cleanup: %d", itemsBefore)

	if itemsBefore == 0 {
		logger.Info("Trash is already empty")
		return nil
	}

	success := false

	logger.Info("Attempting direct removal...")
	_ = runCommandWithTimeout(shortCommandTimeout, "chflags", "-R", "nouchg,noschg,nouappnd", trashPath)

	if err := removeDirectoryContents(trashPath, defaultCommandTimeout); err == nil {
		logger.Info("Direct removal completed")
		success = true
	} else {
		logger.Warn("Direct removal failed: %v", err)
	}

	if !success {
		logger.Info("Attempting find -delete...")
		_ = runCommandWithTimeout(defaultCommandTimeout, "find", trashPath, "-mindepth", "1", "-exec", "chflags", "nouchg,nouappnd", "{}", "+")

		if err := runCommandWithTimeout(defaultCommandTimeout, "find", trashPath, "-mindepth", "1", "-delete"); err == nil {
			logger.Info("Find deletion completed")
			success = true
		} else {
			logger.Warn("Find deletion failed: %v", err)
		}
	}

	if !success {
		logger.Info("Attempting sudo removal for locked files...")
		if err := executeSudoCommand("chflags", "-R", "nouchg,nouappnd", trashPath); err == nil {
			if err := removeDirectoryContents(trashPath, defaultCommandTimeout); err == nil {
				logger.Info("Sudo removal succeeded")
				success = true
			}
		}
		if !success {
			logger.Warn("Sudo removal failed")
		}
	}

	itemsAfter := countDirEntries(trashPath)
	logger.Info("Items in trash after cleanup: %d", itemsAfter)

	if itemsAfter == 0 {
		logger.Info("=== Empty Trash SUCCEEDED ===")
		return nil
	}

	if itemsAfter < itemsBefore {
		logger.Info("Partially cleaned: removed %d items, %d remain", itemsBefore-itemsAfter, itemsAfter)
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

func (m *Model) purgeMemory() error {
	return executeSudoCommand("purge")
}
