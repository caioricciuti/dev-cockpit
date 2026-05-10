package system

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/caioricciuti/dev-cockpit/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemInfo holds all system information
type SystemInfo struct {
	// Hardware
	Model        string
	Chip         string
	CPUCores     int
	MemoryGB     int
	Architecture string

	// OS Info
	OSVersion   string
	BuildNumber string
	Hostname    string
	Uptime      time.Duration
	BootTime    time.Time

	// Storage
	DiskUsagePercent float64
	DiskFree         uint64
	DiskTotal        uint64

	// Performance
	CPUUsage       float64
	MemoryUsage    float64
	CPUTemperature float64
	FanSpeed       int

	// Battery
	BatteryLevel  int
	BatteryCycles int
	BatteryHealth string
	PowerAdapter  bool
}

// Model represents the system module state
type Model struct {
	config     *config.Config
	width      int
	height     int
	info       SystemInfo
	loading    bool
	activeTab  int
	tabs       []string
	lastUpdate time.Time
}

// New creates a new system module
func New(cfg *config.Config) *Model {
	return &Model{
		config:  cfg,
		tabs:    []string{"Overview", "Hardware", "Performance", "Maintenance"},
		loading: true,
	}
}

// Init initializes the module
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchSystemInfo(),
		m.tickCmd(),
	)
}

func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (interface{}, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "l":
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
		case "shift+tab", "h":
			m.activeTab--
			if m.activeTab < 0 {
				m.activeTab = len(m.tabs) - 1
			}
		case "r":
			return m, m.fetchSystemInfo()
		case "1":
			m.activeTab = 0
		case "2":
			m.activeTab = 1
		case "3":
			m.activeTab = 2
		case "4":
			m.activeTab = 3

		// Quick actions based on tab
		case "d":
			if m.activeTab == 3 { // Maintenance tab
				return m, m.runDiskUtility()
			}
		case "s":
			if m.activeTab == 3 { // Maintenance tab
				return m, m.showSMCResetInstructions()
			}
		case "n":
			if m.activeTab == 3 { // Maintenance tab
				return m, m.showNVRAMResetInstructions()
			}
		}

	case systemInfoMsg:
		m.info = msg.info
		m.loading = false
		m.lastUpdate = time.Now()

	case tickMsg:
		return m, m.fetchSystemInfo()
	}

	return m, nil
}

// View renders the module
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	if m.loading {
		return m.renderLoadingScreen()
	}

	// Render header with tabs
	header := m.renderHeader()

	// Calculate available height for content
	headerHeight := lipgloss.Height(header)
	footerHeight := 3
	contentHeight := m.height - headerHeight - footerHeight - 2

	if contentHeight < 10 {
		contentHeight = 10 // Minimum height
	}

	// Render content based on active tab
	var content string
	switch m.activeTab {
	case 0:
		content = m.renderOverview()
	case 1:
		content = m.renderHardware()
	case 2:
		content = m.renderPerformance()
	case 3:
		content = m.renderMaintenance()
	}

	// Apply viewport to prevent overflow
	content = lipgloss.NewStyle().MaxHeight(contentHeight).Render(content)

	footer := m.renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		content,
		footer,
	)
}

func (m *Model) renderHeader() string {
	tabStyle := lipgloss.NewStyle().
		Padding(0, 2)

	activeTabStyle := tabStyle.
		Bold(true).
		Foreground(lipgloss.Color("#000")).
		Background(lipgloss.Color("#00D9FF"))

	inactiveTabStyle := tabStyle.
		Foreground(lipgloss.Color("#888"))

	var tabs []string
	for i, tab := range m.tabs {
		if i == m.activeTab {
			tabs = append(tabs, activeTabStyle.Render(tab))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(tab))
		}
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF")).
		MarginBottom(1)

	// Use safe width for separator (accounting for margins)
	separatorWidth := m.width - 4
	if separatorWidth < 40 {
		separatorWidth = 40
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		titleStyle.Render("🖥️  SYSTEM INFORMATION"),
		tabBar,
		strings.Repeat("─", separatorWidth),
	)
}

