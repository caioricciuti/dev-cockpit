#!/usr/bin/env bash
#
# Dev Cockpit Uninstaller
# Removes Dev Cockpit from macOS
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="devcockpit"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.devcockpit"
FALLBACK_CONFIG_DIR="./.devcockpit"

# Print colored message
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check if Dev Cockpit is running
check_running() {
    if pgrep -x "$BINARY_NAME" > /dev/null; then
        print_warning "Dev Cockpit is currently running"
        read -p "Do you want to stop it? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_info "Stopping Dev Cockpit..."
            pkill -TERM "$BINARY_NAME" || true
            sleep 2
            # Force kill if still running
            if pgrep -x "$BINARY_NAME" > /dev/null; then
                pkill -KILL "$BINARY_NAME" || true
            fi
            print_success "Dev Cockpit stopped"
        else
            print_error "Please stop Dev Cockpit before uninstalling"
            exit 1
        fi
    fi
}

# Remove binary
remove_binary() {
    local binary_path="$INSTALL_DIR/$BINARY_NAME"

    if [[ -f "$binary_path" ]]; then
        print_info "Removing binary from $binary_path..."

        if [[ -w "$INSTALL_DIR" ]]; then
            rm -f "$binary_path"
        else
            print_warning "Requesting administrator privileges to remove binary"
            sudo rm -f "$binary_path"
        fi

        print_success "Binary removed"
    else
        print_info "Binary not found at $binary_path (already removed or never installed)"
    fi
}

# Remove configuration and data
remove_config() {
    if [[ -d "$CONFIG_DIR" ]]; then
        print_info "Found configuration directory: $CONFIG_DIR"

        # Show what will be deleted
        local items_found=false
        if [[ -f "$CONFIG_DIR/config.yaml" ]]; then
            print_info "  - config.yaml"
            items_found=true
        fi
        if [[ -f "$CONFIG_DIR/debug.log" ]]; then
            print_info "  - debug.log"
            items_found=true
        fi
        if [[ -d "$CONFIG_DIR/data" ]]; then
            print_info "  - data directory"
            items_found=true
        fi

        # Show directory size
        local dir_size=$(du -sh "$CONFIG_DIR" 2>/dev/null | cut -f1)
        print_info "  Total size: $dir_size"

        echo ""
        read -p "Remove configuration and data? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$CONFIG_DIR"
            print_success "Configuration directory removed"
        else
            print_info "Configuration directory kept at $CONFIG_DIR"
        fi
    else
        print_info "No configuration directory found at $CONFIG_DIR"
    fi
}

# Remove fallback config directory (if exists in current directory)
remove_fallback_config() {
    if [[ -d "$FALLBACK_CONFIG_DIR" ]]; then
        print_warning "Found fallback config directory: $FALLBACK_CONFIG_DIR"
        read -p "Remove it? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$FALLBACK_CONFIG_DIR"
            print_success "Fallback config directory removed"
        fi
    fi
}

# Remove temporary files
remove_temp_files() {
    print_info "Checking for temporary files..."

    local found_temp=false
    local temp_count=0

    # Check for fallback log in /tmp
    if ls /tmp/devcockpit-debug.log* 2>/dev/null 1>&2; then
        rm -f /tmp/devcockpit-debug.log*
        found_temp=true
        ((temp_count++))
    fi

    # Check for any leftover temp files
    if ls /tmp/devcockpit-* 2>/dev/null 1>&2; then
        local count=$(ls /tmp/devcockpit-* 2>/dev/null | wc -l | tr -d ' ')
        rm -f /tmp/devcockpit-*
        found_temp=true
        temp_count=$((temp_count + count))
    fi

    if [[ "$found_temp" == true ]]; then
        print_success "Removed $temp_count temporary file(s)"
    else
        print_info "No temporary files found"
    fi
}

# Print completion message
print_completion() {
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║  Dev Cockpit uninstalled successfully! ✓   ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}Thank you for using Dev Cockpit!${NC}"
    echo ""
    echo "If you encountered any issues, please report them at:"
    echo "  https://github.com/caioricciuti/dev-cockpit/issues"
    echo ""
    echo "To reinstall in the future:"
    echo "  curl -fsSL https://raw.githubusercontent.com/caioricciuti/dev-cockpit/main/install.sh | bash"
    echo ""
}

# Main uninstallation flow
main() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║      Dev Cockpit Uninstaller v1.0.0       ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
    echo ""

    print_warning "This will remove Dev Cockpit from your system"
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Uninstallation cancelled"
        exit 0
    fi

    echo ""

    check_running
    remove_binary
    remove_config
    remove_fallback_config
    remove_temp_files

    print_completion
}

# Run main function
main
