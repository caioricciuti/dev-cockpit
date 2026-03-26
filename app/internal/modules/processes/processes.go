package processes

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/caioricciuti/dev-cockpit/internal/config"
	"github.com/caioricciuti/dev-cockpit/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SortField defines how processes are sorted
type SortField int

const (
	SortByCPU  SortField = iota
	SortByMem
	SortByPID
	SortByName
)

// Process represents a running system process
type Process struct {
	PID  int
	CPU  float64
	Mem  float64
	RSS  int64 // in KB
	User string
	Name string
}

// Model represents the process manager state
type Model struct {
	config *config.Config
	width  int
	height int

	processes   []Process
	cursor      int
	sortField   SortField
	sortReverse bool

	// Kill confirmation
	confirmKill bool
	killTarget  *Process

	// Status
	output     string
	loading    bool
	lastUpdate time.Time

	// Search
	searching  bool
	searchTerm string
	filtered   []Process
}

// New creates a new process manager module
func New(cfg *config.Config) *Model {
	return &Model{
		config:      cfg,
		sortField:   SortByCPU,
		sortReverse: true,
	}
}

// Init initializes the module
func (m *Model) Init() tea.Cmd {
	return m.refresh()
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

		// Handle kill confirmation
		if m.confirmKill {
			switch msg.String() {
			case "y", "Y":
				target := m.killTarget
				m.confirmKill = false
				m.killTarget = nil
				if target != nil {
					return m, m.killProcess(target.PID)
				}
			default:
				m.confirmKill = false
				m.killTarget = nil
			}
			return m, nil
		}

		// Handle search mode
		if m.searching {
			switch msg.String() {
			case "esc":
				m.searching = false
				m.searchTerm = ""
				m.filtered = nil
				m.cursor = 0
			case "enter":
				m.searching = false
			case "backspace":
				if len(m.searchTerm) > 0 {
					m.searchTerm = m.searchTerm[:len(m.searchTerm)-1]
					m.applyFilter()
				}
			default:
				if len(msg.String()) == 1 {
					m.searchTerm += msg.String()
					m.applyFilter()
					m.cursor = 0
				}
			}
			return m, nil
		}

		// Normal mode
		procs := m.visibleProcesses()
		switch msg.String() {
		case "r":
			return m, m.refresh()
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(procs)-1 {
				m.cursor++
			}
		case "s":
			m.sortField = (m.sortField + 1) % 4
			m.sortProcesses()
			m.cursor = 0
		case "S":
			m.sortReverse = !m.sortReverse
			m.sortProcesses()
		case "x":
			if m.cursor < len(procs) {
				p := procs[m.cursor]
				m.confirmKill = true
				m.killTarget = &p
			}
		case "/":
			m.searching = true
			m.searchTerm = ""
		case "g":
			m.cursor = 0
		case "G":
			if len(procs) > 0 {
				m.cursor = len(procs) - 1
			}
		}

	case processesMsg:
		m.processes = msg.items
		m.output = msg.note
		m.loading = false
		m.lastUpdate = time.Now()
		m.sortProcesses()
		m.applyFilter()
		if m.cursor >= len(m.visibleProcesses()) {
			m.cursor = 0
		}

	case killMsg:
		m.output = msg.note
		m.loading = false
		return m, m.refresh()
	}

	return m, nil
}

