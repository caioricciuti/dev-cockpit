package updater

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Colors for terminal output
const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[0;34m"
	colorNC     = "\033[0m" // No Color
)

// Update performs the complete update process
func Update(opts UpdateOptions) error {
	printBanner()

	// Step 1: Show current version
	printInfo(fmt.Sprintf("Current version: v%s", opts.CurrentVer))

	// Step 2: Check for updates
	printInfo("Checking for updates...")

	release, err := FetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	printInfo(fmt.Sprintf("Latest version: %s", release.TagName))

	// Step 3: Compare versions
	updateAvailable, err := HasUpdate(opts.CurrentVer, release.TagName)
	if err != nil {
		return fmt.Errorf("version comparison failed: %w", err)
	}

	if !updateAvailable {
		printSuccess("Already up to date! You're running the latest version.")
		return nil
	}

	printSuccess("Update available!")
	fmt.Println()

	// If check-only mode, stop here
	if opts.CheckOnly {
		fmt.Printf("Update available: v%s → %s\n", opts.CurrentVer, release.TagName)
		fmt.Println("\nRun 'devcockpit update' to install")
		return nil
	}

	// Show release notes if available
	if release.Body != "" {
		fmt.Printf("Update v%s → %s:\n", opts.CurrentVer, release.TagName)
		// Show first few lines of release notes
		lines := strings.Split(release.Body, "\n")
		for i, line := range lines {
			if i >= 5 { // Limit to first 5 lines
				fmt.Println("  ...")
				break
			}
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}

	// Step 4: Confirm with user (unless --force)
	if !opts.Force {
		if !confirmUpdate() {
			printInfo("Update cancelled. No changes were made.")
			return nil
		}
		fmt.Println()
	}

	// Step 5: Download and verify
	printInfo("Downloading binary...")
	binaryPath, err := DownloadAndVerify(release)
	if err != nil {
		return err
	}

	printSuccess("Downloaded successfully")
	printInfo("Verifying checksum...")
	printSuccess("Checksum verified")
	fmt.Println()

	// Step 6: Install update
	printInfo("Installing update...")
	printWarning("Requesting administrator privileges to install")
	fmt.Println()

	if err := InstallUpdate(binaryPath, defaultInstallPath); err != nil {
		return err
	}

	// Success!
	printCompletion(release.TagName)

	return nil
}

// Helper functions for output

func printBanner() {
	fmt.Println()
	fmt.Printf("%s╔════════════════════════════════════════════╗%s\n", colorBlue, colorNC)
	fmt.Printf("%s║      Dev Cockpit Updater v1.0.0           ║%s\n", colorBlue, colorNC)
	fmt.Printf("%s╚════════════════════════════════════════════╝%s\n", colorBlue, colorNC)
	fmt.Println()
}

func printCompletion(version string) {
	fmt.Println()
	fmt.Printf("%s╔════════════════════════════════════════════╗%s\n", colorGreen, colorNC)
	fmt.Printf("%s║  Dev Cockpit updated to %s! 🚀         ║%s\n", colorGreen, version, colorNC)
	fmt.Printf("%s╚════════════════════════════════════════════╝%s\n", colorGreen, colorNC)
	fmt.Println()
	fmt.Println("Run 'devcockpit --version' to verify")
	fmt.Println()
}

func confirmUpdate() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Continue with update? (y/N): ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
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
