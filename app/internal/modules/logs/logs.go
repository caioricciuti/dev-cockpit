package logs

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/caioricciuti/dev-cockpit/internal/config"
	"github.com/caioricciuti/dev-cockpit/internal/logger"
	"github.com/caioricciuti/dev-cockpit/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewMode for log source tabs
type ViewMode int

const (
	ViewSystem ViewMode = iota
	ViewHomebrew
	ViewDocker
	ViewApp
)

// LogLevel represents parsed log severity
type LogLevel int

const (
	LevelNone LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
)

// LogLine is a parsed log entry
type LogLine struct {
	Content string
	Level   LogLevel
}

// Model represents the log aggregator state
type Model struct {
	config *config.Config
	width  int
	height int

	activeView ViewMode
	views      []string

	// Per-view log lines
	systemLines []LogLine
	brewLines   []LogLine
	dockerLines []LogLine
	appLines    []LogLine

	// Docker state
	containers        []string
	selectedContainer int

	// UI state
	followMode   bool
	searchActive bool
	searchTerm   string
	scrollOffset int
	maxLines     int
	loading      bool
	output       string
}

// New creates a new log aggregator module
func New(cfg *config.Config) *Model {
	maxLines := 500
	if cfg.Modules.Logs.MaxLines > 0 {
		maxLines = cfg.Modules.Logs.MaxLines
	}

	return &Model{
		config:   cfg,
		views:    []string{"System", "Homebrew", "Docker", "App"},
		maxLines: maxLines,
	}
}

