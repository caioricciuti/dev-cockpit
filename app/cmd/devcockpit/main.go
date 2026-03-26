package main

import (
	"fmt"
	"log"
	"os"

	"github.com/caioricciuti/dev-cockpit/internal/app"
	"github.com/caioricciuti/dev-cockpit/internal/cli"
	"github.com/caioricciuti/dev-cockpit/internal/config"
	"github.com/caioricciuti/dev-cockpit/internal/logger"
	"github.com/caioricciuti/dev-cockpit/internal/modules/quickactions"
	"github.com/caioricciuti/dev-cockpit/internal/uninstaller"
	"github.com/caioricciuti/dev-cockpit/internal/updater"
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

	// Check for command line arguments FIRST (before logger initialization)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-v":
			fmt.Printf("Dev Cockpit v%s\n", version)
			os.Exit(0)
		case "cleanup":
			// Handle empty-trash here; other subcommands (list) handled by CLI router
			if len(os.Args) > 2 {
				sub := os.Args[2]
				switch sub {
				case "empty-trash", "--empty-trash":
					if err := quickactions.EmptyTrash(); err != nil {
						fmt.Printf("Empty Trash failed: %v\n", err)
						os.Exit(1)
					}
					fmt.Println("Trash emptied successfully.")
					os.Exit(0)
				}
			}
			// Don't exit — let CLI router handle other cleanup subcommands
		case "help", "--help", "-h":
			showHelp()
			os.Exit(0)
		case "logs", "--logs":
			// Initialize logger just to get the path
			if err := logger.Initialize(false); err != nil {
				log.Fatal("Failed to initialize logger:", err)
			}
			fmt.Printf("Log file location: %s\n", logger.GetLogPath())
			os.Exit(0)
		case "uninstall", "--uninstall":
			// Check for --force flag
			force := false
			for _, arg := range os.Args[2:] {
				if arg == "--force" || arg == "-f" {
					force = true
					break
				}
			}

			// Perform uninstallation
			if err := uninstaller.Uninstall(force); err != nil {
				fmt.Printf("Uninstall failed: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		case "update", "--update":
			// Parse flags
			force := false
			checkOnly := false
			for _, arg := range os.Args[2:] {
				switch arg {
				case "--force", "-f":
					force = true
				case "--check":
					checkOnly = true
				}
			}

			// Perform update
			opts := updater.UpdateOptions{
				Force:      force,
				CheckOnly:  checkOnly,
				CurrentVer: version,
			}

			if err := updater.Update(opts); err != nil {
				fmt.Printf("Update failed: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	// Try CLI commands before launching TUI
	if len(os.Args) > 1 {
		if cli.Run(os.Args[1:], version) {
			os.Exit(0)
		}
	}

	// Initialize logger (only when launching TUI)
	if err := logger.Initialize(debugMode); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.GetLogger().Close()

	// Show debug info only when launching TUI
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

	// Create the main application
	application := app.New(cfg, version)
	defer application.Close()

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
	fmt.Printf(`Dev Cockpit v%s - Development Command Center

USAGE:
  devcockpit [flags]
  devcockpit cleanup empty-trash
  devcockpit uninstall [--force]
  devcockpit update [--check | --force]

AVAILABLE TUI MODULES:
  Dashboard       Real-time system monitoring (CPU, GPU, Memory, Disk, Network)
  Processes       Interactive process manager with search and kill
  Services        Dev service monitor (Homebrew + port detection)
  Quick Actions   One-tap maintenance and optimization tasks
  Cleanup         Free up disk space (caches, logs, trash, downloads)
  Packages        Manage Homebrew, npm, and other package managers
  System          Hardware info, diagnostics, and system details
  Docker          Container management and cleanup
  Network         Interface analysis and connectivity diagnostics
  Security        Firewall, FileVault, and SIP status
  Diagnostics     Health score report with expandable checks
  Logs            Multi-source log aggregator
  Support         Project support and sponsorship information

CLI COMMANDS:
  devcockpit                       Launch interactive TUI
  devcockpit status                Quick system status + health score
  devcockpit diag                  Full diagnostics report
  devcockpit ps [--sort cpu|mem|pid] [--top N]
                                   Process list (default: top 15 by CPU)
  devcockpit services              Homebrew service status
  devcockpit security              Security feature status
  devcockpit cleanup list          Show cache sizes
  devcockpit cleanup empty-trash   Empty the trash (CLI mode)
  devcockpit update                Update to the latest version
  devcockpit update --check        Check for updates without installing
  devcockpit update --force        Update without confirmation prompts
  devcockpit uninstall             Uninstall Dev Cockpit from the system
  devcockpit uninstall --force     Uninstall without confirmation prompts
  devcockpit --help, -h            Show this help message
  devcockpit --version, -v         Show version information
  devcockpit --debug               Launch with debug logging
  devcockpit --logs                Show debug log file location

EXAMPLES:
  devcockpit                      # Start the interactive interface
  devcockpit status               # Quick health check
  devcockpit diag                 # Full diagnostics report
  devcockpit ps --sort mem --top 5  # Top 5 processes by memory
  devcockpit services             # Check dev service status
  devcockpit security             # Check security features
  devcockpit cleanup list         # See cache sizes
  devcockpit --debug              # Launch with live debug output
  devcockpit cleanup empty-trash  # Empty trash from command line
  devcockpit update               # Update to the latest version
  devcockpit uninstall            # Uninstall Dev Cockpit

CONFIGURATION:
  Config: ~/.devcockpit/config.yaml
  Logs:   ~/.devcockpit/debug.log

KEYBOARD SHORTCUTS (in TUI):
  1-9         Jump to module
  Tab         Cycle through modules
  ↑/↓         Navigate lists
  Enter       Select/Execute
  ESC         Go back / Close modal
  q, Ctrl+C   Quit

DOCUMENTATION:
  Website: https://devcockpit.app
  GitHub:  https://github.com/caioricciuti/dev-cockpit
  Issues:  https://github.com/caioricciuti/dev-cockpit/issues

SUPPORT:
  Sponsor: https://github.com/sponsors/caioricciuti
  Donate:  https://buymeacoffee.com/caioricciuti

Pro Tip: Run 'devcockpit' to explore all features interactively!
`, version)
}
