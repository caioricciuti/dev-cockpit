package cli

import (
	"io"
	"os"

	"github.com/charmbracelet/colorprofile"
)

// newOutputWriter wraps w so that ANSI output is downgraded or stripped to
// match what the destination actually supports.
//
// lipgloss v2 removed the global renderer that used to do this on its own, so
// a style now renders truecolor escapes unconditionally. Without this wrapper
// `devcockpit status > report.txt` and `devcockpit ps | grep node` would carry
// escape codes, which they did not under lipgloss v1.
func newOutputWriter(w io.Writer, environ []string) io.Writer {
	return colorprofile.NewWriter(w, environ)
}

// stdout is the destination for every CLI command's output.
var stdout io.Writer = newOutputWriter(os.Stdout, os.Environ())
