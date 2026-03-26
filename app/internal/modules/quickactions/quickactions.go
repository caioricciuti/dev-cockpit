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

// initActions is implemented in platform-specific files:
// actions_darwin.go — macOS-specific actions
// actions_linux.go  — Linux-specific actions

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

func (m *Model) clampSelection() {
	total := len(m.actions)
	if total == 0 {
		m.actionIndex = 0
		return
	}
	if m.actionIndex < 0 {
		m.actionIndex = 0
	}
	if m.actionIndex >= total {
		m.actionIndex = total - 1
	}
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

		// Run common fixes (platform-specific)
		for _, fix := range m.commonFixes() {
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

// Platform-specific action implementations are in:
// actions_darwin.go — macOS-specific actions (getActiveNetworkInterface, resetNetwork, fixBluetooth, etc.)
// actions_linux.go  — Linux-specific actions

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

// countDirEntries counts entries in a directory without shell interpolation
func countDirEntries(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	return len(entries)
}

// removeDirectoryContents removes all entries inside a directory without shell interpolation
func removeDirectoryContents(dirPath string, timeout time.Duration) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dirPath, entry.Name())
		if err := os.RemoveAll(entryPath); err != nil {
			// Fallback to rm -rf with proper arguments (no shell)
			_ = runCommandWithTimeout(timeout, "rm", "-rf", entryPath)
		}
	}
	return nil
}

// countLines counts non-empty lines in a string
func countLines(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n"))
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
