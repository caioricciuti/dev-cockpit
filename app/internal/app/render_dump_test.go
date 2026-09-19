package app

import (
	"os"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestDumpView writes the rendered TUI to a file so it can be inspected
// without a terminal. Skipped unless DEVCOCKPIT_DUMP_VIEW is set to the
// destination path.
//
//	DEVCOCKPIT_DUMP_VIEW=/tmp/view.txt go test -run TestDumpView ./internal/app/
func TestDumpView(t *testing.T) {
	dest := os.Getenv("DEVCOCKPIT_DUMP_VIEW")
	if dest == "" {
		t.Skip("set DEVCOCKPIT_DUMP_VIEW to a path to dump the rendered view")
	}

	width, height := 120, 40
	if w := os.Getenv("DEVCOCKPIT_DUMP_WIDTH"); w != "" {
		switch w {
		case "80":
			width, height = 80, 24
		case "100":
			width, height = 100, 30
		case "160":
			width, height = 160, 50
		}
	}

	m := New(testConfig(t), "test")
	defer m.Close()
	m.Update(tea.WindowSizeMsg{Width: width, Height: height})

	if err := os.WriteFile(dest, []byte(m.View().Content), 0o600); err != nil {
		t.Fatalf("failed to write dump: %v", err)
	}
	t.Logf("wrote %dx%d view to %s", width, height, dest)
}
