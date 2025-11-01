# Dev Cockpit Makefile
# Professional macOS Development Command Center

.PHONY: all build run clean test install uninstall universal

# Variables
BINARY_NAME=devcockpit
VERSION=$(shell cat ../VERSION 2>/dev/null || echo "dev")
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin
MAIN_PACKAGE=./cmd/devcockpit
GOFLAGS=-ldflags="-X main.version=$(VERSION) -s -w"

# Default target
all: clean build

# Build the binary for Apple Silicon only
build:
	@echo "Building Dev Cockpit v$(VERSION) for Apple Silicon..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "🍎 Apple Silicon (arm64) binary ready"

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
	@echo "✅ Dev Cockpit installed to $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "Run 'devcockpit' from anywhere to start"

# Uninstall from system
uninstall:
	@echo "Uninstalling Dev Cockpit..."
	@sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@rm -rf ~/.devcockpit
	@echo "✅ Dev Cockpit uninstalled"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "✅ Clean complete"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run ./...
	@echo "✅ Linting complete"

# Update dependencies
deps:
	@echo "Updating dependencies..."
	@go mod tidy
	@go mod download
	@echo "✅ Dependencies updated"

# Show help
help:
	@echo "Dev Cockpit Makefile - Apple Silicon Edition"
	@echo ""
	@echo "Available targets:"
	@echo "  make build       - Build for Apple Silicon (arm64)"
	@echo "  make run         - Run the application"
	@echo "  make test        - Run tests"
	@echo "  make install     - Install to /usr/local/bin"
	@echo "  make uninstall   - Remove from system"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make fmt         - Format code"
	@echo "  make deps        - Update dependencies"
	@echo "  make release     - Create release build"
	@echo "  make help        - Show this help"

# Development shortcuts
dev: fmt test build run

# Release build (Apple Silicon only)
release: clean test build
	@echo "✅ Release build complete"
	@echo "Binary: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "Size: $(shell du -h $(BUILD_DIR)/$(BINARY_NAME) | cut -f1)"
	@echo "Architecture: Apple Silicon (arm64)"
