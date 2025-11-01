# Dev Cockpit

Dev Cockpit is a Go-powered terminal UI that turns any Apple Silicon Mac into a development and operations command center. The project is now fully open source and sustained by community donations.

## Project Structure

```
dev-cockpit/
├── app/                 # Go module containing the Dev Cockpit binary
│   ├── cmd/devcockpit/ # Main application entry point
│   ├── internal/       # Internal packages and modules
│   ├── Makefile        # Build automation
│   └── go.mod          # Go dependencies
├── LICENSE
└── README.md           # This file
```

## Prerequisites

- macOS 11.0+ running on Apple Silicon (M1/M2/M3)
- Go 1.21 or newer (`brew install go`)
- Xcode Command Line Tools (`xcode-select --install`)

## Installation

### Quick Install (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/caioricciuti/dev-cockpit/main/install.sh | bash
```

This will:
- Download the latest release
- Verify the binary
- Install to `/usr/local/bin/devcockpit`
- Create configuration directory

### Manual Installation

```bash
# Download latest release
curl -L -o devcockpit https://github.com/caioricciuti/dev-cockpit/releases/latest/download/devcockpit-darwin-arm64

# Make executable and install
chmod +x devcockpit
sudo mv devcockpit /usr/local/bin/

# Create config directory
mkdir -p ~/.devcockpit
```

### Usage

```bash
devcockpit              # Launch the TUI
devcockpit --help       # Show help
devcockpit --version    # Show version
devcockpit --debug      # Launch with debug logging
```

### CLI Commands

Dev Cockpit also supports direct CLI commands for quick operations:

```bash
devcockpit cleanup empty-trash    # Empty trash without TUI
```

### Keyboard Shortcuts

**Global Navigation:**
- `Tab` / `Shift+Tab` - Switch between modules
- `Enter` - Focus on selected module
- `Esc` - Exit focused module / Go back
- `Q` - Quit application (from module switcher)
- `?` - Show help

**Module-Specific:**
- `↑/↓` or `k/j` - Navigate lists
- `Space` - Toggle selection (Cleanup module)
- `R` - Refresh/Reload data
- `L` - List packages (Packages module)
- `C` - Cleanup cache (Packages module)
- `U` - Update manager (Packages module)
- `F` - Fix all issues (Quick Actions module)

## Build from Source

```bash
# Clone and enter the repo
git clone https://github.com/caioricciuti/dev-cockpit.git
cd dev-cockpit/app

# Fetch dependencies
make deps

# Build an optimized arm64 binary
make build

# Run the TUI straight from source (handy during development)
make run

# Optional: install system-wide (prompts for sudo)
make install
```

The compiled binary lives at `app/build/devcockpit`. Launch it directly or run `devcockpit` after installation.

## Testing & Tooling

```bash
make test       # unit tests with coverage output
make test-race  # race detector
make fmt        # gofmt across the module
make lint       # requires golangci-lint in PATH
```

## Troubleshooting

### npm/Node not detected

If you're using NVM (Node Version Manager) and npm shows as "unknown":

1. Make sure you have a default Node version set:
   ```bash
   nvm alias default node
   ```

2. Restart Dev Cockpit - it will automatically detect NVM installations

3. If still not detected, check your NVM installation:
   ```bash
   ls ~/.nvm/versions/node/
   ```

### Sudo/Permission issues

Some operations (like system maintenance) require sudo access:

1. Make sure your user has sudo privileges
2. If prompted, enter your password when requested
3. Press Ctrl+C to cancel any sudo prompt if needed

### Terminal compatibility

Dev Cockpit works best with modern terminals:

- **Recommended**: iTerm2, Warp, Terminal.app (macOS 11+)
- **Minimum**: 80x24 terminal size
- **Colors**: 256-color support recommended

If you experience rendering issues:
```bash
devcockpit --debug  # Check debug.log for errors
```

### Docker not connecting

If Docker shows as unavailable:

1. Make sure Docker Desktop is running
2. Check Docker socket exists:
   ```bash
   ls -la /var/run/docker.sock
   ```

3. Verify Docker CLI works:
   ```bash
   docker ps
   ```

## Community Support

Dev Cockpit is free for everyone. If it saves you time, please consider helping to keep development moving:

- Sponsor ongoing work: https://github.com/sponsors/caioricciuti
- Leave a one-time tip: https://buymeacoffee.com/caioricciuti
- Share feedback or bug reports: https://github.com/caioricciuti/dev-cockpit/issues
- Star the repository and spread the word with other Apple Silicon users

## Support

- Issues: https://github.com/caioricciuti/dev-cockpit/issues
- Docs:  https://devcockpit.app/getting-started

Enjoy the cockpit! Feedback and contributions are always welcome.
