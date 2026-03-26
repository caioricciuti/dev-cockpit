package cli

import (
	"fmt"

	"github.com/caioricciuti/dev-cockpit/internal/diagnostics"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

func cmdStatus(version string) {
	printHeader(fmt.Sprintf("Dev Cockpit v%s — Quick Status", version))

	// CPU
	cpuPct, err := cpu.Percent(0, false)
	cpuStr := "N/A"
	if err == nil && len(cpuPct) > 0 {
		cpuStr = fmt.Sprintf("%.1f%%", cpuPct[0])
	}
	printRow("CPU", cpuStr)

	// Memory
	vm, err := mem.VirtualMemory()
	memStr := "N/A"
	if err == nil {
		memStr = fmt.Sprintf("%.1f%% (%.1f GB / %.1f GB)",
			vm.UsedPercent,
			float64(vm.Used)/(1<<30),
			float64(vm.Total)/(1<<30))
	}
	printRow("Memory", memStr)

	// Disk
	du, err := disk.Usage("/")
	diskStr := "N/A"
	if err == nil {
		diskStr = fmt.Sprintf("%.1f%% (%.1f GB free)",
			du.UsedPercent,
			float64(du.Free)/(1<<30))
	}
	printRow("Disk", diskStr)

	fmt.Println()

	// Quick diagnostics score
	report := diagnostics.RunAll()

	gradeStyle := okStyle
	if report.Score < 75 {
		gradeStyle = warnStyle
	}
	if report.Score < 40 {
		gradeStyle = critStyle
	}

	scoreBar := renderScoreBar(report.Score)
	fmt.Printf("  %s %s %s\n\n",
		labelStyle.Render("Health Score"),
		scoreBar,
		gradeStyle.Bold(true).Render(fmt.Sprintf("%d/100 [%s]", report.Score, report.Grade)),
	)

	for _, c := range report.Checks {
		printStatus(c.Category, c.Status, c.Summary)
	}
	fmt.Println()

	if report.Score < 75 {
		fmt.Println(mutedStyle.Render("  Run 'devcockpit diag' for details and suggestions."))
		fmt.Println()
	}
}

func renderScoreBar(score int) string {
	width := 20
	filled := score * width / 100
	if filled > width {
		filled = width
	}

	color := "#0FD976"
	if score < 75 {
		color = "#FFA500"
	}
	if score < 40 {
		color = "#FF6B6B"
	}

	bar := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(
		repeatChar('█', filled),
	)
	empty := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render(
		repeatChar('░', width-filled),
	)
	return bar + empty
}

func repeatChar(ch rune, n int) string {
	r := make([]rune, n)
	for i := range r {
		r[i] = ch
	}
	return string(r)
}
