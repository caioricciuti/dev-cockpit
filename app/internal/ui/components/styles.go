package components

import "github.com/charmbracelet/lipgloss"

// Theme defines the color palette
type Theme struct {
	// Primary colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color

	// Status colors
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
	Info    lipgloss.Color

	// UI colors
	Background lipgloss.Color
	Foreground lipgloss.Color
	Muted      lipgloss.Color
	Border     lipgloss.Color

	// Special
	Highlight lipgloss.Color
}

// DefaultTheme returns the cyberpunk-inspired theme
func DefaultTheme() Theme {
	return Theme{
		Primary:    lipgloss.Color("#00D9FF"), // Cyan
		Secondary:  lipgloss.Color("#FFA500"), // Orange
		Accent:     lipgloss.Color("#FF6AC1"), // Pink

		Success:    lipgloss.Color("#0FD976"), // Green
		Warning:    lipgloss.Color("#FFA500"), // Orange
		Error:      lipgloss.Color("#FF6B6B"), // Red
		Info:       lipgloss.Color("#00D9FF"), // Cyan

		Background: lipgloss.Color("#0A0A0F"), // Very dark blue
		Foreground: lipgloss.Color("#FFFFFF"), // White
		Muted:      lipgloss.Color("#666666"), // Gray
		Border:     lipgloss.Color("#333333"), // Dark gray

		Highlight:  lipgloss.Color("#FFD700"), // Gold
	}
}

// BaseStyles provides common style builders
type BaseStyles struct {
	Theme Theme
}

// NewBaseStyles creates style builders with theme
func NewBaseStyles() *BaseStyles {
	return &BaseStyles{
		Theme: DefaultTheme(),
	}
}

// Title creates a title style
func (s *BaseStyles) Title() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(s.Theme.Primary).
		Padding(0, 1)
}

// Subtitle creates a subtitle style
func (s *BaseStyles) Subtitle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(s.Theme.Secondary).
		Padding(0, 1)
}

// Label creates a label style
func (s *BaseStyles) Label() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(s.Theme.Primary)
}

// Value creates a value style
func (s *BaseStyles) Value() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(s.Theme.Foreground)
}

// Muted creates a muted text style
func (s *BaseStyles) Muted() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(s.Theme.Muted)
}

// Success creates a success message style
func (s *BaseStyles) Success() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(s.Theme.Success).
		Bold(true)
}

// Warning creates a warning message style
func (s *BaseStyles) Warning() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(s.Theme.Warning).
		Bold(true)
}

// Error creates an error message style
func (s *BaseStyles) Error() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(s.Theme.Error).
		Bold(true)
}

// Info creates an info message style
func (s *BaseStyles) Info() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(s.Theme.Info).
		Bold(true)
}

// Box creates a bordered box style with proper dimensions
func (s *BaseStyles) Box(width, height int, borderColor lipgloss.Color) lipgloss.Style {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2)

	if width > 0 {
		style = style.Width(width)
	}

	if height > 0 {
		style = style.Height(height)
	}

	return style
}

// Card creates a card style (box with proper sizing)
func (s *BaseStyles) Card(width, height int) lipgloss.Style {
	return s.Box(width, height, s.Theme.Border)
}

// ActiveCard creates an active/focused card style
func (s *BaseStyles) ActiveCard(width, height int) lipgloss.Style {
	return s.Box(width, height, s.Theme.Primary)
}

// StatusIndicator returns a styled status indicator
func (s *BaseStyles) StatusIndicator(status string, level string) string {
	var color lipgloss.Color
	switch level {
	case "success", "good", "healthy":
		color = s.Theme.Success
	case "warning", "caution":
		color = s.Theme.Warning
	case "error", "critical", "danger":
		color = s.Theme.Error
	case "info", "normal":
		color = s.Theme.Info
	default:
		color = s.Theme.Muted
	}

	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Render("● " + status)
}

// ProgressBar renders a horizontal progress bar
func (s *BaseStyles) ProgressBar(percent float64, width int) string {
	if width <= 0 {
		width = 20
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int(percent * float64(width) / 100)
	if filled > width {
		filled = width
	}

	// Determine color based on percentage
	var barColor lipgloss.Color
	if percent >= 90 {
		barColor = s.Theme.Error
	} else if percent >= 75 {
		barColor = s.Theme.Warning
	} else {
		barColor = s.Theme.Success
	}

	filledBar := lipgloss.NewStyle().
		Foreground(barColor).
		Render(lipgloss.NewStyle().Width(filled).Render("█"))

	emptyBar := lipgloss.NewStyle().
		Foreground(s.Theme.Border).
		Render(lipgloss.NewStyle().Width(width - filled).Render("░"))

	return filledBar + emptyBar
}

// Spinner returns spinner frame
func (s *BaseStyles) Spinner(frame int) string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return lipgloss.NewStyle().
		Foreground(s.Theme.Primary).
		Bold(true).
		Render(frames[frame%len(frames)])
}

// Badge renders a small badge
func (s *BaseStyles) Badge(text string, color lipgloss.Color) string {
	return lipgloss.NewStyle().
		Foreground(color).
		Background(s.Theme.Background).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Render(text)
}

// Hint renders a hint/tip text
func (s *BaseStyles) Hint(text string) string {
	return lipgloss.NewStyle().
		Foreground(s.Theme.Muted).
		Italic(true).
		Render(text)
}

// KeyBinding renders a keyboard shortcut
func (s *BaseStyles) KeyBinding(key, description string) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(s.Theme.Primary).
		Bold(true).
		Render(key)

	descStyle := lipgloss.NewStyle().
		Foreground(s.Theme.Foreground).
		Render(description)

	return keyStyle + " " + descStyle
}
