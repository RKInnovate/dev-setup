# Dev Setup Repository Guide

This document captures the full behavior of the macOS dev-setup project, including install flow, state locations, environment setup, and maintenance practices.

## Repository Contents
- `setup.sh` — primary installer; idempotent and version-aware; provisions Homebrew dependencies, Starship, Hack Nerd Font, Zsh plugins, Flutter wrapper, git-config repo, optional AI CLIs.
- `README.md` — quickstart usage and post-install checklist.
- `AGENTS.md` — contributor guidelines (coding standards, commit/PR expectations).
- `tests/test_setup.sh` — stubbed, no-network regression script.
- `zsh/dev-setup.zsh` — standard Zsh profile snippet to replicate the environment across machines.
- `docs/guide.md` — this detailed reference.
- `docs/hierarchy-diagram.md` — visual map of repo and runtime state layout.

## Install Flow (setup.sh)
1) OS guard: exits unless on macOS.
2) Ensures state dirs: `~/.local/share/dev-setup`, `~/.local/bin`, plugin dir.
3) Version handling: reads/writes `~/.local/share/dev-setup/version`; logs first-run vs update.
4) Homebrew: installs if missing; wires shellenv into `~/.zprofile`.
5) Core formulas: `git`, `node`, `pnpm`, `git-lfs`, `wget`; `uv` via official script.
6) Apps/fonts/prompt: `zed` cask; `starship` with nerd-font-symbols preset; Hack Nerd Font via `homebrew/cask-fonts` tap.
7) Flutter wrapper: clones `BadRat-in/flutter-wrapper`; symlinks `flutterw` to `~/.local/bin`.
8) Git config helper: clones `BadRat-in/git-config` (left manual to apply).
9) Zsh plugins: clones zsh-completions, zsh-syntax-highlighting, zsh-autosuggestions, zsh-history-substring-search, zsh-interactive-cd, zsh-you-should-use.
10) AI CLIs: best-effort Homebrew install for codex, gemini-cli, claude-code; skipped gracefully if unavailable.
11) Post-install notes: writes `POST_INSTALL.txt` under state with manual steps.

## State & Paths
- State: `~/.local/share/dev-setup`
- Shims: `~/.local/bin` (e.g., `flutterw`)
- Version marker: `~/.local/share/dev-setup/version`
- Zsh plugins: `~/.local/share/dev-setup/zsh-plugins/...`
- Starship config: `~/.config/starship.toml` (ensures `display_private = true`)

## Environment Replication
- Source `zsh/dev-setup.zsh` from `~/.zshrc` to standardize Homebrew shellenv, Starship init, aliases, PATH, and plugins (prefers the cloned copies above; falls back to legacy `~/.zsh/...`).
- Manual fonts: set terminal font to **Hack Nerd Font** after install.

## Tests
- Run `bash tests/test_setup.sh` for a safe, stubbed validation (no network, temp HOME). Verifies version recording and Starship package block uniqueness.

## Manual Steps After Install
- Add `~/.local/bin` to PATH (if not already).
- Enable Starship: `eval "$(starship init zsh)"` in `~/.zshrc`.
- Source plugins (or use the provided zsh snippet).
- Run `flutterw doctor`.
- Review `~/.local/share/dev-setup/git-config` before applying to `~/.gitconfig`.
- Authenticate optional AI CLIs if installed (`codex login`, `gemini login`, `claude auth`).

## Policies & Assumptions
- Package managers: pnpm for Node.js, uv for Python; npm/yarn/pip/poetry are intentionally excluded.
- macOS-only; expects curl and git available.
- Installs are idempotent; reruns will skip existing tools and update repos.

## Maintenance Notes
- Update `DEV_SETUP_VERSION` in `setup.sh` when behavior changes; reruns will log detected upgrades.
- When adding tools, prefer Homebrew formulas/casks; if unavailable, document the installer and guard with idempotent checks.
