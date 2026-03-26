# Getting Started

Welcome to Dev Cockpit! This guide will help you get up and running quickly on macOS or Linux.

## Prerequisites

Before installing Dev Cockpit, ensure your system meets these requirements:

**macOS:**
- Apple Silicon Mac (M1/M2/M3/M4 series)
- macOS 11.0 (Big Sur) or later
- Terminal application (iTerm2, kitty, or WezTerm recommended)

**Linux:**
- x86_64 (amd64) or ARM64 (aarch64) processor
- Modern distribution (Ubuntu 20.04+, Fedora 36+, Arch, etc.)
- Terminal with true color support (kitty, alacritty, WezTerm recommended)

**Both:**
- Internet connection (for installation)

## Installation

### Quick Install (Recommended)

The easiest way to install Dev Cockpit is using our installation script:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/caioricciuti/dev-cockpit/main/install.sh)"
```

This script will:
1. Download the latest release binary
2. Install it to `/usr/local/bin/devcockpit`
3. Make it executable
4. Verify the installation

### Manual Installation

If you prefer to install manually:

1. **Download the latest release:**

   Visit the [GitHub Releases page](https://github.com/caioricciuti/dev-cockpit/releases) and download the latest `devcockpit` binary.

2. **Make it executable:**
   ```bash
   chmod +x devcockpit
   ```

3. **Move to system path:**
   ```bash
   sudo mv devcockpit /usr/local/bin/
   ```

4. **Verify installation:**
   ```bash
   devcockpit --version
   ```

### Build from Source

For developers who want to build from source:

1. **Clone the repository:**
   ```bash
   git clone https://github.com/caioricciuti/dev-cockpit.git
   cd dev-cockpit/app
   ```

2. **Install dependencies:**
   ```bash
   make deps
   ```

3. **Build the binary:**
   ```bash
   make build
   ```

4. **Install system-wide (optional):**
   ```bash
   make install
   ```

5. **Or run locally:**
   ```bash
   ./build/devcockpit
   ```

## First Run

After installation, launch Dev Cockpit:

```bash
devcockpit
```

On first run, Dev Cockpit will:
- Create a configuration directory at `~/.devcockpit/`
- Scan your system for installed tools (Homebrew, npm, Docker, etc.)
- Display the main dashboard with system metrics

## Interface Overview

Dev Cockpit uses a Text User Interface (TUI) with the following layout:

```
┌─────────────────────────────────────────────┐
│  Dashboard | Cleanup | Packages | ...       │  ← Module Tabs
├─────────────────────────────────────────────┤
│                                             │
│          Module Content Area                │
│                                             │
│                                             │
└─────────────────────────────────────────────┘
```

### Navigation

**Module Switching:**
- **Number keys (1-9):** Jump directly to a module
- **Tab:** Cycle through modules
- **← →:** Navigate left/right between modules

**Module Navigation:**
- **↑ ↓:** Move up/down in lists
- **Enter:** Select/execute current item
- **Space:** Alternative select key
- **ESC:** Go back / Close modal / Return to switcher

**General:**
- **q or Ctrl+C:** Quit Dev Cockpit
- **?:** Show help (where available)

## Modules

Dev Cockpit includes these modules:

1. **Dashboard** - Real-time system monitoring (CPU, GPU, Memory, Disk, Network)
2. **Cleanup** - Remove system junk and free up disk space
3. **Packages** - Manage Homebrew, npm, and other package managers
4. **Docker** - Monitor and manage Docker containers
5. **Process Manager** - View and manage running processes
6. **Dev Services** - Monitor Homebrew services and dev tools
7. **Quick Actions** - Common system fix tasks (platform-aware)
8. **Network** - Network diagnostics, port scanning, and quality tests
9. **Security** - Security audits (Firewall, Disk Encryption, SIP/Gatekeeper on macOS)
10. **System** - System information and diagnostics
11. **Diagnostics** - Automated health check with A-F grading
12. **Logs** - Aggregated log viewer (system, Homebrew, Docker, app)
13. **Metrics** - Historical performance tracking with SQLite
14. **Support** - Support the project

## Package Manager Detection

Dev Cockpit automatically detects and integrates with:

### Homebrew
- **macOS:** Automatically detected at `/opt/homebrew/bin/brew` (Apple Silicon) or `/usr/local/bin/brew`
- **Linux:** Detected at `/home/linuxbrew/.linuxbrew/bin/brew`
- If not detected, install with:
  ```bash
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  ```

### npm (Node Package Manager)
- Supports **NVM** (Node Version Manager) installations at `~/.nvm/`
- Supports **Homebrew** Node installations
- Supports **system** Node installations
- If using NVM, ensure default is set:
  ```bash
  nvm alias default node
  ```

### Docker
- Requires Docker Desktop (macOS) or Docker Engine (Linux)
- Socket expected at `/var/run/docker.sock`
- **macOS:** Install Docker Desktop from [docker.com](https://www.docker.com/products/docker-desktop)
- **Linux:** Install via your package manager (`apt install docker.io`, `dnf install docker`, etc.)

## Configuration

Dev Cockpit stores its configuration in `~/.devcockpit/`:

```
~/.devcockpit/
├── config.yaml      # Main configuration
└── debug.log        # Debug logs (if --debug enabled)
```

Currently, most settings are auto-detected and don't require manual configuration.

## CLI Commands

Dev Cockpit supports command-line arguments:

**Display help:**
```bash
devcockpit --help
devcockpit -h
```

**Show version:**
```bash
devcockpit --version
devcockpit -v
```

**Enable debug mode:**
```bash
devcockpit --debug
```

**CLI Power Mode:**
```bash
devcockpit status          # Quick system health overview
devcockpit diag            # Full diagnostics report
devcockpit ps              # Top processes by CPU
devcockpit ps --sort mem   # Top processes by memory
devcockpit services        # Homebrew services status
devcockpit security        # Security check
devcockpit cleanup list    # Cache sizes
devcockpit cleanup empty-trash  # Empty trash
```

**Uninstall Dev Cockpit:**
```bash
devcockpit uninstall              # Interactive uninstall with prompts
devcockpit uninstall --force      # Uninstall without confirmation
```

**Show log file location:**
```bash
devcockpit --logs
```

## Tips for Best Experience

1. **Use a modern terminal:**
   - **macOS:** iTerm2, kitty, or WezTerm recommended
   - **Linux:** kitty, alacritty, or WezTerm recommended
   - Needs true color support for the best visual experience

2. **Recommended terminal size:**
   - Minimum: 80 characters × 24 lines
   - Recommended: 120 characters × 40 lines for best experience

3. **Font recommendations:**
   - Fira Code (with ligatures)
   - JetBrains Mono
   - Menlo (default macOS monospace)
   - SF Mono

4. **Enable sudo access:**
   - Some cleanup operations require sudo
   - You'll be prompted when needed
   - Or run with: `sudo devcockpit`

5. **Regular maintenance:**
   - Run cleanup weekly to keep your system healthy
   - Monitor system metrics to catch issues early
   - Update packages regularly through the Packages module

## Next Steps

Now that you have Dev Cockpit installed:

1. **Explore the Dashboard** to see your system metrics in real-time
2. **Run a Cleanup** to free up disk space
3. **Check your Packages** to see what's installed
4. **Customize your terminal** for the best visual experience

## Troubleshooting

If you encounter any issues:

- Check the [Troubleshooting Guide](/troubleshooting)
- Review logs at `~/.devcockpit/debug.log`
- Create an issue on [GitHub](https://github.com/caioricciuti/dev-cockpit/issues)

## Uninstalling

If you need to uninstall Dev Cockpit, use the built-in uninstall command:

```bash
devcockpit uninstall
```

This will:
- Stop Dev Cockpit if running
- Remove the binary from `/usr/local/bin/devcockpit`
- Prompt to remove configuration directory (`~/.devcockpit/`)
- Clean up temporary files

For non-interactive uninstallation:
```bash
devcockpit uninstall --force
```

**Manual uninstallation** (if needed):
```bash
# Remove binary
sudo rm /usr/local/bin/devcockpit

# Remove config directory (optional)
rm -rf ~/.devcockpit
```

## Getting Help

Need assistance?

- **Documentation:** [devcockpit.app](https://devcockpit.app)
- **GitHub Issues:** [Report bugs or request features](https://github.com/caioricciuti/dev-cockpit/issues)
- **Troubleshooting:** See the [Troubleshooting Guide](/troubleshooting)

Welcome to Dev Cockpit - your development command center!
