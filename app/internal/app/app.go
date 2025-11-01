package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caioricciuti/dev-cockpit/internal/config"
	"github.com/caioricciuti/dev-cockpit/internal/logger"
	"github.com/caioricciuti/dev-cockpit/internal/modules/cleanup"
	"github.com/caioricciuti/dev-cockpit/internal/modules/dashboard"
	"github.com/caioricciuti/dev-cockpit/internal/modules/docker"
	"github.com/caioricciuti/dev-cockpit/internal/modules/network"
	"github.com/caioricciuti/dev-cockpit/internal/modules/packages"
	"github.com/caioricciuti/dev-cockpit/internal/modules/quickactions"
	"github.com/caioricciuti/dev-cockpit/internal/modules/security"
	"github.com/caioricciuti/dev-cockpit/internal/modules/support"
	"github.com/caioricciuti/dev-cockpit/internal/modules/system"
	"github.com/caioricciuti/dev-cockpit/internal/ui/components"
	"github.com/caioricciuti/dev-cockpit/internal/ui/events"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Module represents a tab in the application
type Module interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (interface{}, tea.Cmd)
	View() string
	Title() string
	HasOpenModal() bool
}

// Model represents the main application state
type Model struct {
	config        *config.Config
	version       string
	modules       []Module
	activeModule  int
	width         int
	height        int
	showHelp      bool
	showLogs      bool
	moduleFocused bool
	lastUpdate    time.Time
	quitting      bool
	err           error
	logLines      []string
	logLoadErr    error
	maxLogLines   int
	logPath       string
}

// New creates a new application model
func New(cfg *config.Config, version string) *Model {
	m := &Model{
		config:      cfg,
		version:     version,
		lastUpdate:  time.Now(),
		maxLogLines: 200,
		logPath:     logger.GetLogPath(),
	}

	// Initialize modules
	m.initializeModules()

	return m
}

func (m *Model) initializeModules() {
	m.modules = []Module{
		dashboard.New(m.config),
		quickactions.New(m.config),
		cleanup.New(m.config),
		packages.New(m.config),
		system.New(m.config),
		docker.New(m.config),
		network.New(m.config),
		security.New(m.config),
		support.New(),
	}
}

