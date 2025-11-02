package network

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/caioricciuti/dev-cockpit/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gnet "github.com/shirou/gopsutil/v3/net"
)

// ViewMode represents different network views
type ViewMode int

const (
	ViewOverview ViewMode = iota
	ViewPorts
	ViewDiagnostics
	ViewQuality
	ViewTools
)

// DiagnosticMode represents different diagnostic tools
type DiagnosticMode int

const (
	DiagPing DiagnosticMode = iota
	DiagTraceroute
	DiagDNS
)

// ToolMode represents different network tools
type ToolMode int

const (
	ToolWhois ToolMode = iota
)

// PortInfo represents a listening port
type PortInfo struct {
	Command  string
	PID      string
	User     string
	Protocol string
	Address  string
	Port     string
}

// QualityResult represents network quality test results
type QualityResult struct {
	DownloadMbps   float64
	UploadMbps     float64
	LatencyMs      float64
	Responsiveness int
	Interface      string
	Timestamp      time.Time
}

// Model represents the network module state
type Model struct {
	config *config.Config
	width  int
	height int

	// View management
	activeView ViewMode
	views      []string

	// Overview data
	ifaces  []gnet.InterfaceStat
	gateway string
	message string
	cursor  int

	// Port scanner
	listeningPorts []PortInfo
	portsLoading   bool
	portsCursor    int
	portsMessage   string

	// Diagnostics
	diagMode        DiagnosticMode
	diagInputActive bool
	diagInputBuffer string
	diagRunning     bool
	diagOutput      string
	diagTarget      string

	// Quality test
	qualityRunning    bool
	qualityResult     *QualityResult
	qualityMessage    string
	qualityAvailable  bool

	// Tools
	toolMode        ToolMode
	toolInputActive bool
	toolInputBuffer string
	toolRunning     bool
	toolOutput      string
	toolTarget      string

	// General
	errorMsg string
}

// Message types
type netMsg struct {
	ifaces  []gnet.InterfaceStat
	gateway string
	note    string
}

type portsMsg struct {
	ports []PortInfo
	err   error
}

type diagCompleteMsg struct {
	mode   DiagnosticMode
	output string
	target string
	err    error
}

type qualityCompleteMsg struct {
	result *QualityResult
	err    error
}

type toolCompleteMsg struct {
	mode   ToolMode
	output string
	target string
	err    error
}

// New creates a new network module
func New(cfg *config.Config) *Model {
	return &Model{
		config: cfg,
		views:  []string{"Overview", "Ports", "Diagnostics", "Quality", "Tools"},
		qualityAvailable: checkNetworkQualityAvailable(),
	}
}

