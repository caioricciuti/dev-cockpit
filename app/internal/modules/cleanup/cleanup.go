package cleanup

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

// CleanupTarget represents a cleanable target
type CleanupTarget struct {
	Name        string
	Path        string
	Description string
	Size        uint64
	Selected    bool
	Timeout     time.Duration
}

// Model represents the cleanup module state
type Model struct {
	config         *config.Config
	width          int
	height         int
	targets        []CleanupTarget
	cursor         int
	scanning       bool
	cleaning       bool
	results        []CleanupResult
	showingResults bool
	message        string
}

// CleanupResult represents the result of a cleanup operation
type CleanupResult struct {
	Target   string
	Success  bool
	Freed    uint64
	Error    error
	Duration time.Duration
}

// New creates a new cleanup module
func New(cfg *config.Config) *Model {
	homeDir, _ := os.UserHomeDir()

	targets := []CleanupTarget{
		{
			Name:        "User Caches",
			Path:        filepath.Join(homeDir, "Library/Caches"),
			Description: "Application cache files (safe to remove)",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "Trash",
			Path:        filepath.Join(homeDir, ".Trash"),
			Description: "Items in Trash",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "Homebrew Cache",
			Path:        filepath.Join(homeDir, "Library/Caches/Homebrew"),
			Description: "Downloaded Homebrew installers",
			Timeout:     15 * time.Second,
		},
		{
			Name:        "npm Cache",
			Path:        filepath.Join(homeDir, ".npm"),
			Description: "npm package cache",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "Yarn Cache",
			Path:        filepath.Join(homeDir, "Library/Caches/Yarn"),
			Description: "Yarn package cache",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "Go Build Cache",
			Path:        filepath.Join(homeDir, "Library/Caches/go-build"),
			Description: "Go compilation cache",
			Timeout:     5 * time.Second,
		},
		{
			Name:        "Xcode Derived Data",
			Path:        filepath.Join(homeDir, "Library/Developer/Xcode/DerivedData"),
			Description: "Xcode build artifacts (can be large)",
			Timeout:     30 * time.Second,
		},
	}

	return &Model{
		config:   cfg,
		targets:  targets,
		scanning: true,
	}
}

// Init initializes the module
func (m *Model) Init() tea.Cmd {
	return m.scanSizes()
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (interface{}, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		// Handle results screen - any key dismisses
		if m.showingResults {
			m.showingResults = false
			m.results = []CleanupResult{}
			return m, nil
		}

		if m.scanning || m.cleaning {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.targets)-1 {
				m.cursor++
			}

		case " ":
			// Toggle selection
			m.targets[m.cursor].Selected = !m.targets[m.cursor].Selected

		case "a":
			// Select all
			for i := range m.targets {
				m.targets[i].Selected = true
			}
			m.message = "All items selected"

		case "n":
			// Select none
			for i := range m.targets {
				m.targets[i].Selected = false
			}
			m.message = "Selection cleared"

		case "enter":
			// Start cleanup
			hasSelected := false
			for _, target := range m.targets {
				if target.Selected {
					hasSelected = true
					break
				}
			}
			if hasSelected {
				return m, m.performCleanup()
			} else {
				m.message = "⚠ Select at least one item to clean"
			}

		case "r":
			// Rescan sizes
			m.scanning = true
			m.message = "Rescanning..."
			return m, m.scanSizes()
		}

	case scanCompleteMsg:
		m.targets = msg.targets
		m.scanning = false
		m.message = fmt.Sprintf("Found %.2f GB available to clean", float64(m.getTotalSize())/1024/1024/1024)

	case cleanupCompleteMsg:
		m.cleaning = false
		m.results = msg.results
		m.showingResults = true

		// Calculate total freed
		var totalFreed uint64
		successCount := 0
		for _, r := range m.results {
			if r.Success {
				totalFreed += r.Freed
				successCount++
			}
		}

		m.message = fmt.Sprintf("✓ Cleaned %d items, freed %.2f GB", successCount, float64(totalFreed)/1024/1024/1024)

		// Rescan to update sizes
		return m, m.scanSizes()
	}

	return m, nil
}

// View renders the module
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	if m.scanning {
		return m.renderScanning()
	}

	if m.cleaning {
		return m.renderCleaning()
	}

	if len(m.results) > 0 {
		return m.renderResults()
	}

	return m.renderSelection()
}

func (m *Model) renderScanning() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D9FF"))

	var b strings.Builder
	b.WriteString(titleStyle.Render("🧹 CLEANUP"))
	b.WriteString("\n\n")
	b.WriteString("⏳ Scanning directories and calculating sizes...\n")
	b.WriteString("\nThis may take a few seconds.\n")

	return b.String()
}

