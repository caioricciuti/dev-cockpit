package cleanup

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/caioricciuti/dev-cockpit/internal/config"
)

func newTestModel(t *testing.T) *Model {
	t.Helper()
	m := New(&config.Config{})
	m.scanning = false
	if len(m.targets) == 0 {
		t.Skip("no cleanup targets are defined on this platform")
	}
	return m
}

// spaceKey builds the space press the way bubbletea v2 delivers it.
//
// v2 routes space through Keystroke() rather than Text, because its String()
// excludes a single space from the text path, so it arrives as "space" where
// v1 gave " ". The module matched only " ", which silently made the Cleanup
// module's only selection key do nothing after the v2 migration.
func spaceKey() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
}

func TestSpaceKeyStringifiesAsSpaceWord(t *testing.T) {
	// Guards the assumption the fix rests on. If this ever changes back, the
	// binding still works, but the reason for accepting both is gone.
	if got := spaceKey().String(); got != "space" {
		t.Errorf("space key String() = %q, want %q", got, "space")
	}
}

func TestSpaceTogglesSelection(t *testing.T) {
	m := newTestModel(t)
	if m.targets[0].Selected {
		t.Fatal("targets should start unselected")
	}

	m.Update(spaceKey())
	if !m.targets[0].Selected {
		t.Error("space did not select the item under the cursor")
	}

	m.Update(spaceKey())
	if m.targets[0].Selected {
		t.Error("space did not deselect the item under the cursor")
	}
}

func TestSpaceTogglesOnlyTheCursorRow(t *testing.T) {
	m := newTestModel(t)
	if len(m.targets) < 2 {
		t.Skip("needs at least two targets")
	}

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	m.Update(spaceKey())

	if m.targets[0].Selected {
		t.Error("space selected a row the cursor was not on")
	}
	if !m.targets[1].Selected {
		t.Error("space did not select the row the cursor moved to")
	}
}

func TestEnterRefusesWithNothingSelected(t *testing.T) {
	m := newTestModel(t)
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if !strings.Contains(m.message, "at least one") {
		t.Errorf("enter with no selection should explain itself, got %q", m.message)
	}
}

func TestSelectAllAndNone(t *testing.T) {
	m := newTestModel(t)

	m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	for i, target := range m.targets {
		if !target.Selected {
			t.Fatalf("target %d (%s) not selected after 'a'", i, target.Name)
		}
	}

	m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	for i, target := range m.targets {
		if target.Selected {
			t.Fatalf("target %d (%s) still selected after 'n'", i, target.Name)
		}
	}
}
