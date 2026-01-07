#!/bin/sh
# EgenSkriven Installation Script
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/ramtinJ95/EgenSkriven/main/install.sh | sh
#
# Environment variables:
#   INSTALL_DIR  - Installation directory (default: /usr/local/bin or ~/.local/bin)
#   VERSION      - Specific version to install (default: latest)

set -e

# Colors for output (only if terminal supports it)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Print functions
info() { printf "${BLUE}[INFO]${NC} %s\n" "$1"; }
success() { printf "${GREEN}[OK]${NC} %s\n" "$1"; }
warn() { printf "${YELLOW}[WARN]${NC} %s\n" "$1"; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$1" >&2; exit 1; }

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)   OS="linux" ;;
        darwin)  OS="darwin" ;;
        mingw*|msys*|cygwin*) 
            error "Windows detected. Please use the Windows installer or download manually."
            ;;
        *)       error "Unsupported operating system: $OS" ;;
    esac

    case "$ARCH" in
        x86_64|amd64)   ARCH="amd64" ;;
        arm64|aarch64)  ARCH="arm64" ;;
        *)              error "Unsupported architecture: $ARCH" ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    success "Detected platform: $PLATFORM"
}

# Get the latest version from GitHub
get_latest_version() {
    if [ -n "$VERSION" ]; then
        info "Using specified version: $VERSION"
        return
    fi

    info "Fetching latest version..."
    VERSION=$(curl -fsSL "https://api.github.com/repos/ramtinJ95/EgenSkriven/releases/latest" | 
              grep '"tag_name":' | 
              sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ]; then
        error "Could not determine latest version. Please specify VERSION environment variable."
    fi
    
    success "Latest version: $VERSION"
}

# Determine installation directory
get_install_dir() {
    if [ -n "$INSTALL_DIR" ]; then
        info "Using specified install directory: $INSTALL_DIR"
        return
    fi

    # Prefer /usr/local/bin if writable, otherwise ~/.local/bin
    if [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
    else
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
    fi

    success "Install directory: $INSTALL_DIR"
}

# Download and install the binary
install_binary() {
    BINARY_NAME="egenskriven-${PLATFORM}"
    DOWNLOAD_URL="https://github.com/ramtinJ95/EgenSkriven/releases/download/${VERSION}/${BINARY_NAME}"
    TMP_FILE=$(mktemp)

    info "Downloading $BINARY_NAME..."
    curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE" || error "Download failed. Check your internet connection."

    info "Installing to $INSTALL_DIR/egenskriven..."
    mv "$TMP_FILE" "$INSTALL_DIR/egenskriven"
    chmod +x "$INSTALL_DIR/egenskriven"

    success "Installed successfully!"
}

# Verify installation
verify_installation() {
    if command -v egenskriven >/dev/null 2>&1; then
        info "Verifying installation..."
        INSTALLED_VERSION=$(egenskriven version 2>/dev/null | head -1 | awk '{print $2}')
        success "EgenSkriven $INSTALLED_VERSION is ready to use!"
    else
        warn "Installation complete, but 'egenskriven' is not in your PATH."
        echo ""
        echo "Add the following to your shell profile (.bashrc, .zshrc, etc.):"
        echo ""
        echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
        echo "Then restart your shell or run:"
        echo ""
        echo "    source ~/.bashrc  # or ~/.zshrc"
        echo ""
    fi
}

# Print next steps
print_next_steps() {
    echo ""
    echo "=== Next Steps ==="
    echo ""
    echo "1. Start the server:"
    echo "   $ egenskriven serve"
    echo ""
    echo "2. Open the web UI:"
    echo "   http://localhost:8090"
    echo ""
    echo "3. Create your first task:"
    echo "   $ egenskriven add \"My first task\""
    echo ""
    echo "4. Enable shell completions:"
    echo "   $ egenskriven completion --help"
    echo ""
    echo "Documentation: https://github.com/ramtinJ95/EgenSkriven"
    echo ""
}

# Main installation flow
main() {
    echo ""
    echo "=== EgenSkriven Installer ==="
    echo ""

    detect_platform
    get_latest_version
    get_install_dir
    install_binary
    verify_installation
    print_next_steps
}

main
