//go:build linux

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

		cmd := exec.CommandContext(ctx, "journalctl", "--since", "5 minutes ago", "--no-pager", "-q")
		out, err := cmd.Output()
		if err != nil {
			// Fallback: try reading /var/log/syslog or /var/log/messages
			for _, path := range []string{"/var/log/syslog", "/var/log/messages"} {
				lines := tailFile(path, 500)
				if len(lines) > 0 {
					var logLines []LogLine
					for _, l := range lines {
						logLines = append(logLines, LogLine{Content: l, Level: detectLevel(l)})
					}
					return logFetchMsg{
						view:  ViewSystem,
						lines: logLines,
						note:  fmt.Sprintf("Loaded %d system log entries from %s", len(logLines), path),
					}
				}
			}
			return logFetchMsg{
				view: ViewSystem,
				note: fmt.Sprintf("Error: %v (journalctl not available, no syslog found)", err),
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

		// Linuxbrew paths + standard /var/log
		logDirs := []string{
			"/home/linuxbrew/.linuxbrew/var/log",
			"/var/log",
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

		if len(allLines) == 0 {
			allLines = append(allLines, LogLine{
				Content: "No service logs found in /var/log/ or Linuxbrew paths.",
				Level:   LevelInfo,
			})
		}

		return logFetchMsg{
			view:  ViewHomebrew,
			lines: allLines,
			note:  fmt.Sprintf("Loaded %d log entries", len(allLines)),
		}
	}
}
