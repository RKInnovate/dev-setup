#!/usr/bin/env bash
set -euo pipefail

# Dev Setup Bootstrap (macOS only)
# Purpose: Provision a consistent local dev environment in minutes using Homebrew plus curated tools.
# Problem: Team members spend hours configuring git/node/python/flutter/editors manually; this script standardizes it.
# Architectural role: Single entrypoint shell script; delegates installs to Homebrew or official installers, and clones helper repos under ~/.local/share/dev-setup.
# Usage: Run ./setup.sh on macOS; follow POST_INSTALL.txt for manual steps like PATH/zsh config.
# Design choices: Uses ~/.local/{bin,share} for shims/state; keeps clones isolated from repo; idempotent installs via brew checks; leaves git config application manual for safety.
# Assumptions: macOS host with curl and git; user consents to Homebrew installation and cask prompts; network access available during run.

REPO_ROOT="$(cd -- "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEV_SETUP_VERSION="0.2.0"
STATE_DIR="${HOME}/.local/share/dev-setup"
LOCAL_BIN="${HOME}/.local/bin"
PLUGINS_DIR="${STATE_DIR}/zsh-plugins"
VERSION_FILE="${STATE_DIR}/version"

# Function: log
# What: Prints a normalized log line with a dev-setup prefix.
# Why: Provides consistent user feedback during bootstrap steps.
# Params: $1..$n - message tokens to print.
# Returns: 0 always.
# Example: log "Installing brew"
log() {
  printf "[dev-setup] %s\n" "$*"
}

# Function: error
# What: Prints an error message and exits with failure.
# Why: Centralized error handling to stop the script when a required step fails.
# Params: $1..$n - error message tokens to print.
# Returns: Exits with status 1.
# Example: error "Homebrew install failed"
error() {
  printf "[dev-setup] ERROR: %s\n" "$*" >&2
  exit 1
}

# Function: require_macos
# What: Verifies the host OS is macOS (Darwin).
# Why: Homebrew/cask logic in this script is macOS-specific.
# Params: none.
# Returns: Exits if OS is not macOS; otherwise 0.
# Example: require_macos
require_macos() {
  if [ "$(uname -s)" != "Darwin" ]; then
    error "This bootstrap only supports macOS."
  fi
}

# Function: ensure_paths
# What: Creates state, bin, and plugin directories.
# Why: Ensures later steps have write locations without failing on missing directories.
# Params: none.
# Returns: 0 on success.
# Example: ensure_paths
ensure_paths() {
  mkdir -p "$STATE_DIR" "$LOCAL_BIN" "$PLUGINS_DIR"
}

# Function: handle_version_state
# What: Detects first install, re-run, or update based on VERSION_FILE.
# Why: Provides user feedback and supports future upgrade logic.
# Params: none.
# Returns: 0 on success.
# Example: handle_version_state
handle_version_state() {
  if [ ! -f "$VERSION_FILE" ]; then
    log "First-time install detected (version $DEV_SETUP_VERSION)."
    return
  fi

  local previous
  previous="$(cat "$VERSION_FILE")"
  if [ "$previous" = "$DEV_SETUP_VERSION" ]; then
    log "Re-run detected; staying on version $DEV_SETUP_VERSION."
  else
    log "Update detected: $previous -> $DEV_SETUP_VERSION."
  fi
}

# Function: record_version_state
# What: Writes the current dev-setup version marker to VERSION_FILE.
# Why: Persists installed version for future upgrade detection.
# Params: none.
# Returns: 0 on success.
# Example: record_version_state
record_version_state() {
  echo "$DEV_SETUP_VERSION" >"$VERSION_FILE"
}

# Function: install_brew
# What: Installs Homebrew if missing and wires shellenv into ~/.zprofile.
# Why: Homebrew is the package backbone for the rest of the bootstrap.
# Params: none.
# Returns: 0 on success; exits on failure.
# Example: install_brew
install_brew() {
  if command -v brew >/dev/null 2>&1; then
    log "Homebrew already installed."
    return
  fi

  log "Installing Homebrew (will prompt you)..."
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

  # shellcheck disable=SC2016
  if [ -x /opt/homebrew/bin/brew ]; then
    log "Adding Homebrew shellenv for Apple Silicon."
    echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >>"${HOME}/.zprofile"
    eval "$(/opt/homebrew/bin/brew shellenv)"
  elif [ -x /usr/local/bin/brew ]; then
    log "Adding Homebrew shellenv for Intel macs."
    echo 'eval "$(/usr/local/bin/brew shellenv)"' >>"${HOME}/.zprofile"
    eval "$(/usr/local/bin/brew shellenv)"
  else
    error "Homebrew install did not produce a brew binary."
  fi
}