// View renders the process manager
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	styles := components.NewBaseStyles()
	layout := components.NewLayout(m.width, m.height)

	title := styles.Title().Render("⚡ PROCESS MANAGER")

	// Sort indicator
	sortNames := []string{"CPU%", "MEM%", "PID", "NAME"}
	sortDir := "▼"
	if !m.sortReverse {
		sortDir = "▲"
	}
	sortInfo := lipgloss.NewStyle().
		Foreground(styles.Theme.Secondary).Bold(true).
		Render(fmt.Sprintf("Sort: %s %s", sortNames[m.sortField], sortDir))

	help := styles.Muted().Render("[j/k] Navigate  [s] Sort  [S] Reverse  [x] Kill  [/] Search  [r] Refresh")

	// Status line
	statusLine := ""
	if m.loading {
		statusLine = lipgloss.NewStyle().Foreground(styles.Theme.Primary).Render("Loading...")
	} else if m.output != "" {
		statusLine = lipgloss.NewStyle().Foreground(styles.Theme.Success).Render(m.output)
	}

	// Kill confirmation
	if m.confirmKill && m.killTarget != nil {
		statusLine = lipgloss.NewStyle().
			Foreground(styles.Theme.Error).Bold(true).
			Render(fmt.Sprintf("Kill process %d (%s)? [y] confirm / [any] cancel",
				m.killTarget.PID, m.killTarget.Name))
	}

	// Search bar
	searchLine := ""
	if m.searching {
		searchLine = lipgloss.NewStyle().Foreground(styles.Theme.Primary).Bold(true).
			Render(fmt.Sprintf("Search: %s_", m.searchTerm))
	} else if m.searchTerm != "" {
		searchLine = styles.Muted().
			Render(fmt.Sprintf("Filter: \"%s\" (/ to search, esc clears)", m.searchTerm))
	}

	// Column header
	headerStyle := lipgloss.NewStyle().Foreground(styles.Theme.Primary).Bold(true)
	header := headerStyle.Render(
		fmt.Sprintf("  %-7s %-7s %-7s %-10s %-12s %s", "PID", "CPU%", "MEM%", "RSS", "USER", "COMMAND"))

	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("#333")).
		Render(strings.Repeat("─", min(80, layout.ContentWidth-4)))

	// Process list with scrolling
	procs := m.visibleProcesses()
	maxVisible := layout.ContentHeight - 12
	if maxVisible < 5 {
		maxVisible = 5
	}

	startIdx := 0
	if m.cursor >= maxVisible {
		startIdx = m.cursor - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(procs) {
		endIdx = len(procs)
	}

	var lines []string
	for i := startIdx; i < endIdx; i++ {
		p := procs[i]

		cpuColor := styles.Theme.Success
		if p.CPU >= 50 {
			cpuColor = styles.Theme.Error
		} else if p.CPU >= 20 {
			cpuColor = styles.Theme.Warning
		}

		memColor := styles.Theme.Success
		if p.Mem >= 50 {
			memColor = styles.Theme.Error
		} else if p.Mem >= 20 {
			memColor = styles.Theme.Warning
		}

		cpuStr := lipgloss.NewStyle().Foreground(cpuColor).Render(fmt.Sprintf("%-7.1f", p.CPU))
		memStr := lipgloss.NewStyle().Foreground(memColor).Render(fmt.Sprintf("%-7.1f", p.Mem))
		rssStr := formatBytes(p.RSS * 1024)

		name := components.TruncateString(p.Name, 30)
		user := components.TruncateString(p.User, 12)

		line := fmt.Sprintf("%-7d %s %s %-10s %-12s %s",
			p.PID, cpuStr, memStr, rssStr, user, name)

		if i == m.cursor {
			lines = append(lines,
				lipgloss.NewStyle().Foreground(styles.Theme.Primary).Bold(true).
					Render("▶ "+line))
		} else {
			lines = append(lines,
				lipgloss.NewStyle().Foreground(styles.Theme.Foreground).
					Render("  "+line))
		}
	}

	// Scroll indicator
	scrollInfo := ""
	if len(procs) > maxVisible {
		scrollInfo = styles.Muted().
			Render(fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(procs)))
	}

	// Count + last update
	countInfo := styles.Muted().Render(fmt.Sprintf("Total: %d processes", len(m.processes)))
	if m.searchTerm != "" {
		countInfo = styles.Muted().
			Render(fmt.Sprintf("Matched %d of %d", len(procs), len(m.processes)))
	}
	updateInfo := ""
	if !m.lastUpdate.IsZero() {
		updateInfo = styles.Muted().
			Render(fmt.Sprintf("Updated: %s", m.lastUpdate.Format("15:04:05")))
	}

	// Assemble
	sections := []string{title, ""}
	if statusLine != "" {
		sections = append(sections, statusLine, "")
	}
	sections = append(sections,
		lipgloss.JoinHorizontal(lipgloss.Top, help, "    ", sortInfo),
		"")
	if searchLine != "" {
		sections = append(sections, searchLine, "")
	}
	sections = append(sections, header, separator)
	sections = append(sections, lines...)
	if scrollInfo != "" {
		sections = append(sections, "", scrollInfo)
	}
	sections = append(sections, "",
		lipgloss.JoinHorizontal(lipgloss.Top, countInfo, "    ", updateInfo))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return components.Viewport(content, layout.ContentHeight)
}

