package network

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/caioricciuti/dev-cockpit/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gnet "github.com/shirou/gopsutil/v3/net"
)

// Model represents the network module state
type Model struct {
	config  *config.Config
	width   int
	height  int
	ifaces  []gnet.InterfaceStat
	gateway string
	output  string
	cursor  int
}

// New creates a new network module
func New(cfg *config.Config) *Model { return &Model{config: cfg} }

// Init initializes the module
func (m *Model) Init() tea.Cmd { return m.refresh() }

// Update handles messages
func (m *Model) Update(msg tea.Msg) (interface{}, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			return m, m.refresh()
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.ifaces)-1 {
				m.cursor++
			}
		case "p":
			return m, m.ping()
		}
	case netMsg:
		m.ifaces = msg.ifaces
		m.gateway = msg.gateway
		m.output = msg.note
		if m.cursor >= len(m.ifaces) {
			m.cursor = 0
		}
	case actionMsg:
		m.output = msg.note
	}
	return m, nil
}

// View renders the module
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D9FF")).Render("🌐 NETWORK")
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Render("[r] Refresh  [p] Ping  [↑/↓] Select")
	var b strings.Builder
	b.WriteString(title + "\n\n")
	if m.output != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render(m.output) + "\n\n")
	}
	b.WriteString(help + "\n\n")

	if len(m.ifaces) == 0 {
		b.WriteString("No interfaces found.\n")
	} else {
		gw := m.gateway
		if gw == "" {
			gw = "(no default)"
		}
		b.WriteString(fmt.Sprintf("Default Gateway: %s\n\n", gw))

		item := lipgloss.NewStyle().PaddingLeft(2)
		sel := lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#00D9FF")).Bold(true)

		for i, ifc := range m.ifaces {
			addrs := []string{}
			for _, a := range ifc.Addrs {
				addrs = append(addrs, a.Addr)
			}
			line := fmt.Sprintf("%-12s %s", ifc.Name, strings.Join(addrs, ", "))
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
func (m *Model) Title() string { return "Network" }

// HasOpenModal returns true if the module has an open modal/dialog
func (m *Model) HasOpenModal() bool { return false }

type netMsg struct {
	ifaces  []gnet.InterfaceStat
	gateway string
	note    string
}
type actionMsg struct{ note string }

func (m *Model) refresh() tea.Cmd {
	return func() tea.Msg {
		ifaces, _ := gnet.Interfaces()
		gw := getDefaultGateway()
		note := fmt.Sprintf("%d interfaces", len(ifaces))
		return netMsg{ifaces: ifaces, gateway: gw, note: note}
	}
}

func (m *Model) ping() tea.Cmd {
	target := m.gateway
	if target == "" {
		target = "1.1.1.1"
	}
	return func() tea.Msg {
		out, err := exec.Command("ping", "-c", "2", target).CombinedOutput()
		if err != nil {
			return actionMsg{note: fmt.Sprintf("Ping error: %v", err)}
		}
		// short summary
		lines := strings.Split(string(out), "\n")
		if len(lines) > 2 {
			return actionMsg{note: lines[len(lines)-3] + " | " + lines[len(lines)-2]}
		}
		return actionMsg{note: string(out)}
	}
}

func getDefaultGateway() string {
	out, err := exec.Command("sh", "-c", "route get default | awk '/gateway/{print $2}'").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
