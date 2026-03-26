# Acknowledgments 🙏

Dev Cockpit wouldn't be possible without the amazing open-source projects and tools that power it.

## Core Technologies

### Go Programming Language
The foundation of Dev Cockpit, providing fast compilation, excellent concurrency, and cross-platform support.
- Website: [golang.org](https://golang.org)
- License: BSD 3-Clause

### Bubble Tea
A powerful TUI (Text User Interface) framework for Go that makes building interactive terminal applications delightful.
- GitHub: [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
- Creator: Charm
- License: MIT

### Lipgloss
A styling library for terminal output, providing beautiful colors and layouts for the TUI.
- GitHub: [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)
- Creator: Charm
- License: MIT

### gopsutil
Cross-platform library for retrieving system information (CPU, memory, disk, network, etc.)
- GitHub: [shirou/gopsutil](https://github.com/shirou/gopsutil)
- License: BSD 3-Clause

### Viper
Configuration management library supporting multiple formats and live reloading.
- GitHub: [spf13/viper](https://github.com/spf13/viper)
- License: MIT

## Development Tools

### System Tools
Dev Cockpit integrates with platform-native utilities:
- `brew` - Homebrew package manager (macOS + Linuxbrew)
- `npm` - Node package manager
- `docker` - Container management
- **macOS:** `sysctl`, `pmset`, `sw_vers`, `networksetup`, `diskutil`
- **Linux:** `journalctl`, `systemctl`, `nmcli`, `ip`, `resolvectl`, `lsblk`
- Various cross-platform commands (`du`, `ps`, `ping`, `dig`, etc.)

## Documentation

### VitePress
Powers the documentation website with a modern, fast static site generator.
- Website: [vitepress.dev](https://vitepress.dev)
- License: MIT

## Community Contributors

Special thanks to everyone who has:
- ⭐ Starred the repository
- 🍴 Forked and contributed code
- 🐛 Reported bugs and issues
- 💡 Suggested new features
- 📚 Improved documentation
- ❤️ Supported the project through donations

## Cross-Platform

Built for macOS (Apple Silicon) and Linux (x86_64, ARM64), using Go build tags for clean platform separation with zero runtime overhead.

## Inspiration

Dev Cockpit was inspired by the need for a unified, modern command center for development workflows, combining system monitoring, package management, and maintenance tools in one beautiful TUI.

---

::: info
If you'd like to be acknowledged here, contribute to the project on [GitHub](https://github.com/caioricciuti/dev-cockpit)!
:::

## License

Dev Cockpit is released under GPL 3.0. See the [License](/license) page for full details.

All dependencies are used in accordance with their respective licenses.