package dashboard

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/caioricciuti/dev-cockpit/internal/config"
)

func newTestModel(t *testing.T, width int) *Model {
	t.Helper()
	cfg := &config.Config{
		Storage: config.StorageConfig{DataDir: t.TempDir(), MaxHistoryDays: 1},
	}
	m := New(cfg, nil)
	m.Update(tea.WindowSizeMsg{Width: width, Height: 40})
	return m
}

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

// Before the first sample arrives every gauge read "0.0%" next to a green
// "● Normal", which is indistinguishable from a genuinely idle machine.

func TestMetricsSayMeasuringBeforeTheFirstSample(t *testing.T) {
	m := newTestModel(t, 120)
	out := stripANSI(m.renderMetrics())

	if !strings.Contains(out, "measuring") {
		t.Errorf("a freshly started dashboard should say it is measuring:\n%s", out)
	}
	if strings.Contains(out, "0.0%") {
		t.Errorf("an unmeasured gauge must not report 0.0%%:\n%s", out)
	}
	for _, claim := range []string{"● Normal", "● Healthy"} {
		if strings.Contains(out, claim) {
			t.Errorf("an unmeasured gauge must not claim %q:\n%s", claim, out)
		}
	}
}

func TestMetricsShowValuesOnceSampled(t *testing.T) {
	m := newTestModel(t, 120)
	m.updateMetrics(metricsMsg{cpu: []float64{12.5, 17.5}, memory: 64.0, disk: 23.9})

	out := stripANSI(m.renderMetrics())
	if strings.Contains(out, "measuring") {
		t.Errorf("the dashboard should stop measuring once a sample arrives:\n%s", out)
	}
	for _, want := range []string{"15.0%", "64.0%", "23.9%"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the sampled dashboard:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "● Normal") {
		t.Errorf("a 15%% CPU reading should be reported as normal:\n%s", out)
	}
}

func TestMetricsReportThresholds(t *testing.T) {
	tests := []struct {
		name   string
		msg    metricsMsg
		expect string
	}{
		{"cpu critical", metricsMsg{cpu: []float64{92}, memory: 10, disk: 10}, "● Critical"},
		{"cpu high", metricsMsg{cpu: []float64{75}, memory: 10, disk: 10}, "● High"},
		{"memory critical", metricsMsg{cpu: []float64{1}, memory: 95, disk: 10}, "● Critical"},
		{"disk low space", metricsMsg{cpu: []float64{1}, memory: 10, disk: 85}, "● Low Space"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel(t, 120)
			m.updateMetrics(tt.msg)
			if out := stripANSI(m.renderMetrics()); !strings.Contains(out, tt.expect) {
				t.Errorf("expected %q:\n%s", tt.expect, out)
			}
		})
	}
}

// The separator was a hard-coded 60 cells, so it neither filled a wide
// terminal nor fitted a narrow one.
func TestMetricSeparatorFitsTheWidth(t *testing.T) {
	for _, width := range []int{60, 80, 120, 200} {
		m := newTestModel(t, width)
		for _, line := range strings.Split(stripANSI(m.renderMetrics()), "\n") {
			if !strings.Contains(line, "━") {
				continue
			}
			if got := lipgloss.Width(line); got > width {
				t.Errorf("width %d: separator line is %d cells, want <= %d", width, got, got)
			}
			if strings.TrimSpace(line) == "━━━━" {
				t.Errorf("width %d: separator wrapped onto a second line", width)
			}
		}
	}
}
