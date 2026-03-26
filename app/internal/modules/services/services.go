package services

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/caioricciuti/dev-cockpit/internal/config"
	"github.com/caioricciuti/dev-cockpit/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ServiceStatus represents the state of a service
type ServiceStatus int

const (
	StatusUnknown ServiceStatus = iota
	StatusRunning
	StatusStopped
	StatusError
)

// Service represents a detected dev service
type Service struct {
	Name    string
	Status  ServiceStatus
	Port    int
	PID     int
	Type    string // "homebrew", "detected"
	Details string
}

// Model represents the services monitor state
type Model struct {
	config *config.Config
	width  int
	height int

	services   []Service
	cursor     int
	output     string
	loading    bool
	lastUpdate time.Time
	hasBrew    bool

	// Confirmation
	confirmAction bool
	actionTarget  *Service
	actionType    string // "start" or "stop"
}

// New creates a new services monitor module
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
		if m.loading {
			return m, nil
		}

		if m.confirmAction {
			switch msg.String() {
			case "y", "Y":
				target := m.actionTarget
				action := m.actionType
				m.confirmAction = false
				m.actionTarget = nil
				if target != nil {
					return m, m.toggleService(target, action)
				}
			default:
				m.confirmAction = false
				m.actionTarget = nil
			}
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
			if m.cursor < len(m.services)-1 {
				m.cursor++
			}
		case "s", "enter":
			if m.cursor < len(m.services) {
				svc := m.services[m.cursor]
				if svc.Type == "homebrew" {
					action := "start"
					if svc.Status == StatusRunning {
						action = "stop"
					}
					m.confirmAction = true
					m.actionTarget = &svc
					m.actionType = action
				} else {
					m.output = "Only Homebrew services can be managed. Use brew to install services."
				}
			}
		}

	case servicesMsg:
		m.services = msg.items
		m.output = msg.note
		m.loading = false
		m.hasBrew = msg.hasBrew
		m.lastUpdate = time.Now()
		if m.cursor >= len(m.services) {
			m.cursor = 0
		}

	case actionResultMsg:
		m.output = msg.note
		m.loading = false
		return m, m.refresh()
	}

	return m, nil
}

// View renders the services monitor
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	styles := components.NewBaseStyles()
	layout := components.NewLayout(m.width, m.height)

	title := styles.Title().Render("🔌 DEV SERVICES")
	help := styles.Muted().Render("[j/k] Navigate  [s] Start/Stop  [r] Refresh")

	// Status
	statusLine := ""
	if m.loading {
		statusLine = lipgloss.NewStyle().Foreground(styles.Theme.Primary).Render("Scanning services...")
	} else if m.output != "" {
		statusLine = lipgloss.NewStyle().Foreground(styles.Theme.Success).Render(m.output)
	}

	// Confirmation
	if m.confirmAction && m.actionTarget != nil {
		verb := capitalize(m.actionType)
		statusLine = lipgloss.NewStyle().
			Foreground(styles.Theme.Warning).Bold(true).
			Render(fmt.Sprintf("%s %s? [y] confirm / [any] cancel", verb, m.actionTarget.Name))
	}

	var sections []string
	sections = append(sections, title, "")
	if statusLine != "" {
		sections = append(sections, statusLine, "")
	}
	sections = append(sections, help, "")

	if len(m.services) == 0 && !m.loading {
		sections = append(sections,
			styles.Muted().Render("No dev services detected."),
			styles.Muted().Render("Install services via Homebrew (e.g. brew install postgresql) to manage them here."),
			"",
			styles.Muted().Render("Also detects services listening on well-known ports:"),
			styles.Muted().Render("PostgreSQL :5432 | MySQL :3306 | Redis :6379 | MongoDB :27017"),
			styles.Muted().Render("Elasticsearch :9200 | RabbitMQ :5672 | Nginx :80 | ClickHouse :8123"),
		)
	} else {
		// Group by status
		var running, stopped []int
		for i, svc := range m.services {
			if svc.Status == StatusRunning {
				running = append(running, i)
			} else {
				stopped = append(stopped, i)
			}
		}

		if len(running) > 0 {
			sections = append(sections,
				lipgloss.NewStyle().Foreground(styles.Theme.Success).Bold(true).
					Render(fmt.Sprintf("● RUNNING (%d)", len(running))),
				"")
			for _, idx := range running {
				sections = append(sections, m.renderService(idx, styles))
			}
		}

		if len(stopped) > 0 {
			if len(running) > 0 {
				sections = append(sections, "")
			}
			sections = append(sections,
				lipgloss.NewStyle().Foreground(styles.Theme.Muted).Bold(true).
					Render(fmt.Sprintf("○ STOPPED (%d)", len(stopped))),
				"")
			for _, idx := range stopped {
				sections = append(sections, m.renderService(idx, styles))
			}
		}
	}

	// Footer
	sections = append(sections, "")
	var footerParts []string
	footerParts = append(footerParts,
		styles.Muted().Render(fmt.Sprintf("Total: %d services", len(m.services))))
	if !m.hasBrew {
		footerParts = append(footerParts,
			lipgloss.NewStyle().Foreground(styles.Theme.Warning).Render("Homebrew not found"))
	}
	if !m.lastUpdate.IsZero() {
		footerParts = append(footerParts,
			styles.Muted().Render(fmt.Sprintf("Updated: %s", m.lastUpdate.Format("15:04:05"))))
	}
	sections = append(sections, strings.Join(footerParts, "    "))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return components.Viewport(content, layout.ContentHeight)
}

