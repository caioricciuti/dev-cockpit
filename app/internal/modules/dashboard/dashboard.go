package dashboard

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"time"

	"github.com/caioricciuti/dev-cockpit/internal/config"
	"github.com/caioricciuti/dev-cockpit/internal/storage"
	"github.com/caioricciuti/dev-cockpit/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// ViewMode for switching between live and historical views
type ViewMode int

const (
	ViewLive    ViewMode = iota
	ViewHist1h
	ViewHist6h
	ViewHist24h
	ViewHist7d
)

// Model represents the dashboard state
type Model struct {
	config *config.Config
	width  int
	height int

	// System metrics
	cpuPercent    []float64
	cpuHistory    []float64
	memoryPercent float64
	memoryHistory []float64
	diskUsage     float64
	diskHistory   []float64

	// Network metrics
	netStats     []net.IOCountersStat
	netInRate    float64
	netOutRate   float64
	prevNetStats *net.IOCountersStat

	// System info
	hostname   string
	platform   string
	uptime     time.Duration
	numCPU     int
	totalMem   uint64
	lastUpdate time.Time

	// UI state
	selectedMetric int
	showDetails    bool

	// Historical views
	store      *storage.Store
	activeView ViewMode
	histData   []storage.MetricSnapshot
	storeCount int // counter for 10s store tick
}

// New creates a new dashboard module
func New(cfg *config.Config, store *storage.Store) *Model {
	m := &Model{
		config:         cfg,
		cpuHistory:     make([]float64, 60), // 60 seconds of history
		memoryHistory:  make([]float64, 60),
		diskHistory:    make([]float64, 60),
		selectedMetric: 0,
		store:          store,
		activeView:     ViewLive,
	}

	// Initialize system info
	m.updateSystemInfo()

	return m
}

// Init initializes the dashboard
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
		m.fetchMetrics(),
	)
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (interface{}, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			m.activeView = ViewLive
		case "2":
			m.activeView = ViewHist1h
			return m, m.fetchHistorical(time.Hour)
		case "3":
			m.activeView = ViewHist6h
			return m, m.fetchHistorical(6 * time.Hour)
		case "4":
			m.activeView = ViewHist24h
			return m, m.fetchHistorical(24 * time.Hour)
		case "5":
			m.activeView = ViewHist7d
			return m, m.fetchHistorical(7 * 24 * time.Hour)
		case "up", "k":
			m.selectedMetric--
			if m.selectedMetric < 0 {
				m.selectedMetric = 3
			}
		case "down", "j":
			m.selectedMetric = (m.selectedMetric + 1) % 4
		case "enter", " ":
			m.showDetails = !m.showDetails
		case "r":
			if m.activeView != ViewLive {
				return m, m.fetchHistorical(m.viewDuration())
			}
			return m, m.fetchMetrics()
		}

	case metricsMsg:
		m.updateMetrics(msg)
		// Persist metrics to store every 10 ticks (10 seconds)
		m.storeCount++
		if m.storeCount >= 10 && m.store != nil {
			m.storeCount = 0
			avgCPU := 0.0
			for _, c := range m.cpuPercent {
				avgCPU += c
			}
			if len(m.cpuPercent) > 0 {
				avgCPU /= float64(len(m.cpuPercent))
			}
			snap := storage.MetricSnapshot{
				Timestamp:  time.Now(),
				CPUAvg:     avgCPU,
				MemoryPct:  m.memoryPercent,
				DiskPct:    m.diskUsage,
				NetInRate:  m.netInRate,
				NetOutRate: m.netOutRate,
			}
			go m.store.Insert(snap)
		}
		return m, tea.Batch(m.tickCmd(), m.fetchMetrics())

	case historicalMsg:
		m.histData = msg.data

	case tickMsg:
		// Smooth scrolling for graphs
		return m, m.tickCmd()
	}

	return m, nil
}

