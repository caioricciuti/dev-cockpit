package diagnostics

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// CheckPerformance finds heavy processes (>20% CPU or >500 MB RSS)
func CheckPerformance() CheckResult {
	r := CheckResult{Category: "Performance"}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "ps", "-eo", "pid,pcpu,pmem,rss,comm", "-r").Output()
	if err != nil {
		r.Status = Warning
		r.Summary = "Unable to read process list"
		return r
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var heavyCPU, heavyMem []string

	for i, line := range lines {
		if i == 0 {
			continue // skip header
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		cpuPct, _ := strconv.ParseFloat(fields[1], 64)
		rssKB, _ := strconv.Atoi(fields[3])
		rssMB := rssKB / 1024

		name := fields[4]
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}

		if cpuPct > 20 {
			heavyCPU = append(heavyCPU, fmt.Sprintf("%s (%.0f%% CPU)", name, cpuPct))
		}
		if rssMB > 500 {
			heavyMem = append(heavyMem, fmt.Sprintf("%s (%d MB)", name, rssMB))
		}
	}

	for _, p := range heavyCPU {
		r.Details = append(r.Details, "High CPU: "+p)
	}
	for _, p := range heavyMem {
		r.Details = append(r.Details, "High Memory: "+p)
	}

	total := len(heavyCPU) + len(heavyMem)
	switch {
	case total >= 3:
		r.Status = Warning
		r.Summary = fmt.Sprintf("%d heavy processes detected", total)
		r.Suggestions = append(r.Suggestions, "Run 'devcockpit ps --sort cpu' to investigate")
	case total > 0:
		r.Status = OK
		r.Summary = fmt.Sprintf("%d heavy process(es), within normal range", total)
	default:
		r.Status = OK
		r.Summary = "No heavy processes"
	}

	return r
}
