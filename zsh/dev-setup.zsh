#!/usr/bin/env zsh
# File: zsh/dev-setup.zsh
# Purpose: Provide a portable Zsh profile snippet that matches the expected dev environment across macOS systems.
# Problem: Team members need a consistent shell setup (brew env, Starship, aliases, plugin sourcing) without manually recreating it.
# Role: Template to be sourced from ~/.zshrc or equivalent; paths align with tools installed by setup.sh.
# Usage: Add `source /path/to/repo/zsh/dev-setup.zsh` near the top of ~/.zshrc.
# Assumptions: setup.sh was run; Homebrew is present; Starship and plugins are installed into ~/.local/share/dev-setup.

# -------------------------- Homebrew -----------------------------
if [ -x /opt/homebrew/bin/brew ]; then
  eval "$(/opt/homebrew/bin/brew shellenv)"
elif [ -x /usr/local/bin/brew ]; then
  eval "$(/usr/local/bin/brew shellenv)"
fi

# --------------------------- Starship ----------------------------
if command -v starship >/dev/null 2>&1; then
  eval "$(starship init zsh)"
fi

# -------------------------- Zsh Options ---------------------------
[ -f "$HOME/.zshopt" ] && source "$HOME/.zshopt"

# ------------------------- Start Aliases --------------------------
alias list-device="xcrun simctl list 'devices'"
alias flutterflow="$HOME/Library/Application\ Support/io.flutterflow.prod.mac/flutter/bin/flutter"

# Git shortcuts (from git-config installer)
if [ -f "$HOME/.config/git/git_shortcut.zsh" ]; then
  source "$HOME/.config/git/git_shortcut.zsh"
fi

alias ngurl='ngrok http --url=workable-externally-possum.ngrok-free.app'
# ------------------------- End Aliases --------------------------

# ------------------------- Start PATHs --------------------------
export PATH="$HOME/.flutter-wrapper/bin:$PATH"
export PATH="$PATH:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
# ------------------------- End PATHs ----------------------------

# ------------------------- Start ZSH Super Charge  --------------------------
fpath=($HOME/.zsh/zsh-completions/src /opt/homebrew/share/zsh/site-functions /usr/local/share/zsh/site-functions /usr/share/zsh/site-functions /usr/share/zsh/5.9/functions $fpath)

autoload -Uz compinit
compinit -d ~/.zcompdump

[ -f "$HOME/.local/share/dev-setup/zsh-plugins/zsh-completions/src" ] && fpath=("$HOME/.local/share/dev-setup/zsh-plugins/zsh-completions/src" $fpath)

if [ -f "$HOME/.local/share/dev-setup/zsh-plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh" ]; then
  source "$HOME/.local/share/dev-setup/zsh-plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh"
elif [ -f "$HOME/.zsh/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh" ]; then
  source "$HOME/.zsh/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh"
fi

if [ -f "$HOME/.local/share/dev-setup/zsh-plugins/zsh-autosuggestions/zsh-autosuggestions.zsh" ]; then
  source "$HOME/.local/share/dev-setup/zsh-plugins/zsh-autosuggestions/zsh-autosuggestions.zsh"
elif [ -f "$HOME/.zsh/zsh-autosuggestions/zsh-autosuggestions.zsh" ]; then
  source "$HOME/.zsh/zsh-autosuggestions/zsh-autosuggestions.zsh"
fi

if [ -f "$HOME/.local/share/dev-setup/zsh-plugins/zsh-history-substring-search/zsh-history-substring-search.zsh" ]; then
  source "$HOME/.local/share/dev-setup/zsh-plugins/zsh-history-substring-search/zsh-history-substring-search.zsh"
elif [ -f "$HOME/.zsh/zsh-history-substring-search/zsh-history-substring-search.zsh" ]; then
  source "$HOME/.zsh/zsh-history-substring-search/zsh-history-substring-search.zsh"
fi

# Prefer plugin copies managed by setup.sh under ~/.local/share/dev-setup
if [ -f "$HOME/.local/share/dev-setup/zsh-plugins/zsh-you-should-use/zsh-you-should-use.zsh" ]; then
  source "$HOME/.local/share/dev-setup/zsh-plugins/zsh-you-should-use/zsh-you-should-use.zsh"
elif [ -f "$HOME/.zsh/zsh-you-should-use/zsh-you-should-use.plugin.zsh" ]; then
  source "$HOME/.zsh/zsh-you-should-use/zsh-you-should-use.plugin.zsh"
fi

if [ -f "$HOME/.local/share/dev-setup/zsh-plugins/zsh-interactive-cd/zsh-interactive-cd.zsh" ]; then
  source "$HOME/.local/share/dev-setup/zsh-plugins/zsh-interactive-cd/zsh-interactive-cd.zsh"
elif [ -f "$HOME/.zsh/zsh-interactive-cd/zsh-interactive-cd.plugin.zsh" ]; then
  source "$HOME/.zsh/zsh-interactive-cd/zsh-interactive-cd.plugin.zsh"
fi
# ------------------------- End ZSH Super Charge  --------------------------
