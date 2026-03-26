//go:build linux

package system

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func fetchPlatformInfo(info *SystemInfo) {
	// OS version from /etc/os-release
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				info.OSVersion = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
			}
			if strings.HasPrefix(line, "VERSION_ID=") {
				info.BuildNumber = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
			}
		}
	}

	// Hardware model
	if data, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_name"); err == nil {
		info.Model = strings.TrimSpace(string(data))
	}

	// CPU brand
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "model name") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 {
					info.Chip = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	// Battery via /sys/class/power_supply
	batPath := "/sys/class/power_supply/BAT0"
	if _, err := os.Stat(batPath); err == nil {
		if data, err := os.ReadFile(batPath + "/capacity"); err == nil {
			fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &info.BatteryLevel)
		}
		if data, err := os.ReadFile(batPath + "/status"); err == nil {
			status := strings.TrimSpace(string(data))
			info.PowerAdapter = status == "Charging" || status == "Full"
			info.BatteryHealth = "Good"
		}
		if data, err := os.ReadFile(batPath + "/cycle_count"); err == nil {
			fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &info.BatteryCycles)
		}
	}
}

func (m *Model) runDiskUtility() tea.Cmd {
	return func() tea.Msg {
		// Try common Linux disk tools
		for _, tool := range []string{"gnome-disks", "gparted", "kde-disks"} {
			if path, err := exec.LookPath(tool); err == nil {
				exec.Command(path).Start()
				break
			}
		}
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
	return "System Memory"
}

func platformUpdateStatus() string {
	return "Check package manager"
}

func platformArchLabel(arch string) string {
	if strings.Contains(arch, "amd64") {
		return "x86_64 (amd64)"
	}
	if strings.Contains(arch, "arm64") {
		return "ARM64 (aarch64)"
	}
	return arch
}