// Init initializes the module
func (m *Model) Init() tea.Cmd {
	return m.fetchCurrentView()
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (interface{}, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		// Handle search mode
		if m.searchActive {
			switch msg.String() {
			case "esc":
				m.searchActive = false
				m.searchTerm = ""
			case "enter":
				m.searchActive = false
			case "backspace":
				if len(m.searchTerm) > 0 {
					m.searchTerm = m.searchTerm[:len(m.searchTerm)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.searchTerm += msg.String()
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "1":
			m.activeView = ViewSystem
			m.scrollOffset = 0
			return m, m.fetchCurrentView()
		case "2":
			m.activeView = ViewHomebrew
			m.scrollOffset = 0
			return m, m.fetchCurrentView()
		case "3":
			m.activeView = ViewDocker
			m.scrollOffset = 0
			return m, m.fetchCurrentView()
		case "4":
			m.activeView = ViewApp
			m.scrollOffset = 0
			return m, m.fetchCurrentView()
		case "tab":
			m.activeView = (m.activeView + 1) % 4
			m.scrollOffset = 0
			return m, m.fetchCurrentView()
		case "f":
			m.followMode = !m.followMode
			if m.followMode {
				m.scrollOffset = 0
				return m, m.followTick()
			}
		case "/":
			m.searchActive = true
			m.searchTerm = ""
		case "r":
			return m, m.fetchCurrentView()
		case "c":
			// Cycle docker container
			if m.activeView == ViewDocker && len(m.containers) > 0 {
				m.selectedContainer = (m.selectedContainer + 1) % len(m.containers)
				return m, m.fetchCurrentView()
			}
		case "up", "k":
			if m.scrollOffset < len(m.currentLines())-1 {
				m.scrollOffset++
				m.followMode = false
			}
		case "down", "j":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case "g":
			lines := m.currentLines()
			m.scrollOffset = len(lines) - 1
			m.followMode = false
		case "G":
			m.scrollOffset = 0
			m.followMode = true
		}

	case logFetchMsg:
		m.loading = false
		m.output = msg.note
		switch msg.view {
		case ViewSystem:
			m.systemLines = msg.lines
		case ViewHomebrew:
			m.brewLines = msg.lines
		case ViewDocker:
			m.dockerLines = msg.lines
			if msg.containers != nil {
				m.containers = msg.containers
			}
		case ViewApp:
			m.appLines = msg.lines
		}
		if m.followMode {
			m.scrollOffset = 0
		}

	case followTickMsg:
		if m.followMode {
			return m, tea.Batch(m.fetchCurrentView(), m.followTick())
		}
	}

	return m, nil
}

// View renders the log aggregator
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	styles := components.NewBaseStyles()
	layout := components.NewLayout(m.width, m.height)

	title := styles.Title().Render("📋 LOG AGGREGATOR")

	// View tabs
	viewTabs := m.renderViewTabs(styles)

	// Status bar
	statusParts := []string{}
	if m.followMode {
		statusParts = append(statusParts,
			lipgloss.NewStyle().Foreground(styles.Theme.Success).Bold(true).Render("● FOLLOW"))
	}
	if m.loading {
		statusParts = append(statusParts,
			lipgloss.NewStyle().Foreground(styles.Theme.Primary).Render("Loading..."))
	} else if m.output != "" {
		statusParts = append(statusParts,
			styles.Muted().Render(m.output))
	}
	statusLine := strings.Join(statusParts, "  ")

	// Search bar
	searchLine := ""
	if m.searchActive {
		searchLine = lipgloss.NewStyle().Foreground(styles.Theme.Primary).Bold(true).
			Render(fmt.Sprintf("Search: %s_", m.searchTerm))
	} else if m.searchTerm != "" {
		searchLine = styles.Muted().
			Render(fmt.Sprintf("Filter: \"%s\" (/ to edit, esc clears)", m.searchTerm))
	}

	// Help
	help := "[j/k] Scroll  [f] Follow  [/] Search  [1-4] Switch  [r] Refresh"
	if m.activeView == ViewDocker && len(m.containers) > 0 {
		help += "  [c] Cycle container"
	}
	helpLine := styles.Muted().Render(help)

	// Docker container indicator
	containerLine := ""
	if m.activeView == ViewDocker && len(m.containers) > 0 {
		containerLine = lipgloss.NewStyle().Foreground(styles.Theme.Secondary).
			Render(fmt.Sprintf("Container: %s (%d/%d)",
				m.containers[m.selectedContainer],
				m.selectedContainer+1, len(m.containers)))
	}

	// Log content
	lines := m.currentLines()

	// Apply search filter
	if m.searchTerm != "" {
		var filtered []LogLine
		term := strings.ToLower(m.searchTerm)
		for _, l := range lines {
			if strings.Contains(strings.ToLower(l.Content), term) {
				filtered = append(filtered, l)
			}
		}
		lines = filtered
	}

	// Calculate visible window
	maxVisible := layout.ContentHeight - 12
	if maxVisible < 5 {
		maxVisible = 5
	}

	// Scroll from bottom (follow mode = offset 0 = bottom)
	endIdx := len(lines) - m.scrollOffset
	if endIdx < 0 {
		endIdx = 0
	}
	startIdx := endIdx - maxVisible
	if startIdx < 0 {
		startIdx = 0
	}

	var logOutput []string
	for i := startIdx; i < endIdx && i < len(lines); i++ {
		l := lines[i]
		styled := m.styleLine(l, styles)
		logOutput = append(logOutput, styled)
	}

	if len(logOutput) == 0 {
		logOutput = []string{styles.Muted().Render("No log entries found.")}
	}

	// Assemble
	sections := []string{title, "", viewTabs, ""}
	if statusLine != "" {
		sections = append(sections, statusLine)
	}
	if containerLine != "" {
		sections = append(sections, containerLine)
	}
	if searchLine != "" {
		sections = append(sections, searchLine)
	}
	sections = append(sections, helpLine, "")
	sections = append(sections, logOutput...)
	sections = append(sections, "",
		styles.Muted().Render(fmt.Sprintf("%d lines total", len(lines))))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return components.Viewport(content, layout.ContentHeight)
}

// Title returns the module title
func (m *Model) Title() string { return "Logs" }

// HasOpenModal returns whether search is active
func (m *Model) HasOpenModal() bool { return m.searchActive }

// -- Messages --

type logFetchMsg struct {
	view       ViewMode
	lines      []LogLine
	containers []string
	note       string
}

type followTickMsg struct{}

// -- Commands --

func (m *Model) fetchCurrentView() tea.Cmd {
	m.loading = true
	switch m.activeView {
	case ViewSystem:
		return m.fetchSystemLogs()
	case ViewHomebrew:
		return m.fetchBrewLogs()
	case ViewDocker:
		return m.fetchDockerLogs()
	case ViewApp:
		return m.fetchAppLogs()
	}
	return nil
}

func (m *Model) followTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
		return followTickMsg{}
	})
}

// fetchSystemLogs and fetchBrewLogs are in platform-specific files:
// platform_darwin.go — macOS log sources
// platform_linux.go  — Linux log sources (journalctl, /var/log)

func (m *Model) fetchDockerLogs() tea.Cmd {
	selectedIdx := m.selectedContainer
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// List running containers
		listCmd := exec.CommandContext(ctx, "docker", "ps", "--format", "{{.Names}}")
		out, err := listCmd.Output()
		if err != nil {
			return logFetchMsg{
				view: ViewDocker,
				note: "Docker not available or not running",
				lines: []LogLine{{
					Content: "Docker is not running or not installed. Supports Docker Desktop and OrbStack.",
					Level:   LevelWarn,
				}},
			}
		}

		names := strings.Split(strings.TrimSpace(string(out)), "\n")
		var containers []string
		for _, n := range names {
			n = strings.TrimSpace(n)
			if n != "" {
				containers = append(containers, n)
			}
		}

		if len(containers) == 0 {
			return logFetchMsg{
				view:       ViewDocker,
				containers: containers,
				note:       "No running containers",
				lines: []LogLine{{
					Content: "No running Docker containers found.",
					Level:   LevelInfo,
				}},
			}
		}

		// Get logs for selected container
		idx := selectedIdx
		if idx >= len(containers) {
			idx = 0
		}

		logCmd := exec.CommandContext(ctx, "docker", "logs", "--tail", "100", containers[idx])
		logOut, _ := logCmd.CombinedOutput()

		lines := parseLogOutput(string(logOut), 200)

		return logFetchMsg{
			view:       ViewDocker,
			containers: containers,
			lines:      lines,
			note:       fmt.Sprintf("Loaded logs for %s", containers[idx]),
		}
	}
}