func (m *Model) renderFooter() string {
	if m.width == 0 {
		return ""
	}

	help := []string{
		"1-4: Switch Views",
		"Tab/Shift+Tab: Cycle Views",
		"R: Refresh Snapshot",
		"D: Disk First Aid",
		"S: SMC Guide",
		"N: NVRAM Guide",
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		MarginTop(1).
		Render(strings.Join(help, "  |  "))
}

func (m *Model) renderOverview() string {
	style := lipgloss.NewStyle().Padding(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Width(20)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFF"))

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	content := strings.Builder{}

	// System Overview
	content.WriteString(highlightStyle.Render("System") + "\n")
	content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Model:"), valueStyle.Render(m.info.Model)))
	content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Chip:"), valueStyle.Render(m.info.Chip)))
	osLabel := "OS:"
	osValue := m.info.OSVersion
	if m.info.BuildNumber != "" {
		osValue += " (" + m.info.BuildNumber + ")"
	}
	content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render(osLabel), valueStyle.Render(osValue)))
	content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Hostname:"), valueStyle.Render(m.info.Hostname)))
	content.WriteString(fmt.Sprintf("%s %s\n\n", labelStyle.Render("Uptime:"), valueStyle.Render(formatDuration(m.info.Uptime))))

	// Resources
	content.WriteString(highlightStyle.Render("Resources") + "\n")
	content.WriteString(fmt.Sprintf("%s %.1f%%\n", labelStyle.Render("CPU Usage:"), m.info.CPUUsage))
	content.WriteString(fmt.Sprintf("%s %.1f%% (%.1f GB / %d GB)\n",
		labelStyle.Render("Memory:"),
		m.info.MemoryUsage,
		float64(m.info.MemoryGB)*m.info.MemoryUsage/100,
		m.info.MemoryGB))
	content.WriteString(fmt.Sprintf("%s %.1f%% (%.1f GB free)\n\n",
		labelStyle.Render("Disk:"),
		m.info.DiskUsagePercent,
		float64(m.info.DiskFree)/1024/1024/1024))

	// Battery (if applicable)
	if m.info.BatteryLevel > 0 {
		content.WriteString(highlightStyle.Render("Battery") + "\n")
		content.WriteString(fmt.Sprintf("%s %d%%\n", labelStyle.Render("Level:"), m.info.BatteryLevel))
		content.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("Cycles:"), m.info.BatteryCycles))
		content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Health:"), valueStyle.Render(m.info.BatteryHealth)))

		adapterStatus := "Not Connected"
		if m.info.PowerAdapter {
			adapterStatus = "Connected"
		}
		content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Power Adapter:"), valueStyle.Render(adapterStatus)))
	}

	// Last update
	content.WriteString(fmt.Sprintf("\n%s %s",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render("Last updated:"),
		m.lastUpdate.Format("15:04:05")))

	return style.Render(content.String())
}

func (m *Model) renderHardware() string {
	style := lipgloss.NewStyle().Padding(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Width(20)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFF"))

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	content := strings.Builder{}

	content.WriteString(highlightStyle.Render("Hardware Information") + "\n\n")

	// Processor
	content.WriteString(highlightStyle.Render("Processor") + "\n")
	content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Chip:"), valueStyle.Render(m.info.Chip)))
	content.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("CPU Cores:"), m.info.CPUCores))
	content.WriteString(fmt.Sprintf("%s %s\n\n", labelStyle.Render("Architecture:"), valueStyle.Render(m.info.Architecture)))

	// Memory
	content.WriteString(highlightStyle.Render("Memory") + "\n")
	content.WriteString(fmt.Sprintf("%s %d GB\n", labelStyle.Render("Total:"), m.info.MemoryGB))
	memType := platformMemoryType()
	content.WriteString(fmt.Sprintf("%s %s\n\n", labelStyle.Render("Type:"), memType))

	// Storage
	content.WriteString(highlightStyle.Render("Storage") + "\n")
	content.WriteString(fmt.Sprintf("%s %.1f GB\n", labelStyle.Render("Total:"), float64(m.info.DiskTotal)/1024/1024/1024))
	content.WriteString(fmt.Sprintf("%s %.1f GB\n", labelStyle.Render("Available:"), float64(m.info.DiskFree)/1024/1024/1024))
	content.WriteString(fmt.Sprintf("%s %.1f%%\n\n", labelStyle.Render("Used:"), m.info.DiskUsagePercent))

	// Thermal
	if m.info.CPUTemperature > 0 {
		content.WriteString(highlightStyle.Render("Thermal") + "\n")
		content.WriteString(fmt.Sprintf("%s %.1f°C\n", labelStyle.Render("CPU Temperature:"), m.info.CPUTemperature))
		if m.info.FanSpeed > 0 {
			content.WriteString(fmt.Sprintf("%s %d RPM\n", labelStyle.Render("Fan Speed:"), m.info.FanSpeed))
		}
	}

	return style.Render(content.String())
}

