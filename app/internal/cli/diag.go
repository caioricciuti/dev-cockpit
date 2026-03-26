package cli

import (
	"fmt"

	"github.com/caioricciuti/dev-cockpit/internal/diagnostics"
)

func cmdDiag() {
	printHeader("Dev Cockpit — Full Diagnostics Report")

	report := diagnostics.RunAll()

	gradeStyle := okStyle
	if report.Score < 75 {
		gradeStyle = warnStyle
	}
	if report.Score < 40 {
		gradeStyle = critStyle
	}

	scoreBar := renderScoreBar(report.Score)
	fmt.Printf("  %s %s %s\n\n",
		labelStyle.Render("Health Score"),
		scoreBar,
		gradeStyle.Bold(true).Render(fmt.Sprintf("%d/100 [%s]", report.Score, report.Grade)),
	)

	for _, c := range report.Checks {
		printStatus(c.Category, c.Status, c.Summary)

		// Show details
		for _, d := range c.Details {
			fmt.Printf("      %s\n", mutedStyle.Render(d))
		}

		// Show suggestions for non-OK checks
		if c.Status != diagnostics.OK && len(c.Suggestions) > 0 {
			for _, s := range c.Suggestions {
				fmt.Printf("      %s %s\n", warnStyle.Render("→"), s)
			}
		}
		fmt.Println()
	}

	fmt.Printf("  %s\n\n", mutedStyle.Render(fmt.Sprintf("Completed at %s", report.Timestamp.Format("15:04:05"))))
}
