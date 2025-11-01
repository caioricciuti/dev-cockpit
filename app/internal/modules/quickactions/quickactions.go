package quickactions

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/caioricciuti/dev-cockpit/internal/config"
	"github.com/caioricciuti/dev-cockpit/internal/logger"
	sudohelper "github.com/caioricciuti/dev-cockpit/internal/sudo"
	"github.com/caioricciuti/dev-cockpit/internal/ui/events"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Action represents a quick action
type Action struct {
	Name         string
	Description  string
	Category     string
	Command      func() error
	RequiresSudo bool
}

// Model represents the quick actions module state
type Model struct {
	config        *config.Config
	width         int
	height        int
	actions       []Action
	grouped       map[string][]Action
	categories    []string
	categoryIndex int
	actionIndex   int
	running       bool
	runningAction string
	status        string
	statusType    string // "success", "error", "info"
	spinnerFrame  int
}

// New creates a new quick actions module
func New(cfg *config.Config) *Model {
	m := &Model{
		config:  cfg,
		grouped: make(map[string][]Action),
	}
	m.initActions()
	return m
}

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

// Init initializes the module
func (m *Model) Init() tea.Cmd {
	logger.Info("Quick Actions module ready")
	return nil
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (interface{}, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case events.Focus:
		m.running = false
		m.runningAction = ""
		m.status = ""
		m.statusType = ""
		m.spinnerFrame = 0
		m.clampSelection()
	case events.Blur:
		m.running = false
		m.runningAction = ""

	case spinnerTickMsg:
		if m.running {
			m.spinnerFrame = (m.spinnerFrame + 1) % 10
			return m, m.tickSpinner()
		}

	case tea.KeyMsg:
		if m.running {
			return m, nil
		}

		totalActions := len(m.actions)

		switch msg.String() {
		case "up", "k":
			if m.actionIndex > 0 {
				m.actionIndex--
			}
		case "down", "j":
			if m.actionIndex < totalActions-1 {
				m.actionIndex++
			}
		case "enter", " ":
			if m.actionIndex < totalActions {
				return m, m.executeAction(m.actions[m.actionIndex])
			}
		case "f":
			return m, m.fixAllCommon()
		case "g":
			m.actionIndex = 0
		case "G":
			m.actionIndex = totalActions - 1
		}

	case actionCompleteMsg:
		m.running = false
		m.runningAction = ""
		m.spinnerFrame = 0
		m.status = msg.message
		if msg.success {
			m.statusType = "success"
		} else {
			m.statusType = "error"
		}
	}

	return m, nil
}

// View renders the module
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	return m.renderSimpleList()
}

func (m *Model) renderSimpleList() string {
	m.clampSelection()

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF"))

	categoryHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFA500")).
		MarginTop(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDD"))

	// Status line with proper styling
	statusLine := ""
	if m.running {
		spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinner := spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
		statusLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Bold(true).
			Render(fmt.Sprintf("%s Executing: %s", spinner, m.runningAction))
	} else if m.status != "" {
		statusColor := "#0FD976"
		if m.statusType == "error" {
			statusColor = "#FF6B6B"
		} else if m.statusType == "info" {
			statusColor = "#FFA500"
		}
		statusLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color(statusColor)).
			Bold(true).
			Render(m.status)
	}

	// Build single-column list with category headers
	var content []string
	content = append(content, titleStyle.Render("⚡ QUICK ACTIONS"))
	content = append(content, "")

	// Group actions by category for display
	categories := []string{"Performance", "Network", "System", "Cleanup"}
	currentIndex := 0

	for _, category := range categories {
		categoryActions := m.grouped[category]
		if len(categoryActions) == 0 {
			continue
		}

		// Category header
		content = append(content, categoryHeaderStyle.Render(fmt.Sprintf("━━ %s", category)))

		// Actions in this category
		for _, action := range categoryActions {
			prefix := "  "
			if action.RequiresSudo {
				prefix = "  🔒 "
			} else {
				prefix = "    "
			}

			line := prefix + action.Name

			if currentIndex == m.actionIndex {
				content = append(content, selectedStyle.Render("▶ "+line))
			} else {
				content = append(content, itemStyle.Render("  "+line))
			}

			currentIndex++
		}
		content = append(content, "") // spacing between categories
	}

	// Status and help at bottom
	if statusLine != "" {
		content = append(content, "")
		content = append(content, statusLine)
	}
	content = append(content, "")
	content = append(content, helpStyle.Render("↑/↓ Navigate • Enter Execute • F Fix All Common • Esc Back"))

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func (m *Model) renderCategories(width int) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#333")).
		Padding(1, 1).
		Width(width)

	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888"))

	activeStyle := itemStyle.Copy().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	var lines []string
	for i, category := range m.categories {
		count := len(m.grouped[category])
		label := fmt.Sprintf("%s (%d)", category, count)
		if i == m.categoryIndex {
			lines = append(lines, activeStyle.Render("▶ "+label))
		} else {
			lines = append(lines, itemStyle.Render("  "+label))
		}
	}

	return boxStyle.Render(strings.Join(lines, "\n"))
}