# Function: install_formula
# What: Idempotently installs a Homebrew formula.
# Why: Avoids reinstalling and keeps the script re-runnable.
# Params: $1 formula name; $2..$n optional flags.
# Returns: 0 on success.
# Example: install_formula git
install_formula() {
  local formula="$1"
  shift || true
  if brew ls --versions "$formula" >/dev/null 2>&1; then
    log "brew $formula already installed."
  else
    log "Installing brew formula: $formula"
    brew install "$formula" "$@"
  fi
}

# Function: install_cask
# What: Idempotently installs a Homebrew cask.
# Why: Provides GUI apps and some CLIs packaged as casks.
# Params: $1 cask name.
# Returns: 0 on success.
# Example: install_cask zed
install_cask() {
  local cask="$1"
  if brew list --cask "$cask" >/dev/null 2>&1; then
    log "brew cask $cask already installed."
  else
    log "Installing brew cask: $cask"
    brew install --cask "$cask"
  fi
}

# Function: install_uv
# What: Installs uv (Python toolchain manager) via official script.
# Why: uv is the allowed Python package manager per project policy.
# Params: none.
# Returns: 0 if installed; logs advice if PATH needs adjustment.
# Example: install_uv
install_uv() {
  if command -v uv >/dev/null 2>&1; then
    log "uv already installed."
    return
  fi

  log "Installing uv (Python toolchain manager)..."
  curl -LsSf https://astral.sh/uv/install.sh | sh

  if ! command -v uv >/dev/null 2>&1; then
    log "uv install completed; ensure ${HOME}/.local/bin is on your PATH."
  fi
}

# Function: clone_or_update
# What: Clones a git repo to dest_dir or fast-forwards it if already present.
# Why: Keeps helper repositories current without manual steps.
# Params: $1 repo URL; $2 destination directory.
# Returns: 0 on success.
# Example: clone_or_update https://github.com/example/repo ~/.local/share/repo
clone_or_update() {
  local repo_url="$1"
  local dest_dir="$2"

  if [ -d "$dest_dir/.git" ]; then
    log "Updating repo at $dest_dir"
    git -C "$dest_dir" pull --ff-only
  elif [ -d "$dest_dir" ]; then
    log "Skipping clone; $dest_dir exists but is not a git repo."
  else
    log "Cloning $repo_url -> $dest_dir"
    git clone "$repo_url" "$dest_dir"
  fi
}

# Function: link_bin
# What: Creates/updates a symlink in ~/.local/bin pointing at a target binary.
# Why: Exposes helper binaries on PATH without modifying system directories.
# Params: $1 target path; $2 link path.
# Returns: 0 on success.
# Example: link_bin ~/.local/share/tool/bin/tool ~/.local/bin/tool
link_bin() {
  local target="$1"
  local link_name="$2"
  mkdir -p "$(dirname "$link_name")"
  ln -sf "$target" "$link_name"
  log "Linked $link_name -> $target"
}

# Function: install_flutter_wrapper
# What: Clones flutter-wrapper and exposes a flutterw shim to avoid clobbering system Flutter.
# Why: Provides version switching and asset tooling via wrapper while keeping system Flutter optional.
# Params: none.
# Returns: 0 on success.
# Example: install_flutter_wrapper
install_flutter_wrapper() {
  local dest="${STATE_DIR}/flutter-wrapper"
  clone_or_update https://github.com/BadRat-in/flutter-wrapper "$dest"
  if [ -x "$dest/bin/flutter" ]; then
    link_bin "$dest/bin/flutter" "${LOCAL_BIN}/flutterw"
  fi
}

