# Contributing to Dev Cockpit 🙌

::: tip You are awesome!
Thank you for looking into contributing to Dev Cockpit! Your contributions are essential for making this project better. Whether you're fixing a bug, creating a new feature, or improving documentation, we appreciate your effort.
:::

## How to Contribute

### Reporting Bugs 🐛

If you find a bug, please [create an issue](https://github.com/caioricciuti/dev-cockpit/issues/new) in our GitHub repository.

Make sure to include:
1. A clear title describing the bug
2. Steps to reproduce the issue
3. Expected vs. actual behavior
4. macOS version and chip type (M1/M2/M3)
5. Any relevant error messages or logs (check `~/.devcockpit/debug.log`)

### Requesting Features 💡

Have an idea for a new feature? [Create a feature request](https://github.com/caioricciuti/dev-cockpit/issues/new) with:

1. A clear description of the feature
2. Why it's needed or how it improves Dev Cockpit
3. Which module it would belong to (Dashboard, Cleanup, Packages, etc.)
4. Any additional context or mockups if available

### Improving Documentation 📚

::: info
Who in the world doesn't want good documentation!? If you find typos, missing information, or want to expand a section, feel free to submit a pull request.
:::

Documentation lives in two places:
- Main README: `/README.md`
- Documentation site: `/docs/*`

### Code Contributions 💻

Follow these steps to contribute code:

1. Fork the repository to your GitHub account or clone it directly:

```bash
git clone https://github.com/caioricciuti/dev-cockpit.git
cd dev-cockpit/app
```

2. Create a new branch for your contribution:

```bash
git checkout -b feature/my-new-feature
```

3. Make your changes and test them:

```bash
make build
./build/devcockpit  # Test your changes
```

4. Commit your changes with a descriptive message:

```bash
git commit -m "Add new feature: [description]"
```

5. Push to your fork and open a Pull Request:

```bash
git push origin feature/my-new-feature
```

6. Open a pull request from your fork's branch to `main` on the Dev Cockpit repository.

## Development Setup

### Prerequisites

- macOS 11.0+ running on Apple Silicon (M1/M2/M3)
- Go 1.21 or newer (`brew install go`)
- Xcode Command Line Tools (`xcode-select --install`)

### Setting Up the Environment

1. Clone the repository:

```bash
git clone https://github.com/caioricciuti/dev-cockpit.git
cd dev-cockpit/app
```

2. Install Go dependencies:

```bash
make deps
```

3. Build and run:

```bash
make build
./build/devcockpit
```

For development with automatic reload, use:

```bash
make run
```

### Project Structure

```
dev-cockpit/
├── app/                      # Go application
│   ├── cmd/devcockpit/      # Main entry point
│   ├── internal/            # Internal packages
│   │   ├── app/            # Main app logic
│   │   ├── config/         # Configuration
│   │   ├── modules/        # Feature modules
│   │   └── sudo/           # Sudo helper
│   ├── Makefile            # Build automation
│   └── go.mod              # Go dependencies
├── docs/                    # Documentation website
└── README.md               # Main documentation
```

### Code Style Guidelines

To maintain consistency across the codebase:

- Follow standard Go conventions (`gofmt`, `golint`)
- Ensure your code is well-documented with comments where necessary
- Run the formatter before committing:

```bash
make fmt
```

- Run the linter:

```bash
make lint
```

### Testing Requirements

Before submitting a pull request:

1. Run all tests to ensure no existing functionality is broken:

```bash
make test
```

2. Test the binary manually:

```bash
make build
./build/devcockpit
```

3. Test on a clean macOS installation if possible

### Adding a New Module

To add a new module to Dev Cockpit:

1. Create a new directory in `internal/modules/yourmodule/`
2. Implement the Module interface:
   ```go
   type Module interface {
       Init() tea.Cmd
       Update(msg tea.Msg) (interface{}, tea.Cmd)
       View() string
       Title() string
       HasOpenModal() bool
   }
   ```
3. Register it in `internal/app/app.go`
4. Add documentation to `/docs/features.md`

## Community and Support 👥

- Use GitHub Issues for bug reports and feature requests
- Star the repository to show your support
- Share Dev Cockpit with other Apple Silicon users

## License 📄

By contributing to Dev Cockpit, you agree that your contributions will be licensed under GPL 3.0. Check the LICENSE file for more information.

## Thank You! ❤️

Thank you for being part of the Dev Cockpit community. Together, we can build the best macOS development command center!

::: info
Need help getting started? Feel free to reach out through GitHub Issues or check the documentation at [devcockpit.app](https://devcockpit.app)
:::