func (m *Model) fetchAppLogs() tea.Cmd {
	return func() tea.Msg {
		path := logger.GetLogPath()
		lines := tailFile(path, 200)

		var logLines []LogLine
		for _, l := range lines {
			logLines = append(logLines, LogLine{
				Content: l,
				Level:   detectLevel(l),
			})
		}

		if len(logLines) == 0 {
			logLines = append(logLines, LogLine{
				Content: fmt.Sprintf("No log entries in %s", path),
				Level:   LevelInfo,
			})
		}

		return logFetchMsg{
			view:  ViewApp,
			lines: logLines,
			note:  fmt.Sprintf("Loaded %d app log entries", len(logLines)),
		}
	}
}

// -- Rendering helpers --

func (m *Model) renderViewTabs(styles *components.BaseStyles) string {
	labels := []struct {
		key  string
		name string
		view ViewMode
	}{
		{"1", "System", ViewSystem},
		{"2", "Homebrew", ViewHomebrew},
		{"3", "Docker", ViewDocker},
		{"4", "App", ViewApp},
	}

	var tabs []string
	for _, l := range labels {
		label := fmt.Sprintf("[%s] %s", l.key, l.name)
		if l.view == m.activeView {
			tabs = append(tabs, lipgloss.NewStyle().
				Foreground(styles.Theme.Primary).Bold(true).
				Render(label))
		} else {
			tabs = append(tabs, lipgloss.NewStyle().
				Foreground(styles.Theme.Muted).
				Render(label))
		}
	}

	return strings.Join(tabs, "  ")
}

func (m *Model) styleLine(l LogLine, styles *components.BaseStyles) string {
	maxLen := m.width - 8
	if maxLen < 40 {
		maxLen = 40
	}
	content := components.TruncateString(l.Content, maxLen)

	switch l.Level {
	case LevelError:
		return lipgloss.NewStyle().Foreground(styles.Theme.Error).Render(content)
	case LevelWarn:
		return lipgloss.NewStyle().Foreground(styles.Theme.Warning).Render(content)
	case LevelInfo:
		return lipgloss.NewStyle().Foreground(styles.Theme.Primary).Render(content)
	case LevelDebug:
		return lipgloss.NewStyle().Foreground(styles.Theme.Muted).Render(content)
	default:
		return lipgloss.NewStyle().Foreground(styles.Theme.Foreground).Render(content)
	}
}

func (m *Model) currentLines() []LogLine {
	switch m.activeView {
	case ViewSystem:
		return m.systemLines
	case ViewHomebrew:
		return m.brewLines
	case ViewDocker:
		return m.dockerLines
	case ViewApp:
		return m.appLines
	}
	return nil
}

// -- Parsing helpers --

func parseLogOutput(output string, maxLines int) []LogLine {
	rawLines := strings.Split(strings.TrimSpace(output), "\n")

	// Keep last maxLines
	if len(rawLines) > maxLines {
		rawLines = rawLines[len(rawLines)-maxLines:]
	}

	var lines []LogLine
	for _, l := range rawLines {
		l = strings.TrimRight(l, " \t\r")
		if l == "" {
			continue
		}
		lines = append(lines, LogLine{
			Content: l,
			Level:   detectLevel(l),
		})
	}
	return lines
}

func tailFile(path string, maxLines int) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var allLines []string
	scanner := bufio.NewScanner(f)
	// Increase buffer size for long lines
	scanner.Buffer(make([]byte, 64*1024), 256*1024)

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t\r")
		if line != "" {
			allLines = append(allLines, line)
		}
	}

	if len(allLines) > maxLines {
		allLines = allLines[len(allLines)-maxLines:]
	}

	return allLines
}

func detectLevel(line string) LogLevel {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "error") || strings.Contains(lower, "fatal") || strings.Contains(lower, "panic"):
		return LevelError
	case strings.Contains(lower, "warn") || strings.Contains(lower, "warning"):
		return LevelWarn
	case strings.Contains(lower, "info"):
		return LevelInfo
	case strings.Contains(lower, "debug") || strings.Contains(lower, "trace"):
		return LevelDebug
	default:
		return LevelNone
	}
}
