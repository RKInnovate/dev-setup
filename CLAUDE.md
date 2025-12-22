# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## 🎯 Primary Goal: Identical Environment Across All Developer Machines

**Zero tolerance for "it works on my machine" excuses.**

This repository ensures **exact environment replication** across all developer machines through:
- Automated, idempotent installation scripts
- Strict package manager enforcement (pnpm for Node.js, uv for Python)
- Version-locked dependencies and state tracking
- Standardized shell configuration and tooling
- Mandatory pre-commit hooks and CI checks

---

## 📁 Critical Files to Review Before ANY Commit or PR

Before creating **any commits or pull requests**, you MUST review and comply with:

1. **`.git/hooks/commit-msg`** (line 1-140 in this repo)
   - Enforces Conventional Commits format
   - Validates type, scope, subject format
   - Checks subject length (10-72 chars), lowercase start, no trailing period
   - Validates `BREAKING CHANGE:` format if present
   - **Commit messages MUST pass this hook without modification**

2. **`.github/workflows/pr_checks.yml`** (if exists)
   - Defines CI checks, linting, tests required for PRs
   - All generated code must pass these checks
   - **Code that fails CI is NOT ready for commit**

3. **`global_conf/` directory** - Organization-wide development standards:
   - `claude.md` - Mandatory engineering standards
   - `git_best_practices.md` - Git workflow and naming conventions
   - `instructions.md` - Code quality requirements

**Failure to comply with these files means code is NOT ready for commit.**

---

## 📦 Package Manager Policy (ABSOLUTE REQUIREMENT)

**⚠️ This is NON-NEGOTIABLE and STRICTLY ENFORCED ⚠️**

### Allowed Package Managers ONLY

| Language | Required Package Manager | Lockfile |
|----------|--------------------------|----------|
| **Python** | **`uv`** | `uv.lock` |
| **Node.js (JS/TS)** | **`pnpm`** | `pnpm-lock.yaml` |

### FORBIDDEN Package Managers

**NEVER use these tools under ANY circumstances:**

- **Python**: `pip`, `pipenv`, `poetry`, `conda`
- **Node.js**: `npm`, `yarn`, `bun`

### Rules

1. **If a different package manager is detected or requested:**
   - **STOP immediately**
   - **Explicitly warn** that this is **NOT ALLOWED**
   - **DO NOT generate commands or configuration** for incorrect tools
   - Guide user to correct package manager (`uv` or `pnpm`)

2. **Never mix package managers** in the same project or commit

3. **Lockfiles must match** the chosen package manager

4. **Installation commands:**
   ```bash
   # Python - ONLY use uv
   uv pip install <package>
   uv sync

   # Node.js - ONLY use pnpm
   pnpm install
   pnpm add <package>
   ```

---

## 🔧 Project Overview

This is a **macOS-only** development environment bootstrap tool that automates installation of a standardized dev stack using Homebrew. The single entrypoint script (`setup.sh`) provisions tools, editors, fonts, and shell enhancements with **idempotent, version-aware installation logic**.

### Key Components

- **`setup.sh`**: Main bootstrap script (440 lines)
- **`zsh/dev-setup.zsh`**: Portable Zsh profile template
- **`tests/test_setup.sh`**: Stubbed regression tests (no network, safe to run)
- **State directory**: `~/.local/share/dev-setup/` (version marker, cloned repos, plugins)
- **Shims directory**: `~/.local/bin/` (exposed executables like `flutterw`)

### What Gets Installed

- Homebrew (if missing) + core formulas: git, node, pnpm, git-lfs, wget
- uv (Python toolchain manager) via official installer
- Zed editor (cask)
- Starship prompt with Hack Nerd Font
- Flutter wrapper (`flutterw` shim)
- Git config helper repo (manual application)
- Zsh plugins (completions, syntax highlighting, autosuggestions, history search)
- Optional AI CLIs: Codex, Gemini CLI, Claude Code (best-effort)

---

## 🚀 Key Commands

### Bootstrap Installation
```bash
./setup.sh
```
- Installs all dependencies
- Updates cloned repos if already present
- Records version state in `~/.local/share/dev-setup/version`
- **Safe to re-run**: Skips existing tools (idempotent)
- **macOS only**: Exits with error on non-Darwin systems

### Testing (Safe, No Network)
```bash
bash tests/test_setup.sh
```
- Uses temporary HOME and stubbed tools (brew, git, curl, starship, uv)
- Validates version recording and Starship config patching
- Leaves no state after completion

### Linting
```bash
shellcheck setup.sh
```
**Required before committing any shell script changes.**

---

