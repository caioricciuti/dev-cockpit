package cli

import (
	"fmt"

	"github.com/caioricciuti/dev-cockpit/internal/diagnostics"
)

func cmdSecurity() {
	printHeader("Dev Cockpit — Security Status")

	result := diagnostics.CheckSecurity()

	printStatus("Security", result.Status, result.Summary)
	fmt.Println()

	for _, d := range result.Details {
		fmt.Printf("  %s\n", valueStyle.Render(d))
	}

	if len(result.Suggestions) > 0 {
		fmt.Println()
		for _, s := range result.Suggestions {
			fmt.Printf("  %s %s\n", warnStyle.Render("→"), s)
		}
	}
	fmt.Println()
}