func (m *Model) renderSelection() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D9FF"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#DDD"))
	msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#0FD976"))
	controlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))

	var b strings.Builder
	b.WriteString(titleStyle.Render("🧹 CLEANUP"))
	b.WriteString("\n\n")
	b.WriteString("Select items to clean:\n\n")

	// Render targets
	for i, target := range m.targets {
		cursor := "  "
		if i == m.cursor {
			cursor = "▶ "
		}

		checkbox := "[ ]"
		if target.Selected {
			checkbox = "[✓]"
		}

		line := fmt.Sprintf("%s%s %-20s %10s", cursor, checkbox, target.Name, formatBytes(target.Size))

		if i == m.cursor {
			b.WriteString(selectedStyle.Render(line))
			b.WriteString("\n")
			b.WriteString(selectedStyle.Render(fmt.Sprintf("    %s", target.Description)))
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	// Summary
	b.WriteString("\n")
	totalSelected := uint64(0)
	for _, target := range m.targets {
		if target.Selected {
			totalSelected += target.Size
		}
	}
	b.WriteString(fmt.Sprintf("Total to clean: %s\n", formatBytes(totalSelected)))

	// Controls
	b.WriteString("\n")
	b.WriteString(controlStyle.Render("↑/↓ Navigate • Space Toggle • A All • N None • Enter Clean • R Rescan"))

	// Message
	if m.message != "" {
		b.WriteString("\n\n")
		b.WriteString(msgStyle.Render(m.message))
	}

	return b.String()
}

func (m *Model) renderCleaning() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D9FF"))

	var b strings.Builder
	b.WriteString(titleStyle.Render("🧹 CLEANUP IN PROGRESS"))
	b.WriteString("\n\n")
	b.WriteString("⏳ Cleaning selected items...\n")
	b.WriteString("\nPlease wait, this may take a minute.\n")

	return b.String()
}

func (m *Model) renderResults() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D9FF"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#0FD976"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	controlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))

	var b strings.Builder
	b.WriteString(titleStyle.Render("🧹 CLEANUP COMPLETE"))
	b.WriteString("\n\n")

	var totalFreed uint64
	successCount := 0
	failedCount := 0

	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))

	for _, result := range m.results {
		if result.Success {
			successCount++
			totalFreed += result.Freed
			if result.Freed == 0 {
				b.WriteString(warningStyle.Render(fmt.Sprintf("⚠ %s: Nothing to clean (%v)\n",
					result.Target, result.Duration.Round(time.Millisecond))))
			} else {
				b.WriteString(successStyle.Render(fmt.Sprintf("✓ %s: %s freed (%v)\n",
					result.Target, formatBytes(result.Freed), result.Duration.Round(time.Millisecond))))
			}
		} else {
			failedCount++
			b.WriteString(errorStyle.Render(fmt.Sprintf("✗ %s: %v\n", result.Target, result.Error)))
		}
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Summary: %d succeeded, %d failed\n", successCount, failedCount))
	b.WriteString(fmt.Sprintf("Total freed: %s\n", formatBytes(totalFreed)))

	b.WriteString("\n")
	b.WriteString(controlStyle.Render("Press any key to continue"))

	return b.String()
}

// Title returns the module title
func (m *Model) Title() string {
	return "Cleanup"
}

// HasOpenModal returns true if the module has an open modal/dialog
func (m *Model) HasOpenModal() bool {
	return m.showingResults
}

func (m *Model) getTotalSize() uint64 {
	var total uint64
	for _, target := range m.targets {
		total += target.Size
	}
	return total
}

func (m *Model) scanSizes() tea.Cmd {
	// Snapshot targets to avoid mutating shared state from a goroutine
	targets := make([]CleanupTarget, len(m.targets))
	copy(targets, m.targets)

	return func() tea.Msg {
		for i := range targets {
			// Check if path exists
			if _, err := os.Stat(targets[i].Path); os.IsNotExist(err) {
				targets[i].Size = 0
				continue
			}

			// Get size with timeout
			targets[i].Size = getSizeWithTimeout(targets[i].Path, targets[i].Timeout)
		}

		return scanCompleteMsg{targets: targets}
	}
}

func (m *Model) performCleanup() tea.Cmd {
	m.cleaning = true
	m.results = []CleanupResult{}

	return func() tea.Msg {
		var results []CleanupResult

		for _, target := range m.targets {
			if !target.Selected {
				continue
			}

			start := time.Now()

			// Check if path exists
			if _, err := os.Stat(target.Path); os.IsNotExist(err) {
				results = append(results, CleanupResult{
					Target:   target.Name,
					Success:  true,
					Freed:    0,
					Duration: time.Since(start),
				})
				continue
			}

			// Get size before cleanup
			sizeBefore := getSizeWithTimeout(target.Path, target.Timeout)

			// Perform cleanup
			err := cleanTarget(target.Path, target.Timeout)

			// Get size after cleanup
			sizeAfter := uint64(0)
			if err == nil {
				sizeAfter = getSizeWithTimeout(target.Path, target.Timeout)
			}

			freed := sizeBefore - sizeAfter

			results = append(results, CleanupResult{
				Target:   target.Name,
				Success:  err == nil,
				Freed:    freed,
				Error:    err,
				Duration: time.Since(start),
			})
		}

		return cleanupCompleteMsg{results: results}
	}
}

// getSizeWithTimeout calculates directory size with a timeout
func getSizeWithTimeout(path string, timeout time.Duration) uint64 {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "du", "-sk", path)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	var kb uint64
	fmt.Sscanf(string(output), "%d", &kb)
	return kb * 1024
}

// cleanTarget removes all contents of a directory
func cleanTarget(path string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Read directory entries and remove each one individually to avoid shell injection
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		cmd := exec.CommandContext(ctx, "rm", "-rf", entryPath)
		_ = cmd.Run() // Best-effort removal of each entry

		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("cleanup timed out after 60 seconds")
		}
	}

	return nil
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Messages
type scanCompleteMsg struct {
	targets []CleanupTarget
}

type cleanupCompleteMsg struct {
	results []CleanupResult
}