// Init initializes the module
func (m *Model) Init() tea.Cmd {
	return m.refresh()
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (interface{}, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		// Handle input mode for diagnostics
		if m.diagInputActive {
			return m, m.handleDiagInput(msg)
		}

		// Handle input mode for tools
		if m.toolInputActive {
			return m, m.handleToolInput(msg)
		}

		// Global navigation
		switch msg.String() {
		case "1":
			m.activeView = ViewOverview
		case "2":
			m.activeView = ViewPorts
			if len(m.listeningPorts) == 0 && !m.portsLoading {
				return m, m.scanPorts()
			}
		case "3":
			m.activeView = ViewDiagnostics
		case "4":
			if m.qualityAvailable {
				m.activeView = ViewQuality
			}
		case "5":
			m.activeView = ViewTools
		case "tab", "l":
			m.activeView = (m.activeView + 1) % ViewMode(len(m.views))
			if m.activeView == ViewQuality && !m.qualityAvailable {
				m.activeView = (m.activeView + 1) % ViewMode(len(m.views))
			}
		case "shift+tab", "h":
			m.activeView = (m.activeView - 1 + ViewMode(len(m.views))) % ViewMode(len(m.views))
			if m.activeView == ViewQuality && !m.qualityAvailable {
				m.activeView = (m.activeView - 1 + ViewMode(len(m.views))) % ViewMode(len(m.views))
			}
		}

		// View-specific navigation
		switch m.activeView {
		case ViewOverview:
			return m, m.handleOverviewKeys(msg)
		case ViewPorts:
			return m, m.handlePortsKeys(msg)
		case ViewDiagnostics:
			return m, m.handleDiagnosticsKeys(msg)
		case ViewQuality:
			return m, m.handleQualityKeys(msg)
		case ViewTools:
			return m, m.handleToolsKeys(msg)
		}

	case netMsg:
		m.ifaces = msg.ifaces
		m.gateway = msg.gateway
		m.message = msg.note
		if m.cursor >= len(m.ifaces) {
			m.cursor = 0
		}

	case portsMsg:
		m.portsLoading = false
		if msg.err != nil {
			m.portsMessage = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.listeningPorts = msg.ports
			m.portsMessage = fmt.Sprintf("Found %d listening ports", len(msg.ports))
		}

	case diagCompleteMsg:
		m.diagRunning = false
		m.diagTarget = msg.target
		if msg.err != nil {
			m.diagOutput = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.diagOutput = msg.output
		}

	case qualityCompleteMsg:
		m.qualityRunning = false
		if msg.err != nil {
			m.qualityMessage = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.qualityResult = msg.result
			m.qualityMessage = "Test completed successfully"
		}

	case toolCompleteMsg:
		m.toolRunning = false
		m.toolTarget = msg.target
		if msg.err != nil {
			m.toolOutput = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.toolOutput = msg.output
		}
	}

	return m, nil
}

// View renders the module
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var content strings.Builder

	// Title
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D9FF")).Render("🌐 NETWORK")
	content.WriteString(title + "\n\n")

	// Tab navigation
	content.WriteString(m.renderTabs() + "\n")
	content.WriteString(strings.Repeat("─", m.width-4) + "\n\n")

	// View content
	switch m.activeView {
	case ViewOverview:
		content.WriteString(m.renderOverview())
	case ViewPorts:
		content.WriteString(m.renderPorts())
	case ViewDiagnostics:
		content.WriteString(m.renderDiagnostics())
	case ViewQuality:
		content.WriteString(m.renderQuality())
	case ViewTools:
		content.WriteString(m.renderTools())
	}

	// Apply viewport to prevent overflow
	maxHeight := m.height - 4
	if maxHeight < 10 {
		maxHeight = 10
	}

	return lipgloss.NewStyle().MaxHeight(maxHeight).Render(content.String())
}

// Title returns the module title
func (m *Model) Title() string {
	return "Network"
}

// HasOpenModal returns true if the module has an open modal/dialog
func (m *Model) HasOpenModal() bool {
	return m.diagInputActive || m.toolInputActive
}

// renderTabs creates the tab navigation bar
func (m *Model) renderTabs() string {
	var tabs []string

	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF")).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 1)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		Padding(0, 1)

	for i, view := range m.views {
		// Skip quality view if not available
		if ViewMode(i) == ViewQuality && !m.qualityAvailable {
			continue
		}

		if ViewMode(i) == m.activeView {
			tabs = append(tabs, activeStyle.Render(view))
		} else {
			tabs = append(tabs, inactiveStyle.Render(view))
		}
	}

	return strings.Join(tabs, " ")
}

// Overview view handlers
func (m *Model) handleOverviewKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "r":
		return m.refresh()
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.ifaces)-1 {
			m.cursor++
		}
	case "p":
		return m.pingGateway()
	}
	return nil
}