# Function: install_git_config_repo
# What: Clones git-config helper repo for optional application to ~/.gitconfig.
# Why: Centralizes git settings and commit hooks; left manual to avoid overwriting user config silently.
# Params: none.
# Returns: 0 on success.
# Example: install_git_config_repo
install_git_config_repo() {
  local dest="${STATE_DIR}/git-config"
  clone_or_update https://github.com/BadRat-in/git-config "$dest"
  log "Review git config repo at $dest before applying to your ~/.gitconfig."
}

# Function: install_zsh_plugins
# What: Clones core Zsh helper plugins for completion, suggestions, history search, and navigation.
# Why: Ensures consistent shell UX across machines using locally managed copies.
# Params: none.
# Returns: 0 on success.
# Example: install_zsh_plugins
install_zsh_plugins() {
  clone_or_update https://github.com/zsh-users/zsh-completions "${PLUGINS_DIR}/zsh-completions"
  clone_or_update https://github.com/zsh-users/zsh-syntax-highlighting "${PLUGINS_DIR}/zsh-syntax-highlighting"
  clone_or_update https://github.com/zsh-users/zsh-autosuggestions "${PLUGINS_DIR}/zsh-autosuggestions"
  clone_or_update https://github.com/zsh-users/zsh-history-substring-search "${PLUGINS_DIR}/zsh-history-substring-search"
  clone_or_update https://github.com/BadRat-in/zsh-interactive-cd "${PLUGINS_DIR}/zsh-interactive-cd"
  clone_or_update https://github.com/BadRat-in/zsh-you-should-use "${PLUGINS_DIR}/zsh-you-should-use"
}

# Function: ensure_cask_fonts_tap
# What: Ensures the Homebrew cask-fonts tap is available for font casks.
# Why: font-hack-nerd-font resides in homebrew/cask-fonts.
# Params: none.
# Returns: 0 on success.
# Example: ensure_cask_fonts_tap
ensure_cask_fonts_tap() {
  if brew tap | grep -q "^homebrew/cask-fonts$"; then
    log "brew tap homebrew/cask-fonts already present."
  else
    log "Adding brew tap homebrew/cask-fonts."
    brew tap homebrew/cask-fonts
  fi
}

# Function: install_nerd_font_hack
# What: Installs the Hack Nerd Font via Homebrew cask.
# Why: Provides the patched font needed for enhanced prompt glyphs.
# Params: none.
# Returns: 0 on success.
# Example: install_nerd_font_hack
install_nerd_font_hack() {
  ensure_cask_fonts_tap
  install_cask font-hack-nerd-font
}

# Function: ensure_starship_config
# What: Creates or amends ~/.config/starship.toml to use Nerd Font Symbols preset and show private package versions.
# Why: Nerd font glyphs require the preset; package.display_private must be true to surface private package versions.
# Params: none.
# Returns: 0 on success.
# Example: ensure_starship_config
ensure_starship_config() {
  local cfg="${HOME}/.config/starship.toml"
  mkdir -p "$(dirname "$cfg")"

  if [ ! -f "$cfg" ]; then
    if command -v starship >/dev/null 2>&1; then
      log "Generating starship config via preset nerd-font-symbols."
      starship preset nerd-font-symbols -o "$cfg" || log "Preset generation failed; falling back to minimal config."
    fi

    if [ ! -f "$cfg" ]; then
      cat >"$cfg" <<'EOF'
# Generated by dev-setup: fallback config with Nerd Font Symbols and private package display.
[package]
symbol = "󰏗 "
display_private = true
EOF
      log "Created fallback starship config with nerd-font symbols and package display."
      return
    fi
  fi

  # Ensure private packages are visible without duplicating [package] table.
  cfg_path="$cfg" python3 <<PY
from pathlib import Path
import re
import os

cfg = Path(os.environ["cfg_path"])
text = cfg.read_text()

pkg_match = re.search(r"\[package\](.*?)(\n\[|$)", text, re.S)
if pkg_match:
    block = pkg_match.group(1)
    if "display_private" not in block or "display_private = true" not in block:
        new_block = "[package]" + block.rstrip("\n")
        if not new_block.endswith("\n"):
            new_block += "\n"
        new_block += "display_private = true\n"
        start, end = pkg_match.span(0)
        text = text[:start] + new_block + text[end-1:]  # keep following section marker
else:
    if not text.endswith("\n"):
        text += "\n"
    text += "[package]\ndisplay_private = true\n"

cfg.write_text(text)
PY
  log "Ensured package.display_private=true in starship config without duplicating [package]."
}