func (m *Model) renderActions(width int) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D9FF")).
		Padding(1, 2).
		Width(width)

	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDD"))

	activeStyle := itemStyle.Copy().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888"))

	actions := m.visibleActions()
	if len(actions) == 0 {
		return boxStyle.Render("No actions available for this category.")
	}

	var builder strings.Builder
	for i, action := range actions {
		name := action.Name
		if action.RequiresSudo {
			name = "🔒 " + name
		}

		if i == m.actionIndex {
			builder.WriteString(activeStyle.Render("▶ " + name))
			builder.WriteString("\n")
			builder.WriteString(descStyle.Render("   " + action.Description))
		} else {
			builder.WriteString(itemStyle.Render("  " + name))
		}
		if i < len(actions)-1 {
			builder.WriteString("\n")
		}
	}

	return boxStyle.Render(builder.String())
}

func (m *Model) rebuildGroups() {
	groups := make(map[string][]Action)

	all := make([]Action, len(m.actions))
	copy(all, m.actions)
	groups["All"] = all

	for _, category := range []string{"Performance", "Network", "System", "Cleanup"} {
		groups[category] = []Action{}
	}

	for _, action := range m.actions {
		groups[action.Category] = append(groups[action.Category], action)
	}

	m.grouped = groups
}

func (m *Model) visibleActions() []Action {
	if len(m.categories) == 0 {
		return nil
	}
	category := m.categories[m.categoryIndex]
	if category == "All" {
		return m.grouped["All"]
	}
	return m.grouped[category]
}

func (m *Model) clampSelection() {
	actions := m.visibleActions()
	if len(actions) == 0 {
		m.actionIndex = 0
		return
	}
	if m.actionIndex < 0 {
		m.actionIndex = 0
	}
	if m.actionIndex >= len(actions) {
		m.actionIndex = len(actions) - 1
	}
}

func (m *Model) shiftCategory(delta int) {
	total := len(m.categories)
	if total == 0 {
		return
	}
	m.categoryIndex = (m.categoryIndex + delta + total) % total
	m.actionIndex = 0
}

// Title returns the module title
func (m *Model) Title() string {
	return "Quick Actions"
}

// HasOpenModal returns true if the module has an open modal/dialog
func (m *Model) HasOpenModal() bool {
	return false
}

func (m *Model) executeAction(action Action) tea.Cmd {
	m.running = true
	m.runningAction = action.Name
	m.status = ""
	m.statusType = ""
	m.spinnerFrame = 0
	logger.Info("User triggered action: %s", action.Name)

	// Start spinner animation
	spinnerCmd := m.tickSpinner()

	return tea.Batch(spinnerCmd, func() tea.Msg {
		logger.Debug("Executing action: %s (RequiresSudo: %v)", action.Name, action.RequiresSudo)
		err := action.Command()

		success := err == nil
		message := ""
		if err != nil {
			message = fmt.Sprintf("✗ %s failed: %v", action.Name, err)
			logger.Error("Action failed: %s, error: %v", action.Name, err)
		} else {
			message = fmt.Sprintf("✓ %s completed successfully", action.Name)
			logger.Info("Action completed successfully: %s", action.Name)
		}

		return actionCompleteMsg{message: message, success: success}
	})
}

func (m *Model) fixAllCommon() tea.Cmd {
	m.running = true
	m.runningAction = "Fix All Common Issues"
	m.status = ""
	m.statusType = ""
	m.spinnerFrame = 0
	logger.Info("Starting common quick fixes")

	// Start spinner animation
	spinnerCmd := m.tickSpinner()

	return tea.Batch(spinnerCmd, func() tea.Msg {
		fixed := 0
		failed := 0

		// Run common fixes
		fixes := []func() error{
			m.flushDNS,
			m.clearRAM,
			m.fixPermissions,
			m.rebuildLaunchServices,
		}

		for _, fix := range fixes {
			if err := fix(); err == nil {
				fixed++
			} else {
				failed++
			}
		}

		success := failed == 0
		message := fmt.Sprintf("✓ Fixed %d common issues", fixed)
		if failed > 0 {
			success = false
			message = fmt.Sprintf("⚠ Fixed %d issues (%d failed)", fixed, failed)
		}

		return actionCompleteMsg{message: message, success: success}
	})
}

// Action implementations
func (m *Model) killHeavyProcesses() error {
	// Get processes using more than 80% CPU
	cmd := "ps aux | awk '$3 > 80 && NR > 1 {print $2}' | head -5"
	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return err
	}

	pids := strings.Fields(strings.TrimSpace(string(output)))
	killed := 0

	for _, pid := range pids {
		if pid != "" {
			// Kill the process
			if err := exec.Command("kill", "-9", pid).Run(); err == nil {
				killed++
			}
		}
	}

	if killed == 0 {
		return fmt.Errorf("No heavy processes found or killed")
	}

	return nil
}

func (m *Model) clearRAM() error {
	// The ONLY reliable way to clear RAM on macOS is with sudo purge
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
		exec.Command(cmd[0], cmd[1:]...).Run()
	}

	// Restart Dock to apply changes
	return exec.Command("killall", "Dock").Run()
}

func (m *Model) rebuildLaunchServices() error {
	return exec.Command(
		"/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister",
		"-kill", "-r", "-domain", "local", "-domain", "system", "-domain", "user",
	).Run()
}

func (m *Model) fixWiFi() error {
	// Get the correct WiFi interface (Apple Silicon compatible)
	networkInterface := getActiveNetworkInterface()

	// Try multiple interface names for WiFi
	wifiInterfaces := []string{networkInterface, "Wi-Fi", "en0", "en1"}

	success := false
	for _, iface := range wifiInterfaces {
		// Turn WiFi off
		if err := exec.Command("networksetup", "-setairportpower", iface, "off").Run(); err == nil {
			// Use shell sleep command
			exec.Command("sh", "-c", "sleep 2").Run()
			// Turn WiFi on
			if err := exec.Command("networksetup", "-setairportpower", iface, "on").Run(); err == nil {
				success = true
				break
			}
		}
	}

	if !success {
		return fmt.Errorf("Could not reset WiFi - interface not found")
	}

	return nil
}

func (m *Model) flushDNS() error {
	// Proper DNS flush requires sudo for full effectiveness
	err1 := executeSudoCommand("dscacheutil", "-flushcache")
	err2 := executeSudoCommand("killall", "-HUP", "mDNSResponder")

	// Both commands should succeed for proper DNS flush
	if err1 != nil && err2 != nil {
		return fmt.Errorf("DNS flush failed - admin privileges required")
	}

	return nil
}

// Helper function for sudo operations with proper error handling
func executeSudoCommand(command string, args ...string) error {
	fullCmd := fmt.Sprintf("%s %v", command, args)
	logger.Debug("executeSudoCommand: Attempting command: %s", fullCmd)

	// First try without sudo to see if it works
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()

	if err == nil {
		logger.Info("Command succeeded without sudo: %s", fullCmd)
		logger.Debug("Output: %s", string(output))
		return nil
	}

	logger.Debug("Command failed without sudo: %s, error: %v, output: %s", fullCmd, err, string(output))

	logger.Info("Attempting with sudo: sudo %s", fullCmd)
	sudoOutput, sudoErr := sudohelper.Run(command, args...)
	if sudoErr != nil {
		if errors.Is(sudoErr, sudohelper.ErrCancelled) {
			logger.Warn("Sudo authentication cancelled by user for command: %s", fullCmd)
			return fmt.Errorf("administrator approval cancelled")
		}
		logger.Error("Sudo command failed: sudo %s, error: %v", fullCmd, sudoErr)
		if sudoOutput != "" {
			logger.Debug("Sudo output: %s", sudoOutput)
		}
		return sudoErr
	}

	if sudoOutput != "" {
		logger.Debug("Sudo output: %s", sudoOutput)
	}
	logger.Info("Sudo command succeeded: sudo %s", fullCmd)
	return nil
}

// Helper function to run shell command with sudo
func executeSudoShell(shellCmd string) error {
	logger.Debug("executeSudoShell: Attempting shell command: %s", shellCmd)

	// Try without sudo first
	cmd := exec.Command("sh", "-c", shellCmd)
	output, err := cmd.CombinedOutput()

	if err == nil {
		logger.Info("Shell command succeeded without sudo: %s", shellCmd)
		logger.Debug("Output: %s", string(output))
		return nil
	}

	logger.Debug("Shell command failed without sudo: %s, error: %v, output: %s", shellCmd, err, string(output))

	logger.Info("Attempting shell with sudo: sudo %s", shellCmd)
	sudoOutput, sudoErr := sudohelper.RunShell(shellCmd)
	if sudoErr != nil {
		if errors.Is(sudoErr, sudohelper.ErrCancelled) {
			logger.Warn("Sudo authentication cancelled by user for shell command: %s", shellCmd)
			return fmt.Errorf("administrator approval cancelled")
		}
		logger.Error("Sudo shell command failed: %s, error: %v", shellCmd, sudoErr)
		if sudoOutput != "" {
			logger.Debug("Sudo shell output: %s", sudoOutput)
		}
		return sudoErr
	}

	if sudoOutput != "" {
		logger.Debug("Sudo shell output: %s", sudoOutput)
	}
	logger.Info("Sudo shell command succeeded: %s", shellCmd)
	return nil
}

// Helper function to detect active network interface (Apple Silicon compatible)
func getActiveNetworkInterface() string {
	// Get active network interface
	cmd := "route get default | grep interface | awk '{print $2}'"
	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		// Fallback to common interfaces
		interfaces := []string{"en0", "en1", "Wi-Fi", "Ethernet"}
		for _, iface := range interfaces {
			if err := exec.Command("networksetup", "-getinfo", iface).Run(); err == nil {
				return iface
			}
		}
		return "en0" // Default fallback
	}
	return strings.TrimSpace(string(output))
}

func (m *Model) resetNetwork() error {
	networkInterface := getActiveNetworkInterface()
	success := 0

	// Method 1: Use networksetup (no sudo required)
	commands := [][]string{
		{"networksetup", "-setairportpower", networkInterface, "off"},
		{"networksetup", "-setairportpower", networkInterface, "on"},
		{"networksetup", "-setdhcp", networkInterface},
	}

	for _, cmd := range commands {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err == nil {
			success++
		}
	}

	// Method 2: Alternative WiFi reset
	if success == 0 {
		wifiCommands := [][]string{
			{"networksetup", "-setairportpower", "Wi-Fi", "off"},
			{"networksetup", "-setairportpower", "Wi-Fi", "on"},
		}
		for _, cmd := range wifiCommands {
			if err := exec.Command(cmd[0], cmd[1:]...).Run(); err == nil {
				success++
			}
		}
	}

	// Always try to flush DNS regardless
	m.flushDNS()

	if success == 0 {
		return fmt.Errorf("Network reset may need admin privileges")
	}

	return nil
}

func (m *Model) fixBluetooth() error {
	// Kill the Bluetooth daemon (requires sudo)
	err1 := executeSudoCommand("pkill", "bluetoothd")

	// Remove user Bluetooth preferences
	homeDir, _ := os.UserHomeDir()
	btPlist := filepath.Join(homeDir, "Library/Preferences/com.apple.Bluetooth.plist")
	os.Remove(btPlist)

	// Restart Bluetooth daemon
	err2 := executeSudoShell("launchctl load /System/Library/LaunchDaemons/com.apple.bluetoothd.plist")

	if err1 != nil && err2 != nil {
		return fmt.Errorf("Bluetooth restart requires admin privileges")
	}

	return nil
}

func (m *Model) fixAudio() error {
	// The proper way to fix audio issues is to restart the audio daemon
	return executeSudoCommand("killall", "coreaudiod")
}

func (m *Model) resetSMC() error {
	// SMC reset is hardware-specific
	// This would need to detect Mac model and provide instructions
	return fmt.Errorf("SMC reset requires manual intervention: Shut down, press Shift-Control-Option on left side + power button")
}

func (m *Model) resetNVRAM() error {
	// NVRAM reset requires admin privileges - provide instructions instead
	return fmt.Errorf("NVRAM reset requires manual steps: Shut down Mac, then hold Option+Command+P+R during startup")
}

func (m *Model) fixSpotlight() error {
	// System-wide Spotlight reindexing requires sudo
	err1 := executeSudoCommand("mdutil", "-i", "off", "/")
	if err1 != nil {
		return err1
	}

	err2 := executeSudoCommand("mdutil", "-E", "/")
	if err2 != nil {
		return err2
	}

	return executeSudoCommand("mdutil", "-i", "on", "/")
}

func (m *Model) fixTimeMachine() error {
	success := 0

	// Method 1: Try tmutil without sudo first
	if err := exec.Command("tmutil", "destinationinfo").Run(); err == nil {
		success++
	}

	// Method 2: Check Time Machine status
	if err := exec.Command("tmutil", "status").Run(); err == nil {
		success++
	}

	// Method 3: Use AppleScript to open Time Machine preferences
	script := `osascript -e 'tell application "System Preferences"
		reveal pane "com.apple.preference.TimeMachine"
		activate
	end tell'`
	if err := exec.Command("sh", "-c", script).Run(); err == nil {
		success++
	}

	// Method 4: Clear Time Machine preferences (user level)
	homeDir, _ := os.UserHomeDir()
	tmPrefs := filepath.Join(homeDir, "Library/Preferences/com.apple.TimeMachine.plist")
	if _, err := os.Stat(tmPrefs); err == nil {
		os.Remove(tmPrefs)
		success++
	}

	if success == 0 {
		return fmt.Errorf("Time Machine management needs admin privileges")
	}

	return nil
}

func (m *Model) fixPermissions() error {
	homeDir, _ := os.UserHomeDir()

	// Fix home directory permissions (be more careful with recursive chmod)
	exec.Command("sh", "-c", fmt.Sprintf("chmod 755 %s", homeDir)).Run()

	// Repair disk permissions (note: this command may not work on newer macOS)
	exec.Command("diskutil", "repairPermissions", "/").Run()

	// Alternative: First Aid on disk
	return exec.Command("diskutil", "verifyDisk", "/").Run()
}

// emptyTrashInternal performs robust trash clean with multiple fallbacks.
func emptyTrashInternal() error {
	logger.Info("=== Starting Empty Trash operation ===")

	homeDir, _ := os.UserHomeDir()
	trashPath := filepath.Join(homeDir, ".Trash")
	logger.Debug("Trash path: %s", trashPath)

	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		logger.Error("Trash directory not found: %s", trashPath)
		return fmt.Errorf("Trash directory not found")
	}

	// Counting helper
	countCmd := fmt.Sprintf("find %s -mindepth 1 2>/dev/null | wc -l", trashPath)
	beforeOutput, _ := exec.Command("sh", "-c", countCmd).Output()
	var itemsBefore int
	fmt.Sscanf(strings.TrimSpace(string(beforeOutput)), "%d", &itemsBefore)
	logger.Info("Items in trash before cleanup: %d", itemsBefore)

	success := false

	// 0) Finder AppleScript first
	asCmd := `osascript -e 'tell application "Finder" to empty trash'`
	if out, err := exec.Command("sh", "-c", asCmd).CombinedOutput(); err == nil {
		logger.Info("Finder emptied trash successfully")
		success = true
	} else {
		logger.Warn("Finder AppleScript failed: %v", err)
		_ = out
	}

	// 1) Remove flags and delete
	if !success {
		chflagsCmd := fmt.Sprintf("chflags -R nouchg,noschg %s/* 2>/dev/null", trashPath)
		_ = exec.Command("sh", "-c", chflagsCmd).Run()

		rmCmd := fmt.Sprintf("rm -rf %s/* 2>/dev/null", trashPath)
		if out, err := exec.Command("sh", "-c", rmCmd).CombinedOutput(); err == nil {
			logger.Info("Removed trash contents with rm")
			success = true
		} else {
			logger.Warn("rm failed: %v", err)
			_ = out
		}
	}

	// 2) Use find to delete with flag clearing
	if !success {
		findCmd := fmt.Sprintf("find %s -mindepth 1 -exec chflags -R nouchg,noschg {} + -delete 2>/dev/null", trashPath)
		if out, err := exec.Command("sh", "-c", findCmd).CombinedOutput(); err == nil {
			logger.Info("Deleted trash contents with find")
			success = true
		} else {
			logger.Warn("find -delete failed: %v", err)
			_ = out
		}
	}

	// 3) As last resort, try sudo removal
	if !success {
		sudoCmd := fmt.Sprintf("chflags -R nouchg,noschg %s/* 2>/dev/null; rm -rf %s/* 2>/dev/null", trashPath, trashPath)
		if err := executeSudoShell(sudoCmd); err == nil {
			logger.Info("Sudo removal succeeded")
			success = true
		} else {
			logger.Error("Sudo removal failed: %v", err)
		}
	}

	// VERIFY
	afterOutput, _ := exec.Command("sh", "-c", countCmd).Output()
	var itemsAfter int
	fmt.Sscanf(strings.TrimSpace(string(afterOutput)), "%d", &itemsAfter)
	logger.Info("Items in trash after cleanup: %d", itemsAfter)

	if itemsAfter == 0 {
		logger.Info("=== Empty Trash SUCCEEDED ===")
		return nil
	}
	if itemsBefore == 0 {
		logger.Warn("Trash was already empty")
		return nil
	}
	return fmt.Errorf("Failed to empty trash: %d items remain", itemsAfter)
}

// EmptyTrash exposes the operation for CLI usage.
func EmptyTrash() error { return emptyTrashInternal() }

func (m *Model) emptyTrash() error { return emptyTrashInternal() }

func (m *Model) cleanDownloads() error {
	homeDir, _ := os.UserHomeDir()
	downloadsPath := filepath.Join(homeDir, "Downloads")

	// Remove files older than 30 days
	return exec.Command("find", downloadsPath, "-mtime", "+30", "-delete").Run()
}

func (m *Model) purgeMemory() error {
	// Same as clearRAM - requires sudo purge
	return executeSudoCommand("purge")
}

// Message types
type actionCompleteMsg struct {
	message string
	success bool
}

type spinnerTickMsg struct{}

func (m *Model) tickSpinner() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}
