//go:build darwin

package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func fetchPlatformInfo(info *SystemInfo) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if output, err := exec.CommandContext(ctx, "sw_vers", "-productVersion").Output(); err == nil {
		info.OSVersion = strings.TrimSpace(string(output))
	}
	if output, err := exec.CommandContext(ctx, "sw_vers", "-buildVersion").Output(); err == nil {
		info.BuildNumber = strings.TrimSpace(string(output))
	}
	if output, err := exec.CommandContext(ctx, "sysctl", "-n", "hw.model").Output(); err == nil {
		info.Model = strings.TrimSpace(string(output))
	}
	if output, err := exec.CommandContext(ctx, "sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
		brand := strings.TrimSpace(string(output))
		if strings.Contains(brand, "Apple") {
			info.Chip = brand
		} else {
			info.Chip = brand
		}
	}

	if output, err := exec.CommandContext(ctx, "pmset", "-g", "batt").Output(); err == nil {
		batteryStr := string(output)
		if strings.Contains(batteryStr, "%") {
			parts := strings.Split(batteryStr, "%")
			if len(parts) > 0 {
				for i := len(parts[0]) - 1; i >= 0; i-- {
					if parts[0][i] < '0' || parts[0][i] > '9' {
						if i < len(parts[0])-1 {
							fmt.Sscanf(parts[0][i+1:], "%d", &info.BatteryLevel)
						}
						break
					}
				}
			}
		}
		info.PowerAdapter = strings.Contains(batteryStr, "AC Power")
		if strings.Contains(batteryStr, "Normal") {
			info.BatteryHealth = "Normal"
		} else if strings.Contains(batteryStr, "Service") {
			info.BatteryHealth = "Service Recommended"
		} else {
			info.BatteryHealth = "Good"
		}
	}

	if output, err := exec.CommandContext(ctx, "system_profiler", "SPPowerDataType").Output(); err == nil {
		for _, line := range strings.Split(string(output), "\n") {
			if strings.Contains(line, "Cycle Count:") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &info.BatteryCycles)
				}
			}
		}
	}
}

func (m *Model) runDiskUtility() tea.Cmd {
	return func() tea.Msg {
		exec.Command("open", "-a", "Disk Utility").Run()
		return nil
	}
}

func (m *Model) showSMCResetInstructions() tea.Cmd {
	return func() tea.Msg { return nil }
}

func (m *Model) showNVRAMResetInstructions() tea.Cmd {
	return func() tea.Msg { return nil }
}

func platformMemoryType() string {
	return "Unified Memory"
}

func platformUpdateStatus() string {
	return "Check System Settings"
}

func platformArchLabel(arch string) string {
	if strings.Contains(arch, "arm64") {
		return "Apple Silicon (arm64)"
	}
	return arch
}
