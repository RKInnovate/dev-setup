# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## 🎯 Primary Goal: Identical Environment Across All Developer Machines

**Zero tolerance for "it works on my machine" excuses.**

This repository ensures **exact environment replication** across all developer machines through:
- Automated, three-stage progressive setup (5min critical → 10min full → 15min polish)
- Parallel task execution (8+ concurrent tasks) for maximum speed
- Version-locked dependencies (Brewfile.lock.json + versions.lock)
- Strict package manager enforcement (pnpm for Node.js, uv for Python)
- Self-updating binary with GitHub releases
- Comprehensive verification and diagnostics

---

## 📁 Critical Files to Review Before ANY Commit or PR

Before creating **any commits or pull requests**, you MUST review and comply with:

1. **`.git/hooks/commit-msg`**
   - Enforces Conventional Commits format
   - Validates type, scope, subject format
   - **Commit messages MUST pass this hook without modification**

2. **`.github/workflows/ci.yml`** and **`.github/workflows/release.yml`**
   - CI: Runs tests, linting, build validation on every push/PR
   - Release: Automates binary builds and GitHub releases on version tags
   - All code must pass CI checks before merge

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
| **Go** | **`go mod`** | `go.sum` |

### FORBIDDEN Package Managers

**NEVER use these tools under ANY circumstances:**

- **Python**: `pip`, `pipenv`, `poetry`, `conda`
- **Node.js**: `npm`, `yarn`, `bun`
- **Go**: Manual vendor management without go.mod

### Rules

1. **If a different package manager is detected or requested:**
   - **STOP immediately**
   - **Explicitly warn** that this is **NOT ALLOWED**
   - **DO NOT generate commands or configuration** for incorrect tools
   - Guide user to correct package manager (`uv`, `pnpm`, or `go mod`)

2. **Never mix package managers** in the same project or commit

3. **Lockfiles must match** the chosen package manager

---

## 🔧 Project Overview

This is a **macOS development environment bootstrap tool** written in Go that reduces developer setup time from **days to 30 minutes** (productive in just 5 minutes!). It uses a three-stage progressive installation approach with parallel task execution.

### Key Technology Stack

- **Language**: Go 1.21+
- **CLI Framework**: Cobra (professional CLI with subcommands)
- **Configuration**: YAML (stage configs) + TOML (version locking)
- **Deployment**: Single binary, auto-updates via GitHub releases
- **CI/CD**: GitHub Actions (tests, linting, releases)
- **Package Manager**: Homebrew (for macOS tools)

### What Gets Installed

**Stage 1 (5 min, blocking - developer productive):**
- Homebrew + Git, Node, pnpm, Python, uv
- Zed editor
- Essential CLI tools

**Stage 2 (10 min, background - full stack):**
- Flutter wrapper (`flutterw`)
- Zsh plugins (completions, syntax highlighting, etc.)
- Starship prompt
- Development utilities

**Stage 3 (15 min, background - polish):**
- Fonts (Hack Nerd Font)
- AI CLIs (Codex, Gemini, Claude)
- Docker and optional tools

---

## 🚀 Key Commands

### Build & Run

```bash
# Build binary
make build

# Build for all architectures
make build-all

# Install to /usr/local/bin
make install

# Run without installing
make run
```

### Testing & Quality

```bash
# Run all tests
make test

# Run linter
make lint

# Run with verbose output
go test -v ./...

# Run specific test
go test ./internal/installer -run TestParallelExecutor -v

# Check test coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Using devsetup

```bash
# Install development environment
devsetup install

# Fast mode (Stage 1 only - 5 minutes)
devsetup install --fast

# Skip optional tools (Stages 1-2 only)
devsetup install --skip-optional

# Dry run (see what would be installed)
devsetup install --dry-run

# Verify environment matches versions.lock
devsetup verify
devsetup verify --fix  # Auto-fix mismatches

# Run diagnostics
devsetup doctor

# Check installation status
devsetup status

# Update devsetup binary
devsetup update
devsetup update --check  # Check without installing

