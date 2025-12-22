#!/usr/bin/env bash
# File: bootstrap.sh
# Purpose: Minimal bootstrap script to download and run devsetup Go binary
# Problem: Need simple one-liner to kick off installation without pre-installed dependencies
# Role: Downloads devsetup binary from GitHub releases, makes executable, runs install command
# Usage: curl -fsSL https://raw.githubusercontent.com/rkinnovate/dev-setup/main/bootstrap.sh | bash
# Design choices: Minimal dependencies (only bash, curl); detects architecture; cleans up on error
# Assumptions: macOS host; curl available; internet access; GitHub releases exist

set -euo pipefail

# Version to download (can be overridden with DEVSETUP_VERSION env var)
VERSION="${DEVSETUP_VERSION:-latest}"

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
BINARY_PATH="${TEMP_DIR}/devsetup"

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
if ! curl -fsSL "$DOWNLOAD_URL" -o "$BINARY_PATH"; then
  echo "❌ Failed to download devsetup binary"
  echo ""
  echo "Troubleshooting:"
  echo "  • Check your internet connection"
  echo "  • Verify release exists: https://github.com/rkinnovate/dev-setup/releases"
  echo "  • Try specifying version: DEVSETUP_VERSION=v0.4.0 bash bootstrap.sh"
  exit 1
fi

# Make executable
chmod +x "$BINARY_PATH"

echo "✅ Downloaded devsetup binary"
echo ""
echo "Starting installation..."
echo ""

# Run installer
"$BINARY_PATH" install

# Installation complete
echo ""
echo "╔════════════════════════════════════════════════════════╗"
echo "║                                                        ║"
echo "║   🎉 Installation Complete!                            ║"
echo "║                                                        ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""
echo "Next steps:"
echo "  1. Restart your terminal (or run: source ~/.zshrc)"
echo "  2. Verify installation: devsetup verify"
echo "  3. Check status: devsetup status"
echo ""
echo "The devsetup binary has been removed."
echo "To install permanently:"
echo "  git clone https://github.com/rkinnovate/dev-setup"
echo "  cd dev-setup"
echo "  make install"
echo ""
echo "Happy coding! 🚀"
