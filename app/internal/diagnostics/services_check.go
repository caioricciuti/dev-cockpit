package diagnostics

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// CheckServices checks Homebrew service statuses
func CheckServices() CheckResult {
	r := CheckResult{Category: "Services"}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "brew", "services", "list").Output()
	if err != nil {
		// brew not installed or services unavailable
		r.Status = OK
		r.Summary = "Homebrew services not available"
		return r
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var running, stopped, errored int

	for i, line := range lines {
		if i == 0 {
			continue // skip header
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := fields[0]
		status := strings.ToLower(fields[1])

		switch status {
		case "started":
			running++
			r.Details = append(r.Details, fmt.Sprintf("[running] %s", name))
		case "error":
			errored++
			r.Details = append(r.Details, fmt.Sprintf("[error]   %s", name))
		default:
			stopped++
			r.Details = append(r.Details, fmt.Sprintf("[stopped] %s", name))
		}
	}

	total := running + stopped + errored
	switch {
	case errored > 0:
		r.Status = Warning
		r.Summary = fmt.Sprintf("%d running, %d error, %d stopped (of %d)", running, errored, stopped, total)
		r.Suggestions = append(r.Suggestions, "Run 'brew services list' to see error details")
	default:
		r.Status = OK
		r.Summary = fmt.Sprintf("%d running, %d stopped (of %d)", running, stopped, total)
	}

	return r
}
