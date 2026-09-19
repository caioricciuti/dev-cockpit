---
layout: home

hero:
  name: "Get Under the Hood of Your System"
  tagline: "Dev cockpit was made for developers who want to see what's happening real-time."
  image:
    src: /logo.png
    alt: Dev Cockpit Logo
  actions:
    - theme: brand
      text: Get Started
      link: /getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/caioricciuti/dev-cockpit

features:
  - title: Thirteen tools, one window
    details: Dashboard, processes, services, Docker, network, packages, logs and more, without leaving the terminal or juggling six commands.

  - title: Real-time monitoring
    details: CPU, GPU, memory, disk and network, with history kept locally in SQLite so you can see what changed rather than only what is.

  - title: Works as a CLI too
    details: Every read-only view has a command-line equivalent. Output is plain text when piped, so it composes with grep, scripts and anything else.

  - title: Reclaim disk space
    details: Caches, logs, trash and downloads are scanned and sized first. You pick what goes, individually, before anything is deleted.

  - title: Health score, not a wall of numbers
    details: One graded report across disk, storage, performance, network, services and security, with the reasoning and a suggested fix behind each check.

  - title: Verified updates
    details: Releases ship with SHA-256 checksums, and the installer and updater both refuse to install anything that does not match.
---

## Screenshots

<ScreenshotGallery />

## Quick Installation

### Installer Script
Run the following command in your terminal to install Dev Cockpit:

```bash
curl -fsSL https://raw.githubusercontent.com/caioricciuti/dev-cockpit/main/install.sh | bash
```


## Requirements

**macOS:**
- Apple Silicon Mac (M1, M1 Pro, M1 Max, M2, etc.)
- macOS 11.0 (Big Sur) or later
- Terminal app (iTerm2, kitty, or WezTerm recommended)

**Linux:**
- x86_64 (amd64) or ARM64 (aarch64) processor
- Ubuntu 20.04+, Fedora 36+, Arch, or other modern distribution
- Terminal with true color support

## Support
If you find Dev Cockpit useful, please consider supporting the project:
- ⭐ [Star on GitHub](https://github.com/caioricciuti/dev-cockpit)

## License

Dev Cockpit is open source software [licensed under GPL-3.0](/license).

---

[![Buy Me A Coffee](https://img.buymeacoffee.com/button-api/?text=Buy%20me%20a%20coffee&emoji=&slug=caioricciuti&button_colour=FF813F&font_colour=ffffff&font_family=Cookie&outline_colour=000000&coffee_colour=FFDD00)](https://buymeacoffee.com/caioricciuti)