func (m *Model) renderPerformance() string {
	style := lipgloss.NewStyle().Padding(1)

	content := strings.Builder{}

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	// CPU Usage bar
	content.WriteString(highlightStyle.Render("CPU Usage") + "\n")
	content.WriteString(m.renderProgressBar(40, m.info.CPUUsage/100))
	content.WriteString(fmt.Sprintf(" %.1f%%\n\n", m.info.CPUUsage))

	// Memory Usage bar
	content.WriteString(highlightStyle.Render("Memory Usage") + "\n")
	content.WriteString(m.renderProgressBar(40, m.info.MemoryUsage/100))
	content.WriteString(fmt.Sprintf(" %.1f%% (%.1f GB / %d GB)\n\n",
		m.info.MemoryUsage,
		float64(m.info.MemoryGB)*m.info.MemoryUsage/100,
		m.info.MemoryGB))

	// Disk Usage bar
	content.WriteString(highlightStyle.Render("Disk Usage") + "\n")
	content.WriteString(m.renderProgressBar(40, m.info.DiskUsagePercent/100))
	content.WriteString(fmt.Sprintf(" %.1f%%\n\n", m.info.DiskUsagePercent))

	// Performance Tips
	content.WriteString(highlightStyle.Render("Performance Tips") + "\n")
	tips := []string{}

	if m.info.CPUUsage > 80 {
		tips = append(tips, "⚠️  High CPU usage detected - check Activity Monitor")
	}
	if m.info.MemoryUsage > 85 {
		tips = append(tips, "⚠️  High memory usage - consider closing unused apps")
	}
	if m.info.DiskUsagePercent > 90 {
		tips = append(tips, "⚠️  Low disk space - run cleanup tools")
	}
	if m.info.CPUTemperature > 80 {
		tips = append(tips, "⚠️  High CPU temperature - check ventilation")
	}

	if len(tips) == 0 {
		tips = append(tips, "✅ System performance is good")
	}

	for _, tip := range tips {
		content.WriteString(fmt.Sprintf("  %s\n", tip))
	}

	return style.Render(content.String())
}

func (m *Model) renderMaintenance() string {
	style := lipgloss.NewStyle().Padding(1)

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	actionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500"))

	content := strings.Builder{}

	content.WriteString(highlightStyle.Render("System Maintenance") + "\n\n")

	// Quick Actions
	content.WriteString(highlightStyle.Render("Quick Actions") + "\n")
	content.WriteString(actionStyle.Render("[D]") + " Run Disk Utility First Aid\n")
	content.WriteString(actionStyle.Render("[S]") + " SMC Reset Instructions\n")
	content.WriteString(actionStyle.Render("[N]") + " NVRAM Reset Instructions\n")
	content.WriteString(actionStyle.Render("[R]") + " Refresh System Info\n\n")

	// Maintenance Tasks
	content.WriteString(highlightStyle.Render("Recommended Maintenance") + "\n")

	tasks := []struct {
		task   string
		status string
		color  string
	}{
		{"OS Updates", platformUpdateStatus(), "#FFA500"},
		{"Disk Verification", "Press [D] to run", "#888"},
		{"Storage Optimization", m.getStorageStatus(), m.getStorageStatusColor()},
		{"Battery Health", m.info.BatteryHealth, "#0F0"},
	}

	for _, task := range tasks {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(task.color))
		content.WriteString(fmt.Sprintf("  • %-25s %s\n", task.task, statusStyle.Render(task.status)))
	}

	content.WriteString("\n" + highlightStyle.Render("System Integrity") + "\n")

	// Show boot time
	content.WriteString(fmt.Sprintf("  Boot Time: %s\n", m.info.BootTime.Format("Jan 2, 15:04")))
	content.WriteString(fmt.Sprintf("  Uptime: %s\n", formatDuration(m.info.Uptime)))

	// Show architecture info
	content.WriteString(fmt.Sprintf("  Architecture: %s\n", platformArchLabel(m.info.Architecture)))

	return style.Render(content.String())
}

