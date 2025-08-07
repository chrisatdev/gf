#!/bin/bash

# GF Installation Script
# Installs the gf command globally

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if running as root for global installation
if [[ $EUID -eq 0 ]]; then
    INSTALL_DIR="/usr/local/bin"
    print_info "Installing globally to $INSTALL_DIR"
else
    # Install to user's local bin
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    print_info "Installing locally to $INSTALL_DIR"
fi

# Copy the gf script
if [[ -f "gf" ]]; then
    cp gf "$INSTALL_DIR/gf"
    chmod +x "$INSTALL_DIR/gf"
    print_success "GF command installed successfully!"
else
    print_error "gf script not found in current directory"
    exit 1
fi

# Check if directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    print_info "Adding $INSTALL_DIR to PATH in your shell profile"

    # Detect shell and add to appropriate profile
    if [[ -n "$ZSH_VERSION" ]]; then
        echo 'export PATH="$HOME/.local/bin:$PATH"' >>~/.zshrc
        print_info "Added to ~/.zshrc"
    elif [[ -n "$BASH_VERSION" ]]; then
        echo 'export PATH="$HOME/.local/bin:$PATH"' >>~/.bashrc
        print_info "Added to ~/.bashrc"
    fi

    print_info "Please restart your terminal or run: source ~/.bashrc (or ~/.zshrc)"
fi

print_success "Installation complete!"
print_info "Try: gf -h to see all available commands"