// View renders the dashboard
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Use Layout system to calculate available space
	layout := components.NewLayout(m.width, m.height)

	// View tabs
	viewTabs := m.renderViewTabs()

	var sections []string
	if m.activeView == ViewLive {
		sections = []string{
			viewTabs,
			"",
			m.renderSystemInfo(),
			"",
			m.renderMetrics(),
			"",
			m.renderAdvancedMetrics(),
		}
	} else {
		sections = []string{
			viewTabs,
			"",
			m.renderHistoricalView(),
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Apply viewport with proper height constraint to prevent overflow
	return components.Viewport(content, layout.ContentHeight)
}

func (m *Model) renderSystemInfo() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF")).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Width(12)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFF"))

	// Build info lines vertically with proper spacing
	infoLines := []string{
		headerStyle.Render("📊 SYSTEM DASHBOARD"),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("Hostname:"), valueStyle.Render(m.hostname)),
		fmt.Sprintf("%s %s", labelStyle.Render("Platform:"), valueStyle.Render(m.platform)),
		fmt.Sprintf("%s %d cores", labelStyle.Render("CPUs:"), m.numCPU),
		fmt.Sprintf("%s %.1f GB", labelStyle.Render("Memory:"), float64(m.totalMem)/1024/1024/1024),
		fmt.Sprintf("%s %s", labelStyle.Render("Uptime:"), valueStyle.Render(m.formatUptime())),
		"",
	}

	return lipgloss.JoinVertical(lipgloss.Left, infoLines...)
}

func (m *Model) renderMetrics() string {
	// Calculate average CPU
	avgCPU := 0.0
	for _, cpu := range m.cpuPercent {
		avgCPU += cpu
	}
	if len(m.cpuPercent) > 0 {
		avgCPU /= float64(len(m.cpuPercent))
	}

	// Separator line
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#444"))
	separator := separatorStyle.Render(strings.Repeat("━", 60))

	// Label styles
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFF"))

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0FD976"))

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500"))

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B"))

	// Build metrics lines
	lines := []string{
		separator,
		"",
	}

	// CPU Metric
	cpuStatus := "● Normal"
	cpuStatusStyle := statusStyle
	if avgCPU >= 85 {
		cpuStatus = "● Critical"
		cpuStatusStyle = errorStyle
	} else if avgCPU >= 70 {
		cpuStatus = "● High"
		cpuStatusStyle = warningStyle
	}
	lines = append(lines,
		labelStyle.Render("⚡ CPU: ")+valueStyle.Render(fmt.Sprintf("%.1f%%", avgCPU)),
		m.renderProgressBar(avgCPU)+" "+cpuStatusStyle.Render(cpuStatus),
		"",
	)

	// Memory Metric
	memStatus := "● Healthy"
	memStatusStyle := statusStyle
	if m.memoryPercent >= 90 {
		memStatus = "● Critical"
		memStatusStyle = errorStyle
	} else if m.memoryPercent >= 75 {
		memStatus = "● High"
		memStatusStyle = warningStyle
	}
	lines = append(lines,
		labelStyle.Render("💾 Memory: ")+valueStyle.Render(fmt.Sprintf("%.1f%%", m.memoryPercent)),
		m.renderProgressBar(m.memoryPercent)+" "+memStatusStyle.Render(memStatus),
		"",
	)

	// Disk Metric
	diskStatus := "● Healthy"
	diskStatusStyle := statusStyle
	if m.diskUsage >= 90 {
		diskStatus = "● Critical"
		diskStatusStyle = errorStyle
	} else if m.diskUsage >= 80 {
		diskStatus = "● Low Space"
		diskStatusStyle = warningStyle
	}
	lines = append(lines,
		labelStyle.Render("💿 Disk: ")+valueStyle.Render(fmt.Sprintf("%.1f%%", m.diskUsage)),
		m.renderProgressBar(m.diskUsage)+" "+diskStatusStyle.Render(diskStatus),
		"",
	)

	// Network Metric
	totalRate := (m.netInRate + m.netOutRate) / 1024 / 1024
	netStatus := "Idle"
	if totalRate > 10 {
		netStatus = "High Activity"
	} else if totalRate > 1 {
		netStatus = "Active"
	} else if totalRate > 0.1 {
		netStatus = "Light"
	}

	netSubStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888"))
	lines = append(lines,
		labelStyle.Render("🌐 Network: ")+valueStyle.Render(netStatus),
		"  "+netSubStyle.Render(fmt.Sprintf("▼ Down: %.1f KB/s", m.netInRate/1024)),
		"  "+netSubStyle.Render(fmt.Sprintf("▲ Up: %.1f KB/s", m.netOutRate/1024)),
		"",
	)

	lines = append(lines, separator)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderProgressBar creates a simple ASCII progress bar
func (m *Model) renderProgressBar(percent float64) string {
	barWidth := 30
	filled := int(math.Round(percent / 100 * float64(barWidth)))
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}

	// Choose color based on percentage
	barColor := "#0FD976" // Green
	if percent >= 85 {
		barColor = "#FF6B6B" // Red
	} else if percent >= 70 {
		barColor = "#FFA500" // Orange
	}

	filledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#333"))

	bar := "["
	bar += filledStyle.Render(strings.Repeat("█", filled))
	bar += emptyStyle.Render(strings.Repeat("░", barWidth-filled))
	bar += "]"

	return bar
}

