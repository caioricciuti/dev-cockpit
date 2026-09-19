package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// The footer used to clamp its spacer to zero when the terminal was too
// narrow, which ran the shortcut hints straight into the clock ("Q Quit⟳
// 16:08:41") and pushed the footer onto a second line, eating a row of
// content. It now drops hints, then the clock, rather than overflowing.

func TestFooterNeverExceedsTerminalWidth(t *testing.T) {
	widths := []int{20, 30, 40, 50, 60, 72, 80, 100, 120, 160, 200}

	for _, w := range widths {
		m := New(testConfig(t), "test")
		m.Update(tea.WindowSizeMsg{Width: w, Height: 40})

		footer := m.renderFooter()
		for i, line := range strings.Split(footer, "\n") {
			if got := lipgloss.Width(line); got > w {
				t.Errorf("width %d: footer line %d is %d cells wide, want <= %d\n%q",
					w, i, got, w, line)
			}
		}
		m.Close()
	}
}

func TestFooterStaysTwoLines(t *testing.T) {
	// One border row plus one content row. A third line means it wrapped.
	widths := []int{20, 40, 80, 120, 200}

	for _, w := range widths {
		m := New(testConfig(t), "test")
		m.Update(tea.WindowSizeMsg{Width: w, Height: 40})

		lines := strings.Split(m.renderFooter(), "\n")
		if len(lines) != 2 {
			t.Errorf("width %d: footer is %d lines, want 2:\n%s", w, len(lines), m.renderFooter())
		}
		m.Close()
	}
}

func TestFooterKeepsHintsAndClockApart(t *testing.T) {
	// The reported symptom was the clock glyph touching the last hint.
	widths := []int{40, 60, 80, 100, 120, 160}

	for _, w := range widths {
		m := New(testConfig(t), "test")
		m.Update(tea.WindowSizeMsg{Width: w, Height: 40})

		plain := stripANSI(m.renderFooter())
		if idx := strings.Index(plain, "⟳"); idx > 0 {
			if before := plain[:idx]; !strings.HasSuffix(before, " ") {
				t.Errorf("width %d: clock is not separated from the text before it: %q",
					w, plain[max(0, idx-24):idx+12])
			}
		}
		m.Close()
	}
}

func TestFooterShowsFullHintsWhenThereIsRoom(t *testing.T) {
	m := New(testConfig(t), "test")
	defer m.Close()
	m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})

	plain := stripANSI(m.renderFooter())
	for _, want := range []string{"Tab Switch", "Enter Focus", "Q Quit", "⟳"} {
		if !strings.Contains(plain, want) {
			t.Errorf("a 140-column footer should contain %q:\n%s", want, plain)
		}
	}
}

func TestFooterDropsHintsBeforeItWraps(t *testing.T) {
	m := New(testConfig(t), "test")
	defer m.Close()
	m.Update(tea.WindowSizeMsg{Width: 60, Height: 40})

	plain := stripANSI(m.renderFooter())
	if strings.Contains(plain, "Enter Focus") {
		t.Errorf("a 60-column footer should have dropped the long hints:\n%s", plain)
	}
	if !strings.Contains(plain, "Dev Cockpit") {
		t.Errorf("the version should survive at every width:\n%s", plain)
	}
}

// stripANSI removes SGR sequences so assertions can look at the text.
func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			i = j + 1
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
