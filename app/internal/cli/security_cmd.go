package cli

import (
	"fmt"

	"github.com/caioricciuti/dev-cockpit/internal/diagnostics"
)

func cmdSecurity() {
	printHeader("Dev Cockpit — Security Status")

	result := diagnostics.CheckSecurity()

	printStatus("Security", result.Status, result.Summary)
	fmt.Fprintln(stdout)

	for _, d := range result.Details {
		fmt.Fprintf(stdout, "  %s\n", valueStyle.Render(d))
	}

	if len(result.Suggestions) > 0 {
		fmt.Fprintln(stdout)
		for _, s := range result.Suggestions {
			fmt.Fprintf(stdout, "  %s %s\n", warnStyle.Render("→"), s)
		}
	}
	fmt.Fprintln(stdout)
}