func (m *Model) renderOverview() string {
	var b strings.Builder

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Render("[R]efresh  [P]ing gateway  [↑/↓]Navigate  [1-5]Switch views")
	b.WriteString(help + "\n\n")

	if m.message != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render(m.message) + "\n\n")
	}

	if len(m.ifaces) == 0 {
		b.WriteString("No interfaces found.\n")
	} else {
		gw := m.gateway
		if gw == "" {
			gw = "(no default)"
		}
		b.WriteString(fmt.Sprintf("Default Gateway: %s\n\n", gw))

		item := lipgloss.NewStyle().PaddingLeft(2)
		sel := lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#00D9FF")).Bold(true)

		b.WriteString("Network Interfaces:\n")
		for i, ifc := range m.ifaces {
			addrs := []string{}
			for _, a := range ifc.Addrs {
				addrs = append(addrs, a.Addr)
			}
			line := fmt.Sprintf("%-12s %s", ifc.Name, strings.Join(addrs, ", "))
			if i == m.cursor {
				b.WriteString(sel.Render("▶ " + line))
			} else {
				b.WriteString(item.Render("  " + line))
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m *Model) refresh() tea.Cmd {
	return func() tea.Msg {
		ifaces, _ := gnet.Interfaces()
		gw := getDefaultGateway()
		note := fmt.Sprintf("%d interfaces found", len(ifaces))
		return netMsg{ifaces: ifaces, gateway: gw, note: note}
	}
}

func (m *Model) pingGateway() tea.Cmd {
	target := m.gateway
	if target == "" {
		target = "1.1.1.1"
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "/sbin/ping", "-c", "2", target)
		out, err := cmd.CombinedOutput()
		if err != nil {
			errMsg := string(out)
			if errMsg == "" {
				errMsg = err.Error()
			}
			return netMsg{note: fmt.Sprintf("Ping error: %v", errMsg)}
		}
		// Short summary
		lines := strings.Split(string(out), "\n")
		if len(lines) > 2 {
			return netMsg{note: lines[len(lines)-3] + " | " + lines[len(lines)-2]}
		}
		return netMsg{note: string(out)}
	}
}

// Port scanner handlers
func (m *Model) handlePortsKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "r":
		return m.scanPorts()
	case "up", "k":
		if m.portsCursor > 0 {
			m.portsCursor--
		}
	case "down", "j":
		if m.portsCursor < len(m.listeningPorts)-1 {
			m.portsCursor++
		}
	}
	return nil
}

func (m *Model) renderPorts() string {
	var b strings.Builder

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Render("[R]efresh  [↑/↓]Navigate  [1-5]Switch views")
	b.WriteString(help + "\n\n")

	if m.portsLoading {
		b.WriteString("⏳ Scanning listening ports...\n")
		return b.String()
	}

	if m.portsMessage != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render(m.portsMessage) + "\n\n")
	}

	if len(m.listeningPorts) == 0 {
		b.WriteString("No listening ports found. Press [R] to scan.\n")
	} else {
		b.WriteString("LISTENING PORTS (TCP):\n\n")

		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D9FF"))
		b.WriteString(headerStyle.Render(fmt.Sprintf("%-15s %-8s %-10s %-8s %s\n", "COMMAND", "PID", "USER", "PORT", "ADDRESS")))

		item := lipgloss.NewStyle().PaddingLeft(2)
		sel := lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#00D9FF")).Bold(true)

		for i, port := range m.listeningPorts {
			line := fmt.Sprintf("%-15s %-8s %-10s %-8s %s", port.Command, port.PID, port.User, port.Port, port.Address)
			if i == m.portsCursor {
				b.WriteString(sel.Render("▶ " + line))
			} else {
				b.WriteString(item.Render("  " + line))
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m *Model) scanPorts() tea.Cmd {
	m.portsLoading = true
	m.portsMessage = "Scanning..."

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "lsof", "-iTCP", "-sTCP:LISTEN", "-n", "-P")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return portsMsg{err: err}
		}

		ports := parseListeningPorts(string(output))
		return portsMsg{ports: ports}
	}
}

func parseListeningPorts(output string) []PortInfo {
	var ports []PortInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines[1:] { // Skip header
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		// Parse address:port from last field
		addr := fields[8]
		parts := strings.Split(addr, ":")
		if len(parts) < 2 {
			continue
		}

		port := PortInfo{
			Command:  fields[0],
			PID:      fields[1],
			User:     fields[2],
			Protocol: "TCP",
			Address:  strings.Join(parts[:len(parts)-1], ":"),
			Port:     parts[len(parts)-1],
		}

		ports = append(ports, port)
	}

	return ports
}