func (m *Model) renderAdvancedMetrics() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF")).
		Padding(0, 1)

	insightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDD"))

	header := headerStyle.Render("💡 System Insights")

	insights, score := m.generateAdvancedInsights()

	// Format score with color
	scoreColor := "#0FD976"
	if score < 50 {
		scoreColor = "#FF6B6B"
	} else if score < 70 {
		scoreColor = "#FFA500"
	}
	scoreText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(scoreColor)).
		Bold(true).
		Render(fmt.Sprintf("Performance Score: %d/100", score))

	// Build insight lines
	insightLines := []string{
		"",
		header,
		"",
		scoreText,
		"",
	}

	for _, insight := range insights {
		insightLines = append(insightLines, insightStyle.Render(insight))
	}

	return lipgloss.JoinVertical(lipgloss.Left, insightLines...)
}

func (m *Model) updateSystemInfo() {
	info, _ := host.Info()
	if info != nil {
		m.hostname = info.Hostname
	}
	m.platform = runtime.GOOS
	m.numCPU = runtime.NumCPU()

	if v, err := mem.VirtualMemory(); err == nil {
		m.totalMem = v.Total
	}

	if uptime, err := host.Uptime(); err == nil {
		m.uptime = time.Duration(uptime) * time.Second
	}
}

func (m *Model) updateMetrics(msg metricsMsg) {
	// Update CPU
	m.cpuPercent = msg.cpu
	avgCPU := 0.0
	for _, cpu := range m.cpuPercent {
		avgCPU += cpu
	}
	if len(m.cpuPercent) > 0 {
		avgCPU /= float64(len(m.cpuPercent))
	}
	m.cpuHistory = append(m.cpuHistory[1:], avgCPU)

	// Update Memory
	m.memoryPercent = msg.memory
	m.memoryHistory = append(m.memoryHistory[1:], m.memoryPercent)

	// Update Disk
	m.diskUsage = msg.disk
	m.diskHistory = append(m.diskHistory[1:], m.diskUsage)

	// Update Network
	m.netStats = msg.network
	if len(msg.network) > 0 && m.prevNetStats != nil {
		// Calculate rate (bytes per second)
		timeDiff := time.Since(m.lastUpdate).Seconds()
		if timeDiff > 0 {
			m.netInRate = float64(msg.network[0].BytesRecv-m.prevNetStats.BytesRecv) / timeDiff
			m.netOutRate = float64(msg.network[0].BytesSent-m.prevNetStats.BytesSent) / timeDiff
		}
	}
	if len(msg.network) > 0 {
		m.prevNetStats = &msg.network[0]
	}

	m.lastUpdate = time.Now()
}