func (m *Model) renderLoadingScreen() string {
	loadingStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return loadingStyle.Render("⏳ Gathering system information...")
}

func (m *Model) renderProgressBar(width int, percent float64) string {
	if percent > 1 {
		percent = 1
	}
	if percent < 0 {
		percent = 0
	}

	filled := int(float64(width) * percent)
	empty := width - filled

	// Color based on percentage
	var color string
	if percent < 0.5 {
		color = "#0F0" // Green
	} else if percent < 0.8 {
		color = "#FFA500" // Orange
	} else {
		color = "#F00" // Red
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render(fmt.Sprintf("[%s]", bar))
}

// Title returns the module title
func (m *Model) Title() string {
	return "System"
}

// HasOpenModal returns true if the module has an open modal/dialog
func (m *Model) HasOpenModal() bool {
	return false
}

// fetchSystemInfo gathers cross-platform system info via gopsutil,
// then delegates to fetchPlatformInfo() for OS-specific data.
func (m *Model) fetchSystemInfo() tea.Cmd {
	return func() tea.Msg {
		info := SystemInfo{}

		info.Architecture = runtime.GOARCH
		info.CPUCores = runtime.NumCPU()

		if vmStat, err := mem.VirtualMemory(); err == nil {
			info.MemoryGB = int(vmStat.Total / 1024 / 1024 / 1024)
			info.MemoryUsage = vmStat.UsedPercent
		}

		if hostname, err := os.Hostname(); err == nil {
			info.Hostname = hostname
		}

		if hostInfo, err := host.Info(); err == nil {
			info.Uptime = time.Duration(hostInfo.Uptime) * time.Second
			info.BootTime = time.Unix(int64(hostInfo.BootTime), 0)
		}

		if cpuPercent, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercent) > 0 {
			info.CPUUsage = cpuPercent[0]
		}

		if diskStat, err := disk.Usage("/"); err == nil {
			info.DiskUsagePercent = diskStat.UsedPercent
			info.DiskFree = diskStat.Free
			info.DiskTotal = diskStat.Total
		}

		// Platform-specific: OS version, hardware model, chip, battery
		fetchPlatformInfo(&info)

		return systemInfoMsg{info: info}
	}
}

func (m *Model) getStorageStatus() string {
	if m.info.DiskUsagePercent > 90 {
		return "Critical - Clean up needed"
	} else if m.info.DiskUsagePercent > 80 {
		return "Warning - Consider cleanup"
	}
	return "Good"
}

func (m *Model) getStorageStatusColor() string {
	if m.info.DiskUsagePercent > 90 {
		return "#F00"
	} else if m.info.DiskUsagePercent > 80 {
		return "#FFA500"
	}
	return "#0F0"
}

// runDiskUtility, showSMCResetInstructions, showNVRAMResetInstructions,
// fetchPlatformInfo, platformMemoryType, platformUpdateStatus, platformArchLabel
// are implemented in platform-specific files (platform_darwin.go / platform_linux.go)

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// Message types
type tickMsg time.Time

type systemInfoMsg struct {
	info SystemInfo
}
