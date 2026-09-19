package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Tab width used to be the available width divided by the number of modules,
// which cut every label to ten cells regardless of how much room there was.
// On a 160-column terminal "Dashboard" still rendered as "Dashb...".

func renderTabsAt(t *testing.T, width int) string {
	t.Helper()
	m := New(testConfig(t), "test")
	t.Cleanup(func() { m.Close() })
	m.Update(tea.WindowSizeMsg{Width: width, Height: 40})
	return stripANSI(m.renderTabs())
}

func TestTabLabelsAreNotTruncatedWhenThereIsRoom(t *testing.T) {
	for _, width := range []int{120, 160, 200} {
		tabs := renderTabsAt(t, width)
		for _, label := range []string{"Dashboard", "Quick Actions", "Diagnostics", "Processes"} {
			if !strings.Contains(tabs, label) {
				t.Errorf("width %d: tab bar should show %q in full:\n%s", width, label, tabs)
			}
		}
		if strings.Contains(tabs, "...") {
			t.Errorf("width %d: tab bar should not truncate at this width:\n%s", width, tabs)
		}
	}
}

func TestTabBarStaysWithinTerminalWidth(t *testing.T) {
	for _, width := range []int{40, 60, 80, 100, 120, 160, 200} {
		tabs := renderTabsAt(t, width)
		for i, line := range strings.Split(tabs, "\n") {
			if got := lipgloss.Width(line); got > width {
				t.Errorf("width %d: tab row %d is %d cells wide, want <= %d",
					width, i, got, width)
			}
		}
	}
}

func TestTabBarKeepsToTwoRowsWhenItCan(t *testing.T) {
	// Every tab row costs a row of content, so a wide terminal should not
	// spend more than two on navigation.
	for _, width := range []int{120, 160, 200} {
		tabs := strings.TrimRight(renderTabsAt(t, width), "\n")
		rows := 0
		for _, line := range strings.Split(tabs, "\n") {
			if strings.TrimSpace(line) != "" {
				rows++
			}
		}
		// Two tab rows plus the bar's bottom border.
		if rows > 3 {
			t.Errorf("width %d: tab bar uses %d rows, want at most 3:\n%s", width, rows, tabs)
		}
	}
}

func TestActiveTabIsMarked(t *testing.T) {
	tabs := renderTabsAt(t, 160)
	if !strings.Contains(tabs, "◎") {
		t.Errorf("the active tab should carry its marker:\n%s", tabs)
	}
}
