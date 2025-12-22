# File: Brewfile
# Purpose: Declarative definition of all Homebrew packages for the development environment
# Problem: Need single source of truth for Homebrew dependencies with version locking
# Role: Defines all formulas and casks; used by `brew bundle` for parallel installation
# Usage: Run `brew bundle install` to install all packages; `brew bundle dump` to update
# Design choices: Organized by category; taps listed first; uses Brewfile.lock.json for version locking
# Assumptions: Homebrew installed; user has permissions to install casks

# ===================================================================
# Taps (Additional package repositories)
# ===================================================================
tap "homebrew/bundle"
tap "homebrew/cask-fonts"

# ===================================================================
# Critical formulas (Stage 1 - required immediately)
# ===================================================================
brew "git"
brew "node"
brew "pnpm"
brew "python@3.11"

# ===================================================================
# Development tools (Stage 2 - full stack)
# ===================================================================
brew "git-lfs"
brew "wget"
brew "starship"
brew "jq"
brew "ripgrep"
brew "fd"
brew "bat"
brew "eza"

# ===================================================================
# Optional tools (Stage 3 - polish)
# ===================================================================
brew "fzf"
brew "tree"
brew "htop"
brew "tldr"

# ===================================================================
# Casks (GUI Applications)
# ===================================================================

# Editors
cask "zed"

# Fonts
cask "font-hack-nerd-font"

# ===================================================================
# Notes:
# ===================================================================
# - Run `brew bundle install --no-upgrade` to install without upgrading
# - Run `brew bundle install --parallel` for faster installation
# - Versions locked in Brewfile.lock.json (commit this file!)
# - To update versions: `brew bundle dump --force`
