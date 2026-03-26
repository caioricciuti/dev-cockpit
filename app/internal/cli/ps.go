package cli

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type process struct {
	pid  int
	cpu  float64
	mem  float64
	rss  int // KB
	user string
	name string
}

func cmdPS(args []string) {
	sortField := "cpu"
	topN := 15

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--sort":
			if i+1 < len(args) {
				sortField = args[i+1]
				i++
			}
		case "--top":
			if i+1 < len(args) {
				n, err := strconv.Atoi(args[i+1])
				if err == nil && n > 0 {
					topN = n
				}
				i++
			}
		}
	}

	printHeader(fmt.Sprintf("Dev Cockpit — Top %d Processes (sorted by %s)", topN, sortField))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "ps", "-eo", "pid,pcpu,pmem,rss,user,comm", "-r").Output()
	if err != nil {
		exitErr("Failed to read process list")
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var procs []process

	for i, line := range lines {
		if i == 0 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		pid, _ := strconv.Atoi(fields[0])
		cpuPct, _ := strconv.ParseFloat(fields[1], 64)
		memPct, _ := strconv.ParseFloat(fields[2], 64)
		rss, _ := strconv.Atoi(fields[3])

		name := fields[5]
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}

		procs = append(procs, process{
			pid:  pid,
			cpu:  cpuPct,
			mem:  memPct,
			rss:  rss,
			user: fields[4],
			name: name,
		})
	}

	// Sort
	switch sortField {
	case "mem":
		sort.Slice(procs, func(i, j int) bool { return procs[i].rss > procs[j].rss })
	case "pid":
		sort.Slice(procs, func(i, j int) bool { return procs[i].pid < procs[j].pid })
	default: // cpu
		sort.Slice(procs, func(i, j int) bool { return procs[i].cpu > procs[j].cpu })
	}

	if len(procs) > topN {
		procs = procs[:topN]
	}

	// Print table
	fmt.Printf("  %-7s  %-6s  %-6s  %-9s  %-12s  %s\n",
		mutedStyle.Render("PID"),
		mutedStyle.Render("CPU%"),
		mutedStyle.Render("MEM%"),
		mutedStyle.Render("RSS"),
		mutedStyle.Render("USER"),
		mutedStyle.Render("NAME"),
	)
	fmt.Printf("  %s\n", mutedStyle.Render(strings.Repeat("─", 60)))

	for _, p := range procs {
		cpuColor := okStyle
		if p.cpu > 50 {
			cpuColor = critStyle
		} else if p.cpu > 20 {
			cpuColor = warnStyle
		}

		rssMB := float64(p.rss) / 1024
		rssStr := fmt.Sprintf("%.0f MB", rssMB)

		fmt.Printf("  %-7d  %s  %-6.1f  %-9s  %-12s  %s\n",
			p.pid,
			cpuColor.Render(fmt.Sprintf("%-6.1f", p.cpu)),
			p.mem,
			rssStr,
			p.user,
			p.name,
		)
	}
	fmt.Println()
}
