#!/usr/bin/env bash
# File: bootstrap.sh
# Purpose: Minimal bootstrap script to download and install devsetup Go binary
# Problem: Need simple one-liner to kick off installation without pre-installed dependencies
# Role: Downloads devsetup binary from GitHub releases, installs to ~/.local/bin, runs install
# Usage: curl -fsSL https://raw.githubusercontent.com/rkinnovate/dev-setup/main/bootstrap.sh | bash
# Design choices: Minimal dependencies (only bash, curl); detects architecture; installs to user bin
# Assumptions: macOS host; curl available; internet access; GitHub releases exist

set -euo pipefail

# Version to download (can be overridden with DEVSETUP_VERSION env var)
VERSION="${DEVSETUP_VERSION:-latest}"

# Installation directory (user-local, no sudo required)
INSTALL_DIR="${HOME}/.local/bin"

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)
    BINARY_ARCH="amd64"
    ;;
  arm64)
    BINARY_ARCH="arm64"
    ;;
  *)
    echo "❌ Unsupported architecture: $ARCH"
    echo "   Supported: x86_64 (Intel), arm64 (Apple Silicon)"
    exit 1
    ;;
esac

# Binary name
BINARY_NAME="devsetup-darwin-${BINARY_ARCH}"

# Download URL
if [ "$VERSION" = "latest" ]; then
  DOWNLOAD_URL="https://github.com/rkinnovate/dev-setup/releases/latest/download/${BINARY_NAME}"
else
  DOWNLOAD_URL="https://github.com/rkinnovate/dev-setup/releases/download/${VERSION}/${BINARY_NAME}"
fi

# Temporary download location
TEMP_DIR="$(mktemp -d)"
TEMP_BINARY="${TEMP_DIR}/devsetup"

# Cleanup function
cleanup() {
  rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

echo "╔════════════════════════════════════════════════════════╗"
echo "║                                                        ║"
echo "║   DEV-SETUP: Zero to Productive in 5 Minutes           ║"
echo "║                                                        ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""
echo "Detected architecture: $ARCH"
echo "Downloading devsetup ($VERSION)..."
echo ""

# Download binary
if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_BINARY"; then
  echo "❌ Failed to download devsetup binary"
  echo ""
  echo "Troubleshooting:"
  echo "  • Check your internet connection"
  echo "  • Verify release exists: https://github.com/rkinnovate/dev-setup/releases"
  echo "  • Try specifying version: DEVSETUP_VERSION=v0.4.0 bash bootstrap.sh"
  exit 1
fi

# Make executable
chmod +x "$TEMP_BINARY"

echo "✅ Downloaded devsetup binary"
echo ""

# Create installation directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Install binary
echo "Installing devsetup to $INSTALL_DIR..."
mv "$TEMP_BINARY" "$INSTALL_DIR/devsetup"

echo "✅ Installed devsetup to $INSTALL_DIR/devsetup"
echo ""

# Check if ~/.local/bin is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  echo "⚠️  $INSTALL_DIR is not in your PATH"
  echo "   Adding to ~/.zshrc..."
  echo ""

  # Add to .zshrc if not already there
  ZSHRC="$HOME/.zshrc"
  PATH_LINE="export PATH=\"\$HOME/.local/bin:\$PATH\""

  if [ -f "$ZSHRC" ] && ! grep -q "\.local/bin" "$ZSHRC"; then
    echo "" >> "$ZSHRC"
    echo "# Added by dev-setup bootstrap" >> "$ZSHRC"
    echo "$PATH_LINE" >> "$ZSHRC"
    echo "✅ Added $INSTALL_DIR to PATH in ~/.zshrc"
  fi

  # Add to current session
  export PATH="$INSTALL_DIR:$PATH"
fi

echo "Starting installation..."
echo ""

# Run installer
devsetup install

# Installation complete
echo ""
echo "╔════════════════════════════════════════════════════════╗"
echo "║                                                        ║"
echo "║   🎉 Installation Complete!                            ║"
echo "║                                                        ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""
echo "The devsetup binary is installed at: $INSTALL_DIR/devsetup"
echo ""
echo "Next steps:"
echo "  1. Restart your terminal (or run: source ~/.zshrc)"
echo "  2. Run configuration: devsetup setup"
echo "  3. Verify installation: devsetup verify"
echo "  4. Check status: devsetup status"
echo ""
echo "Update devsetup anytime with: devsetup update"
echo ""
echo "Happy coding! 🚀"
