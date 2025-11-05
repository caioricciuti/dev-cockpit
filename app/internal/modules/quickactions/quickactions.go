package quickactions

import (
	"context"
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

const (
	// Default timeout for commands to prevent hanging
	defaultCommandTimeout = 30 * time.Second
	shortCommandTimeout   = 5 * time.Second
	longCommandTimeout    = 60 * time.Second
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
		if err := runCommandWithTimeout(shortCommandTimeout, cmd[0], cmd[1:]...); err != nil {
			logger.Warn("Failed to set animation preference: %v", err)
		}
	}

	// Restart Dock to apply changes
	logger.Info("Restarting Dock to apply animation changes")
	return runCommandWithTimeout(shortCommandTimeout, "killall", "Dock")
}

func (m *Model) rebuildLaunchServices() error {
	// Try the lsregister command with timeout
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

	// Try to get the Wi-Fi service name (more reliable on Apple Silicon)
	wifiServices := []string{"Wi-Fi", "WiFi", "AirPort"}

	success := false
	for _, service := range wifiServices {
		logger.Debug("Trying WiFi service: %s", service)

		// Turn WiFi off
		if err := runCommandWithTimeout(shortCommandTimeout, "networksetup", "-setairportpower", service, "off"); err == nil {
			logger.Info("WiFi turned off for service: %s", service)

			// Wait 2 seconds
			time.Sleep(2 * time.Second)

			// Turn WiFi on
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

	// Alternative: Try with interface names if service names failed
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

	// First try without sudo to see if it works (with timeout)
	output, err := runCommandWithTimeoutOutput(defaultCommandTimeout, command, args...)

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

	// Try without sudo first (with timeout)
	output, err := runShellWithTimeoutOutput(defaultCommandTimeout, shellCmd)

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
	logger.Debug("Detecting active network interface")

	// Method 1: Get default route interface
	cmd := "route -n get default 2>/dev/null | grep 'interface:' | awk '{print $2}'"
	output, err := runShellWithTimeoutOutput(shortCommandTimeout, cmd)
	if err == nil && len(output) > 0 {
		iface := strings.TrimSpace(string(output))
		if iface != "" {
			logger.Debug("Found interface via route: %s", iface)
			return iface
		}
	}

	// Method 2: Try common interface names
	interfaces := []string{"en0", "en1", "en2"}
	for _, iface := range interfaces {
		// Check if interface is up
		checkCmd := fmt.Sprintf("ifconfig %s 2>/dev/null | grep -q 'status: active'", iface)
		if err := runShellWithTimeout(shortCommandTimeout, checkCmd); err == nil {
			logger.Debug("Found active interface: %s", iface)
			return iface
		}
	}

	// Method 3: Try networksetup
	services := []string{"Wi-Fi", "Ethernet", "en0", "en1"}
	for _, service := range services {
		if err := runCommandWithTimeout(shortCommandTimeout, "networksetup", "-getinfo", service); err == nil {
			logger.Debug("Found interface via networksetup: %s", service)
			return service
		}
	}

	// Final fallback
	logger.Warn("Could not determine network interface, using default: en0")
	return "en0"
}

func (m *Model) resetNetwork() error {
	logger.Info("Starting complete network reset")
	success := 0

	// Step 1: Flush DNS first
	logger.Info("Flushing DNS cache")
	if err := m.flushDNS(); err == nil {
		success++
	}

	// Step 2: Reset WiFi via service names (most reliable)
	wifiServices := []string{"Wi-Fi", "WiFi"}
	for _, service := range wifiServices {
		logger.Debug("Trying to reset WiFi service: %s", service)
		if err := runCommandWithTimeout(shortCommandTimeout, "networksetup", "-setairportpower", service, "off"); err == nil {
			time.Sleep(2 * time.Second)
			if err := runCommandWithTimeout(shortCommandTimeout, "networksetup", "-setairportpower", service, "on"); err == nil {
				logger.Info("Successfully reset WiFi via service: %s", service)
				success++

				// Try to renew DHCP
				if err := runCommandWithTimeout(defaultCommandTimeout, "networksetup", "-setdhcp", service); err == nil {
					logger.Info("Renewed DHCP for: %s", service)
					success++
				}
				break
			}
		}
	}

	// Step 3: Try via interface if service method didn't work
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

	// Method 1: Kill and restart bluetoothd (modern approach)
	logger.Info("Attempting to restart Bluetooth daemon")
	err1 := executeSudoCommand("pkill", "-9", "bluetoothd")

	// Give it time to restart automatically
	if err1 == nil {
		logger.Info("Bluetooth daemon killed, waiting for automatic restart")
		time.Sleep(3 * time.Second)
		return nil
	}

	// Method 2: Toggle Bluetooth via blueutil if available
	logger.Info("Trying alternative Bluetooth toggle method")
	toggleCmd := "blueutil -p 0 && sleep 2 && blueutil -p 1"
	if err := runShellWithTimeout(defaultCommandTimeout, toggleCmd); err == nil {
		logger.Info("Bluetooth toggled successfully via blueutil")
		return nil
	}

	// Method 3: Remove preferences (user-level, no sudo needed)
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

	// Kill Core Audio daemon (requires sudo, auto-restarts)
	err := executeSudoCommand("killall", "-9", "coreaudiod")
	if err != nil {
		logger.Error("Failed to restart audio daemon: %v", err)
		return fmt.Errorf("audio reset requires admin privileges")
	}

	logger.Info("Core Audio daemon restarted")
	// Give it a moment to restart
	time.Sleep(2 * time.Second)
	return nil
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
	logger.Info("Starting Spotlight reindex")

	// System-wide Spotlight reindexing requires sudo
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

	// Wait a moment before re-enabling
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

	// Method 1: Verify Time Machine status (works without sudo)
	logger.Info("Checking Time Machine status")
	if err := runCommandWithTimeout(shortCommandTimeout, "tmutil", "status"); err == nil {
		logger.Info("Time Machine is accessible")
		success++
	}

	// Method 2: List destinations (works without sudo)
	if err := runCommandWithTimeout(shortCommandTimeout, "tmutil", "destinationinfo"); err == nil {
		logger.Info("Time Machine destinations accessible")
		success++
	}

	// Method 3: Clear Time Machine preferences (user level only, no System Preferences opening)
	homeDir, _ := os.UserHomeDir()
	tmPrefs := filepath.Join(homeDir, "Library/Preferences/com.apple.TimeMachine.plist")
	if _, err := os.Stat(tmPrefs); err == nil {
		logger.Info("Removing Time Machine preferences file")
		if err := os.Remove(tmPrefs); err == nil {
			success++
		}
	}

	// Method 4: Try to start a manual backup (requires destination configured)
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

	// Fix home directory permissions (non-recursive, safer)
	logger.Info("Fixing home directory permissions")
	chmodCmd := fmt.Sprintf("chmod 755 %s", homeDir)
	if err := runShellWithTimeout(shortCommandTimeout, chmodCmd); err == nil {
		logger.Info("Home directory permissions fixed")
		success++
	} else {
		logger.Warn("Failed to fix home permissions: %v", err)
	}

	// Fix common subdirectories
	commonDirs := []string{
		filepath.Join(homeDir, "Desktop"),
		filepath.Join(homeDir, "Documents"),
		filepath.Join(homeDir, "Downloads"),
	}

	for _, dir := range commonDirs {
		if _, err := os.Stat(dir); err == nil {
			chmodCmd := fmt.Sprintf("chmod 755 %s", dir)
			if err := runShellWithTimeout(shortCommandTimeout, chmodCmd); err == nil {
				logger.Debug("Fixed permissions for: %s", dir)
				success++
			}
		}
	}

	// Verify disk (read-only, safe operation)
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

// emptyTrashInternal performs robust trash clean with multiple fallbacks.
func emptyTrashInternal() error {
	logger.Info("=== Starting Empty Trash operation ===")

	homeDir, _ := os.UserHomeDir()
	trashPath := filepath.Join(homeDir, ".Trash")
	logger.Debug("Trash path: %s", trashPath)

	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		logger.Error("Trash directory not found: %s", trashPath)
		return fmt.Errorf("trash directory not found")
	}

	// Count items before (with timeout)
	countCmd := fmt.Sprintf("find %s -mindepth 1 2>/dev/null | wc -l", trashPath)
	beforeOutput, _ := runShellWithTimeoutOutput(shortCommandTimeout, countCmd)
	var itemsBefore int
	fmt.Sscanf(strings.TrimSpace(string(beforeOutput)), "%d", &itemsBefore)
	logger.Info("Items in trash before cleanup: %d", itemsBefore)

	if itemsBefore == 0 {
		logger.Info("Trash is already empty")
		return nil
	}

	success := false

	// Method 1: Direct rm -rf (fastest, works without user interaction)
	logger.Info("Attempting direct removal...")
	chflagsCmd := fmt.Sprintf("chflags -R nouchg,noschg,nouappnd,noschg %s/* 2>/dev/null || true", trashPath)
	runShellWithTimeout(shortCommandTimeout, chflagsCmd)

	rmCmd := fmt.Sprintf("rm -rf %s/* %s/.[!.]* 2>/dev/null || true", trashPath, trashPath)
	if err := runShellWithTimeout(defaultCommandTimeout, rmCmd); err == nil {
		logger.Info("Direct removal completed")
		success = true
	} else {
		logger.Warn("Direct rm failed: %v", err)
	}

	// Method 2: Use find with deletion (handles stubborn files)
	if !success {
		logger.Info("Attempting find -delete...")
		// First clear flags
		findFlagsCmd := fmt.Sprintf("find %s -mindepth 1 -exec chflags nouchg,nouappnd {} + 2>/dev/null || true", trashPath)
		runShellWithTimeout(defaultCommandTimeout, findFlagsCmd)

		// Then delete
		findDelCmd := fmt.Sprintf("find %s -mindepth 1 -delete 2>/dev/null", trashPath)
		if err := runShellWithTimeout(defaultCommandTimeout, findDelCmd); err == nil {
			logger.Info("Find deletion completed")
			success = true
		} else {
			logger.Warn("Find deletion failed: %v", err)
		}
	}

	// Method 3: Try with sudo for locked files
	if !success {
		logger.Info("Attempting sudo removal for locked files...")
		sudoCmd := fmt.Sprintf("chflags -R nouchg,nouappnd %s/* 2>/dev/null; rm -rf %s/* %s/.[!.]* 2>/dev/null", trashPath, trashPath, trashPath)
		if err := executeSudoShell(sudoCmd); err == nil {
			logger.Info("Sudo removal succeeded")
			success = true
		} else {
			logger.Warn("Sudo removal failed: %v", err)
		}
	}

	// Verify results
	afterOutput, _ := runShellWithTimeoutOutput(shortCommandTimeout, countCmd)
	var itemsAfter int
	fmt.Sscanf(strings.TrimSpace(string(afterOutput)), "%d", &itemsAfter)
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

	// Verify Downloads directory exists
	if _, err := os.Stat(downloadsPath); os.IsNotExist(err) {
		logger.Error("Downloads directory not found: %s", downloadsPath)
		return fmt.Errorf("downloads directory not found")
	}

	// Count files before cleanup
	countCmd := fmt.Sprintf("find %s -type f -mtime +30 2>/dev/null | wc -l", downloadsPath)
	beforeOutput, _ := runShellWithTimeoutOutput(shortCommandTimeout, countCmd)
	var filesBefore int
	fmt.Sscanf(strings.TrimSpace(string(beforeOutput)), "%d", &filesBefore)
	logger.Info("Files older than 30 days in Downloads: %d", filesBefore)

	if filesBefore == 0 {
		logger.Info("No old files to clean in Downloads")
		return nil
	}

	// Remove files older than 30 days (with timeout to prevent hanging)
	logger.Info("Removing files older than 30 days from Downloads")
	deleteCmd := fmt.Sprintf("find %s -type f -mtime +30 -delete 2>/dev/null", downloadsPath)
	err := runShellWithTimeout(defaultCommandTimeout, deleteCmd)

	if err != nil {
		logger.Error("Failed to clean Downloads: %v", err)
		return fmt.Errorf("failed to clean downloads: %v", err)
	}

	// Count files after cleanup
	afterOutput, _ := runShellWithTimeoutOutput(shortCommandTimeout, countCmd)
	var filesAfter int
	fmt.Sscanf(strings.TrimSpace(string(afterOutput)), "%d", &filesAfter)

	removed := filesBefore - filesAfter
	logger.Info("Removed %d old files from Downloads", removed)

	if removed == 0 && filesBefore > 0 {
		return fmt.Errorf("could not remove old files (they may be in use)")
	}

	return nil
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

// Helper function to run command with timeout
func runCommandWithTimeout(timeout time.Duration, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		logger.Warn("Command timed out after %v: %s %v", timeout, name, args)
		return fmt.Errorf("command timed out")
	}

	return err
}

// Helper function to run command with timeout and get output
func runCommandWithTimeoutOutput(timeout time.Duration, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		logger.Warn("Command timed out after %v: %s %v", timeout, name, args)
		return nil, fmt.Errorf("command timed out")
	}

	return output, err
}

// Helper function to run shell command with timeout
func runShellWithTimeout(timeout time.Duration, shellCmd string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", shellCmd)
	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		logger.Warn("Shell command timed out after %v: %s", timeout, shellCmd)
		return fmt.Errorf("command timed out")
	}

	return err
}

// Helper function to run shell command with timeout and get output
func runShellWithTimeoutOutput(timeout time.Duration, shellCmd string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", shellCmd)
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		logger.Warn("Shell command timed out after %v: %s", timeout, shellCmd)
		return nil, fmt.Errorf("command timed out")
	}

	return output, err
}