// Diagnostics handlers
func (m *Model) handleDiagnosticsKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "p":
		m.diagMode = DiagPing
		m.diagInputActive = true
		m.diagInputBuffer = ""
	case "t":
		m.diagMode = DiagTraceroute
		m.diagInputActive = true
		m.diagInputBuffer = ""
	case "d":
		m.diagMode = DiagDNS
		m.diagInputActive = true
		m.diagInputBuffer = ""
	}
	return nil
}

func (m *Model) handleDiagInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.diagInputActive = false
		m.diagInputBuffer = ""
		return nil
	case "enter":
		if m.diagInputBuffer == "" {
			return nil
		}
		target := m.diagInputBuffer
		m.diagInputActive = false
		m.diagInputBuffer = ""

		if !isValidTarget(target) {
			m.diagOutput = "Invalid target. Use domain name or IP address."
			return nil
		}

		switch m.diagMode {
		case DiagPing:
			return m.executePing(target)
		case DiagTraceroute:
			return m.executeTraceroute(target)
		case DiagDNS:
			return m.executeDNS(target)
		}
	case "backspace":
		if len(m.diagInputBuffer) > 0 {
			m.diagInputBuffer = m.diagInputBuffer[:len(m.diagInputBuffer)-1]
		}
	default:
		// Add printable characters
		if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
			m.diagInputBuffer += msg.String()
		}
	}
	return nil
}

func (m *Model) renderDiagnostics() string {
	var b strings.Builder

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Render("[P]ing  [T]raceroute  [D]NS Lookup  [1-5]Switch views")
	b.WriteString(help + "\n\n")

	// Mode indicator
	modeMap := map[DiagnosticMode]string{
		DiagPing:       "Ping",
		DiagTraceroute: "Traceroute",
		DiagDNS:        "DNS Lookup",
	}

	b.WriteString(fmt.Sprintf("Mode: %s\n\n", modeMap[m.diagMode]))

	// Input box
	b.WriteString(m.renderInputBox("Enter target (domain or IP)") + "\n\n")

	if m.diagInputActive {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Render("Press ENTER to start, ESC to cancel") + "\n\n")
	}

	// Results
	if m.diagRunning {
		b.WriteString("⏳ Running diagnostic...\n")
	} else if m.diagOutput != "" {
		b.WriteString(fmt.Sprintf("Last Result (target: %s):\n", m.diagTarget))
		outputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#aaa")).MaxHeight(20)
		b.WriteString(outputStyle.Render(m.diagOutput) + "\n")
	}

	return b.String()
}

func (m *Model) renderInputBox(placeholder string) string {
	text := m.diagInputBuffer
	if m.toolInputActive {
		text = m.toolInputBuffer
	}

	if text == "" && !m.diagInputActive && !m.toolInputActive {
		text = placeholder
	}

	cursor := ""
	if m.diagInputActive || m.toolInputActive {
		cursor = "▊"
	}

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D9FF")).
		Padding(0, 1).
		Width(min(m.width-8, 50))

	return inputStyle.Render(text + cursor)
}

func (m *Model) executePing(target string) tea.Cmd {
	m.diagRunning = true

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "/sbin/ping", "-c", "4", target)
		output, err := cmd.CombinedOutput()

		if err != nil {
			errMsg := string(output)
			if errMsg == "" {
				errMsg = err.Error()
			}
			return diagCompleteMsg{
				mode:   DiagPing,
				output: "",
				target: target,
				err:    fmt.Errorf("%s", errMsg),
			}
		}

		return diagCompleteMsg{
			mode:   DiagPing,
			output: string(output),
			target: target,
			err:    nil,
		}
	}
}

