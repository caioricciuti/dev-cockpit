package diagnostics

import (
	"sync"
	"time"
)

// Severity represents the health level of a check
type Severity int

const (
	OK       Severity = iota
	Warning
	Critical
)

func (s Severity) String() string {
	switch s {
	case OK:
		return "OK"
	case Warning:
		return "WARN"
	case Critical:
		return "CRIT"
	default:
		return "?"
	}
}

// CheckResult holds the outcome of a single diagnostic check
type CheckResult struct {
	Category    string
	Status      Severity
	Summary     string
	Details     []string
	Suggestions []string
}

// Report holds all diagnostic results with a computed score
type Report struct {
	Checks    []CheckResult
	Score     int
	Grade     string
	Timestamp time.Time
}

// Category weights for scoring (must sum to 100)
var weights = map[string]int{
	"Disk":        20,
	"Storage":     15,
	"Performance": 25,
	"Network":     10,
	"Services":    10,
	"Security":    20,
}

// RunAll executes all diagnostic checks concurrently and returns a scored report
func RunAll() Report {
	checks := make([]CheckResult, 6)
	var wg sync.WaitGroup
	wg.Add(6)

	go func() { defer wg.Done(); checks[0] = CheckDisk() }()
	go func() { defer wg.Done(); checks[1] = CheckStorage() }()
	go func() { defer wg.Done(); checks[2] = CheckPerformance() }()
	go func() { defer wg.Done(); checks[3] = CheckNetwork() }()
	go func() { defer wg.Done(); checks[4] = CheckServices() }()
	go func() { defer wg.Done(); checks[5] = CheckSecurity() }()

	wg.Wait()

	score := calculateScore(checks)
	return Report{
		Checks:    checks,
		Score:     score,
		Grade:     CalculateGrade(score),
		Timestamp: time.Now(),
	}
}

func calculateScore(checks []CheckResult) int {
	total := 0
	for _, c := range checks {
		w := weights[c.Category]
		switch c.Status {
		case OK:
			total += w
		case Warning:
			total += w / 2
		case Critical:
			// 0 points
		}
	}
	return total
}

// CalculateGrade maps a 0-100 score to a letter grade
func CalculateGrade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 75:
		return "B"
	case score >= 60:
		return "C"
	case score >= 40:
		return "D"
	default:
		return "F"
	}
}
