package diagnostics

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/disk"
)

// CheckDisk checks disk usage via gopsutil
func CheckDisk() CheckResult {
	r := CheckResult{Category: "Disk"}

	usage, err := disk.Usage("/")
	if err != nil {
		r.Status = Critical
		r.Summary = "Unable to read disk usage"
		r.Suggestions = append(r.Suggestions, "Check disk permissions")
		return r
	}

	pct := usage.UsedPercent
	freeGB := float64(usage.Free) / (1 << 30)
	totalGB := float64(usage.Total) / (1 << 30)
	usedGB := float64(usage.Used) / (1 << 30)

	r.Details = append(r.Details,
		fmt.Sprintf("Total: %.1f GB", totalGB),
		fmt.Sprintf("Used: %.1f GB (%.1f%%)", usedGB, pct),
		fmt.Sprintf("Free: %.1f GB", freeGB),
	)

	switch {
	case pct >= 95:
		r.Status = Critical
		r.Summary = fmt.Sprintf("%.1f%% used — only %.1f GB free", pct, freeGB)
		r.Suggestions = append(r.Suggestions,
			"Run 'devcockpit cleanup list' to find large caches",
			"Empty Trash with 'devcockpit cleanup empty-trash'",
		)
	case pct >= 85:
		r.Status = Warning
		r.Summary = fmt.Sprintf("%.1f%% used (%.1f GB free)", pct, freeGB)
		r.Suggestions = append(r.Suggestions, "Consider freeing up disk space")
	default:
		r.Status = OK
		r.Summary = fmt.Sprintf("%.1f%% used (%.1f GB free)", pct, freeGB)
	}

	return r
}
