#!/usr/bin/env bash
#
# Dev Cockpit Installer
# Install Dev Cockpit for macOS and Linux
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="caioricciuti/dev-cockpit"
BINARY_NAME="devcockpit"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.devcockpit"

# Print colored message
print_info() {
    echo -e "${BLUE}i${NC} $1"
}

print_success() {
    echo -e "${GREEN}+${NC} $1"
}

print_error() {
    echo -e "${RED}x${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}!${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        darwin)
            OS="darwin"
            ;;
        linux)
            OS="linux"
            ;;
        *)
            print_error "Unsupported operating system: $OS"
            print_error "Dev Cockpit supports macOS and Linux"
            exit 1
            ;;
    esac

    case "$ARCH" in
        arm64|aarch64)
            ARCH="arm64"
            ;;
        x86_64|amd64)
            ARCH="amd64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    # Darwin only supports arm64
    if [[ "$OS" == "darwin" && "$ARCH" != "arm64" ]]; then
        print_error "macOS builds are only available for Apple Silicon (arm64)"
        print_error "Detected architecture: $ARCH"
        exit 1
    fi

    PLATFORM_BINARY="${BINARY_NAME}-${OS}-${ARCH}"
    print_info "Detected platform: ${OS}/${ARCH}"
}

# Get latest release info from GitHub
get_latest_release() {
    print_info "Fetching latest release information..."

    LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [[ -z "$LATEST_TAG" ]]; then
        print_warning "Could not fetch latest release, using default version"
        LATEST_TAG="v1.0.0"
    fi

    print_info "Latest version: $LATEST_TAG"
}

# Download binary
download_binary() {
    print_info "Downloading Dev Cockpit..."

    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/${PLATFORM_BINARY}"
    TEMP_FILE="/tmp/${BINARY_NAME}-$$"

    if command -v curl &> /dev/null; then
        curl -L -f -o "$TEMP_FILE" "$DOWNLOAD_URL" 2>&1 | grep -v "^$" || {
            print_error "Failed to download Dev Cockpit"
            print_error "URL: $DOWNLOAD_URL"
            exit 1
        }
    elif command -v wget &> /dev/null; then
        wget -q -O "$TEMP_FILE" "$DOWNLOAD_URL" || {
            print_error "Failed to download Dev Cockpit"
            exit 1
        }
    else
        print_error "Neither curl nor wget is available"
        exit 1
    fi

    print_success "Downloaded successfully"
}

# Verify checksum
verify_checksum() {
    print_info "Verifying checksum..."

    CHECKSUM_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/${PLATFORM_BINARY}.sha256"

    if curl -s -L -f "$CHECKSUM_URL" -o /tmp/devcockpit-checksum-$$ 2>/dev/null; then
        EXPECTED_CHECKSUM=$(cat /tmp/devcockpit-checksum-$$ | awk '{print $1}')

        # Use appropriate checksum tool
        if command -v sha256sum &> /dev/null; then
            ACTUAL_CHECKSUM=$(sha256sum "$TEMP_FILE" | awk '{print $1}')
        elif command -v shasum &> /dev/null; then
            ACTUAL_CHECKSUM=$(shasum -a 256 "$TEMP_FILE" | awk '{print $1}')
        else
            print_warning "No checksum tool found, skipping verification"
            rm -f /tmp/devcockpit-checksum-$$
            return
        fi

        if [[ "$EXPECTED_CHECKSUM" == "$ACTUAL_CHECKSUM" ]]; then
            print_success "Checksum verified"
        else
            print_error "Checksum verification failed!"
            print_error "Expected: $EXPECTED_CHECKSUM"
            print_error "Actual:   $ACTUAL_CHECKSUM"
            print_error "The downloaded file may be corrupted or tampered with."
            print_error "Aborting installation for security."
            rm -f /tmp/devcockpit-checksum-$$
            rm -f "$TEMP_FILE"
            exit 1
        fi
        rm -f /tmp/devcockpit-checksum-$$
    else
        print_warning "Checksum file not available, skipping verification"
    fi
}

# Install binary
install_binary() {
    print_info "Installing to $INSTALL_DIR..."

    chmod +x "$TEMP_FILE"

    if [[ -w "$INSTALL_DIR" ]]; then
        mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    else
        print_warning "Requesting administrator privileges to install to $INSTALL_DIR"
        sudo mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
        sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi

    print_success "Installed to $INSTALL_DIR/$BINARY_NAME"
}

# Create config directory
create_config_dir() {
    if [[ ! -d "$CONFIG_DIR" ]]; then
        print_info "Creating configuration directory..."
        mkdir -p "$CONFIG_DIR"
        print_success "Created $CONFIG_DIR"
    fi
}

# Print success message
print_completion() {
    echo ""
    echo -e "${GREEN}Dev Cockpit installed successfully!${NC}"
    echo ""
    echo -e "${BLUE}Get started:${NC}"
    echo "  devcockpit              # Launch the TUI"
    echo "  devcockpit --help       # Show help"
    echo "  devcockpit --version    # Show version"
    echo ""
    echo -e "${BLUE}Support the project:${NC}"
    echo "  https://github.com/sponsors/caioricciuti"
    echo "  https://buymeacoffee.com/caioricciuti"
    echo ""
}

# Main installation flow
main() {
    echo ""
    echo -e "${BLUE}Dev Cockpit Installer${NC}"
    echo ""

    detect_platform
    get_latest_release
    download_binary
    verify_checksum
    install_binary
    create_config_dir
    print_completion
}

main