# Function: install_starship
# What: Installs the Starship prompt via Homebrew and ensures config for private packages.
# Why: Provides a modern prompt that can show private package versions.
# Params: none.
# Returns: 0 on success.
# Example: install_starship
install_starship() {
  install_formula starship
  ensure_starship_config
}

# Function: install_optional_ai_clis
# What: Attempts brew installs for Codex, Gemini CLI, and Claude Code when available.
# Why: Automates AI agent CLIs without relying on npm (project policy forbids npm).
# Params: none.
# Returns: 0 on success; logs and skips if formula/cask not found.
# Example: install_optional_ai_clis
install_optional_ai_clis() {
  if brew info --formula codex >/dev/null 2>&1; then
    install_formula codex
  elif brew info --cask codex >/dev/null 2>&1; then
    install_cask codex
  else
    log "Skipping Codex CLI; brew formula/cask not found (add tap manually if available)."
  fi

  if brew info --formula gemini-cli >/dev/null 2>&1; then
    install_formula gemini-cli
  elif brew info --cask gemini-cli >/dev/null 2>&1; then
    install_cask gemini-cli
  else
    log "Skipping Gemini CLI; brew formula/cask not found."
  fi

  if brew info --cask claude-code >/dev/null 2>&1; then
    install_cask claude-code
  elif brew info --formula claude-code >/dev/null 2>&1; then
    install_formula claude-code
  else
    if brew tap anthropic/claude >/dev/null 2>&1 && brew info --cask claude-code >/dev/null 2>&1; then
      install_cask claude-code
    else
      log "Skipping Claude Code; brew cask/formula not found even after anthropic/claude tap."
    fi
  fi
}

# Function: write_post_install_notes
# What: Writes manual follow-up guidance to a state file.
# Why: Captures steps that cannot be automated safely (PATH, shell sourcing, git config).
# Params: none.
# Returns: 0 on success.
# Example: write_post_install_notes
write_post_install_notes() {
  cat >"${STATE_DIR}/POST_INSTALL.txt" <<'EOF'
Next steps (manual):
- Ensure ~/.local/bin is on PATH, e.g. add to ~/.zprofile:
    export PATH="$HOME/.local/bin:$PATH"
- For flutter-wrapper, run: flutterw doctor
- To wire zsh plugins, add to ~/.zshrc:
    source "${HOME}/.local/share/dev-setup/zsh-plugins/zsh-interactive-cd/zsh-interactive-cd.zsh"
    source "${HOME}/.local/share/dev-setup/zsh-plugins/zsh-you-should-use/zsh-you-should-use.zsh"
- Review git configuration at "${HOME}/.local/share/dev-setup/git-config" before applying.
- If brew installed Codex, Gemini CLI, or Claude Code, run their auth steps (codex login, gemini login, claude auth).
- Switch your terminal font to "Hack Nerd Font" (install provided). In iTerm2: Profiles -> Text -> Font.
- Enable starship in ~/.zshrc (after PATH): eval "$(starship init zsh)"
- Starship uses the nerd-font-symbols preset; package.display_private=true so private package versions show.
- Current dev-setup version is recorded in ~/.local/share/dev-setup/version.
EOF
  log "Wrote manual follow-ups to ${STATE_DIR}/POST_INSTALL.txt"
}

# Function: main
# What: Orchestrates the bootstrap flow in dependency order.
# Why: Provides a single entrypoint for the provisioning sequence.
# Params: CLI args (unused).
# Returns: 0 on success; exits early on failures.
# Example: main "$@"
main() {
  require_macos
  ensure_paths
  handle_version_state
  install_brew

  install_formula git
  install_formula node
  install_formula pnpm
  install_uv
  install_formula git-lfs
  install_formula wget

  install_cask zed
  install_starship
  install_nerd_font_hack

  install_flutter_wrapper
  install_git_config_repo
  install_zsh_plugins
  install_optional_ai_clis

  write_post_install_notes
  record_version_state

  log "Bootstrap completed."
}

main "$@"
