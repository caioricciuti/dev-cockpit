package uninstaller

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/caioricciuti/dev-cockpit/internal/sudo"
)

const (
	binaryName         = "devcockpit"
	installDir         = "/usr/local/bin"
	configDirName      = ".devcockpit"
	fallbackConfigDir  = "./.devcockpit"
)

// Colors for terminal output
const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[0;34m"
	colorNC     = "\033[0m" // No Color
)

// Uninstall performs the complete uninstallation process
func Uninstall(force bool) error {
	printBanner()

	if !force {
		if !confirmUninstall() {
			printInfo("Uninstallation cancelled")
			return nil
		}
		fmt.Println()
	}

	if err := checkRunning(); err != nil {
		return err
	}

	if err := removeBinary(); err != nil {
		return err
	}

	if err := removeConfig(); err != nil {
		return err
	}

	removeFallbackConfig()
	removeTempFiles()

	printCompletion()
	return nil
}

func printBanner() {
	fmt.Println()
	fmt.Printf("%s╔════════════════════════════════════════════╗%s\n", colorBlue, colorNC)
	fmt.Printf("%s║      Dev Cockpit Uninstaller v1.0.0       ║%s\n", colorBlue, colorNC)
	fmt.Printf("%s╚════════════════════════════════════════════╝%s\n", colorBlue, colorNC)
	fmt.Println()
}

func confirmUninstall() bool {
	printWarning("This will remove Dev Cockpit from your system")
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Are you sure you want to continue? (y/N): ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func checkRunning() error {
	// Check if Dev Cockpit is running
	cmd := exec.Command("pgrep", "-x", binaryName)
	if err := cmd.Run(); err == nil {
		// Process is running
		printWarning("Dev Cockpit is currently running")
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Do you want to stop it? (y/N): ")
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "y" || response == "yes" {
			printInfo("Stopping Dev Cockpit...")

			// Try graceful termination first
			exec.Command("pkill", "-TERM", binaryName).Run()

			// Wait a moment
			exec.Command("sleep", "2").Run()

			// Force kill if still running
			if cmd := exec.Command("pgrep", "-x", binaryName); cmd.Run() == nil {
				exec.Command("pkill", "-KILL", binaryName).Run()
			}

			printSuccess("Dev Cockpit stopped")
		} else {
			return fmt.Errorf("please stop Dev Cockpit before uninstalling")
		}
	}
	return nil
}

func removeBinary() error {
	binaryPath := filepath.Join(installDir, binaryName)

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		printInfo(fmt.Sprintf("Binary not found at %s (already removed or never installed)", binaryPath))
		return nil
	}

	printInfo(fmt.Sprintf("Removing binary from %s...", binaryPath))

	// Try to remove without sudo first
	if err := os.Remove(binaryPath); err != nil {
		// Need sudo
		printWarning("Requesting administrator privileges to remove binary")
		if _, err := sudo.Run("rm", "-f", binaryPath); err != nil {
			return fmt.Errorf("failed to remove binary: %w", err)
		}
	}

	printSuccess("Binary removed")
	return nil
}

func removeConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, configDirName)

	// Check if config directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		printInfo(fmt.Sprintf("No configuration directory found at %s", configDir))
		return nil
	}

	printInfo(fmt.Sprintf("Found configuration directory: %s", configDir))

	// Show what will be deleted
	configFile := filepath.Join(configDir, "config.yaml")
	debugLog := filepath.Join(configDir, "debug.log")
	dataDir := filepath.Join(configDir, "data")

	if _, err := os.Stat(configFile); err == nil {
		printInfo("  - config.yaml")
	}
	if _, err := os.Stat(debugLog); err == nil {
		printInfo("  - debug.log")
	}
	if _, err := os.Stat(dataDir); err == nil {
		printInfo("  - data directory")
	}

	// Show directory size
	cmd := exec.Command("du", "-sh", configDir)
	if output, err := cmd.Output(); err == nil {
		size := strings.Fields(string(output))[0]
		printInfo(fmt.Sprintf("  Total size: %s", size))
	}

	fmt.Println()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Remove configuration and data? (y/N): ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "y" || response == "yes" {
		if err := os.RemoveAll(configDir); err != nil {
			return fmt.Errorf("failed to remove config directory: %w", err)
		}
		printSuccess("Configuration directory removed")
	} else {
		printInfo(fmt.Sprintf("Configuration directory kept at %s", configDir))
	}

	return nil
}

func removeFallbackConfig() {
	// Check if fallback config exists
	if _, err := os.Stat(fallbackConfigDir); os.IsNotExist(err) {
		return
	}

	printWarning(fmt.Sprintf("Found fallback config directory: %s", fallbackConfigDir))
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Remove it? (y/N): ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "y" || response == "yes" {
		if err := os.RemoveAll(fallbackConfigDir); err == nil {
			printSuccess("Fallback config directory removed")
		}
	}
}

func removeTempFiles() {
	printInfo("Checking for temporary files...")

	foundTemp := false
	tempCount := 0

	// Check for temp files matching pattern
	files, err := filepath.Glob("/tmp/devcockpit-*")
	if err == nil && len(files) > 0 {
		for _, file := range files {
			if err := os.Remove(file); err == nil {
				foundTemp = true
				tempCount++
			}
		}
	}

	if foundTemp {
		printSuccess(fmt.Sprintf("Removed %d temporary file(s)", tempCount))
	} else {
		printInfo("No temporary files found")
	}
}

func printCompletion() {
	fmt.Println()
	fmt.Printf("%s╔════════════════════════════════════════════╗%s\n", colorGreen, colorNC)
	fmt.Printf("%s║  Dev Cockpit uninstalled successfully! ✓   ║%s\n", colorGreen, colorNC)
	fmt.Printf("%s╚════════════════════════════════════════════╝%s\n", colorGreen, colorNC)
	fmt.Println()
	fmt.Printf("%sThank you for using Dev Cockpit!%s\n", colorBlue, colorNC)
	fmt.Println()
	fmt.Println("If you encountered any issues, please report them at:")
	fmt.Println("  https://github.com/caioricciuti/dev-cockpit/issues")
	fmt.Println()
	fmt.Println("To reinstall in the future:")
	fmt.Println("  curl -fsSL https://raw.githubusercontent.com/caioricciuti/dev-cockpit/main/install.sh | bash")
	fmt.Println()
}

func printInfo(msg string) {
	fmt.Printf("%sℹ%s %s\n", colorBlue, colorNC, msg)
}

func printSuccess(msg string) {
	fmt.Printf("%s✓%s %s\n", colorGreen, colorNC, msg)
}

func printWarning(msg string) {
	fmt.Printf("%s⚠%s %s\n", colorYellow, colorNC, msg)
}

func printError(msg string) {
	fmt.Printf("%s✗%s %s\n", colorRed, colorNC, msg)
}