func (m *Model) formatUptime() string {
	hours := int(m.uptime.Hours())
	days := hours / 24
	hours = hours % 24

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	return fmt.Sprintf("%dh %dm", hours, int(m.uptime.Minutes())%60)
}

func (m *Model) generateAdvancedInsights() ([]string, int) {
	cpuForecast, cpuSlope := m.forecastUsage(m.cpuHistory, 60)
	memForecast, memSlope := m.forecastUsage(m.memoryHistory, 60)
	diskLevel := m.diskUsage
	cpuTrend := describeTrend(cpuSlope)
	memTrend := describeTrend(memSlope)
	cpuVolatility := calculateVolatility(m.cpuHistory)
	memVolatility := calculateVolatility(m.memoryHistory)
	avgVolatility := (cpuVolatility + memVolatility) / 2
	cpuSaturation := timeToThreshold(m.cpuHistory, 85)
	memSaturation := timeToThreshold(m.memoryHistory, 90)

	riskLabel, riskReason := operationalRisk(cpuForecast, memForecast, diskLevel, avgVolatility)
	recommendations := recommendActions(riskLabel, cpuForecast, memForecast, memSaturation, cpuSaturation)

	insights := []string{
		fmt.Sprintf("• CPU 60s forecast: %.1f%% (%s, σ=%.1f)", cpuForecast, cpuTrend, cpuVolatility),
		fmt.Sprintf("• Memory 60s forecast: %.1f%% (%s, σ=%.1f)", memForecast, memTrend, memVolatility),
	}

	if cpuSaturation > 0 {
		insights = append(insights, fmt.Sprintf("• CPU headroom: ~%s to reach 85%% load", formatShortDuration(cpuSaturation)))
	} else {
		insights = append(insights, "• CPU headroom: Stable - no saturation trend detected")
	}
	if memSaturation > 0 {
		insights = append(insights, fmt.Sprintf("• Memory headroom: ~%s to reach 90%% load", formatShortDuration(memSaturation)))
	} else {
		insights = append(insights, "• Memory headroom: Stable - ample capacity available")
	}

	insights = append(insights, fmt.Sprintf("• Operational risk: %s - %s", riskLabel, riskReason))

	if len(recommendations) > 0 {
		insights = append(insights, fmt.Sprintf("• Recommended action: %s", recommendations[0]))
	}
	if len(recommendations) > 1 {
		insights = append(insights, fmt.Sprintf("• Next step: %s", recommendations[1]))
	}

	score := calculatePerformanceScore(cpuForecast, memForecast, diskLevel, avgVolatility)
	return insights, score
}

func (m *Model) forecastUsage(history []float64, horizonSeconds int) (float64, float64) {
	if len(history) == 0 {
		return 0, 0
	}

	window := len(history)
	if window > 45 {
		window = 45
	}
	recent := history[len(history)-window:]
	slope, intercept := linearRegression(recent)
	forecastIndex := float64(window - 1 + horizonSeconds)
	predicted := intercept + slope*forecastIndex
	return clamp(predicted, 0, 100), slope
}

func linearRegression(values []float64) (float64, float64) {
	n := len(values)
	if n == 0 {
		return 0, 0
	}

	var sumX, sumY, sumXX, sumXY float64
	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXX += x * x
		sumXY += x * y
	}

	denominator := float64(n)*sumXX - sumX*sumX
	if denominator == 0 {
		return 0, values[n-1]
	}

	slope := (float64(n)*sumXY - sumX*sumY) / denominator
	intercept := (sumY - slope*sumX) / float64(n)
	return slope, intercept
}

func describeTrend(slope float64) string {
	perMinute := slope * 60
	switch {
	case perMinute > 8:
		return "surging"
	case perMinute > 3:
		return "rising"
	case perMinute < -8:
		return "dropping"
	case perMinute < -3:
		return "cooling"
	default:
		return "stable"
	}
}

func calculateVolatility(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var mean float64
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	var variance float64
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	return math.Sqrt(variance)
}

