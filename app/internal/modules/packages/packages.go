package packages

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/caioricciuti/dev-cockpit/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PackageManager represents a package manager
type PackageManager struct {
	Name        string
	Binary      string
	Installed   bool
	Version     string
	PackageCount int
	Outdated    int
	CacheSize   string
}

// Model represents the packages module state
type Model struct {
	config        *config.Config
	width         int
	height        int
	managers      []PackageManager
	cursor        int
	loading       bool
	executing     bool
	output        string
	message       string
	showingList   bool
	showingOutput bool
	packageList   []string
	listScroll    int
	searchFilter  string
}

// New creates a new packages module
func New(cfg *config.Config) *Model {
	return &Model{
		config:  cfg,
		loading: true,
	}
}

// Init initializes the module
func (m *Model) Init() tea.Cmd {
	return m.detectManagers()
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (interface{}, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		// Handle output screen - any key dismisses
		if m.showingOutput {
			m.showingOutput = false
			m.output = ""
			return m, nil
		}

		// Handle package list modal separately
		if m.showingList {
			key := strings.ToLower(msg.String())
			switch key {
			case "esc", "q":
				m.showingList = false
				m.packageList = nil
				m.listScroll = 0
				m.searchFilter = ""
				return m, nil

			case "up", "k":
				if m.listScroll > 0 {
					m.listScroll--
				}

			case "down", "j":
				if m.listScroll < len(m.getFilteredPackages())-1 {
					m.listScroll++
				}

			case "backspace":
				if len(m.searchFilter) > 0 {
					m.searchFilter = m.searchFilter[:len(m.searchFilter)-1]
					m.listScroll = 0
				}

			default:
				// Add character to search filter
				if len(msg.String()) == 1 && msg.String() >= " " {
					m.searchFilter += msg.String()
					m.listScroll = 0
				}
			}
			return m, nil
		}

		if m.loading || m.executing {
			return m, nil
		}

		key := strings.ToLower(msg.String())

		switch key {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.managers)-1 {
				m.cursor++
			}

		case "c":
			// Cleanup cache for current manager
			if m.cursor < len(m.managers) && m.managers[m.cursor].Installed {
				return m, m.cleanupCache()
			}

		case "l":
			// List packages for current manager
			if m.cursor < len(m.managers) && m.managers[m.cursor].Installed {
				return m, m.listPackages()
			}

		case "o":
			// Show outdated packages
			if m.cursor < len(m.managers) && m.managers[m.cursor].Installed {
				return m, m.showOutdated()
			}

		case "u":
			// Update package manager (not packages)
			if m.cursor < len(m.managers) && m.managers[m.cursor].Installed {
				return m, m.updateManager()
			}

		case "r":
			// Refresh/rescan
			m.loading = true
			m.output = ""
			m.message = ""
			return m, m.detectManagers()
		}

	case detectCompleteMsg:
		m.managers = msg.managers
		m.loading = false
		installedCount := 0
		for _, mgr := range m.managers {
			if mgr.Installed {
				installedCount++
			}
		}
		m.message = fmt.Sprintf("Found %d package manager(s)", installedCount)

	case actionCompleteMsg:
		m.executing = false
		m.output = msg.output
		m.message = msg.message
		m.showingOutput = true

	case actionStartMsg:
		m.executing = true
		m.output = msg.message
		m.message = msg.message

	case packageListMsg:
		m.executing = false
		m.showingList = true
		m.packageList = msg.packages
		m.listScroll = 0
		m.searchFilter = ""
		m.message = fmt.Sprintf("Loaded %d packages from %s", len(msg.packages), msg.manager)
	}

	return m, nil
}

// View renders the module
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	if m.showingList {
		return m.renderPackageList()
	}

	if m.loading {
		return m.renderLoading()
	}

	return m.renderManagers()
}

func (m *Model) renderLoading() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D9FF"))

	var b strings.Builder
	b.WriteString(titleStyle.Render("📦 PACKAGE MANAGEMENT"))
	b.WriteString("\n\n")
	b.WriteString("⏳ Detecting package managers...\n")

	return b.String()
}

func (m *Model) renderManagers() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D9FF"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#DDD"))
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))
	msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#0FD976"))
	controlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))

	var b strings.Builder
	b.WriteString(titleStyle.Render("📦 PACKAGE MANAGEMENT"))
	b.WriteString("\n\n")

	if m.executing {
		b.WriteString("⏳ Executing command...\n\n")
		if m.output != "" {
			// Show output in a box
			outputBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444")).
				Padding(1).
				Width(m.width - 4).
				Render(m.output)
			b.WriteString(outputBox)
		}
		return b.String()
	}

	for i, mgr := range m.managers {
		status := "✗ Not Installed"
		statusStyle := grayStyle
		if mgr.Installed {
			status = fmt.Sprintf("✓ Installed (%s)", mgr.Version)
			statusStyle = normalStyle
		}

		cursor := "  "
		if i == m.cursor {
			cursor = "▶ "
		}

		line := fmt.Sprintf("%s%s: %s", cursor, mgr.Name, status)

		if i == m.cursor {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(statusStyle.Render(line))
		}
		b.WriteString("\n")

		// Show details for installed managers
		if mgr.Installed {
			details := fmt.Sprintf("    Packages: %d", mgr.PackageCount)
			if mgr.Outdated > 0 {
				details += fmt.Sprintf(" (%d outdated)", mgr.Outdated)
			}
			if mgr.CacheSize != "" {
				details += fmt.Sprintf(", Cache: %s", mgr.CacheSize)
			}

			if i == m.cursor {
				b.WriteString(selectedStyle.Render(details))
			} else {
				b.WriteString(grayStyle.Render(details))
			}
			b.WriteString("\n")

			// Show actions for selected manager
			if i == m.cursor {
				actions := "    [C]leanup cache  [L]ist packages  [O]utdated  [U]pdate"
				b.WriteString(selectedStyle.Render(actions))
				b.WriteString("\n")
			}
		}

		b.WriteString("\n")
	}

	// Controls
	b.WriteString(controlStyle.Render("↑/↓ Navigate • C/L/O/U Actions • R Refresh"))

	// Message
	if m.message != "" {
		b.WriteString("\n\n")
		b.WriteString(msgStyle.Render(m.message))
	}

	return b.String()
}

// Title returns the module title
func (m *Model) Title() string {
	return "Packages"
}

// HasOpenModal returns true if the module has an open modal/dialog
func (m *Model) HasOpenModal() bool {
	return m.showingList || m.showingOutput
}

func (m *Model) detectManagers() tea.Cmd {
	return func() tea.Msg {
		var managers []PackageManager

		// Check Homebrew
		brew := PackageManager{
			Name:   "Homebrew",
			Binary: "brew",
		}

		if checkBinary("brew", 2*time.Second) {
			brew.Installed = true
			brew.Version = getBrewVersion()
			brew.PackageCount = getBrewPackageCount()
			brew.Outdated = getBrewOutdatedCount()
			brew.CacheSize = getBrewCacheSize()
		}

		managers = append(managers, brew)

		// Check npm
		npm := PackageManager{
			Name:   "npm",
			Binary: "npm",
		}

		if checkBinary("npm", 2*time.Second) {
			npm.Installed = true
			npm.Version = getNpmVersion()
			npm.PackageCount = getNpmGlobalCount()
			npm.CacheSize = getNpmCacheSize()
		}

		managers = append(managers, npm)

		return detectCompleteMsg{managers: managers}
	}
}

