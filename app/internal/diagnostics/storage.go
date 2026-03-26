package diagnostics

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// cacheDirs is defined in platform-specific files:
// dirs_darwin.go — macOS cache paths
// dirs_linux.go  — Linux cache paths

// CheckStorage scans common cache directories for size
func CheckStorage() CheckResult {
	r := CheckResult{Category: "Storage"}

	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		r.Status = Warning
		r.Summary = "Cannot determine home directory"
		return r
	}

	var totalBytes uint64
	var bigCaches []string
	found := 0

	for _, cd := range cacheDirs {
		full := filepath.Join(homeDir, cd.path)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			continue
		}
		found++
		size := getSizeWithTimeout(full, 5*time.Second)
		totalBytes += size
		if size > 1<<30 { // > 1 GB
			bigCaches = append(bigCaches, fmt.Sprintf("%s: %.1f GB", cd.name, float64(size)/(1<<30)))
		}
		r.Details = append(r.Details, fmt.Sprintf("%s: %s", cd.name, formatBytes(size)))
	}

	totalGB := float64(totalBytes) / (1 << 30)

	switch {
	case len(bigCaches) >= 3 || totalGB > 10:
		r.Status = Warning
		r.Summary = fmt.Sprintf("%d caches, %.1f GB total — %d over 1 GB", found, totalGB, len(bigCaches))
		r.Suggestions = append(r.Suggestions, "Run 'devcockpit cleanup list' for details")
	default:
		r.Status = OK
		r.Summary = fmt.Sprintf("%d caches, %.1f GB total", found, totalGB)
	}

	return r
}

func getSizeWithTimeout(path string, timeout time.Duration) uint64 {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "du", "-sk", path).Output()
	if err != nil {
		return 0
	}
	var kb uint64
	fmt.Sscanf(string(out), "%d", &kb)
	return kb * 1024
}

func formatBytes(b uint64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