func timeToThreshold(history []float64, threshold float64) time.Duration {
	if len(history) < 2 {
		return 0
	}

	window := len(history)
	if window > 45 {
		window = 45
	}
	recent := history[len(history)-window:]
	slope, _ := linearRegression(recent)
	current := recent[len(recent)-1]

	if slope <= 0 || current >= threshold {
		return 0
	}

	seconds := (threshold - current) / slope
	if seconds <= 0 || math.IsInf(seconds, 0) || math.IsNaN(seconds) {
		return 0
	}

	return time.Duration(seconds) * time.Second
}

func operationalRisk(cpuForecast, memForecast, diskUsage, volatility float64) (string, string) {
	riskScore := (cpuForecast/100)*0.35 + (memForecast/100)*0.35 + (diskUsage/100)*0.2 + (clamp(volatility, 0, 25)/25)*0.1

	reasons := []string{}
	if cpuForecast >= 85 {
		reasons = append(reasons, "CPU forecast above 85%")
	} else if cpuForecast >= 70 {
		reasons = append(reasons, "CPU trend rising")
	}
	if memForecast >= 85 {
		reasons = append(reasons, "Memory headroom tight")
	} else if memForecast >= 70 {
		reasons = append(reasons, "Memory pressure increasing")
	}
	if diskUsage >= 90 {
		reasons = append(reasons, "Disk nearly full")
	}
	if volatility >= 12 {
		reasons = append(reasons, "Workload volatility elevated")
	}

	var label string
	switch {
	case riskScore >= 0.75:
		label = "Critical"
	case riskScore >= 0.55:
		label = "Elevated"
	default:
		label = "Nominal"
	}

	reason := "Within optimal operating bounds"
	if len(reasons) > 0 {
		reason = strings.Join(reasons, "; ")
	}

	return label, reason
}

func recommendActions(risk string, cpuForecast, memForecast float64, memSaturation, cpuSaturation time.Duration) []string {
	actions := []string{}
	switch risk {
	case "Critical":
		actions = append(actions, "Close runaway processes (Activity Monitor > CPU) and pause heavy containers")
		if memForecast >= 85 || memSaturation > 0 {
			actions = append(actions, "Free ≥5GB of memory or restart memory-intensive apps to avoid swapping")
		} else {
			actions = append(actions, "Use Quick Actions › Deep Clean to reclaim disk and cache headroom")
		}
	case "Elevated":
		actions = append(actions, "Monitor top workloads and enable Low Power Mode during spikes")
		if memSaturation > 0 {
			actions = append(actions, "Schedule a restart or memory purge within the hour")
		} else if cpuSaturation > 0 {
			actions = append(actions, "Consider scaling background jobs to another machine or cloud runner")
		} else {
			actions = append(actions, "Archive large build artefacts to keep disk usage below 80%")
		}
	default:
		actions = append(actions, "All indicators nominal - capture this snapshot as a baseline")
		if cpuForecast > 60 {
			actions = append(actions, "Resource trend rising: plan capacity for upcoming workloads")
		}
	}
	return actions
}

func formatShortDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}

	seconds := int(math.Round(d.Seconds()))
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	seconds = seconds % 60
	if minutes < 60 {
		if seconds == 0 {
			return fmt.Sprintf("%dm", minutes)
		}
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	hours := minutes / 60
	minutes = minutes % 60
	if minutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func calculatePerformanceScore(cpuForecast, memForecast, diskUsage, volatility float64) int {
	score := 100

	cpuPenalty := clamp(cpuForecast-50, 0, 50) * 0.6
	memPenalty := clamp(memForecast-55, 0, 45) * 0.7
	diskPenalty := clamp(diskUsage-70, 0, 30) * 0.5
	volatilityPenalty := clamp(volatility, 0, 25) * 1.2

	score -= int(cpuPenalty + memPenalty + diskPenalty + volatilityPenalty)
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

// Title returns the module title
func (m *Model) Title() string {
	return "Dashboard"
}

// HasOpenModal returns true if the module has an open modal/dialog
func (m *Model) HasOpenModal() bool {
	return false
}

// Messages
type metricsMsg struct {
	cpu     []float64
	memory  float64
	disk    float64
	network []net.IOCountersStat
}

type historicalMsg struct {
	data []storage.MetricSnapshot
}

type tickMsg time.Time

func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) fetchMetrics() tea.Cmd {
	return func() tea.Msg {
		// Fetch CPU
		cpuPercent, _ := cpu.Percent(time.Second, true)

		// Fetch Memory
		var memPercent float64
		if memInfo, err := mem.VirtualMemory(); err == nil && memInfo != nil {
			memPercent = memInfo.UsedPercent
		}

		// Fetch Disk
		var diskPercent float64
		if diskInfo, err := disk.Usage("/"); err == nil && diskInfo != nil {
			diskPercent = diskInfo.UsedPercent
		}

		// Fetch Network
		netInfo, _ := net.IOCounters(false)

		return metricsMsg{
			cpu:     cpuPercent,
			memory:  memPercent,
			disk:    diskPercent,
			network: netInfo,
		}
	}
}

func (m *Model) fetchHistorical(since time.Duration) tea.Cmd {
	if m.store == nil {
		return nil
	}
	return func() tea.Msg {
		data, err := m.store.Query(since)
		if err != nil {
			return historicalMsg{}
		}
		return historicalMsg{data: data}
	}
}

func (m *Model) viewDuration() time.Duration {
	switch m.activeView {
	case ViewHist1h:
		return time.Hour
	case ViewHist6h:
		return 6 * time.Hour
	case ViewHist24h:
		return 24 * time.Hour
	case ViewHist7d:
		return 7 * 24 * time.Hour
	default:
		return time.Hour
	}
}

func (m *Model) renderViewTabs() string {
	styles := components.NewBaseStyles()

	labels := []struct {
		key  string
		name string
		view ViewMode
	}{
		{"1", "Live", ViewLive},
		{"2", "1h", ViewHist1h},
		{"3", "6h", ViewHist6h},
		{"4", "24h", ViewHist24h},
		{"5", "7d", ViewHist7d},
	}

	var tabs []string
	for _, l := range labels {
		label := fmt.Sprintf("[%s] %s", l.key, l.name)
		if l.view == m.activeView {
			tabs = append(tabs, lipgloss.NewStyle().
				Foreground(styles.Theme.Primary).Bold(true).
				Render(label))
		} else {
			tabs = append(tabs, lipgloss.NewStyle().
				Foreground(styles.Theme.Muted).
				Render(label))
		}
	}

	return strings.Join(tabs, "  ")
}

func (m *Model) renderHistoricalView() string {
	styles := components.NewBaseStyles()

	viewNames := map[ViewMode]string{
		ViewHist1h:  "Last 1 Hour",
		ViewHist6h:  "Last 6 Hours",
		ViewHist24h: "Last 24 Hours",
		ViewHist7d:  "Last 7 Days",
	}

	title := styles.Title().Render(fmt.Sprintf("📈 %s", viewNames[m.activeView]))

	if m.store == nil {
		return lipgloss.JoinVertical(lipgloss.Left, title, "",
			styles.Muted().Render("Metrics store not available."))
	}

	if len(m.histData) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, title, "",
			styles.Muted().Render("No historical data yet. Metrics are recorded every 10 seconds."),
			styles.Muted().Render("Check back shortly."))
	}

	// Calculate sparkline width
	sparkWidth := 60
	if m.width > 0 {
		sparkWidth = m.width - 30
		if sparkWidth > 100 {
			sparkWidth = 100
		}
		if sparkWidth < 30 {
			sparkWidth = 30
		}
	}

	// Extract metric arrays
	cpuVals := make([]float64, len(m.histData))
	memVals := make([]float64, len(m.histData))
	diskVals := make([]float64, len(m.histData))
	netVals := make([]float64, len(m.histData))
	for i, s := range m.histData {
		cpuVals[i] = s.CPUAvg
		memVals[i] = s.MemoryPct
		diskVals[i] = s.DiskPct
		netVals[i] = (s.NetInRate + s.NetOutRate) / 1024 // KB/s
	}

	labelStyle := lipgloss.NewStyle().Foreground(styles.Theme.Primary).Bold(true).Width(12)
	statStyle := styles.Muted()

	sections := []string{title, ""}

	// Time range info
	first := m.histData[0].Timestamp.Format("15:04:05")
	last := m.histData[len(m.histData)-1].Timestamp.Format("15:04:05")
	sections = append(sections,
		statStyle.Render(fmt.Sprintf("Range: %s — %s  (%d data points)", first, last, len(m.histData))),
		"")

	// CPU sparkline
	sections = append(sections,
		labelStyle.Render("⚡ CPU %"),
		renderSparkline(cpuVals, sparkWidth, 100),
		statStyle.Render(fmt.Sprintf("   min: %.1f%%  avg: %.1f%%  max: %.1f%%", minVal(cpuVals), avgVal(cpuVals), maxVal(cpuVals))),
		"")

	// Memory sparkline
	sections = append(sections,
		labelStyle.Render("💾 Memory %"),
		renderSparkline(memVals, sparkWidth, 100),
		statStyle.Render(fmt.Sprintf("   min: %.1f%%  avg: %.1f%%  max: %.1f%%", minVal(memVals), avgVal(memVals), maxVal(memVals))),
		"")

	// Disk sparkline
	sections = append(sections,
		labelStyle.Render("💿 Disk %"),
		renderSparkline(diskVals, sparkWidth, 100),
		statStyle.Render(fmt.Sprintf("   min: %.1f%%  avg: %.1f%%  max: %.1f%%", minVal(diskVals), avgVal(diskVals), maxVal(diskVals))),
		"")

	// Network sparkline (auto-scaled)
	maxNet := maxVal(netVals)
	if maxNet < 1 {
		maxNet = 1
	}
	sections = append(sections,
		labelStyle.Render("🌐 Net KB/s"),
		renderSparkline(netVals, sparkWidth, maxNet),
		statStyle.Render(fmt.Sprintf("   min: %.1f  avg: %.1f  max: %.1f KB/s", minVal(netVals), avgVal(netVals), maxVal(netVals))),
		"")

	sections = append(sections,
		styles.Muted().Render("[r] Refresh  [1] Back to Live"))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderSparkline renders an ASCII sparkline chart