func (m *Model) cleanupCache() tea.Cmd {
	mgr := m.managers[m.cursor]

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var cmd *exec.Cmd
		switch mgr.Binary {
		case "brew":
			cmd = exec.CommandContext(ctx, "brew", "cleanup", "-s", "--prune=all")
		case "npm":
			cmd = exec.CommandContext(ctx, "npm", "cache", "clean", "--force")
		default:
			return actionCompleteMsg{
				output:  "",
				message: fmt.Sprintf("✗ Unknown package manager: %s", mgr.Name),
			}
		}

		setCommandPath(cmd)
		output, err := cmd.CombinedOutput()

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return actionCompleteMsg{
					output:  "",
					message: fmt.Sprintf("✗ %s cleanup timed out after 60 seconds", mgr.Name),
				}
			}
			return actionCompleteMsg{
				output:  string(output),
				message: fmt.Sprintf("✗ %s cleanup failed: %v", mgr.Name, err),
			}
		}

		return actionCompleteMsg{
			output:  string(output),
			message: fmt.Sprintf("✓ %s cache cleaned successfully", mgr.Name),
		}
	}
}

func (m *Model) listPackages() tea.Cmd {
	mgr := m.managers[m.cursor]

	return tea.Batch(
		func() tea.Msg {
			return actionStartMsg{message: fmt.Sprintf("Listing %s packages...", mgr.Name)}
		},
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var cmd *exec.Cmd
			switch mgr.Binary {
			case "brew":
				cmd = exec.CommandContext(ctx, "brew", "list", "--versions")
			case "npm":
				cmd = exec.CommandContext(ctx, "npm", "list", "-g", "--depth=0")
			default:
				return actionCompleteMsg{
					output:  "",
					message: fmt.Sprintf("✗ Unknown package manager: %s", mgr.Name),
				}
			}

			setCommandPath(cmd)
			output, err := cmd.CombinedOutput()

			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					return actionCompleteMsg{
						output:  string(output),
						message: fmt.Sprintf("✗ Listing timed out after 10 seconds"),
					}
				}
				return actionCompleteMsg{
					output:  string(output),
					message: fmt.Sprintf("✗ Failed to list packages: %v", err),
				}
			}

			// Parse package list
			packages := parsePackageList(mgr.Binary, string(output))

			return packageListMsg{
				packages: packages,
				manager:  mgr.Name,
			}
		},
	)
}

func (m *Model) showOutdated() tea.Cmd {
	mgr := m.managers[m.cursor]

	return tea.Batch(
		func() tea.Msg {
			return actionStartMsg{message: fmt.Sprintf("Checking %s for outdated packages...", mgr.Name)}
		},
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			var cmd *exec.Cmd
			switch mgr.Binary {
			case "brew":
				cmd = exec.CommandContext(ctx, "brew", "outdated")
			case "npm":
				cmd = exec.CommandContext(ctx, "npm", "outdated", "-g")
			default:
				return actionCompleteMsg{
					output:  "",
					message: fmt.Sprintf("✗ Unknown package manager: %s", mgr.Name),
				}
			}

			setCommandPath(cmd)
			output, err := cmd.CombinedOutput()

			if err != nil && ctx.Err() != context.DeadlineExceeded {
				// Command error is okay - might mean no outdated packages
				if len(output) == 0 {
					return actionCompleteMsg{
						output:  "No outdated packages",
						message: "✓ All packages are up to date",
					}
				}
			}

			if ctx.Err() == context.DeadlineExceeded {
				return actionCompleteMsg{
					output:  string(output),
					message: "✗ Check timed out after 15 seconds",
				}
			}

			return actionCompleteMsg{
				output:  string(output),
				message: fmt.Sprintf("Found %d outdated package(s)", mgr.Outdated),
			}
		},
	)
}

func (m *Model) updateManager() tea.Cmd {
	mgr := m.managers[m.cursor]

	return tea.Batch(
		func() tea.Msg {
			return actionStartMsg{message: fmt.Sprintf("Updating %s...", mgr.Name)}
		},
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			var cmd *exec.Cmd
			switch mgr.Binary {
			case "brew":
				cmd = exec.CommandContext(ctx, "brew", "update")
			case "npm":
				cmd = exec.CommandContext(ctx, "npm", "install", "-g", "npm@latest")
			default:
				return actionCompleteMsg{
					output:  "",
					message: fmt.Sprintf("✗ Unknown package manager: %s", mgr.Name),
				}
			}

			setCommandPath(cmd)
			output, err := cmd.CombinedOutput()

			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					return actionCompleteMsg{
						output:  string(output),
						message: fmt.Sprintf("✗ Update timed out after 30 seconds"),
					}
				}
				return actionCompleteMsg{
					output:  string(output),
					message: fmt.Sprintf("✗ Update failed: %v", err),
				}
			}

			return actionCompleteMsg{
				output:  string(output),
				message: fmt.Sprintf("✓ %s updated successfully", mgr.Name),
			}
		},
	)
}