## 🏗️ Architecture Deep Dive

### Core Design Principles (setup.sh)

1. **Idempotency**: All install functions check existence before acting
   - `install_formula()` checks `brew ls --versions` before installing
   - `install_cask()` checks `brew list --cask` before installing
   - `clone_or_update()` fast-forwards existing repos or clones new ones

2. **Version Tracking**: Records `DEV_SETUP_VERSION` (line 13) in state file
   - First run: Creates `~/.local/share/dev-setup/version`
   - Re-run: Detects version, logs scenario (first-run/update/re-run)
   - Update detection: Compares previous version with current

3. **Error Handling**: `set -euo pipefail` (line 2)
   - `error()` helper (lines 35-38) exits on critical failures
   - Validates macOS-only with `require_macos()` (lines 46-50)

4. **State Isolation**: User-level paths prevent system pollution
   - State: `~/.local/share/dev-setup/`
   - Bins: `~/.local/bin/`
   - Config: `~/.config/starship.toml`

### Critical Functions Reference

| Function | Lines | Purpose |
|----------|-------|---------|
| `main()` | 412-440 | Orchestrates install flow in dependency order |
| `install_brew()` | 99-120 | Installs Homebrew, wires shellenv to `~/.zprofile` |
| `install_formula()` | 128-137 | Idempotent brew formula installer |
| `install_cask()` | 145-153 | Idempotent brew cask installer |
| `clone_or_update()` | 181-194 | Clone repo or fast-forward existing |
| `link_bin()` | 202-208 | Symlink executables to `~/.local/bin` |
| `ensure_starship_config()` | 283-332 | Create/patch starship config (sets `display_private=true`, avoids duplicate `[package]` blocks) |
| `install_optional_ai_clis()` | 351-379 | Best-effort AI CLI installs, graceful skip if unavailable |

### State & Path Locations

```
~/.local/
├── bin/                          # Exposed executables (on PATH)
│   └── flutterw -> ../share/dev-setup/flutter-wrapper/bin/flutter
├── share/
│   └── dev-setup/                # State directory
│       ├── version               # Version marker (DEV_SETUP_VERSION)
│       ├── POST_INSTALL.txt      # Manual steps after bootstrap
│       ├── flutter-wrapper/      # Cloned from BadRat-in/flutter-wrapper
│       ├── git-config/           # Cloned from BadRat-in/git-config (manual apply)
│       └── zsh-plugins/          # Cloned Zsh plugins
│           ├── zsh-completions/
│           ├── zsh-syntax-highlighting/
│           ├── zsh-autosuggestions/
│           ├── zsh-history-substring-search/
│           ├── zsh-interactive-cd/
│           └── zsh-you-should-use/
~/.config/
└── starship.toml                 # Starship config (patched by script)
```

### Zsh Profile Template (zsh/dev-setup.zsh)

Portable snippet for sourcing from `~/.zshrc`. Provides:
- Homebrew shellenv wiring (Intel/Apple Silicon detection)
- Starship prompt initialization
- PATH setup (flutter-wrapper, standard bins)
- Zsh plugin sourcing (prefers `~/.local/share/dev-setup/zsh-plugins/`)
- Git shortcuts and aliases

**Usage:** Add to `~/.zshrc`:
```zsh
source /path/to/repo/zsh/dev-setup.zsh
```

---

## 📝 Git Workflow & Naming Standards

### Branch Naming (Canonical Format)

**Format:** `<type>/<ISSUE_NUMS>-<short-kebab-description>`

**Allowed types:**
- `feat`, `fix`, `chore`, `docs`, `refactor`, `perf`, `test`, `ci`, `style`, `deploy`, `hotfix`

**Examples:**
```
feat/3-add-login-api
fix/7-null-pointer-dashboard
chore/12-update-deps
feat/3-4-add-reporting-endpoints  # Multiple issues
hotfix/15-critical-security-patch
```