# Show version
devsetup --version
```

### One-Line Install (for end users)

```bash
curl -fsSL https://raw.githubusercontent.com/rkinnovate/dev-setup/main/bootstrap.sh | bash
```

---

## 🏗️ Architecture Deep Dive

### Project Structure

```
dev-setup/
├── cmd/devsetup/              # CLI entry point
│   └── main.go               # Cobra commands (install, verify, doctor, update)
├── internal/                  # Internal packages (not importable)
│   ├── config/               # Configuration loading and validation
│   │   ├── models.go        # Data structures (StageConfig, Task, VersionsLock)
│   │   ├── loader.go        # YAML/TOML parsers
│   │   └── loader_test.go   # Config tests
│   ├── installer/            # Installation orchestration
│   │   ├── installer.go     # High-level stage management
│   │   ├── parallel.go      # Parallel task execution engine
│   │   ├── installer_test.go
│   │   └── parallel_test.go
│   ├── updater/              # Self-update functionality
│   │   ├── updater.go       # GitHub releases integration
│   │   └── updater_test.go
│   └── ui/                   # Terminal UI (progress bars, colors)
│       └── progress.go
├── configs/                   # Stage configuration files
│   ├── stage1.yaml          # Critical path (5 min)
│   ├── stage2.yaml          # Full stack (10 min)
│   └── stage3.yaml          # Polish (15 min)
├── .github/workflows/         # CI/CD pipelines
│   ├── ci.yml               # Tests, linting, builds
│   └── release.yml          # Automated releases
├── Brewfile                   # Homebrew package declarations
├── versions.lock              # Version pinning (TOML)
├── bootstrap.sh               # One-line installer script
├── Makefile                   # Build automation
├── go.mod / go.sum            # Go dependencies
├── README.md                  # User documentation
├── CONTRIBUTING.md            # Developer guide
└── CLAUDE.md                  # This file

```

### Core Design Principles

1. **Three-Stage Progressive Setup**
   - Stage 1: Developer can start coding after 5 minutes
   - Stages 2-3: Complete in background while developer works
   - Minimizes time-to-productivity

2. **Parallel Task Execution**
   - Tasks in same `parallel_group` run concurrently
   - Semaphore pattern limits concurrency (max 8 concurrent)
   - Sequential tasks run one at a time
   - Goroutines + channels for coordination

3. **Version Locking**
   - `Brewfile.lock.json`: Locks Homebrew package versions
   - `versions.lock`: TOML file for git repos and tools
   - Ensures identical versions across all machines

4. **Idempotency**
   - All operations safe to re-run
   - Checks existence before installing
   - Updates existing installations

5. **Self-Updating**
   - Checks GitHub releases for new versions
   - Downloads and atomically replaces binary
   - Preserves backup of old version

### Key Packages Reference

| Package | Purpose |
|---------|---------|
| `cmd/devsetup` | CLI entry point with Cobra commands |
| `internal/config` | Configuration loading (YAML, TOML, Brewfile) |
| `internal/installer` | Stage orchestration and parallel execution |
| `internal/updater` | Self-update via GitHub releases |
| `internal/ui` | Rich terminal UI (progress bars, colors) |

### State & Path Locations

```
~/.local/
├── bin/                          # Exposed executables (on PATH)
│   └── flutterw -> ../share/dev-setup/flutter-wrapper/bin/flutter
├── share/
│   └── dev-setup/                # State directory
│       ├── state.json            # Installation state
│       ├── flutter-wrapper/      # Cloned from rkinnovate/flutter-wrapper
│       ├── git-config/           # Cloned from rkinnovate/git-config
│       └── zsh-plugins/          # Cloned Zsh plugins
│           ├── zsh-completions/
│           ├── zsh-syntax-highlighting/
│           ├── zsh-autosuggestions/
│           ├── zsh-history-substring-search/
│           ├── zsh-interactive-cd/
│           └── zsh-you-should-use/
~/.config/
└── starship.toml                 # Starship prompt config
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
```

### Commit Message Format (Conventional Commits - ENFORCED)

**Format:**
```
<type>(<scope>)?: <subject>

<body> (optional)

<footer(s)> (optional)
```

**Rules:**
- Type: Same as branch types (lowercase)
- Scope: Optional module/area (e.g., `installer`, `config`, `updater`)
- Subject: Imperative mood, lowercase start, 10-72 chars, no trailing period
- Body: Explain what and why (not how)
- Footer: Issue references (`Fixes #N`) or `BREAKING CHANGE:`

**Examples:**

✅ **GOOD:**
```
feat(installer): add parallel task execution

Implements goroutine-based parallel execution with
semaphore pattern to limit concurrency to 8 tasks.

Fixes #15
```

```
fix(updater): handle GitHub API redirects

GitHub API can return redirects causing update check
to fail. Now properly follows redirects.

Fixes #23
```

---

## 🧩 Code Documentation & Commenting Rules (MANDATORY)

**Documentation is NOT optional, even for simple code.**

### 1. File-Level Documentation (Required for EVERY file)

Every Go file **must begin** with a detailed comment block:

```go
// File: internal/installer/parallel.go
// Purpose: Parallel task execution engine with concurrency control
// Problem: Need to execute multiple tasks concurrently without overwhelming system
// Role: Core execution engine that runs tasks in parallel groups with semaphore pattern
// Usage: Create ParallelExecutor, call Execute() with task slice
// Design choices: Goroutines + channels; semaphore limits concurrency; retries with backoff
// Assumptions: Bash available for command execution; tasks are independent within groups
```

### 2. Function/Method Documentation (Required for EVERY function)

**Example from installer.go:**
```go
// RunStage executes a single installation stage
// What: Loads stage config, executes tasks via parallel executor, updates state
// Why: Main entry point for stage execution with complete error handling
// Params: stageConfigPath - path to stage YAML file (e.g. "configs/stage1.yaml")
// Returns: Error if stage failed, nil if successful
// Example: err := installer.RunStage("configs/stage1.yaml")
// Edge cases: Creates state directory if missing; handles partial failures
func (i *Installer) RunStage(stageConfigPath string) error {
    // Implementation
}
```

### 3. Inline Comments

**Explain WHY, not just WHAT:**

✅ **GOOD:**
```go
// Condition check failed - skip task execution
if !p.checkCondition(ctx, task.Condition) {
    return TaskResult{Output: "Skipped (condition not met)"}
}
```

❌ **BAD:**
```go
// Check condition
if !p.checkCondition(ctx, task.Condition) {
    return TaskResult{Output: "Skipped (condition not met)"}
}
```

---

## 🧪 Linting & Testing (NON-NEGOTIABLE)

**Claude must ALWAYS ensure code quality before committing.**

### Pre-Commit Checklist (MANDATORY)

**Before ANY commit or PR:**

1. **Run all tests:**
   ```bash
   make test
   # OR
   go test ./...
   ```

2. **Run linter:**
   ```bash
   make lint
   # OR
   golangci-lint run
   ```

3. **Verify build:**
   ```bash
   make build
   make build-all  # Test both architectures
   ```

4. **Check test coverage:**
   ```bash
   go test -v -race -coverprofile=coverage.out ./...
   go tool cover -func=coverage.out
   ```

### ⚠️ NEVER Commit:
- Failing tests
- Lint violations
- Code without tests
- Breaking changes (without BREAKING CHANGE footer)
- Code that doesn't build

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
   - Proper error handling
   - Test coverage

3. **Write tests:**
   ```go
   // File: internal/installer/parallel_test.go
   func TestParallelExecutor_ExecuteSequential(t *testing.T) {
       ui := &mockUI{}
       executor := NewParallelExecutor(4, 30*time.Second, ui)

       tasks := []config.Task{
           {Name: "Task 1", Command: "echo test"},
       }

       err := executor.Execute(tasks)
       if err != nil {
           t.Fatalf("Expected no error, got: %v", err)
       }
   }
   ```

4. **Test changes:**
   ```bash
   # Run tests
   make test

   # Run linter
   make lint

   # Build binary
   make build

   # Test actual installation (optional)
   ./devsetup install --dry-run
   ```

5. **Commit with proper format:**
   ```bash
   git add .
   git commit -m "feat(installer): add new feature

   Adds new functionality with proper error handling
   and test coverage.

   Fixes #42"
   ```

6. **Create PR:**
   ```bash
   git push -u origin feat/42-add-feature-name
   # Then create PR via GitHub
   ```

---

## 🔨 Adding New Tools

### 1. Add to Brewfile

```ruby
# Brewfile
brew "new-tool"
```

### 2. Lock Homebrew versions

```bash
brew bundle install
brew bundle lock --lockfile=Brewfile.lock.json
```

### 3. Add to versions.lock

```toml
# versions.lock
[homebrew.formulas.new-tool]
version = "1.2.3"
tap = "homebrew/core"
```

### 4. Add installation task to stage config

```yaml
# configs/stage2.yaml
tasks:
  - name: "Install new-tool"
    command: "brew install new-tool"
    parallel_group: "homebrew-tools"
    required: false
    timeout: 120s
    condition: "! command -v new-tool"
```

### For Git Repositories

```toml
# versions.lock
[git_repos.repo-name]
url = "https://github.com/user/repo.git"
commit = "abc123def456"
path = "~/dev/repo-name"
shallow = true
stage = 2
```

```yaml
# Stage config
- name: "Clone repo-name"
  command: |
    if [ ! -d ~/dev/repo-name ]; then
      git clone --depth=1 https://github.com/user/repo.git ~/dev/repo-name
      cd ~/dev/repo-name && git checkout abc123def456
    fi
  parallel_group: "git-repos"
  required: false
  timeout: 180s
```

