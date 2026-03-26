package support

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/caioricciuti/dev-cockpit/internal/ui/events"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	sponsorsURL  = "https://github.com/sponsors/caioricciuti"
	buyCoffeeURL = "https://buymeacoffee.com/caioricciuti"
)

type Model struct {
	width        int
	height       int
	status       string
	selectedItem int
}

func New() *Model {
	return &Model{}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (interface{}, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case events.Focus:
		m.status = ""
	case events.Blur:
		m.status = ""

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedItem > 0 {
				m.selectedItem--
			}
		case "down", "j":
			if m.selectedItem < 1 {
				m.selectedItem++
			}
		case "1":
			return m, m.openURL("GitHub Sponsors", sponsorsURL)
		case "2":
			return m, m.openURL("Buy Me a Coffee", buyCoffeeURL)
		case "c":
			// Copy selected URL to clipboard
			url := sponsorsURL
			if m.selectedItem == 1 {
				url = buyCoffeeURL
			}
			return m, m.copyToClipboard(url)
		case "enter", " ":
			// Open selected item
			if m.selectedItem == 0 {
				return m, m.openURL("GitHub Sponsors", sponsorsURL)
			} else {
				return m, m.openURL("Buy Me a Coffee", buyCoffeeURL)
			}
		}

	case supportMsg:
		m.status = msg.note
	}

	return m, nil
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading support dashboard..."
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF"))

	paragraphStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDD"))

	selectedCardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D9FF")).
		Padding(1, 2).
		Width(m.width - 10)

	unselectedCardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444")).
		Padding(1, 2).
		Width(m.width - 10)

	optionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	urlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666"))

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0FD976")).
		Bold(true)

	controlsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666"))

	// Build content simply
	content := []string{
		titleStyle.Render("💛 SUPPORT DEV COCKPIT"),
		"",
		paragraphStyle.Render("Help keep this project alive! Support via:"),
		"",
	}

	// GitHub Sponsors card
	card1Style := unselectedCardStyle
	prefix1 := "  "
	if m.selectedItem == 0 {
		card1Style = selectedCardStyle
		prefix1 = "▶ "
	}
	card1 := card1Style.Render(
		optionStyle.Render(prefix1+"[1] GitHub Sponsors") + "\n" +
			urlStyle.Render("    "+sponsorsURL),
	)
	content = append(content, card1, "")

	// Buy Me a Coffee card
	card2Style := unselectedCardStyle
	prefix2 := "  "
	if m.selectedItem == 1 {
		card2Style = selectedCardStyle
		prefix2 = "▶ "
	}
	card2 := card2Style.Render(
		optionStyle.Render(prefix2+"[2] Buy Me a Coffee") + "\n" +
			urlStyle.Render("    "+buyCoffeeURL),
	)
	content = append(content, card2)

	// Controls
	content = append(content, "", controlsStyle.Render("↑/↓ Navigate • Enter Open • C Copy • 1/2 Quick • Esc Back"))

	// Status
	if m.status != "" {
		content = append(content, "", statusStyle.Render(m.status))
	}

	// Join content
	finalContent := lipgloss.JoinVertical(lipgloss.Left, content...)

	// Apply viewport to prevent overflow
	maxHeight := m.height - 4 // Account for margins
	if maxHeight < 10 {
		maxHeight = 10
	}

	return lipgloss.NewStyle().MaxHeight(maxHeight).Render(finalContent)
}

func (m *Model) Title() string {
	return "Support"
}

// HasOpenModal returns true if the module has an open modal/dialog
func (m *Model) HasOpenModal() bool {
	return false
}

func (m *Model) openURL(label, url string) tea.Cmd {
	return func() tea.Msg {
		var opener string
		switch runtime.GOOS {
		case "darwin":
			opener = "open"
		case "linux":
			opener = "xdg-open"
		default:
			return supportMsg{note: fmt.Sprintf("Link: %s", url)}
		}

		cmd := exec.Command(opener, url)
		if err := cmd.Start(); err != nil {
			return supportMsg{note: fmt.Sprintf("Unable to open %s: %v", label, err)}
		}
		return supportMsg{note: fmt.Sprintf("Opening %s in your browser…", label)}
	}
}

func (m *Model) copyToClipboard(url string) tea.Cmd {
	return func() tea.Msg {
		var clipCmd string
		switch runtime.GOOS {
		case "darwin":
			clipCmd = "pbcopy"
		case "linux":
			// Try xclip first, fall back to xsel
			if _, err := exec.LookPath("xclip"); err == nil {
				clipCmd = "xclip"
			} else if _, err := exec.LookPath("xsel"); err == nil {
				clipCmd = "xsel"
			} else {
				return supportMsg{note: fmt.Sprintf("URL: %s (install xclip for clipboard support)", url)}
			}
		default:
			return supportMsg{note: fmt.Sprintf("URL: %s", url)}
		}

		var cmd *exec.Cmd
		if clipCmd == "xclip" {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if clipCmd == "xsel" {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			cmd = exec.Command(clipCmd)
		}

		stdin, err := cmd.StdinPipe()
		if err != nil {
			return supportMsg{note: fmt.Sprintf("Failed to access clipboard: %v", err)}
		}

		if err := cmd.Start(); err != nil {
			return supportMsg{note: fmt.Sprintf("Failed to copy: %v", err)}
		}

		stdin.Write([]byte(url))
		stdin.Close()
		cmd.Wait()

		return supportMsg{note: fmt.Sprintf("Copied to clipboard: %s", url)}
	}
}

type supportMsg struct {
	note string
}