**Rules:**
- Use **issue numbers** (no `#`), joined by hyphens for multiple issues
- Keep descriptions **short and descriptive** (avoid generic words like "update")
- **hotfix/** for urgent changes requiring immediate merge

### Commit Message Format (Conventional Commits - ENFORCED)

**Format:**
```
<type>(<scope>)?: <subject>

<body> (optional)

<footer(s)> (optional)
```

**Type:** Same as branch types (lowercase)
**Scope:** Optional, short module/area identifier (e.g., `auth`, `api`, `ui`)
**Subject:**
- Imperative mood, present tense
- Lowercase start (enforced by hook)
- 10-72 characters (enforced by hook)
- No trailing period (enforced by hook)

**Body:** Explain **what and why** (not how), wrap at ~72 chars

**Footer:**
- Issue references: `Fixes #3` or `Fixes #3, #4` (auto-closes on merge)
- Breaking changes: `BREAKING CHANGE: <description>`

**Examples:**

✅ **GOOD:**
```
feat(auth): add JWT login endpoint

Adds /auth/login endpoint using JWT for stateless sessions.
This implements the initial user login flow.

Fixes #3
```

```
fix(ui): prevent crash on missing avatar

Check for null avatar and fallback to initials.

Fixes #7, #8
```

```
refactor(api): split user service into modules

BREAKING CHANGE: user service endpoints changed from /user/v1/* to /user/v2/*
```

❌ **BAD:**
```
Update stuff                     # No type, vague
feat: Add feature               # Capital letter in subject
fix: fixed bug.                 # Past tense, trailing period
feature: new login              # Wrong type (use 'feat')
feat(auth): add                 # Subject too short (<10 chars)
```

### PR Title & Description

**Title format:** `<type>: <short summary> (#ISSUE)`

**Examples:**
```
feat: add JWT login endpoint (#3)
fix: prevent null avatar crash (#7)
```

**PR Description Template:**
```markdown
### Summary
Short description of the changes.

### Changes
- Bullet list of main changes
- Any migrations / DB changes

### Related Issues
Fixes #3

### QA / Testing
Steps to reproduce / test plan.

### Screenshots / Notes
Attach screenshots or important notes here.
```

---

## 🧩 Code Documentation & Commenting Rules (MANDATORY)

**Documentation is NOT optional, even for simple code.**

### 1. File-Level Documentation (Required for EVERY file)

Every file **must begin** with a detailed DocString or header comment:

```bash
#!/usr/bin/env bash
# File: path/to/script.sh
# Purpose: [What this file does]
# Problem: [What problem it solves]
# Role: [Its role in overall architecture]
# Usage: [How and when to use it]
# Design choices: [Important architectural decisions]
# Assumptions: [Required preconditions or constraints]
```

**Example from setup.sh (lines 1-10):**
```bash
#!/usr/bin/env bash
set -euo pipefail

# Dev Setup Bootstrap (macOS only)
# Purpose: Provision a consistent local dev environment in minutes using Homebrew plus curated tools.
# Problem: Team members spend hours configuring git/node/python/flutter/editors manually; this script standardizes it.
# Architectural role: Single entrypoint shell script; delegates installs to Homebrew or official installers, and clones helper repos under ~/.local/share/dev-setup.
# Usage: Run ./setup.sh on macOS; follow POST_INSTALL.txt for manual steps like PATH/zsh config.
# Design choices: Uses ~/.local/{bin,share} for shims/state; keeps clones isolated from repo; idempotent installs via brew checks; leaves git config application manual for safety.
# Assumptions: macOS host with curl and git; user consents to Homebrew installation and cask prompts; network access available during run.
```

### 2. Function/Method Documentation (Required for EVERY function)

**Example from setup.sh (lines 19-27):**
```bash
# Function: log
# What: Prints a normalized log line with a dev-setup prefix.
# Why: Provides consistent user feedback during bootstrap steps.
# Params: $1..$n - message tokens to print.
# Returns: 0 always.
# Example: log "Installing brew"
log() {
  printf "[dev-setup] %s\n" "$*"
}
```

**Required elements:**
- **What**: Function's purpose
- **Why**: Reason it exists
- **Params**: Each parameter with name, type, purpose
- **Returns**: Return value(s)
- **Example**: Usage example (where meaningful)
- **Edge cases**: Constraints or special conditions

**This applies to:**
- Utility functions
- Private/internal helpers
- Small or "obvious" methods

### 3. Inline Comments

**Explain WHY, not just WHAT:**

✅ **GOOD:**
```bash
# Nerd font glyphs require preset; package.display_private shows private package versions
ensure_starship_config
```

❌ **BAD:**
```bash
# Call ensure starship config
ensure_starship_config
```

---

## 🧪 Linting & Testing (NON-NEGOTIABLE)

**Claude must ALWAYS ensure code quality before committing.**

### Pre-Commit Checklist (MANDATORY)

**Before ANY commit or PR:**

1. **Run all relevant linters:**
   ```bash
   # Shell scripts
   shellcheck setup.sh tests/test_setup.sh

   # JS/TS (if applicable)
   pnpm run lint
   pnpm run format

   # Python (if applicable)
   uv run ruff check .
   uv run black --check .
   ```

2. **Run all test suites:**
   ```bash
   # This repo
   bash tests/test_setup.sh

   # General pattern
   pnpm test         # Node.js projects
   uv run pytest     # Python projects
   ```

3. **Fix ALL errors:**
   - **DO NOT suppress errors** unless explicitly allowed
   - **DO NOT commit failing or flaky tests**
   - **DO NOT commit lint violations**

### ⚠️ NEVER Commit:
- Failing tests
- Lint violations
- Breaking changes (without BREAKING CHANGE footer)
- Code that hasn't been linted
- Code that hasn't been tested

---

## 🔒 Commit & PR Quality Standards

### Commit Requirements

Commits must be:
- **Small**: Single logical change per commit
- **Focused**: One purpose, one commit
- **Logically grouped**: Related changes together

### PR Requirements

PRs should:
- **Clearly describe** what changed and why
- **Reference relevant files** or architectural concerns
- **Include test results** and validation steps
- **Avoid unrelated changes** (no scope creep)
- **Pass all CI checks** before requesting review

---

## 🛠️ Development Workflow

### Making Changes

1. **Create feature branch:**
   ```bash
   git checkout -b feat/42-add-feature-name
   ```

2. **Make changes following:**
   - File-level documentation (every file)
   - Function-level documentation (every function)
   - Inline comments (explain WHY)
   - Two-space indent (shell scripts)
   - Idempotent operations

3. **Test changes:**
   ```bash
   # For setup.sh changes
   shellcheck setup.sh
   bash tests/test_setup.sh

   # Test actual bootstrap (optional, safe to re-run)
   ./setup.sh
   ```

4. **Commit with proper format:**
   ```bash
   git add .
   git commit -m "feat(installer): add new tool support

   Adds idempotent installation for <tool-name>.

   Fixes #42"
   ```
   - Commit message will be validated by `.git/hooks/commit-msg`
   - Must pass Conventional Commits format

5. **Create PR:**
   ```bash
   git push -u origin feat/42-add-feature-name
   # Then create PR via GitHub with proper template
   ```

### When Adding New Tools to setup.sh

**Pattern for Homebrew formula:**
```bash
# In main() function (line 412-440), add:
install_formula <formula-name>
```

**Pattern for Git repository clone:**
```bash
# Create function (follow existing patterns):
install_my_tool() {
  local dest="${STATE_DIR}/my-tool"
  clone_or_update https://github.com/user/my-tool "$dest"

  # Optional: symlink binary if needed
  if [ -x "$dest/bin/tool" ]; then
    link_bin "$dest/bin/tool" "${LOCAL_BIN}/tool"
  fi
}

# Call in main():
install_my_tool
```

**Don't forget:**
- Update README.md "What gets installed" section
- Update `POST_INSTALL.txt` template if manual steps required
- Bump `DEV_SETUP_VERSION` (line 13) if behavior changes
- Add tests to `tests/test_setup.sh` if applicable

---

## 🔐 Environment Consistency Requirements

To ensure "it works exactly the same on every machine":

### 1. Version Locking
- **Homebrew**: Formulas install latest stable (acceptable variance)
- **Python packages**: Use `uv.lock` (commit to repo)
- **Node packages**: Use `pnpm-lock.yaml` (commit to repo)
- **Dev-setup version**: Tracked in `~/.local/share/dev-setup/version`

### 2. State Management
- All state under `~/.local/share/dev-setup/` (never in repo)
- Shims under `~/.local/bin/` (never in repo)
- User must add `~/.local/bin` to PATH (documented in POST_INSTALL.txt)

### 3. Configuration Standardization
- Starship config: Standardized via `ensure_starship_config()`
- Zsh plugins: Cloned to known locations
- Git config: Provided but manual application (safety)
- Terminal font: Hack Nerd Font installed, user applies

### 4. Testing Strategy
- **Local validation**: `bash tests/test_setup.sh` (no network, safe)
- **Real bootstrap test**: `./setup.sh` (idempotent, safe to re-run)
- **Shellcheck**: `shellcheck setup.sh` (catches shell errors)

### 5. Documentation
- `POST_INSTALL.txt`: Generated with manual steps
- `README.md`: User-facing quickstart
- `CLAUDE.md`: This file (development standards)
- `docs/guide.md`: Detailed reference

---

## 📚 Common Development Tasks

### Testing Without Side Effects
```bash
# Stubbed tests (no network, temporary HOME)
bash tests/test_setup.sh
```

### Testing Actual Bootstrap (Safe to Re-run)
```bash
# Full bootstrap on your machine
./setup.sh

# Script is idempotent:
# - Skips existing formulas/casks
# - Fast-forwards existing git repos
# - Updates version marker
```

### Linting Before Commit
```bash
shellcheck setup.sh
shellcheck tests/test_setup.sh
shellcheck zsh/dev-setup.zsh
```

### Updating Dev-Setup Version
```bash
# In setup.sh line 13:
DEV_SETUP_VERSION="0.3.0"  # Bump version

# Script will detect upgrade on next run
```

### Modifying Starship Config Behavior
Edit `ensure_starship_config()` function (lines 283-332). The inline Python script handles TOML patching to avoid duplicate `[package]` sections.

### Adding Zsh Plugin
```bash
# In install_zsh_plugins() function (lines 242-249):
clone_or_update https://github.com/user/plugin "${PLUGINS_DIR}/plugin-name"

# In zsh/dev-setup.zsh, add sourcing:
if [ -f "$HOME/.local/share/dev-setup/zsh-plugins/plugin-name/plugin.zsh" ]; then
  source "$HOME/.local/share/dev-setup/zsh-plugins/plugin-name/plugin.zsh"
fi
```

---

## ⚠️ Important Constraints & Assumptions

### Platform Requirements
- **macOS ONLY**: Script exits on non-Darwin systems (line 46-50)
- **Requires**: `curl`, `git` (assumed present on macOS)
- **Network required**: Homebrew install, git clones, uv installer

### User Interactions
- **Homebrew install prompts** user (first-time only)
- **No silent system modifications** (user consent required)
- **Manual steps documented** in `POST_INSTALL.txt`

### Manual Steps After Bootstrap
1. Add `~/.local/bin` to PATH in `~/.zprofile`
2. Source `zsh/dev-setup.zsh` from `~/.zshrc`
3. Switch terminal font to **Hack Nerd Font**
4. Run `flutterw doctor` (downloads Flutter)
5. Review and apply `~/.local/share/dev-setup/git-config`
6. Authenticate AI CLIs if installed (`codex login`, `gemini login`, `claude auth`)

---

## ✅ Pre-Commit Validation Checklist

Before committing ANY code change, verify:

- [ ] Shellcheck passes: `shellcheck setup.sh`
- [ ] Tests pass: `bash tests/test_setup.sh`
- [ ] File-level documentation present (ALL files)
- [ ] Function-level documentation present (ALL functions)
- [ ] Inline comments explain WHY (not just what)
- [ ] Commit message follows Conventional Commits format
- [ ] Commit message has type, scope (optional), subject (10-72 chars, lowercase)
- [ ] Commit message references issue (`Fixes #N`)
- [ ] No usage of forbidden package managers (npm, yarn, pip, poetry, bun)
- [ ] Only `pnpm` (Node.js) and `uv` (Python) used
- [ ] Idempotent operations (safe to re-run)
- [ ] No breaking changes without `BREAKING CHANGE:` footer
- [ ] README.md updated if user-facing changes
- [ ] `DEV_SETUP_VERSION` bumped if behavior changes

---

## 🚫 Absolute Prohibitions

**NEVER:**
- Use npm, yarn, bun, pip, pipenv, poetry, conda
- Commit code without file-level documentation
- Commit code without function-level documentation
- Suppress linting errors without explicit reason
- Commit failing tests
- Skip running tests before commit
- Use commit messages without type prefix
- Use commit messages with uppercase subject start
- Use commit messages shorter than 10 characters
- Create branches without descriptive names (e.g., `feat/42` - must be `feat/42-description`)
- Make system-wide modifications (use `~/.local/` instead)
- Auto-apply git config (leave manual for safety)

---

## 📖 Reference Documentation Files

- **README.md**: User quickstart, what gets installed, post-install steps
- **AGENTS.md**: Contributor guidelines, coding style, commit expectations
- **docs/guide.md**: Detailed install flow, state paths, environment replication
- **docs/hierarchy-diagram.md**: Visual repo and runtime state layout
- **global_conf/claude.md**: Mandatory engineering standards (THIS IS LAW)
- **global_conf/git_best_practices.md**: Git workflow standards (THIS IS LAW)
- **global_conf/instructions.md**: Code quality requirements (THIS IS LAW)

---

## 🎓 Key Takeaways

1. **Environment consistency is paramount** - exact replication across machines
2. **Package managers are STRICTLY enforced** - only pnpm and uv
3. **Documentation is MANDATORY** - file-level, function-level, inline comments
4. **Testing is NON-NEGOTIABLE** - lint and test before every commit
5. **Git workflow is standardized** - branch naming, commit format, PR templates
6. **Idempotency is required** - all operations safe to re-run
7. **State isolation is enforced** - user-level paths only
8. **Commit hooks are law** - must pass Conventional Commits validation