func (m *Model) executeTraceroute(target string) tea.Cmd {
	m.diagRunning = true

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "/usr/sbin/traceroute", "-m", "15", target)
		output, err := cmd.CombinedOutput()

		if err != nil {
			errMsg := string(output)
			if errMsg == "" {
				errMsg = err.Error()
			}
			return diagCompleteMsg{
				mode:   DiagTraceroute,
				output: "",
				target: target,
				err:    fmt.Errorf("%s", errMsg),
			}
		}

		return diagCompleteMsg{
			mode:   DiagTraceroute,
			output: string(output),
			target: target,
			err:    nil,
		}
	}
}

func (m *Model) executeDNS(target string) tea.Cmd {
	m.diagRunning = true

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Try dig first, fall back to nslookup
		cmd := exec.CommandContext(ctx, "/usr/bin/dig", "+short", target)
		output, err := cmd.CombinedOutput()

		if err != nil {
			// Fall back to nslookup
			cmd = exec.CommandContext(ctx, "/usr/bin/nslookup", target)
			output, err = cmd.CombinedOutput()

			if err != nil {
				errMsg := string(output)
				if errMsg == "" {
					errMsg = err.Error()
				}
				return diagCompleteMsg{
					mode:   DiagDNS,
					output: "",
					target: target,
					err:    fmt.Errorf("%s", errMsg),
				}
			}
		}

		return diagCompleteMsg{
			mode:   DiagDNS,
			output: string(output),
			target: target,
			err:    nil,
		}
	}
}

// Quality test handlers
func (m *Model) handleQualityKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "s":
		if !m.qualityRunning {
			return m.executeQualityTest()
		}
	}
	return nil
}

func (m *Model) renderQuality() string {
	var b strings.Builder

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Render("[S]tart test  [1-5]Switch views")
	b.WriteString(help + "\n\n")

	b.WriteString("NETWORK QUALITY TEST\n\n")

	if m.qualityRunning {
		b.WriteString("⏳ Testing network performance... (15-20 seconds)\n\n")
		b.WriteString("Please wait while we measure your connection quality.\n")
	} else if m.qualityResult != nil {
		b.WriteString("Last Test Results:\n\n")

		resultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF")).Bold(true)

		b.WriteString(fmt.Sprintf("  Download:       %s\n", resultStyle.Render(fmt.Sprintf("%.1f Mbps", m.qualityResult.DownloadMbps))))
		b.WriteString(fmt.Sprintf("  Upload:         %s\n", resultStyle.Render(fmt.Sprintf("%.1f Mbps", m.qualityResult.UploadMbps))))
		b.WriteString(fmt.Sprintf("  Latency:        %s\n", resultStyle.Render(fmt.Sprintf("%.1f ms", m.qualityResult.LatencyMs))))

		rpm := m.qualityResult.Responsiveness
		rpmStatus := "Good"
		if rpm < 100 {
			rpmStatus = "Poor"
		} else if rpm < 300 {
			rpmStatus = "Fair"
		}
		b.WriteString(fmt.Sprintf("  Responsiveness: %s (%s)\n", resultStyle.Render(fmt.Sprintf("%d RPM", rpm)), rpmStatus))
		b.WriteString(fmt.Sprintf("  Interface:      %s\n", m.qualityResult.Interface))
		b.WriteString(fmt.Sprintf("  Tested:         %s\n\n", m.qualityResult.Timestamp.Format("15:04:05")))
	} else {
		b.WriteString("Press [S] to start a network quality test.\n\n")
		b.WriteString("This will measure:\n")
		b.WriteString("  • Download and upload throughput\n")
		b.WriteString("  • Network latency\n")
		b.WriteString("  • Responsiveness (RPM - Requests Per Minute)\n\n")
		b.WriteString("Note: Test takes 15-20 seconds to complete.\n")
	}

	if m.qualityMessage != "" {
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render(m.qualityMessage) + "\n")
	}

	return b.String()
}

