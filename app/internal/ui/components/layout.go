package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Layout manages terminal space and prevents overflow
type Layout struct {
	Width  int
	Height int

	// Reserved space (tabs, footer, hints)
	TabsHeight   int
	FooterHeight int
	HintHeight   int

	// Calculated available content space
	ContentWidth  int
	ContentHeight int
}

// NewLayout creates a layout manager that calculates usable space
func NewLayout(width, height int) *Layout {
	l := &Layout{
		Width:        width,
		Height:       height,
		TabsHeight:   4, // tabs with border
		FooterHeight: 3, // footer with border
		HintHeight:   0, // only shown when not focused
	}

	// Calculate available content space (with safety margin)
	l.ContentWidth = width - 4  // margin on sides
	if l.ContentWidth < 40 {
		l.ContentWidth = 40 // minimum
	}

	// Total height minus reserved areas minus spacing
	l.ContentHeight = height - l.TabsHeight - l.FooterHeight - 2
	if l.ContentHeight < 10 {
		l.ContentHeight = 10 // minimum
	}

	return l
}

// WithHint recalculates layout when hint is shown
func (l *Layout) WithHint() *Layout {
	newLayout := *l
	newLayout.HintHeight = 3
	newLayout.ContentHeight = l.Height - l.TabsHeight - l.FooterHeight - l.HintHeight - 2
	if newLayout.ContentHeight < 10 {
		newLayout.ContentHeight = 10
	}
	return &newLayout
}

// SplitHorizontal splits available width into N columns with spacing
func (l *Layout) SplitHorizontal(columns int, spacing int) []int {
	if columns <= 0 {
		return []int{}
	}

	totalSpacing := spacing * (columns - 1)
	availableWidth := l.ContentWidth - totalSpacing

	columnWidth := availableWidth / columns
	if columnWidth < 20 {
		columnWidth = 20 // minimum column width
	}

	widths := make([]int, columns)
	for i := range widths {
		widths[i] = columnWidth
	}

	return widths
}

// BoxDimensions calculates inner dimensions accounting for border and padding
type BoxDimensions struct {
	OuterWidth  int
	OuterHeight int
	InnerWidth  int
	InnerHeight int
	PaddingX    int
	PaddingY    int
}

// CalculateBoxDimensions returns dimensions with border and padding accounted for
func CalculateBoxDimensions(outerWidth, outerHeight, paddingX, paddingY int, hasBorder bool) BoxDimensions {
	dims := BoxDimensions{
		OuterWidth:  outerWidth,
		OuterHeight: outerHeight,
		PaddingX:    paddingX,
		PaddingY:    paddingY,
	}

	// Subtract border (2 chars per side if rounded/normal border)
	borderOffset := 0
	if hasBorder {
		borderOffset = 2
	}

	// Inner dimensions = outer - border - padding
	dims.InnerWidth = outerWidth - borderOffset - (paddingX * 2)
	if dims.InnerWidth < 10 {
		dims.InnerWidth = 10
	}

	dims.InnerHeight = outerHeight - borderOffset - (paddingY * 2)
	if dims.InnerHeight < 3 {
		dims.InnerHeight = 3
	}

	return dims
}

// Viewport truncates content to fit available height (prevents overflow)
func Viewport(content string, maxHeight int) string {
	if maxHeight <= 0 {
		return ""
	}

	lines := lipgloss.NewStyle().MaxHeight(maxHeight).Render(content)
	return lines
}

// TruncateString truncates text with ellipsis to fit width
func TruncateString(text string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(text)
	if len(runes) <= width {
		return text
	}

	if width <= 3 {
		return string(runes[:width])
	}

	return string(runes[:width-3]) + "..."
}