func renderSparkline(values []float64, width int, maxRange float64) string {
	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	if len(values) == 0 {
		return ""
	}

	// Downsample to fit width
	sampled := downsample(values, width)

	var result strings.Builder
	for _, v := range sampled {
		ratio := v / maxRange
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}

		idx := int(ratio * float64(len(blocks)-1))

		// Color based on percentage (for 0-100 range)
		pct := ratio * 100
		color := "#0FD976" // green
		if pct >= 85 {
			color = "#FF6B6B" // red
		} else if pct >= 70 {
			color = "#FFA500" // orange
		}

		styled := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(string(blocks[idx]))
		result.WriteString(styled)
	}

	return result.String()
}

// downsample reduces a slice to the target length by averaging buckets
func downsample(values []float64, targetLen int) []float64 {
	if len(values) <= targetLen {
		return values
	}

	result := make([]float64, targetLen)
	bucketSize := float64(len(values)) / float64(targetLen)

	for i := range targetLen {
		start := int(float64(i) * bucketSize)
		end := int(float64(i+1) * bucketSize)
		if end > len(values) {
			end = len(values)
		}
		if start >= end {
			if start < len(values) {
				result[i] = values[start]
			}
			continue
		}

		sum := 0.0
		for j := start; j < end; j++ {
			sum += values[j]
		}
		result[i] = sum / float64(end-start)
	}

	return result
}

func minVal(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	m := vals[0]
	for _, v := range vals[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func maxVal(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

func avgVal(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}