// Init initializes the application
func (m *Model) Init() tea.Cmd {
	// Initialize the first module
	if len(m.modules) > 0 {
		return m.modules[0].Init()
	}
	return nil
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Forward RAW size to modules (they handle their own layout)
		// Don't pre-adjust sizes or we get double reduction!
		for _, module := range m.modules {
			_, cmd := module.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case tea.KeyMsg:
		key := msg.String()
		// Normalize to lowercase for case-insensitive commands
		keyLower := strings.ToLower(key)

		switch key {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}

		// Handle help/logs screens first
		if m.showHelp {
			switch keyLower {
			case "esc", "q":
				m.showHelp = false
			}
			return m, tea.Batch(cmds...)
		}

		if m.showLogs {
			if keyLower == "esc" || keyLower == "q" {
				m.showLogs = false
			}
			return m, tea.Batch(cmds...)
		}

		// If module is focused, it gets ALL keys
		if m.moduleFocused {
			// Pass key to the focused module first
			if m.activeModule < len(m.modules) {
				_, cmd := m.modules[m.activeModule].Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}

			// If ESC was pressed and module doesn't have open modals, unfocus
			if key == "esc" && m.activeModule < len(m.modules) {
				if !m.modules[m.activeModule].HasOpenModal() {
					m.moduleFocused = false
					if _, cmd := m.modules[m.activeModule].Update(events.Blur{}); cmd != nil {
						cmds = append(cmds, cmd)
					}
				}
			}

			return m, tea.Batch(cmds...)
		}

		// Global commands (only when NOT focused on a module)
		switch keyLower {
		case "q":
			m.quitting = true
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			if m.showHelp {
				m.showLogs = false
			}
			return m, tea.Batch(cmds...)
		case "l":
			if m.showLogs {
				m.showLogs = false
			} else {
				m.showLogs = true
				m.refreshLogs()
			}
			return m, tea.Batch(cmds...)
		}

		if len(m.modules) == 0 {
			return m, tea.Batch(cmds...)
		}

		switch key {
		case "tab", "right":
			m.activeModule = (m.activeModule + 1) % len(m.modules)
			if init := m.modules[m.activeModule].Init(); init != nil {
				cmds = append(cmds, init)
			}
		case "shift+tab", "left":
			m.activeModule = m.activeModule - 1
			if m.activeModule < 0 {
				m.activeModule = len(m.modules) - 1
			}
			if init := m.modules[m.activeModule].Init(); init != nil {
				cmds = append(cmds, init)
			}
		case "home":
			m.activeModule = 0
			if init := m.modules[m.activeModule].Init(); init != nil {
				cmds = append(cmds, init)
			}
		case "end":
			m.activeModule = len(m.modules) - 1
			if init := m.modules[m.activeModule].Init(); init != nil {
				cmds = append(cmds, init)
			}
		case "enter":
			m.moduleFocused = true
			if m.activeModule < len(m.modules) {
				if _, cmd := m.modules[m.activeModule].Update(events.Focus{}); cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

		return m, tea.Batch(cmds...)

	case tickMsg:
		m.lastUpdate = time.Now()
		// Update active module
		if m.activeModule < len(m.modules) {
			_, cmd := m.modules[m.activeModule].Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		if m.showLogs {
			m.refreshLogs()
		}
		cmds = append(cmds, doTick())

	default:
		// Pass other messages to the active module
		if m.activeModule < len(m.modules) {
			_, cmd := m.modules[m.activeModule].Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the application
func (m *Model) View() string {
	if m.quitting {
		return "Thanks for using Dev Cockpit!\n"
	}

	// Create layout manager to calculate available space
	layout := components.NewLayout(m.width, m.height)
	if !m.moduleFocused && !m.showLogs {
		layout = layout.WithHint()
	}

	// Handle overlays (they take full screen)
	if m.showHelp {
		return m.renderHelp()
	}

	if m.showLogs {
		return m.renderLogOverlay(layout)
	}

	// Render main UI
	tabs := m.renderTabs()
	footer := m.renderFooter()

	// Render module content with available space
	moduleContent := ""
	if m.activeModule < len(m.modules) {
		moduleContent = m.modules[m.activeModule].View()
	}

	// Add hint if not focused
	finalContent := moduleContent
	if !m.moduleFocused {
		hint := m.renderHint(layout.ContentWidth)
		finalContent = lipgloss.JoinVertical(lipgloss.Top, hint, "", moduleContent)
	}

	// Constrain content to prevent overflow
	constrainedContent := lipgloss.NewStyle().
		Width(layout.ContentWidth).
		MaxHeight(layout.ContentHeight).
		Padding(0, 2).
		Render(components.Viewport(finalContent, layout.ContentHeight))

	// Stack everything
	return lipgloss.JoinVertical(
		lipgloss.Top,
		tabs,
		constrainedContent,
		footer,
	)
}

func (m *Model) renderTabs() string {
	styles := components.NewBaseStyles()

	// Calculate fixed width for each tab to prevent jumping
	numTabs := len(m.modules)
	if numTabs == 0 {
		return ""
	}

	// Reserve space for borders and padding in tab bar
	availableWidth := m.width - 8 // margins and borders
	tabWidth := (availableWidth / numTabs) - 2 // spacing between tabs
	if tabWidth < 12 {
		tabWidth = 12 // minimum width
	}

	var tabs []string

	// All tabs have SAME dimensions - only colors change
	for i, module := range m.modules {
		label := module.Title()

		// Determine tab style based on state
		var style lipgloss.Style
		if i == m.activeModule {
			if m.moduleFocused {
				// Focused: white text, cyan background
				label = "◉ " + label
				style = lipgloss.NewStyle().
					Width(tabWidth).
					Bold(true).
					Foreground(styles.Theme.Foreground).
					Background(styles.Theme.Primary).
					Padding(0, 1).
					Align(lipgloss.Center)
			} else {
				// Active but not focused: cyan text, dark background
				label = "◎ " + label
				style = lipgloss.NewStyle().
					Width(tabWidth).
					Bold(true).
					Foreground(styles.Theme.Primary).
					Background(lipgloss.Color("#1A1A2E")).
					Padding(0, 1).
					Align(lipgloss.Center)
			}
		} else {
			// Inactive: gray text, dark background
			style = lipgloss.NewStyle().
				Width(tabWidth).
				Foreground(styles.Theme.Muted).
				Background(styles.Theme.Background).
				Padding(0, 1).
				Align(lipgloss.Center)
		}

		tabs = append(tabs, style.Render(components.TruncateString(label, tabWidth-2)))
	}

	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	return lipgloss.NewStyle().
		Width(m.width).
		Background(styles.Theme.Background).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(styles.Theme.Primary).
		Padding(1, 2).
		Render(tabRow)
}

func (m *Model) renderFooter() string {
	styles := components.NewBaseStyles()

	footerStyle := lipgloss.NewStyle().
		Width(m.width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(styles.Theme.Primary).
		Background(styles.Theme.Background).
		Foreground(styles.Theme.Muted).
		Padding(0, 2)

	versionStyle := lipgloss.NewStyle().
		Foreground(styles.Theme.Primary).
		Bold(true)

	shortcutsStyle := lipgloss.NewStyle().
		Foreground(styles.Theme.Foreground)

	statusStyle := lipgloss.NewStyle().
		Foreground(styles.Theme.Success).
		Bold(true)

	focusIndicator := ""
	if m.moduleFocused {
		focusIndicator = lipgloss.NewStyle().
			Foreground(styles.Theme.Primary).
			Bold(true).
			Render(" [FOCUSED]")
	}

	shortcuts := "Tab Switch • Enter Focus • Esc Back • ? Help • L Logs • Q Quit"
	info := versionStyle.Render(fmt.Sprintf("Dev Cockpit v%s", m.version)) + focusIndicator
	left := fmt.Sprintf("%s  │  %s", info, shortcutsStyle.Render(shortcuts))
	status := statusStyle.Render(fmt.Sprintf("⟳ %s", m.lastUpdate.Format("15:04:05")))

	// Calculate spacing dynamically
	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(status)
	spacer := m.width - leftLen - rightLen - 6
	if spacer < 0 {
		spacer = 0
	}

	return footerStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			left,
			lipgloss.NewStyle().Width(spacer).Render(""),
			status,
		),
	)
}

func (m *Model) refreshLogs() {
	path := m.logPath
	if path == "" {
		path = logger.GetLogPath()
		m.logPath = path
	}

	data, err := os.ReadFile(path)
	if err != nil {
		m.logLoadErr = err
		m.logLines = nil
		return
	}

	lines := strings.Split(string(data), "\n")
	var trimmed []string
	for _, line := range lines {
		line = strings.TrimRight(line, " \t")
		if strings.TrimSpace(line) == "" {
			continue
		}
		trimmed = append(trimmed, line)
	}

	if len(trimmed) > m.maxLogLines {
		trimmed = trimmed[len(trimmed)-m.maxLogLines:]
	}

	m.logLoadErr = nil
	m.logLines = trimmed
}

func (m *Model) renderLogOverlay(layout *components.Layout) string {
	boxWidth := layout.ContentWidth - 6
	if boxWidth > 120 {
		boxWidth = 120
	}
	if boxWidth < 60 {
		boxWidth = 60
	}

	styles := components.NewBaseStyles()

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Theme.Primary)

	infoStyle := lipgloss.NewStyle().
		Foreground(styles.Theme.Muted)

	contentStyle := lipgloss.NewStyle().
		Foreground(styles.Theme.Foreground)

	var builder strings.Builder
	builder.WriteString(headerStyle.Render("📋 Log Viewer"))
	builder.WriteString("\n")
	location := fmt.Sprintf("File: %s", m.logPath)
	builder.WriteString(infoStyle.Render(location))
	builder.WriteString("\n")
	builder.WriteString(infoStyle.Render("Press 'l' to close"))
	builder.WriteString("\n\n")

	if m.logLoadErr != nil {
		builder.WriteString(lipgloss.NewStyle().Foreground(styles.Theme.Error).Render(
			fmt.Sprintf("Unable to read log: %v", m.logLoadErr),
		))
		builder.WriteString("\n")
	} else if len(m.logLines) == 0 {
		builder.WriteString(infoStyle.Render("No log entries captured yet."))
		builder.WriteString("\n")
	} else {
		for _, line := range m.logLines {
			truncated := components.TruncateString(line, boxWidth-4)
			builder.WriteString(contentStyle.Render(truncated))
			builder.WriteString("\n")
		}
	}

	maxHeight := layout.ContentHeight - 6
	if maxHeight < 12 {
		maxHeight = 12
	}

	box := lipgloss.NewStyle().
		Width(boxWidth).
		MaxHeight(maxHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Theme.Border).
		Padding(1, 2).
		Render(components.Viewport(builder.String(), maxHeight-4))

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

func (m *Model) renderHint(width int) string {
	styles := components.NewBaseStyles()

	hintStyle := lipgloss.NewStyle().
		Foreground(styles.Theme.Warning).
		Background(lipgloss.Color("#1A1A2E")).
		Bold(true).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Theme.Warning)

	hint := hintStyle.Render("⚠️  Press ENTER to enable commands in this module  ⚠️")

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(hint)
}

func (m *Model) renderHelp() string {
	boxStyle := lipgloss.NewStyle().
		Width(m.width - 10).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D9FF")).
		Background(lipgloss.Color("#0F1419")).
		Padding(2, 4)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF")).
		Background(lipgloss.Color("#1A1A2E")).
		Padding(0, 2).
		MarginBottom(1)

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500")).
		Bold(true).
		MarginTop(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDD"))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("⌘ DEV COCKPIT HELP"),
		"",
		sectionStyle.Render("NAVIGATION (GLOBAL):"),
		fmt.Sprintf("  %s  Switch modules", keyStyle.Render("Tab / Shift+Tab")),
		fmt.Sprintf("  %s          Focus current module", keyStyle.Render("Enter")),
		fmt.Sprintf("  %s            Leave focused module", keyStyle.Render("Esc")),
		"",
		sectionStyle.Render("COMMANDS:"),
		fmt.Sprintf("  %s            Close current dialog", keyStyle.Render("q")),
		fmt.Sprintf("  %s            Quit application", keyStyle.Render("Q")),
		fmt.Sprintf("  %s            Toggle this help", keyStyle.Render("?")),
		fmt.Sprintf("  %s            Toggle logs overlay", keyStyle.Render("l")),
		fmt.Sprintf("  %s            Refresh current view", keyStyle.Render("r")),
		"",
		sectionStyle.Render("INSIDE MODULES:"),
		descStyle.Render("  Follow the on-screen hints for module-specific controls"),
		"",
		sectionStyle.Render("SUPPORT:"),
		descStyle.Render("  Navigate to the Support tab for contribution links"),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render("Press 'q' or 'Esc' to close help"),
	)

	centered := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		boxStyle.Render(content),
	)

	return centered
}

// tickMsg is sent every second to update the display
type tickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
