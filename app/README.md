# Dev Cockpit v4.0 🚀

**Professional macOS Development Command Center for Apple Silicon**

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-GPLv3-blue)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-macOS%20Apple%20Silicon-000000?logo=apple)](https://www.apple.com/macos/)
[![Architecture](https://img.shields.io/badge/Architecture-arm64%20Only-FF6B6B)](https://support.apple.com/en-us/HT211814)
[![Support](https://img.shields.io/badge/Support-Donations%20Welcome-ff69b4)](https://buymeacoffee.com/caioricciuti)

Dev Cockpit is a professional-grade terminal user interface (TUI) application specifically designed for Apple Silicon Macs (M1/M2/M3). Built with Go and the Bubble Tea framework, it provides comprehensive system monitoring, Docker management, security scanning, and much more in a beautiful, responsive interface optimized for ARM64 architecture.

## ✨ Features

Every module ships unlocked in the Community Edition:

- 📊 **Real-time System Dashboard** – CPU, memory, disk, and network monitoring with historical trends
- 🖥️ **Deep System Insights** – Hardware inventory, process viewer, and performance scoring for Apple Silicon
- ⚡ **Quick Actions & Cleanup** – One-tap maintenance tasks, cache sweeping, and storage recovery tools
- 🐳 **Docker Management** – Container lifecycle control, log viewing, and health checks from the terminal
- 🌐 **Network Analysis** – Interface overview, default gateway detection, and quick connectivity diagnostics
- 🔒 **Security Snapshot** – Firewall/FileVault/SIP status at a glance with actionable guidance
- 📦 **Package Manager Hub** – Audit Homebrew, npm, pip, cargo, and more in one place
- 🛠️ **Automation-friendly CLI** – Invoke focused workflows (e.g., cleanup) directly with command flags

## 📸 Screenshots

```
╔══════════════════════════════════════════════════════════════════╗
║ DEV COCKPIT │ Community Edition v4.0               OPEN SOURCE  ║
╟──────────────────────────────────────────────────────────────────╢
║ CPU: 23.4% │ Disk: 67% │ Net: 180 Mbps │ Uptime: 5d 14h         ║
╚══════════════════════════════════════════════════════════════════╝

Dashboard  System  Docker  Network  Security

┌─ CPU Usage: 23.4% ─────────┐  ┌─ Memory Usage: 45.2% ──────┐
│ ████████░░░░░░░░░░░░░░░░░░ │  │ ██████████████░░░░░░░░░░░░ │
│ Core 0: 45.2%               │  │ Used: 8.3 GB / Total: 16 GB│
│ Core 1: 12.1%               │  │                            │
└────────────────────────────┘  └────────────────────────────┘

Tab: Switch Module | ?: Help | S: Support | l: Logs | Q: Quit
```

## 📱 System Requirements

**⚠️ IMPORTANT: Apple Silicon Mac Required**

- **Processor**: Apple M1, M2, M3, or newer
- **macOS**: 11.0 Big Sur or later
- **Architecture**: ARM64 (Apple Silicon) only
- **Memory**: 4GB RAM minimum
- **Storage**: 100MB available space

> **Note**: This application is optimized exclusively for Apple Silicon Macs. Intel-based Macs are not supported.

## 🚀 Installation

### Quick Install (Recommended)
```bash
curl -sSL https://raw.githubusercontent.com/caioricciuti/dev-cockpit/main/install.sh | bash
```

### Manual Download
```bash
# Download latest release
curl -L -o devcockpit https://github.com/caioricciuti/dev-cockpit/releases/latest/download/devcockpit-darwin-arm64

# Install
chmod +x devcockpit
sudo mv devcockpit /usr/local/bin/
```

### From Source
```bash
# Clone the repository
git clone https://github.com/caioricciuti/dev-cockpit.git
cd dev-cockpit/app

# Build for Apple Silicon
make build

# Install to system
make install
```

## 🎮 Usage

### Starting Dev Cockpit
```bash
# Launch the TUI
devcockpit

# Show version
devcockpit version
```

### Keyboard Shortcuts
- `Tab` / `Shift+Tab` - Switch modules
- `Enter` - Focus the active module
- `Esc` - Return to the module switcher
- `?` - Show help overlay
- `l` - Toggle log overlay
- `Q` - Quit the application
- Module-specific keys (arrows, numbers, etc.) are shown inside each view once focused

## ❤️ Support the Project

Dev Cockpit thrives on community energy. If it streamlines your daily work, you can help in a few ways:

- Become a sponsor: https://github.com/sponsors/caioricciuti
- Send a one-time donation: https://buymeacoffee.com/caioricciuti
- File issues & ideas: https://github.com/caioricciuti/dev-cockpit/issues
- Star the repository and share it with your team

## 🛠️ Development

### Prerequisites
- Go 1.21 or later
- macOS 11.0+ (Big Sur or later)
- Apple Silicon Mac (M1/M2/M3)
- Xcode Command Line Tools (for building)

### Building from Source
```bash
# Clone repository
git clone https://github.com/caioricciuti/dev-cockpit.git
cd dev-cockpit

# Install dependencies
go mod download

# Build for current architecture
make build

# Run development version
make run

# Run tests
make test
```

### Project Structure
```
dev-cockpit/
├── cmd/devcockpit/       # Main application entry point
├── internal/
│   ├── app/              # Core application logic
│   ├── config/           # Configuration management
│   └── modules/          # Feature modules
│       ├── dashboard/    # System monitoring
│       ├── docker/       # Docker management
│       ├── network/      # Network analysis
│       └── security/     # Security scanning
├── pkg/                  # Reusable packages
├── go.mod                # Go module definition
├── Makefile              # Build automation
└── README.md             # This file
```

## 🔧 Configuration

Configuration file is located at `~/.devcockpit/config.yaml`:

```yaml
# Theme settings
theme: dark
update_interval: 1000

# UI settings
ui:
  color_scheme: cyberpunk
  animation_speed: 60
  mouse_enabled: true

# Module settings
modules:
  dashboard:
    refresh_rate: 1
    graph_height: 10

  docker:
    socket_path: /var/run/docker.sock
    auto_refresh: true
```

## 🐛 Troubleshooting

### Common Issues

**Q: Application won't start**
A: Ensure you have the correct permissions and Go version:
```bash
go version  # Should be 1.21+
chmod +x /usr/local/bin/devcockpit
```

**Q: Docker module shows no containers**
A: Check Docker daemon is running:
```bash
docker ps  # Should work without errors
```


## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📬 Support

- **Email**: support@devcockpit.dev
- **Documentation**: [https://devcockpit.dev/docs](https://devcockpit.dev/docs)
- **Issues**: [GitHub Issues](https://github.com/caioricciuti/dev-cockpit/issues)
- **Donations**: [https://buymeacoffee.com/caioricciuti](https://buymeacoffee.com/caioricciuti)

## 📄 License

Dev Cockpit is open source under the GPL v3. See [LICENSE](LICENSE) for details and the README for ways to support ongoing development.

## 🙏 Acknowledgments

Built with these amazing libraries:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [gopsutil](https://github.com/shirou/gopsutil) - System information
- [Viper](https://github.com/spf13/viper) - Configuration management

---

Made with ❤️ by [Caio Ricciuti](https://github.com/caioricciuti)