// Helper functions

// setCommandPath sets the user's full shell PATH on the command, including NVM support
func setCommandPath(cmd *exec.Cmd) {
	homeDir := os.Getenv("HOME")
	paths := []string{}

	// Check for NVM and add its paths
	nvmDir := filepath.Join(homeDir, ".nvm")
	if _, err := os.Stat(nvmDir); err == nil {
		// Try to find the default/current Node version
		defaultVersion := filepath.Join(nvmDir, "alias", "default")
		if versionBytes, err := os.ReadFile(defaultVersion); err == nil {
			version := strings.TrimSpace(string(versionBytes))
			nvmBin := filepath.Join(nvmDir, "versions", "node", version, "bin")
			if _, err := os.Stat(nvmBin); err == nil {
				paths = append(paths, nvmBin)
			}
		}

		// Also check current symlink
		currentBin := filepath.Join(nvmDir, "current", "bin")
		if _, err := os.Stat(currentBin); err == nil {
			paths = append(paths, currentBin)
		}
	}

	// Add common Homebrew paths
	paths = append(paths, "/opt/homebrew/bin", "/usr/local/bin")

	// Get user's shell PATH
	shellPath := exec.Command("sh", "-c", "echo $PATH")
	if output, err := shellPath.Output(); err == nil && len(output) > 0 {
		paths = append(paths, strings.TrimSpace(string(output)))
	}

	// Add system paths
	paths = append(paths, "/usr/bin", "/bin", "/usr/sbin", "/sbin")

	// Set environment with combined PATH
	cmd.Env = []string{
		"PATH=" + strings.Join(paths, ":"),
		"HOME=" + homeDir,
		"USER=" + os.Getenv("USER"),
		"NVM_DIR=" + nvmDir,
	}
}

func checkBinary(name string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "which", name)
	err := cmd.Run()
	return err == nil
}

func getBrewVersion() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "brew", "--version")
	setCommandPath(cmd)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		// Extract version from "Homebrew 4.1.9"
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	return "unknown"
}

func getBrewPackageCount() int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "brew", "list", "--formula")
	setCommandPath(cmd)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	return count
}

func getBrewOutdatedCount() int {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "brew", "outdated")
	setCommandPath(cmd)
	output, _ := cmd.Output()

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	return count
}

func getBrewCacheSize() string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	homeDir := os.Getenv("HOME")
	cachePath := filepath.Join(homeDir, "Library/Caches/Homebrew")

	cmd := exec.CommandContext(ctx, "du", "-sh", cachePath)
	setCommandPath(cmd)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	parts := strings.Fields(string(output))
	if len(parts) > 0 {
		return parts[0]
	}

	return "unknown"
}

func getNpmVersion() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "npm", "--version")
	setCommandPath(cmd)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(output))
}

func getNpmGlobalCount() int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "npm", "list", "-g", "--depth=0")
	setCommandPath(cmd)
	output, _ := cmd.CombinedOutput()

	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and header/footer lines
		if line == "" || strings.HasPrefix(line, "/") || strings.Contains(line, "node_modules") {
			continue
		}
		// Count lines that start with ├──, └── or contain @ version
		if strings.HasPrefix(line, "├──") || strings.HasPrefix(line, "└──") ||
		   (strings.Contains(line, "@") && !strings.HasPrefix(line, "npm")) {
			count++
		}
	}

	return count
}

func getNpmCacheSize() string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	homeDir := os.Getenv("HOME")
	cachePath := filepath.Join(homeDir, ".npm")

	cmd := exec.CommandContext(ctx, "du", "-sh", cachePath)
	setCommandPath(cmd)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	parts := strings.Fields(string(output))
	if len(parts) > 0 {
		return parts[0]
	}

	return "unknown"
}

