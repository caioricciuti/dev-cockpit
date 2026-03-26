package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// cacheLocations is defined in platform-specific files:
// locations_darwin.go — macOS cache paths
// locations_linux.go  — Linux cache paths

func cmdCleanupList() {
	printHeader("Dev Cockpit — Cache Sizes")

	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		exitErr("Cannot determine home directory")
	}

	fmt.Printf("  %-25s  %s\n",
		mutedStyle.Render("CACHE"),
		mutedStyle.Render("SIZE"),
	)
	fmt.Printf("  %s\n", mutedStyle.Render(strings.Repeat("─", 40)))

	var totalBytes uint64

	for _, loc := range cacheLocations {
		full := filepath.Join(homeDir, loc.path)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			continue
		}

		size := getDirSize(full)
		totalBytes += size

		sizeStr := formatSize(size)
		sizeStyle := mutedStyle
		if size > 1<<30 { // > 1 GB
			sizeStyle = warnStyle
		}
		if size > 5<<30 { // > 5 GB
			sizeStyle = critStyle
		}

		fmt.Printf("  %-25s  %s\n", loc.name, sizeStyle.Render(sizeStr))
	}

	fmt.Printf("  %s\n", mutedStyle.Render(strings.Repeat("─", 40)))
	fmt.Printf("  %-25s  %s\n\n",
		labelStyle.Bold(true).Render("Total"),
		labelStyle.Bold(true).Render(formatSize(totalBytes)),
	)
}

func getDirSize(path string) uint64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "du", "-sk", path).Output()
	if err != nil {
		return 0
	}
	var kb uint64
	fmt.Sscanf(string(out), "%d", &kb)
	return kb * 1024
}

func formatSize(b uint64) string {
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
