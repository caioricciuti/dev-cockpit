package cli

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func cmdServices() {
	printHeader("Dev Cockpit — Services")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "brew", "services", "list").Output()
	if err != nil {
		fmt.Fprintln(stdout, mutedStyle.Render("  Homebrew services not available"))
		fmt.Fprintln(stdout)
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) <= 1 {
		fmt.Fprintln(stdout, mutedStyle.Render("  No services found"))
		fmt.Fprintln(stdout)
		return
	}

	var running, stopped, errored int

	fmt.Fprintf(stdout, "  %-20s  %-10s  %s\n",
		mutedStyle.Render("SERVICE"),
		mutedStyle.Render("STATUS"),
		mutedStyle.Render("DETAILS"),
	)
	fmt.Fprintf(stdout, "  %s\n", mutedStyle.Render(strings.Repeat("─", 50)))

	for i, line := range lines {
		if i == 0 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		name := fields[0]
		status := strings.ToLower(fields[1])
		detail := ""
		if len(fields) > 2 {
			detail = strings.Join(fields[2:], " ")
		}

		var statusStyled string
		switch status {
		case "started":
			running++
			statusStyled = okStyle.Render("running")
		case "error":
			errored++
			statusStyled = critStyle.Render("error")
		default:
			stopped++
			statusStyled = mutedStyle.Render("stopped")
		}

		fmt.Fprintf(stdout, "  %-20s  %-10s  %s\n", name, statusStyled, mutedStyle.Render(detail))
	}

	fmt.Fprintln(stdout)
	total := running + stopped + errored
	fmt.Fprintf(stdout, "  %s\n\n",
		mutedStyle.Render(fmt.Sprintf("%d total: %d running, %d stopped, %d error", total, running, stopped, errored)),
	)
}
