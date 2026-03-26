package diagnostics

import (
	"fmt"
	"strings"

	"github.com/caioricciuti/dev-cockpit/internal/config"
	diag "github.com/caioricciuti/dev-cockpit/internal/diagnostics"
	"github.com/caioricciuti/dev-cockpit/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type reportMsg struct {
	report diag.Report
}

// Model is the TUI diagnostics module
type Model struct {
	config   *config.Config
	report   *diag.Report
	loading  bool
	cursor   int
	expanded map[int]bool
	width    int
	height   int
	frame    int
}

// New creates a new diagnostics module
func New(cfg *config.Config) *Model {
	return &Model{
		config:   cfg,
		expanded: make(map[int]bool),
	}
}

func (m *Model) Title() string { return "Diagnostics" }

func (m *Model) HasOpenModal() bool { return false }

func (m *Model) Init() tea.Cmd {
	m.loading = true
	m.frame = 0
	return runDiagnostics
}

func runDiagnostics() tea.Msg {
	return reportMsg{report: diag.RunAll()}
}

func (m *Model) Update(msg tea.Msg) (interface{}, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case reportMsg:
		m.report = &msg.report
		m.loading = false

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.report != nil && m.cursor < len(m.report.Checks)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			m.expanded[m.cursor] = !m.expanded[m.cursor]
		case "r":
			m.loading = true
			m.expanded = make(map[int]bool)
			return m, runDiagnostics
		}

	default:
		if m.loading {
			m.frame++
		}
	}

	return m, nil
}

func (m *Model) View() string {
	styles := components.NewBaseStyles()
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Theme.Primary)

	b.WriteString(titleStyle.Render("DIAGNOSTICS"))

	if m.loading {
		b.WriteString(fmt.Sprintf("  %s Scanning...\n\n", styles.Spinner(m.frame)))
		return b.String()
	}

	if m.report == nil {
		b.WriteString("\n\n  Press 'r' to run diagnostics\n")
		return b.String()
	}

	// Score display
	gradeColor := styles.Theme.Success
	if m.report.Score < 75 {
		gradeColor = styles.Theme.Warning
	}
	if m.report.Score < 40 {
		gradeColor = styles.Theme.Error
	}

	scoreStyle := lipgloss.NewStyle().Bold(true).Foreground(gradeColor)
	bar := styles.ProgressBar(float64(m.report.Score), 25)

	b.WriteString(fmt.Sprintf("    Score: %s %s\n\n",
		bar,
		scoreStyle.Render(fmt.Sprintf("%d/100 [%s]", m.report.Score, m.report.Grade)),
	))

	// Check rows
	for i, c := range m.report.Checks {
		prefix := "  "
		if i == m.cursor {
			prefix = lipgloss.NewStyle().Foreground(styles.Theme.Primary).Bold(true).Render("> ")
		}

		icon, sevStyle := statusStyle(c.Status, styles)
		arrow := " "
		if m.expanded[i] {
			arrow = lipgloss.NewStyle().Foreground(styles.Theme.Muted).Render("v")
		} else if len(c.Details) > 0 || len(c.Suggestions) > 0 {
			arrow = lipgloss.NewStyle().Foreground(styles.Theme.Muted).Render(">")
		}

		category := lipgloss.NewStyle().Width(14).Render(c.Category)
		status := sevStyle.Width(6).Render(c.Status.String())
		summary := lipgloss.NewStyle().Foreground(styles.Theme.Muted).Render(c.Summary)

		b.WriteString(fmt.Sprintf("%s%s %s %s %s  %s\n", prefix, arrow, icon, category, status, summary))

		// Expanded details
		if m.expanded[i] {
			detailStyle := lipgloss.NewStyle().Foreground(styles.Theme.Muted)
			suggStyle := lipgloss.NewStyle().Foreground(styles.Theme.Warning)

			for _, d := range c.Details {
				b.WriteString(fmt.Sprintf("        %s\n", detailStyle.Render(d)))
			}
			if c.Status != diag.OK && len(c.Suggestions) > 0 {
				for _, s := range c.Suggestions {
					b.WriteString(fmt.Sprintf("        %s %s\n", suggStyle.Render("->"), s))
				}
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.Hint("[j/k] Navigate  [Enter] Expand  [r] Re-scan"))
	b.WriteString("\n")

	return b.String()
}

func statusStyle(sev diag.Severity, styles *components.BaseStyles) (string, lipgloss.Style) {
	switch sev {
	case diag.Warning:
		return lipgloss.NewStyle().Foreground(styles.Theme.Warning).Render("!"),
			lipgloss.NewStyle().Foreground(styles.Theme.Warning).Bold(true)
	case diag.Critical:
		return lipgloss.NewStyle().Foreground(styles.Theme.Error).Render("x"),
			lipgloss.NewStyle().Foreground(styles.Theme.Error).Bold(true)
	default:
		return lipgloss.NewStyle().Foreground(styles.Theme.Success).Render("*"),
			lipgloss.NewStyle().Foreground(styles.Theme.Success).Bold(true)
	}
}
