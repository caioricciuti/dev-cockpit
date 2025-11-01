package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// MetricCard renders a metric display card (for dashboard)
type MetricCard struct {
	Title       string
	Icon        string
	Value       string
	Percent     float64
	Status      string
	StatusLevel string // "success", "warning", "error"
	SubValues   []string
	Width       int
	Height      int
}

// Render creates the metric card view
func (m *MetricCard) Render() string {
	styles := NewBaseStyles()

	// Determine border color based on status level
	var borderColor lipgloss.Color
	switch m.StatusLevel {
	case "error":
		borderColor = styles.Theme.Error
	case "warning":
		borderColor = styles.Theme.Warning
	default:
		borderColor = styles.Theme.Primary
	}

	// Calculate inner dimensions
	dims := CalculateBoxDimensions(m.Width, m.Height, 2, 1, true)

	// Build content
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Theme.Primary).
		Width(dims.InnerWidth).
		Render(m.Icon + " " + m.Title)

	value := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Theme.Foreground).
		Width(dims.InnerWidth).
		Render(m.Value)

	// Progress bar
	progressBar := ""
	if m.Percent >= 0 && m.Percent <= 100 {
		barWidth := dims.InnerWidth
		if barWidth > 30 {
			barWidth = 30
		}
		progressBar = styles.ProgressBar(m.Percent, barWidth)
	}

	// Status indicator
	statusLine := ""
	if m.Status != "" {
		statusLine = styles.StatusIndicator(m.Status, m.StatusLevel)
	}

	// Sub-values (optional)
	subLines := []string{}
	for _, sub := range m.SubValues {
		truncated := TruncateString(sub, dims.InnerWidth)
		subLines = append(subLines, lipgloss.NewStyle().
			Foreground(styles.Theme.Muted).
			Render(truncated))
	}

	// Assemble content
	content := []string{title, value}
	if progressBar != "" {
		content = append(content, "", progressBar)
	}
	if statusLine != "" {
		content = append(content, "", statusLine)
	}
	if len(subLines) > 0 {
		content = append(content, "")
		content = append(content, subLines...)
	}

	innerContent := lipgloss.JoinVertical(lipgloss.Left, content...)

	// Apply box with proper dimensions
	box := lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Render(Viewport(innerContent, dims.InnerHeight))

	return box
}

// InfoCard renders an informational card
type InfoCard struct {
	Title   string
	Lines   []string
	Width   int
	Height  int
	Focused bool
}

// Render creates the info card view
func (i *InfoCard) Render() string {
	styles := NewBaseStyles()

	borderColor := styles.Theme.Border
	if i.Focused {
		borderColor = styles.Theme.Primary
	}

	// Calculate inner dimensions
	dims := CalculateBoxDimensions(i.Width, i.Height, 2, 1, true)

	// Title
	title := ""
	if i.Title != "" {
		title = lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.Theme.Primary).
			Width(dims.InnerWidth).
			Render(i.Title)
	}

	// Lines
	contentLines := []string{}
	if title != "" {
		contentLines = append(contentLines, title, "")
	}

	for _, line := range i.Lines {
		truncated := TruncateString(line, dims.InnerWidth)
		contentLines = append(contentLines, lipgloss.NewStyle().
			Foreground(styles.Theme.Foreground).
			Width(dims.InnerWidth).
			Render(truncated))
	}

	innerContent := lipgloss.JoinVertical(lipgloss.Left, contentLines...)

	// Apply box
	box := lipgloss.NewStyle().
		Width(i.Width).
		Height(i.Height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Render(Viewport(innerContent, dims.InnerHeight))

	return box
}

// ListCard renders a selectable list
type ListCard struct {
	Title        string
	Items        []string
	SelectedItem int
	Width        int
	Height       int
}

// Render creates the list card view
func (l *ListCard) Render() string {
	styles := NewBaseStyles()

	// Calculate inner dimensions
	dims := CalculateBoxDimensions(l.Width, l.Height, 2, 1, true)

	// Title
	title := ""
	if l.Title != "" {
		title = lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.Theme.Primary).
			Width(dims.InnerWidth).
			Render(l.Title)
	}

	// Items
	contentLines := []string{}
	if title != "" {
		contentLines = append(contentLines, title, "")
	}

	for i, item := range l.Items {
		truncated := TruncateString(item, dims.InnerWidth-3)

		if i == l.SelectedItem {
			line := lipgloss.NewStyle().
				Foreground(styles.Theme.Primary).
				Bold(true).
				Width(dims.InnerWidth).
				Render("▶ " + truncated)
			contentLines = append(contentLines, line)
		} else {
			line := lipgloss.NewStyle().
				Foreground(styles.Theme.Foreground).
				Width(dims.InnerWidth).
				Render("  " + truncated)
			contentLines = append(contentLines, line)
		}
	}

	innerContent := lipgloss.JoinVertical(lipgloss.Left, contentLines...)

	// Apply box
	box := lipgloss.NewStyle().
		Width(l.Width).
		Height(l.Height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Theme.Primary).
		Padding(1, 2).
		Render(Viewport(innerContent, dims.InnerHeight))

	return box
}

// StatusCard renders a status message card
type StatusCard struct {
	Type    string // "success", "error", "warning", "info"
	Message string
	Width   int
}

// Render creates the status card view
func (s *StatusCard) Render() string {
	if s.Message == "" {
		return ""
	}

	styles := NewBaseStyles()

	var style lipgloss.Style
	var icon string

	switch s.Type {
	case "success":
		style = styles.Success()
		icon = "✓"
	case "error":
		style = styles.Error()
		icon = "✗"
	case "warning":
		style = styles.Warning()
		icon = "⚠"
	default:
		style = styles.Info()
		icon = "ℹ"
	}

	content := TruncateString(fmt.Sprintf("%s %s", icon, s.Message), s.Width)

	return style.Width(s.Width).Render(content)
}

// Grid renders items in a grid layout
type Grid struct {
	Items   []string
	Columns int
	Spacing int
	Width   int
}

// Render creates a grid view
func (g *Grid) Render() string {
	if len(g.Items) == 0 || g.Columns <= 0 {
		return ""
	}

	// Calculate column width
	totalSpacing := g.Spacing * (g.Columns - 1)
	availableWidth := g.Width - totalSpacing
	colWidth := availableWidth / g.Columns

	if colWidth < 10 {
		colWidth = 10
	}

	rows := []string{}
	for i := 0; i < len(g.Items); i += g.Columns {
		rowItems := []string{}
		for j := 0; j < g.Columns && i+j < len(g.Items); j++ {
			item := g.Items[i+j]
			rowItems = append(rowItems, lipgloss.NewStyle().
				Width(colWidth).
				Render(item))
		}

		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			rowItems...,
		)

		// Add spacing between columns
		if g.Spacing > 0 && len(rowItems) > 1 {
			spacedRow := []string{}
			for idx, r := range strings.Split(row, "\n") {
				if idx > 0 {
					spacedRow = append(spacedRow, r)
				} else {
					spacedRow = append(spacedRow, r)
				}
			}
			row = strings.Join(spacedRow, strings.Repeat(" ", g.Spacing))
		}

		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
