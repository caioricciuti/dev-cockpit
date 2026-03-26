//go:build darwin

package logs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) fetchSystemLogs() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "log", "show", "--last", "5m", "--style", "compact")
		out, err := cmd.Output()
		if err != nil {
			return logFetchMsg{
				view: ViewSystem,
				note: fmt.Sprintf("Error: %v", err),
			}
		}

		lines := parseLogOutput(string(out), 500)
		return logFetchMsg{
			view:  ViewSystem,
			lines: lines,
			note:  fmt.Sprintf("Loaded %d system log entries", len(lines)),
		}
	}
}

func (m *Model) fetchBrewLogs() tea.Cmd {
	return func() tea.Msg {
		var allLines []LogLine

		logDirs := []string{
			"/opt/homebrew/var/log",
			filepath.Join(os.Getenv("HOME"), "Library/Logs/Homebrew"),
		}

		for _, dir := range logDirs {
			entries, err := os.ReadDir(dir)
			if err != nil {
				continue
			}

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				name := entry.Name()
				if !strings.HasSuffix(name, ".log") {
					continue
				}

				path := filepath.Join(dir, name)
				lines := tailFile(path, 50)
				svcName := strings.TrimSuffix(name, ".log")
				for _, l := range lines {
					allLines = append(allLines, LogLine{
						Content: fmt.Sprintf("[%s] %s", svcName, l),
						Level:   detectLevel(l),
					})
				}
			}
		}

		// Also try reading logs from individual service log dirs
		svcLogDir := "/opt/homebrew/var/log"
		entries, _ := os.ReadDir(svcLogDir)
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			subDir := filepath.Join(svcLogDir, entry.Name())
			subEntries, _ := os.ReadDir(subDir)
			for _, sub := range subEntries {
				if sub.IsDir() || !strings.HasSuffix(sub.Name(), ".log") {
					continue
				}
				path := filepath.Join(subDir, sub.Name())
				lines := tailFile(path, 30)
				svcName := entry.Name()
				for _, l := range lines {
					allLines = append(allLines, LogLine{
						Content: fmt.Sprintf("[%s] %s", svcName, l),
						Level:   detectLevel(l),
					})
				}
			}
		}

		if len(allLines) == 0 {
			allLines = append(allLines, LogLine{
				Content: "No Homebrew service logs found. Services log to /opt/homebrew/var/log/",
				Level:   LevelInfo,
			})
		}

		return logFetchMsg{
			view:  ViewHomebrew,
			lines: allLines,
			note:  fmt.Sprintf("Loaded %d Homebrew log entries", len(allLines)),
		}
	}
}