func (m *Model) executeQualityTest() tea.Cmd {
	m.qualityRunning = true
	m.qualityMessage = "Running test..."

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "networkQuality", "-c")
		output, err := cmd.Output()

		if err != nil {
			return qualityCompleteMsg{err: err}
		}

		// Parse JSON output
		var rawResult map[string]interface{}
		if err := json.Unmarshal(output, &rawResult); err != nil {
			return qualityCompleteMsg{err: err}
		}

		result := &QualityResult{
			Timestamp: time.Now(),
		}

		// Extract values from JSON
		if dl, ok := rawResult["dl_throughput"].(float64); ok {
			result.DownloadMbps = (dl / 1000000.0) * 8
		}
		if ul, ok := rawResult["ul_throughput"].(float64); ok {
			result.UploadMbps = (ul / 1000000.0) * 8
		}
		if lat, ok := rawResult["base_rtt"].(float64); ok {
			result.LatencyMs = lat
		}
		if rpm, ok := rawResult["responsiveness"].(float64); ok {
			result.Responsiveness = int(rpm)
		}
		if iface, ok := rawResult["interface_name"].(string); ok {
			result.Interface = iface
		}

		return qualityCompleteMsg{result: result}
	}
}

// Tools handlers
func (m *Model) handleToolsKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "w":
		m.toolMode = ToolWhois
		m.toolInputActive = true
		m.toolInputBuffer = ""
	}
	return nil
}

func (m *Model) handleToolInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.toolInputActive = false
		m.toolInputBuffer = ""
		return nil
	case "enter":
		if m.toolInputBuffer == "" {
			return nil
		}
		target := m.toolInputBuffer
		m.toolInputActive = false
		m.toolInputBuffer = ""

		if !isValidTarget(target) {
			m.toolOutput = "Invalid target. Use a domain name."
			return nil
		}

		switch m.toolMode {
		case ToolWhois:
			return m.executeWhois(target)
		}
	case "backspace":
		if len(m.toolInputBuffer) > 0 {
			m.toolInputBuffer = m.toolInputBuffer[:len(m.toolInputBuffer)-1]
		}
	default:
		// Add printable characters
		if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
			m.toolInputBuffer += msg.String()
		}
	}
	return nil
}

func (m *Model) renderTools() string {
	var b strings.Builder

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Render("[W]hois  [1-5]Switch views")
	b.WriteString(help + "\n\n")

	b.WriteString("NETWORK TOOLS\n\n")

	// Mode indicator
	toolMap := map[ToolMode]string{
		ToolWhois: "Whois",
	}
	b.WriteString(fmt.Sprintf("Tool: %s\n\n", toolMap[m.toolMode]))

	// Input box
	b.WriteString(m.renderInputBox("Enter domain name") + "\n\n")

	if m.toolInputActive {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Render("Press ENTER to query, ESC to cancel") + "\n\n")
	}

	// Results
	if m.toolRunning {
		b.WriteString("⏳ Running query...\n")
	} else if m.toolOutput != "" {
		b.WriteString(fmt.Sprintf("Results (target: %s):\n", m.toolTarget))
		outputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#aaa")).MaxHeight(20)
		b.WriteString(outputStyle.Render(m.toolOutput) + "\n")
	}

	return b.String()
}

func (m *Model) executeWhois(target string) tea.Cmd {
	m.toolRunning = true

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "/usr/bin/whois", target)
		output, err := cmd.CombinedOutput()

		if err != nil {
			errMsg := string(output)
			if errMsg == "" {
				errMsg = err.Error()
			}
			return toolCompleteMsg{
				mode:   ToolWhois,
				output: "",
				target: target,
				err:    fmt.Errorf("%s", errMsg),
			}
		}

		return toolCompleteMsg{
			mode:   ToolWhois,
			output: string(output),
			target: target,
			err:    nil,
		}
	}
}

// Helper functions
func getDefaultGateway() string {
	out, err := exec.Command("sh", "-c", "route get default | awk '/gateway/{print $2}'").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func isValidTarget(target string) bool {
	if target == "" {
		return false
	}
	// Allow: domain names, IPv4, IPv6
	// Simple regex for basic validation
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9.-:]+$`)
	return validPattern.MatchString(target)
}

func checkNetworkQualityAvailable() bool {
	_, err := exec.LookPath("networkQuality")
	return err == nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