func (m *Model) renderService(idx int, styles *components.BaseStyles) string {
	svc := m.services[idx]

	// Status indicator
	var statusIcon string
	var statusColor lipgloss.Color
	switch svc.Status {
	case StatusRunning:
		statusIcon = "●"
		statusColor = styles.Theme.Success
	case StatusStopped:
		statusIcon = "○"
		statusColor = styles.Theme.Muted
	case StatusError:
		statusIcon = "✗"
		statusColor = styles.Theme.Error
	default:
		statusIcon = "?"
		statusColor = styles.Theme.Muted
	}

	indicator := lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon)

	// Name
	nameStyle := lipgloss.NewStyle().Foreground(styles.Theme.Foreground).Bold(true)
	if svc.Status != StatusRunning {
		nameStyle = lipgloss.NewStyle().Foreground(styles.Theme.Muted)
	}
	name := nameStyle.Render(fmt.Sprintf("%-20s", components.TruncateString(svc.Name, 20)))

	// Details
	var details []string
	if svc.Port > 0 {
		details = append(details, fmt.Sprintf(":%d", svc.Port))
	}
	if svc.PID > 0 {
		details = append(details, fmt.Sprintf("PID %d", svc.PID))
	}
	if svc.Details != "" {
		details = append(details, svc.Details)
	}

	detailStr := styles.Muted().Render(strings.Join(details, "  "))

	// Type badge
	typeBadge := ""
	if svc.Type == "homebrew" {
		typeBadge = lipgloss.NewStyle().Foreground(styles.Theme.Secondary).Render(" [brew]")
	} else if svc.Type == "detected" {
		typeBadge = lipgloss.NewStyle().Foreground(styles.Theme.Info).Render(" [port]")
	}

	line := fmt.Sprintf("  %s %s %s%s", indicator, name, detailStr, typeBadge)

	if idx == m.cursor {
		return lipgloss.NewStyle().Foreground(styles.Theme.Primary).Bold(true).
			Render("▶" + line[1:])
	}
	return line
}

// Title returns the module title
func (m *Model) Title() string { return "Services" }

// HasOpenModal returns true if confirmation dialog is open
func (m *Model) HasOpenModal() bool { return m.confirmAction }

// -- Messages --

type servicesMsg struct {
	items   []Service
	note    string
	hasBrew bool
}

type actionResultMsg struct {
	note string
}

// Well-known dev service ports
var knownPorts = []struct {
	name string
	port int
}{
	{"PostgreSQL", 5432},
	{"MySQL", 3306},
	{"Redis", 6379},
	{"MongoDB", 27017},
	{"Elasticsearch", 9200},
	{"RabbitMQ", 5672},
	{"Memcached", 11211},
	{"Nginx", 80},
	{"MinIO", 9000},
	{"ClickHouse", 8123},
}

// -- Commands --

func (m *Model) refresh() tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		var allServices []Service
		seen := make(map[string]bool)
		hasBrew := false

		// 1. Check Homebrew services
		if _, err := exec.LookPath("brew"); err == nil {
			hasBrew = true
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			out, err := exec.CommandContext(ctx, "brew", "services", "list").Output()
			cancel()

			if err == nil {
				lines := strings.Split(strings.TrimSpace(string(out)), "\n")
				for i, line := range lines {
					if i == 0 {
						continue
					}
					fields := strings.Fields(line)
					if len(fields) < 2 {
						continue
					}

					name := fields[0]
					status := StatusStopped
					if fields[1] == "started" {
						status = StatusRunning
					} else if fields[1] == "error" {
						status = StatusError
					}

					port := matchKnownPort(name)
					pid := 0
					if status == StatusRunning && port > 0 {
						pid = getPIDForPort(port)
					}

					allServices = append(allServices, Service{
						Name:   name,
						Status: status,
						Port:   port,
						PID:    pid,
						Type:   "homebrew",
					})
					seen[strings.ToLower(name)] = true
				}
			}
		}

		// 2. Scan well-known ports for non-Homebrew services
		for _, kp := range knownPorts {
			nameKey := strings.ToLower(strings.Split(kp.name, " ")[0])
			// Skip if already found via brew
			alreadySeen := false
			for k := range seen {
				if strings.Contains(k, nameKey) {
					alreadySeen = true
					break
				}
			}
			if alreadySeen {
				continue
			}

			pid := getPIDForPort(kp.port)
			if pid > 0 {
				allServices = append(allServices, Service{
					Name:   kp.name,
					Status: StatusRunning,
					Port:   kp.port,
					PID:    pid,
					Type:   "detected",
				})
			}
		}

		note := fmt.Sprintf("Found %d services", len(allServices))
		return servicesMsg{items: allServices, note: note, hasBrew: hasBrew}
	}
}

func (m *Model) toggleService(svc *Service, action string) tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "brew", "services", action, svc.Name)
		if out, err := cmd.CombinedOutput(); err != nil {
			return actionResultMsg{
				note: fmt.Sprintf("Failed to %s %s: %v %s", action, svc.Name, err, string(out)),
			}
		}

		verb := capitalize(action) + "ed"
		if action == "stop" {
			verb = "Stopped"
		}
		return actionResultMsg{
			note: fmt.Sprintf("%s %s", verb, svc.Name),
		}
	}
}

// -- Helpers --

func matchKnownPort(brewName string) int {
	lower := strings.ToLower(brewName)
	for _, kp := range knownPorts {
		keyword := strings.ToLower(strings.Split(kp.name, " ")[0])
		if strings.Contains(lower, keyword) {
			return kp.port
		}
	}
	return 0
}

func getPIDForPort(port int) int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "lsof", "-ti", fmt.Sprintf(":%d", port)).Output()
	if err != nil {
		return 0
	}

	pidStr := strings.TrimSpace(strings.Split(string(out), "\n")[0])
	pid, _ := strconv.Atoi(pidStr)
	return pid
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
