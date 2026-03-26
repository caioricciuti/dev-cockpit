package cli

import (
	"fmt"
	"os"

	"github.com/caioricciuti/dev-cockpit/internal/diagnostics"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D9FF")).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#333333")).
			PaddingBottom(1).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Width(18).
			Foreground(lipgloss.Color("#FFFFFF"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	okStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#0FD976"))
	warnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	critStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))

	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
)

// Run routes CLI commands. Returns true if a command was handled.
func Run(args []string, version string) bool {
	if len(args) == 0 {
		return false
	}

	switch args[0] {
	case "status":
		cmdStatus(version)
	case "diag":
		cmdDiag()
	case "ps":
		cmdPS(args[1:])
	case "services":
		cmdServices()
	case "security":
		cmdSecurity()
	case "cleanup":
		// Only handle "cleanup list"; let main.go handle "cleanup empty-trash"
		if len(args) > 1 && args[1] == "list" {
			cmdCleanupList()
		} else {
			return false
		}
	default:
		return false
	}

	return true
}

func printHeader(title string) {
	fmt.Println(headerStyle.Render(title))
}

func printStatus(label string, sev diagnostics.Severity, detail string) {
	icon := okStyle.Render("✓")
	sevStr := okStyle.Render("OK  ")
	switch sev {
	case diagnostics.Warning:
		icon = warnStyle.Render("⚠")
		sevStr = warnStyle.Render("WARN")
	case diagnostics.Critical:
		icon = critStyle.Render("✗")
		sevStr = critStyle.Render("CRIT")
	}
	fmt.Printf("  %s %s %s  %s\n", icon, labelStyle.Render(label), sevStr, valueStyle.Render(detail))
}

func printRow(label, value string) {
	fmt.Printf("  %s %s\n", labelStyle.Render(label), valueStyle.Render(value))
}

func exitErr(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}
