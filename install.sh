#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="sdexmon/sdexmon"
BINARY_NAME="sdexmon"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case "$os" in
        linux) OS="linux" ;;
        darwin) OS="darwin" ;;
        mingw*|msys*|cygwin*) OS="windows" ;;
        *) log_error "Unsupported operating system: $os"; exit 1 ;;
    esac
    
    case "$arch" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) log_error "Unsupported architecture: $arch"; exit 1 ;;
    esac
    
    log_info "Detected platform: ${OS}_${ARCH}"
}

# Get latest release version
get_latest_version() {
    log_info "Fetching latest release..."
    VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ]; then
        log_error "Failed to get latest version"
        exit 1
    fi
    
    log_info "Latest version: $VERSION"
}

# Create wrapper script for proper environment setup
create_wrapper() {
    local wrapper_path="${INSTALL_DIR}/${BINARY_NAME}"
    local binary_path="${INSTALL_DIR}/.${BINARY_NAME}-bin"
    local temp_wrapper=$(mktemp)
    
    log_info "Creating wrapper script..."
    
    # Create wrapper with proper environment in temp location
    cat > "$temp_wrapper" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail

# Safe defaults for running sdexmon
export HORIZON_URL="${HORIZON_URL:-https://horizon.stellar.org}"
export DEBUG="${DEBUG:-true}"

# Set terminal window title
printf '\033]0;sdexmon\007'

# Set fixed terminal size (140 columns x 60 rows)
if command -v tput >/dev/null 2>&1; then
  printf '\e[8;60;140t'
fi

# Run the actual binary
exec "$(dirname "$0")/.sdexmon-bin" "$@"
EOF
    
    # Move existing binary to hidden location if reinstalling
    if [ -f "$wrapper_path" ] && [ ! -L "$wrapper_path" ]; then
        sudo mv "$wrapper_path" "$binary_path" 2>/dev/null || true
    fi
    
    # Install wrapper using sudo
    if ! sudo cp "$temp_wrapper" "$wrapper_path"; then
        log_error "Failed to install wrapper script"
        rm -f "$temp_wrapper"
        return 1
    fi
    
    rm -f "$temp_wrapper"
    sudo chmod 755 "$wrapper_path"
    log_info "✅ Wrapper script created"
}

# Download and install binary
install_binary() {
    local filename="${BINARY_NAME}_${VERSION#v}_${OS}_${ARCH}"
    local archive="${filename}.tar.gz"
    local url="https://github.com/${REPO}/releases/download/${VERSION}/${archive}"
    local temp_dir=$(mktemp -d)
    
    log_info "Downloading $url..."
    
    if ! curl -L -o "${temp_dir}/${archive}" "$url"; then
        log_error "Failed to download binary"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    log_info "Extracting archive..."
    if ! tar -xzf "${temp_dir}/${archive}" -C "$temp_dir"; then
        log_error "Failed to extract archive"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    # Create install directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        log_warn "Install directory $INSTALL_DIR doesn't exist. Creating..."
        sudo mkdir -p "$INSTALL_DIR"
    fi
    
    # Install binary with hidden name (wrapper will call it)
    log_info "Installing to $INSTALL_DIR..."
    if ! sudo cp "${temp_dir}/${BINARY_NAME}" "$INSTALL_DIR/.${BINARY_NAME}-bin"; then
        log_error "Failed to install binary. Do you have permission to write to $INSTALL_DIR?"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    # Make executable
    sudo chmod 755 "$INSTALL_DIR/.${BINARY_NAME}-bin"
    
    # Create wrapper script
    create_wrapper
    
    # Cleanup
    rm -rf "$temp_dir"
    
    log_info "✅ Successfully installed ${BINARY_NAME} to $INSTALL_DIR"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        log_info "✅ Installation verified!"
        echo ""
        echo "Run '${BINARY_NAME}' to start monitoring Stellar DEX markets."
        echo "Run '${BINARY_NAME} --version' to check the version."
        echo ""
        echo "For help and documentation, visit: https://github.com/${REPO}"
    else
        log_warn "Binary installed but not found in PATH."
        echo "You may need to add $INSTALL_DIR to your PATH or run: $INSTALL_DIR/$BINARY_NAME"
    fi
}

# Main installation flow
main() {
    echo "🚀 sdexmon Installer"
    echo "===================="
    echo ""
    
    # Check dependencies
    if ! command -v curl >/dev/null 2>&1; then
        log_error "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v tar >/dev/null 2>&1; then
        log_error "tar is required but not installed"
        exit 1
    fi
    
    detect_platform
    get_latest_version
    install_binary
    verify_installation
}

# Handle options
case "${1:-}" in
    -h|--help)
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  -h, --help     Show this help message"
        echo ""
        echo "Environment variables:"
        echo "  INSTALL_DIR    Installation directory (default: /usr/local/bin)"
        echo ""
        echo "Example:"
        echo "  INSTALL_DIR=~/.local/bin $0"
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac