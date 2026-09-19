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
curl -fsSL https://raw.githubusercontent.com/caioricciuti/dev-cockpit/main/install.sh | bash
```

This script will:
1. Detect your platform and download the matching release binary
2. Download the published SHA-256 checksum and verify the binary against it
3. Install it to `/usr/local/bin/devcockpit` and make it executable

Verification is not optional. If the checksum is missing, unreadable or
does not match, the installer aborts and installs nothing.

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

Dev Cockpit is thirteen modules behind one window: a tab bar across the top,
the active module below it.

```
┌─────────────────────────────────────────────┐
│  Dashboard | Processes | Services | ...     │  ← Module tabs
├─────────────────────────────────────────────┤
│                                             │
│            Module content                   │
│                                             │
└─────────────────────────────────────────────┘
```

### Navigation

The interface has **two levels**, and this is the thing worth learning first.

At the **browse** level you move between modules. Press `Enter` to **focus**
one, and from then on that module receives every key. `Esc` hands control
back to browsing.

If a key seems to do nothing, you are usually at the wrong level. Press `Esc`
until you are browsing again.

| Key | Action |
| --- | --- |
| `Tab`, `→` | Next module |
| `Shift+Tab`, `←` | Previous module |
| `Home` / `End` | First / last module |
| `Enter` | Focus the current module |
| `Esc` | Leave a focused module |
| `?` | Help overlay |
| `l` | Log overlay |
| `q` | Quit |
| `Ctrl+C` | Quit from anywhere, including inside a module |

Inside a focused module, `j`/`k` or the arrows move, `Enter` selects and `r`
refreshes. Number keys switch that module's own sub-views; they do not jump
between modules.

Every module and its individual keys is listed in [Modules](/modules).

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

Configuration lives in `~/.devcockpit/`:

```
~/.devcockpit/
├── config.yaml      # Settings, all optional
├── debug.log        # Written when --debug is used
└── data/            # Metrics history
```

Every setting has a default, so the file is optional and anything you leave
out keeps its default. The full schema is in
[Configuration](/configuration).

## CLI Commands

Every read-only view has a command-line equivalent, and output is plain text
when piped or redirected:

```bash
devcockpit status    # health score and key metrics
devcockpit diag      # full diagnostics report
devcockpit ps        # top processes by CPU
```

All commands and flags are in the [CLI Reference](/cli).

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

- **[Modules](/modules)** — what each of the thirteen screens does, and the
  keys it responds to
- **[CLI Reference](/cli)** — every command, flag and output format
- **[Configuration](/configuration)** — every setting and its default
- **[Troubleshooting](/troubleshooting)** — when something misbehaves

If you only do one thing next, press `?` inside the app for the help overlay,
then `Enter` on the Dashboard to focus it.

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
