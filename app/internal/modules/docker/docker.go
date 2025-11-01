package docker

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/caioricciuti/dev-cockpit/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Container struct {
	ID     string
	Name   string
	Image  string
	Status string
	State  string
}

// Model represents the Docker module state
type Model struct {
	config     *config.Config
	width      int
	height     int
	containers []Container
	cursor     int
	output     string
	runningCmd bool
	dockerOK   bool
}

// New creates a new Docker module
func New(cfg *config.Config) *Model {
	return &Model{config: cfg}
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
		if m.runningCmd {
			return m, nil
		}
		switch msg.String() {
		case "r":
			return m, m.refresh()
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.containers)-1 {
				m.cursor++
			}
		case "s":
			if m.cursor < len(m.containers) {
				return m, m.toggleStartStop(m.containers[m.cursor])
			}
		case "l":
			if m.cursor < len(m.containers) {
				return m, m.tailLogs(m.containers[m.cursor])
			}
		}
	case containersMsg:
		m.containers = msg.items
		m.output = msg.note
		m.dockerOK = msg.ok
		if m.cursor >= len(m.containers) {
			m.cursor = 0
		}
		m.runningCmd = false
	case actionMsg:
		m.output = msg.note
		m.runningCmd = false
	}
	return m, nil
}

// View renders the module
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D9FF")).Render("🐳 DOCKER")
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Render("[r] Refresh  [s] Start/Stop  [l] Logs")

	if !m.dockerOK {
		msg := "Docker not available. Install Docker Desktop and ensure the daemon is running."
		return lipgloss.JoinVertical(lipgloss.Top, title, "", msg)
	}

	var b strings.Builder
	b.WriteString(title + "\n\n")
	if m.output != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render(m.output))
		b.WriteString("\n\n")
	}
	b.WriteString(help + "\n\n")

	if len(m.containers) == 0 {
		b.WriteString("No containers found.\n")
	} else {
		item := lipgloss.NewStyle().PaddingLeft(2)
		sel := lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#00D9FF")).Bold(true)

		for i, c := range m.containers {
			line := fmt.Sprintf("%-20s %-18s %-10s %s", truncate(c.Name, 20), truncate(c.Image, 18), c.State, c.Status)
			if i == m.cursor {
				b.WriteString(sel.Render("▶ " + line))
			} else {
				b.WriteString(item.Render("  " + line))
			}
			b.WriteString("\n")
		}
	}

	// Apply viewport to prevent overflow
	content := b.String()
	maxHeight := m.height - 4 // Account for margins
	if maxHeight < 10 {
		maxHeight = 10
	}

	return lipgloss.NewStyle().MaxHeight(maxHeight).Render(content)
}

// Title returns the module title
func (m *Model) Title() string { return "Docker" }

// HasOpenModal returns true if the module has an open modal/dialog
func (m *Model) HasOpenModal() bool { return false }

// Messages
type containersMsg struct {
	items []Container
	note  string
	ok    bool
}
type actionMsg struct{ note string }

func (m *Model) refresh() tea.Cmd {
	m.runningCmd = true
	return func() tea.Msg {
		if _, err := exec.LookPath("docker"); err != nil {
			return containersMsg{ok: false, note: "docker CLI not found"}
		}
		out, err := exec.Command("docker", "ps", "-a", "--format", "{{.ID}}|{{.Names}}|{{.Image}}|{{.Status}}|{{.State}}").Output()
		if err != nil {
			return containersMsg{ok: false, note: "Docker daemon not reachable"}
		}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		items := []Container{}
		for _, l := range lines {
			if strings.TrimSpace(l) == "" {
				continue
			}
			parts := strings.SplitN(l, "|", 5)
			if len(parts) < 5 {
				continue
			}
			items = append(items, Container{ID: parts[0], Name: parts[1], Image: parts[2], Status: parts[3], State: parts[4]})
		}
		note := fmt.Sprintf("%d containers", len(items))
		return containersMsg{items: items, note: note, ok: true}
	}
}

func (m *Model) toggleStartStop(c Container) tea.Cmd {
	m.runningCmd = true
	return func() tea.Msg {
		var cmd *exec.Cmd
		if c.State == "running" {
			cmd = exec.Command("docker", "stop", c.ID)
		} else {
			cmd = exec.Command("docker", "start", c.ID)
		}
		if out, err := cmd.CombinedOutput(); err != nil {
			return actionMsg{note: fmt.Sprintf("Error: %v: %s", err, string(out))}
		}
		// Refresh after action
		return actionMsg{note: fmt.Sprintf("Toggled %s", c.Name)}
	}
}

func (m *Model) tailLogs(c Container) tea.Cmd {
	m.runningCmd = true
	return func() tea.Msg {
		cmd := exec.Command("docker", "logs", "--tail", "50", c.ID)
		pipe, err := cmd.StdoutPipe()
		if err != nil {
			return actionMsg{note: err.Error()}
		}
		_ = cmd.Start()
		scanner := bufio.NewScanner(pipe)
		var b strings.Builder
		for scanner.Scan() {
			b.WriteString(scanner.Text())
			b.WriteString("\n")
		}
		_ = cmd.Wait()
		out := b.String()
		if len(out) > 500 {
			lines := strings.Split(out, "\n")
			if len(lines) > 10 {
				out = strings.Join(lines[len(lines)-10:], "\n")
			}
		}
		return actionMsg{note: fmt.Sprintf("Logs for %s:\n%s", c.Name, out)}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
