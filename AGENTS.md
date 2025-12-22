<!--
File: AGENTS.md
Purpose: Contributor-facing guidelines for the dev-setup bootstrap repository.
Problem: Give a concise, single-page reference so contributors know how to structure changes, run checks, and open PRs.
Role: Acts as the canonical contribution guide alongside README and repo-wide Codex rules in ~/.codex/instructions.md.
Usage: Read before making changes; keep this file updated when workflows evolve.
Assumptions: macOS-focused project; package managers restricted to pnpm (Node) and uv (Python); shell scripts are primary artifacts.
-->

# Repository Guidelines

## Project Structure & Module Organization
- `setup.sh`: Single entrypoint macOS bootstrap script; installs Homebrew dependencies, uv, pnpm, Zed, flutter-wrapper shim, git-config clone, Zsh plugins, and optional AI CLIs.
- `README.md`: User-facing run/usage notes and post-install steps.
- `AGENTS.md`: Contributor guide (this document).
- State created at runtime under `~/.local/share/dev-setup` and shims under `~/.local/bin`; not tracked in git.

## Build, Test, and Development Commands
- Run bootstrap locally: `./setup.sh` (macOS only).
- Shell lint (recommended): `shellcheck setup.sh`.
- Format Markdown (optional): `prettier --check README.md AGENTS.md` if you have pnpm/prettier available.

## Coding Style & Naming Conventions
- Shell: `bash`, `set -euo pipefail`, clear function headers, prefer idempotent operations. Two-space indent matches current script.
- Package managers: **pnpm** for Node, **uv** for Python; do not introduce npm/yarn/pip/poetry.
- Paths: Prefer `${HOME}/.local/{bin,share}` for user-level installs; avoid system-wide writes.
- Symlink helpers into `~/.local/bin` using `link_bin` pattern already in `setup.sh`.

## Testing Guidelines
- No automated test suite yet; at minimum run `shellcheck setup.sh` after edits.
- If adding scripts, provide a simple self-check (e.g., `script.sh --help` or a dry-run flag) and document it here and in README.

## Commit & Pull Request Guidelines
- Follow Conventional Commits enforced by the git-config hook (e.g., `feat: add ai cli installs`, `fix: handle missing brew tap`).
- Keep commits small and scoped; include doc updates with behavior changes.
- PRs should state: what changed, why, how to validate (commands run), and any manual steps for users.

## Security & Configuration Tips
- Do not auto-apply user git config; keep actions manual (as in current script).
- Avoid storing secrets; ensure auth steps for AI CLIs are manual (`codex login`, `gemini login`, `claude auth`).
- Prefer HTTPS clones; no SSH key assumptions.
