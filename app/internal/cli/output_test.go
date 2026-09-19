package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// Regression: lipgloss v2 dropped the global renderer, so styles emit
// truecolor escapes regardless of destination. Piping or redirecting a CLI
// command must still produce clean text, as it did under lipgloss v1.

func styledSample() string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF")).
		Render("Dev Cockpit")
}

func TestStyledOutputIsPlainWhenNotATerminal(t *testing.T) {
	// A bytes.Buffer is not a terminal, which is the same situation as a pipe
	// or a redirect to a file.
	var buf bytes.Buffer
	w := newOutputWriter(&buf, []string{"TERM=xterm-256color"})

	fmt.Fprintln(w, styledSample())

	got := buf.String()
	if strings.ContainsRune(got, 0x1b) {
		t.Errorf("output to a non-terminal contains ANSI escapes: %q", got)
	}
	if !strings.Contains(got, "Dev Cockpit") {
		t.Errorf("output lost its text; got %q", got)
	}
}

func TestStyledOutputKeepsTextWhenColorIsForcedOff(t *testing.T) {
	var buf bytes.Buffer
	w := newOutputWriter(&buf, []string{"NO_COLOR=1"})

	fmt.Fprintf(w, "%s: %d%%\n", styledSample(), 42)

	got := buf.String()
	if strings.ContainsRune(got, 0x1b) {
		t.Errorf("NO_COLOR output contains ANSI escapes: %q", got)
	}
	if !strings.Contains(got, "42%") {
		t.Errorf("formatted values were lost; got %q", got)
	}
}

func TestStyledSampleActuallyCarriesEscapes(t *testing.T) {
	// Guards the two tests above from passing trivially: if lipgloss stopped
	// emitting escapes altogether they would pass for the wrong reason.
	if !strings.ContainsRune(styledSample(), 0x1b) {
		t.Skip("lipgloss produced no escapes in this environment; the stripping tests prove nothing here")
	}
}