// Messages
type detectCompleteMsg struct {
	managers []PackageManager
}

type actionCompleteMsg struct {
	output  string
	message string
}

type actionStartMsg struct {
	message string
}

type packageListMsg struct {
	packages []string
	manager  string
}

// Helper functions for package list modal

func (m *Model) getFilteredPackages() []string {
	if m.searchFilter == "" {
		return m.packageList
	}

	filtered := []string{}
	searchLower := strings.ToLower(m.searchFilter)
	for _, pkg := range m.packageList {
		if strings.Contains(strings.ToLower(pkg), searchLower) {
			filtered = append(filtered, pkg)
		}
	}
	return filtered
}

func parsePackageList(binary, output string) []string {
	packages := []string{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch binary {
		case "brew":
			// Brew format: "package_name version"
			if !strings.HasPrefix(line, "=") && !strings.Contains(line, "Formulae") {
				packages = append(packages, line)
			}

		case "npm":
			// npm format: various tree formats
			// Skip header lines and paths
			if strings.HasPrefix(line, "/") || strings.Contains(line, "node_modules") {
				continue
			}
			// Remove tree characters
			line = strings.TrimPrefix(line, "├── ")
			line = strings.TrimPrefix(line, "└── ")
			line = strings.TrimPrefix(line, "│   ")
			line = strings.TrimSpace(line)

			if line != "" && !strings.HasPrefix(line, "npm") {
				packages = append(packages, line)
			}
		}
	}

	return packages
}

func (m *Model) renderPackageList() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D9FF"))
	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#DDD"))
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))
	searchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#0FD976"))

	filtered := m.getFilteredPackages()

	var b strings.Builder
	b.WriteString(titleStyle.Render("📦 PACKAGE LIST"))
	b.WriteString("\n\n")

	// Search bar
	if m.searchFilter != "" {
		b.WriteString(searchStyle.Render(fmt.Sprintf("🔍 Filter: %s", m.searchFilter)))
	} else {
		b.WriteString(borderStyle.Render("Type to filter packages..."))
	}
	b.WriteString(fmt.Sprintf(" (%d/%d packages)\n\n", len(filtered), len(m.packageList)))

	// Calculate visible area
	maxVisible := m.height - 8 // Leave room for header, footer
	if maxVisible < 5 {
		maxVisible = 5
	}

	// Ensure scroll is within bounds
	if m.listScroll >= len(filtered) && len(filtered) > 0 {
		m.listScroll = len(filtered) - 1
	}
	if m.listScroll < 0 {
		m.listScroll = 0
	}

	// Calculate visible window
	start := m.listScroll
	end := start + maxVisible
	if end > len(filtered) {
		end = len(filtered)
	}
	if start > 0 && end-start < maxVisible {
		start = end - maxVisible
		if start < 0 {
			start = 0
		}
	}

	// Render visible packages
	if len(filtered) == 0 {
		b.WriteString(borderStyle.Render("No packages match your filter"))
	} else {
		for i := start; i < end; i++ {
			cursor := "  "
			if i == m.listScroll {
				cursor = "▶ "
			}

			line := cursor + filtered[i]
			if i == m.listScroll {
				b.WriteString(highlightStyle.Render(line))
			} else {
				b.WriteString(normalStyle.Render(line))
			}
			b.WriteString("\n")
		}

		// Scroll indicator
		if len(filtered) > maxVisible {
			percentage := float64(m.listScroll) / float64(len(filtered)-1) * 100
			b.WriteString("\n")
			b.WriteString(borderStyle.Render(fmt.Sprintf("Showing %d-%d of %d (%.0f%%)", start+1, end, len(filtered), percentage)))
		}
	}

	// Controls
	b.WriteString("\n\n")
	b.WriteString(borderStyle.Render("↑/↓ Scroll • Type to filter • Backspace Clear • Esc/Q Close"))

	return b.String()
}
