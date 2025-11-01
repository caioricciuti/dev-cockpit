package main

import (
	"fmt"
	"log"
	"os"

	"github.com/caioricciuti/dev-cockpit/internal/app"
	"github.com/caioricciuti/dev-cockpit/internal/config"
	"github.com/caioricciuti/dev-cockpit/internal/logger"
	"github.com/caioricciuti/dev-cockpit/internal/modules/quickactions"
	tea "github.com/charmbracelet/bubbletea"
)

// version is injected at build time via -ldflags
var version = "dev"

func main() {
	// Debug logging off by default; enable with --debug
	debugMode := false
	for _, arg := range os.Args {
		switch arg {
		case "--debug":
			debugMode = true
		case "--no-debug":
			debugMode = false
		}
	}

	// Initialize logger
	if err := logger.Initialize(debugMode); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.GetLogger().Close()

	fmt.Printf("Debug logging: %v\n", debugMode)
	fmt.Printf("Log file: %s\n", logger.GetLogPath())
	if debugMode {
		fmt.Println("Tail logs in another terminal with:")
		fmt.Printf("  tail -f %s\n\n", logger.GetLogPath())
	} else {
		fmt.Println("Run with --debug to stream logs to the console.")
		fmt.Println()
	}

	logger.Info("Starting Dev Cockpit v%s", version)

	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration: %v", err)
		log.Fatal("Failed to load configuration:", err)
	}
	logger.Info("Configuration loaded successfully")

	// Check for command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-v":
			fmt.Printf("Dev Cockpit v%s\n", version)
			os.Exit(0)
		case "cleanup":
			// Minimal CLI for testing cleanup operations
			if len(os.Args) > 2 {
				sub := os.Args[2]
				switch sub {
				case "empty-trash", "--empty-trash":
					// Use quickactions implementation
					if err := quickactions.EmptyTrash(); err != nil {
						fmt.Printf("Empty Trash failed: %v\n", err)
						os.Exit(1)
					}
					fmt.Println("Trash emptied successfully.")
					os.Exit(0)
				}
			}
			fmt.Println("Usage: devcockpit cleanup empty-trash")
			os.Exit(1)
		case "help", "--help", "-h":
			showHelp()
			os.Exit(0)
		case "logs", "--logs":
			fmt.Printf("Log file location: %s\n", logger.GetLogPath())
			os.Exit(0)
		}
	}

	// Create the main application
	application := app.New(cfg, version)

	// Initialize Bubble Tea program
	p := tea.NewProgram(
		application,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Fatal("Error running program:", err)
	}
}

func showHelp() {
	fmt.Printf(`
Dev Cockpit v%s - Professional macOS Development Command Center

Usage:
  devcockpit              Launch the interactive TUI
  devcockpit --debug      Launch with debug logging enabled
  devcockpit version      Show version information
  devcockpit logs         Show log file location
  devcockpit help         Show this help message

Debug Mode:
  Debug logging is OFF by default; use --debug to enable
  tail -f ~/.devcockpit/debug.log    View logs in real-time

Keyboard Shortcuts (in app):
  Tab/Shift+Tab    Navigate between modules
  ←/→/Home/End     Navigate modules
  Enter            Select/Execute
  Esc              Return to module switcher
  l                Toggle logs overlay
  q                Quit current view
  Q                Quit application
  ?                Show help

For more information: https://devcockpit.dev
Support: support@devcockpit.dev
`, version)
}