---

## 📚 Common Development Tasks

### Testing Specific Components

```bash
# Test config loader
go test ./internal/config -v

# Test parallel executor
go test ./internal/installer -run TestParallelExecutor -v

# Test updater
go test ./internal/updater -v

# Test with race detector
go test -race ./...
```

### Debugging

```bash
# Verbose test output
go test -v ./...

# Run with dry-run
./devsetup install --dry-run

# Run diagnostics
./devsetup doctor
```

### Modifying Stage Configurations

Stage configs are YAML files in `configs/`:

```yaml
name: "Stage Name"
timeout: 30m
parallel: true

tasks:
  - name: "Task Name"
    command: "shell command"
    parallel_group: "group-name"  # Tasks in same group run concurrently
    required: true                 # Fail stage if this fails
    timeout: 60s
    retry_count: 2
    condition: "test command"      # Skip if condition fails

post_stage:
  message: "Stage complete!"
```

### Updating Dependencies

```bash
# Update Go modules
go get -u ./...
go mod tidy

# Verify build still works
make build
make test
```

---

## 🚀 Release Process

### Creating a Release

1. **Ensure all tests pass:**
   ```bash
   make test
   make lint
   make build-all
   ```

2. **Create git tag:**
   ```bash
   git tag -a v0.5.0 -m "Release v0.5.0: Add feature X"
   git push origin v0.5.0
   ```

3. **GitHub Actions automatically:**
   - Runs CI tests
   - Builds binaries (darwin-arm64, darwin-amd64)
   - Generates SHA256 checksums
   - Creates GitHub release
   - Publishes release notes

4. **Users can update with:**
   ```bash
   devsetup update
   ```

---

## ⚠️ Important Constraints & Assumptions

### Platform Requirements
- **macOS ONLY**: Primary development platform
- **Go 1.21+**: Required for building
- **Homebrew**: Will be installed if missing
- **Network required**: For downloads and git clones

### Design Decisions
- **Single binary**: No runtime dependencies
- **User-level installation**: No sudo required
- **State in ~/.local/**: Never modifies system directories
- **Idempotent**: Safe to re-run all operations
- **Version locking**: Ensures consistency across machines

---

## ✅ Pre-Commit Validation Checklist

Before committing ANY code change, verify:

- [ ] Tests pass: `make test`
- [ ] Linter passes: `make lint`
- [ ] Build succeeds: `make build`
- [ ] File-level documentation present (ALL files)
- [ ] Function-level documentation present (ALL functions)
- [ ] Inline comments explain WHY
- [ ] Test coverage for new code (>80%)
- [ ] Commit message follows Conventional Commits
- [ ] Commit references issue (`Fixes #N`)
- [ ] No forbidden package managers (npm, yarn, pip, poetry)
- [ ] Only `pnpm` (Node.js), `uv` (Python), `go mod` (Go)
- [ ] README.md updated if user-facing changes
- [ ] CONTRIBUTING.md updated if workflow changes

---

## 🚫 Absolute Prohibitions

**NEVER:**
- Use npm, yarn, bun, pip, pipenv, poetry, conda
- Commit code without file-level documentation
- Commit code without function-level documentation
- Commit code without tests
- Suppress linting errors without reason
- Commit failing tests
- Skip running tests before commit
- Use commit messages without type prefix
- Make system-wide modifications (use ~/.local/)
- Ignore error returns
- Use panic() except in truly unrecoverable situations

---

## 📖 Reference Documentation

- **README.md**: User quickstart and feature overview
- **CONTRIBUTING.md**: Comprehensive developer guide
- **AGENTS.md**: Agent coordination and task planning
- **global_conf/claude.md**: Mandatory engineering standards
- **global_conf/git_best_practices.md**: Git workflow standards
- **global_conf/instructions.md**: Code quality requirements

---

## 🎓 Key Takeaways

1. **Environment consistency is paramount** - exact replication across machines
2. **Package managers are STRICTLY enforced** - only pnpm, uv, go mod
3. **Documentation is MANDATORY** - file, function, inline comments
4. **Testing is NON-NEGOTIABLE** - test everything, >80% coverage
5. **Three-stage progressive setup** - productive in 5 minutes
6. **Parallel execution** - maximize speed with goroutines
7. **Version locking** - Brewfile.lock.json + versions.lock
8. **Self-updating** - binary updates via GitHub releases
9. **Go best practices** - proper error handling, clean code
10. **CI/CD automated** - tests, builds, releases all automated
