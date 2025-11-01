package security

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/caioricciuti/dev-cockpit/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the security module state
type Model struct {
	config     *config.Config
	width      int
	height     int
	firewall   string
	filevault  string
	sip        string
	gatekeeper string
	output     string
}

// New creates a new security module
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
		}
	case secMsg:
		m.firewall = msg.firewall
		m.filevault = msg.filevault
		m.sip = msg.sip
		m.gatekeeper = msg.gatekeeper
		m.output = msg.note
	}
	return m, nil
}

// View renders the module
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D9FF")).Render("🔐 SECURITY")
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Render("[r] Refresh")
	var b strings.Builder
	b.WriteString(title + "\n\n")
	if m.output != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render(m.output) + "\n\n")
	}
	b.WriteString(help + "\n\n")

	b.WriteString(fmt.Sprintf("Firewall:   %s\n", m.firewall))
	b.WriteString(fmt.Sprintf("FileVault:  %s\n", m.filevault))
	b.WriteString(fmt.Sprintf("SIP:        %s\n", m.sip))
	b.WriteString(fmt.Sprintf("Gatekeeper: %s\n", m.gatekeeper))

	// Apply viewport to prevent overflow
	content := b.String()
	maxHeight := m.height - 4 // Account for margins
	if maxHeight < 10 {
		maxHeight = 10
	}

	return lipgloss.NewStyle().MaxHeight(maxHeight).Render(content)
}

// Title returns the module title
func (m *Model) Title() string { return "Security" }

// HasOpenModal returns true if the module has an open modal/dialog
func (m *Model) HasOpenModal() bool { return false }

type secMsg struct {
	firewall, filevault, sip, gatekeeper, note string
}

func (m *Model) refresh() tea.Cmd {
	return func() tea.Msg {
		fw := readFirewall()
		fv := readCmd("fdesetup", "status")
		sip := readCmd("csrutil", "status")
		gk := readCmd("spctl", "--status")
		return secMsg{firewall: fw, filevault: fv, sip: sip, gatekeeper: gk, note: "Security status refreshed"}
	}
}

func readFirewall() string {
	out, err := exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate").CombinedOutput()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	out, err = exec.Command("defaults", "read", "/Library/Preferences/com.apple.alf", "globalstate").CombinedOutput()
	if err != nil {
		return "Unknown"
	}
	s := strings.TrimSpace(string(out))
	switch s {
	case "0":
		return "Firewall is disabled (0)"
	case "1":
		return "Firewall is enabled (1)"
	case "2":
		return "Firewall is enabled for essential services (2)"
	default:
		return "Firewall state: " + s
	}
}

func readCmd(name string, args ...string) string {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Sprintf("%s error", name)
	}
	return strings.TrimSpace(string(out))
}
