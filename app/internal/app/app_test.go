package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/caioricciuti/dev-cockpit/internal/config"
)

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		Theme:          "default",
		UpdateInterval: 1,
		LogLevel:       "error",
		Storage: config.StorageConfig{
			DataDir:        t.TempDir(),
			MaxHistoryDays: 1,
		},
	}
}

func newSizedModel(t *testing.T) *Model {
	t.Helper()
	m := New(testConfig(t), "test")
	t.Cleanup(func() { m.Close() })

	// Nothing renders until the model knows how big the terminal is.
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return m
}

// bubbletea v2 replaced View() string with View() tea.View, and moved alt
// screen and mouse mode off ProgramOption and onto that View. If either
// regressed the program would still compile and still run, but would render
// into the primary screen buffer with no mouse support.

func TestViewRendersContent(t *testing.T) {
	m := newSizedModel(t)

	v := m.View()
	if strings.TrimSpace(v.Content) == "" {
		t.Fatal("View returned empty content for a 120x40 terminal")
	}
}

func TestViewRequestsAltScreenAndMouse(t *testing.T) {
	m := newSizedModel(t)

	v := m.View()
	if !v.AltScreen {
		t.Error("View does not request the alternate screen buffer")
	}
	if v.MouseMode != tea.MouseModeCellMotion {
		t.Errorf("View MouseMode = %v, want MouseModeCellMotion", v.MouseMode)
	}
}

func TestQuittingViewIsStillAValidView(t *testing.T) {
	m := newSizedModel(t)
	m.quitting = true

	v := m.View()
	if !strings.Contains(v.Content, "Dev Cockpit") {
		t.Errorf("quitting view lost its message; got %q", v.Content)
	}
}

func TestModelSatisfiesTeaModel(t *testing.T) {
	// Compile-time guarantee that the v2 interface is still met. A signature
	// drift here is exactly what broke during the v1 to v2 migration.
	var _ tea.Model = (*Model)(nil)
}