// Title returns the module title
func (m *Model) Title() string { return "Processes" }

// HasOpenModal returns true if confirmation dialog is open
func (m *Model) HasOpenModal() bool { return m.confirmKill || m.searching }

// -- Messages --

type processesMsg struct {
	items []Process
	note  string
}

type killMsg struct {
	note string
}

// -- Commands --

func (m *Model) refresh() tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		out, err := exec.CommandContext(ctx, "ps", "-eo", "pid,pcpu,pmem,rss,user,comm", "-r").Output()
		if err != nil {
			return processesMsg{note: fmt.Sprintf("Error: %v", err)}
		}

		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		var items []Process

		for i, line := range lines {
			if i == 0 {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) < 6 {
				continue
			}

			pid, _ := strconv.Atoi(fields[0])
			cpuPct, _ := strconv.ParseFloat(fields[1], 64)
			memPct, _ := strconv.ParseFloat(fields[2], 64)
			rss, _ := strconv.ParseInt(fields[3], 10, 64)
			user := fields[4]
			name := strings.Join(fields[5:], " ")

			// Extract binary name from full path
			if idx := strings.LastIndex(name, "/"); idx >= 0 {
				name = name[idx+1:]
			}

			items = append(items, Process{
				PID:  pid,
				CPU:  cpuPct,
				Mem:  memPct,
				RSS:  rss,
				User: user,
				Name: name,
			})
		}

		return processesMsg{
			items: items,
			note:  fmt.Sprintf("Loaded %d processes", len(items)),
		}
	}
}

func (m *Model) killProcess(pid int) tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "kill", "-15", strconv.Itoa(pid))
		if out, err := cmd.CombinedOutput(); err != nil {
			return killMsg{note: fmt.Sprintf("Failed to kill PID %d: %v %s", pid, err, string(out))}
		}
		return killMsg{note: fmt.Sprintf("Sent SIGTERM to PID %d", pid)}
	}
}

// -- Helpers --

func (m *Model) sortProcesses() {
	sort.Slice(m.processes, func(i, j int) bool {
		var less bool
		switch m.sortField {
		case SortByCPU:
			less = m.processes[i].CPU < m.processes[j].CPU
		case SortByMem:
			less = m.processes[i].Mem < m.processes[j].Mem
		case SortByPID:
			less = m.processes[i].PID < m.processes[j].PID
		case SortByName:
			less = strings.ToLower(m.processes[i].Name) < strings.ToLower(m.processes[j].Name)
		}
		if m.sortReverse {
			return !less
		}
		return less
	})
	m.applyFilter()
}

func (m *Model) applyFilter() {
	if m.searchTerm == "" {
		m.filtered = nil
		return
	}

	term := strings.ToLower(m.searchTerm)
	m.filtered = nil
	for _, p := range m.processes {
		if strings.Contains(strings.ToLower(p.Name), term) ||
			strings.Contains(strings.ToLower(p.User), term) ||
			strings.Contains(strconv.Itoa(p.PID), term) {
			m.filtered = append(m.filtered, p)
		}
	}
}

func (m *Model) visibleProcesses() []Process {
	if m.filtered != nil {
		return m.filtered
	}
	return m.processes
}

func formatBytes(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.0f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
