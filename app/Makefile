# Dev Cockpit Makefile
# Cross-platform Development Command Center

.PHONY: all build build-linux build-linux-arm64 build-all run clean test install uninstall

# Variables
BINARY_NAME=devcockpit
VERSION=$(shell cat ../VERSION 2>/dev/null || echo "dev")
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin
MAIN_PACKAGE=./cmd/devcockpit
GOFLAGS=-ldflags="-X main.version=$(VERSION) -s -w"

# Auto-detect OS and ARCH
GOOS_LOCAL=$(shell go env GOOS)
GOARCH_LOCAL=$(shell go env GOARCH)

# Default target
all: clean build

# Build for current platform
build:
	@echo "Building Dev Cockpit v$(VERSION) for $(GOOS_LOCAL)/$(GOARCH_LOCAL)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=$(GOOS_LOCAL) GOARCH=$(GOARCH_LOCAL) go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for macOS arm64
build-darwin:
	@echo "Building for darwin/arm64..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)

# Build for Linux amd64
build-linux:
	@echo "Building for linux/amd64..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)

# Build for Linux arm64
build-linux-arm64:
	@echo "Building for linux/arm64..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)

# Build all platforms
build-all: clean build-darwin build-linux build-linux-arm64
	@echo "All platform builds complete:"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)-*

# Run the application
run:
	@go run $(MAIN_PACKAGE)

# Run with race detector (development)
run-race:
	@go run -race $(MAIN_PACKAGE)

# Run tests
test:
	@echo "Running tests..."
	@go test -v -cover ./...

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@go test -race -v ./...

# Install the binary to system
install: build
	@echo "Installing Dev Cockpit to $(INSTALL_DIR)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Dev Cockpit installed to $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "Run 'devcockpit' from anywhere to start"

# Uninstall from system
uninstall:
	@echo "Uninstalling Dev Cockpit..."
	@sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@rm -rf ~/.devcockpit
	@echo "Dev Cockpit uninstalled"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run ./...

# Update dependencies
deps:
	@echo "Updating dependencies..."
	@go mod tidy
	@go mod download

# Show help
help:
	@echo "Dev Cockpit Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make build             - Build for current platform"
	@echo "  make build-darwin      - Build for macOS arm64"
	@echo "  make build-linux       - Build for Linux amd64"
	@echo "  make build-linux-arm64 - Build for Linux arm64"
	@echo "  make build-all         - Build all platforms"
	@echo "  make run               - Run the application"
	@echo "  make test              - Run tests"
	@echo "  make install           - Install to /usr/local/bin"
	@echo "  make uninstall         - Remove from system"
	@echo "  make clean             - Clean build artifacts"
	@echo "  make fmt               - Format code"
	@echo "  make deps              - Update dependencies"
	@echo "  make help              - Show this help"

# Development shortcuts
dev: fmt test build run

# Release build (all platforms)
release: clean test build-all
	@echo "Release build complete"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)-*
