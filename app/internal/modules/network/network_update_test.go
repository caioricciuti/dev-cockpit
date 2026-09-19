package network

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// These cover the module's state machine: which view is active after a key,
// and how the tab cycle behaves when a view is hidden. This is the surface a
// bubbletea major upgrade is most likely to break silently, since key
// decoding changes but the code still compiles.

func key(s string) tea.KeyMsg {
	switch s {
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func newTestModel(qualityAvailable bool) *Model {
	return &Model{
		views:            []string{"Overview", "Ports", "Diagnostics", "Quality", "Tools"},
		qualityAvailable: qualityAvailable,
	}
}

func TestNumberKeysSelectView(t *testing.T) {
	tests := []struct {
		key  string
		want ViewMode
	}{
		{"1", ViewOverview},
		{"2", ViewPorts},
		{"3", ViewDiagnostics},
		{"4", ViewQuality},
		{"5", ViewTools},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			m := newTestModel(true)
			m.Update(key(tt.key))
			if m.activeView != tt.want {
				t.Errorf("after key %q activeView = %d, want %d", tt.key, m.activeView, tt.want)
			}
		})
	}
}

func TestQualityViewIsUnreachableWhenUnavailable(t *testing.T) {
	m := newTestModel(false)
	m.Update(key("4"))
	if m.activeView == ViewQuality {
		t.Error("key 4 selected the Quality view even though it is unavailable")
	}
}

func TestTabCyclesForward(t *testing.T) {
	m := newTestModel(true)

	want := []ViewMode{ViewPorts, ViewDiagnostics, ViewQuality, ViewTools, ViewOverview}
	for i, w := range want {
		m.Update(key("tab"))
		if m.activeView != w {
			t.Fatalf("tab press %d: activeView = %d, want %d", i+1, m.activeView, w)
		}
	}
}

func TestTabSkipsQualityWhenUnavailable(t *testing.T) {
	m := newTestModel(false)
	m.activeView = ViewDiagnostics

	m.Update(key("tab"))
	if m.activeView == ViewQuality {
		t.Fatal("tab landed on the Quality view even though it is unavailable")
	}
	if m.activeView != ViewTools {
		t.Errorf("tab from Diagnostics = %d, want ViewTools (%d)", m.activeView, ViewTools)
	}
}

func TestShiftTabCyclesBackward(t *testing.T) {
	m := newTestModel(true)
	m.activeView = ViewOverview

	m.Update(key("shift+tab"))
	if m.activeView != ViewTools {
		t.Errorf("shift+tab from Overview = %d, want ViewTools (%d) by wrapping", m.activeView, ViewTools)
	}
}

func TestShiftTabSkipsQualityWhenUnavailable(t *testing.T) {
	m := newTestModel(false)
	m.activeView = ViewTools

	m.Update(key("shift+tab"))
	if m.activeView == ViewQuality {
		t.Fatal("shift+tab landed on the Quality view even though it is unavailable")
	}
	if m.activeView != ViewDiagnostics {
		t.Errorf("shift+tab from Tools = %d, want ViewDiagnostics (%d)", m.activeView, ViewDiagnostics)
	}
}

func TestWindowSizeIsRecorded(t *testing.T) {
	m := newTestModel(true)
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if m.width != 120 || m.height != 40 {
		t.Errorf("got width=%d height=%d, want 120x40", m.width, m.height)
	}
}

func TestInputModeSwallowsNavigationKeys(t *testing.T) {
	m := newTestModel(true)
	m.activeView = ViewDiagnostics
	m.diagInputActive = true

	m.Update(key("1"))

	if m.activeView != ViewDiagnostics {
		t.Errorf("a navigation key changed the view while the input field was active; activeView = %d", m.activeView)
	}
}

func TestNetMsgClampsCursor(t *testing.T) {
	m := newTestModel(true)
	m.cursor = 9

	m.Update(netMsg{gateway: "192.168.1.1", note: "0 interfaces found"})

	if m.cursor != 0 {
		t.Errorf("cursor = %d after a refresh returning no interfaces, want 0", m.cursor)
	}
	if m.gateway != "192.168.1.1" {
		t.Errorf("gateway = %q, want it taken from the message", m.gateway)
	}
}

func TestPortsMsgClearsLoading(t *testing.T) {
	m := newTestModel(true)
	m.portsLoading = true

	m.Update(portsMsg{ports: []PortInfo{{Command: "node", Port: "3000"}}})

	if m.portsLoading {
		t.Error("portsLoading is still true after the ports message arrived")
	}
	if len(m.listeningPorts) != 1 {
		t.Errorf("listeningPorts has %d entries, want 1", len(m.listeningPorts))
	}
}

func TestHasOpenModalTracksInputFields(t *testing.T) {
	m := newTestModel(true)
	if m.HasOpenModal() {
		t.Error("HasOpenModal is true with no input field active")
	}

	m.diagInputActive = true
	if !m.HasOpenModal() {
		t.Error("HasOpenModal is false while the diagnostics input is active")
	}

	m.diagInputActive = false
	m.toolInputActive = true
	if !m.HasOpenModal() {
		t.Error("HasOpenModal is false while the tools input is active")
	}
}
