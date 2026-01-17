# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## 🎯 Primary Goal: Identical Environment Across All Developer Machines

**Zero tolerance for "it works on my machine" excuses.**

This repository ensures **exact environment replication** across all developer machines through:
- Automated install → setup → verify workflow
- Idempotent tool installation (check before install, never reinstall)
- Interactive post-install configuration with remote-first/local-fallback
- Parallel task execution within groups for maximum speed
- Git submodules for version-locked external dependencies
- Strict package manager enforcement (pnpm for Node.js, uv for Python)
- Self-updating binary with GitHub releases
- Accurate verification without false positives

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

This is a **macOS development environment bootstrap tool** written in Go that reduces developer setup time from **days to 30 minutes**. It uses a clean install → setup → verify workflow with idempotent operations and parallel execution.

### Key Technology Stack

- **Language**: Go 1.21+
- **CLI Framework**: Cobra (professional CLI with subcommands)
- **Configuration**: YAML (tools.yaml + setup.yaml)
- **Dependencies**: Git submodules for external tools (claude-standard-env, git-config, flutter-wrapper, zsh plugins)
- **Deployment**: Single binary, auto-updates via GitHub releases
- **CI/CD**: GitHub Actions (tests, linting, releases)
- **Package Manager**: Homebrew (for macOS tools)

### Architecture: Install → Setup → Verify

**1. Install (`devsetup install`)**
- Installs all tools from `tools.yaml`
- Idempotent: checks if tool exists before installing (never reinstalls)
- Parallel execution within groups for speed
- Dependency resolution via topological sort
- State tracking in `~/.local/share/devsetup/state.json`

**2. Setup (`devsetup setup`)**
- Post-install configuration from `setup.yaml`
- Interactive prompts for API keys and configuration
- Remote-first with local fallback (prefers latest, works offline)
- File operations: edits .zshrc, starship.toml, etc.
- Runs setup scripts from git submodules

**3. Verify (`devsetup verify`)**
- Accurate verification without false positives
- Checks actual tool existence and versions
- Verifies configuration files have expected content
- No reliance on state alone - runs real checks

**4. Status (`devsetup status`)**
- Shows installed tools with versions
- Displays configured tasks
- Calculates accurate completion percentage
- Suggests next steps

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
# Complete installation workflow
devsetup install          # Install all tools from tools.yaml
devsetup setup            # Configure tools (interactive)
devsetup verify           # Verify installation and configuration
devsetup status           # Show current environment status

# Dry run mode (see what would happen without doing it)
devsetup install --dry-run
devsetup setup --dry-run

# Run diagnostics
devsetup doctor

# Update devsetup binary
devsetup update
devsetup update --check  # Check without installing

# Show version
devsetup --version

# Show help
devsetup --help
devsetup install --help
devsetup setup --help
```

### One-Line Install (for end users)

```bash
# Bootstrap script downloads devsetup binary to ~/.local/bin and runs install
curl -fsSL https://raw.githubusercontent.com/rkinnovate/dev-setup/main/bootstrap.sh | bash

# After bootstrap completes, the devsetup binary is installed permanently
# You can then run: devsetup setup, devsetup verify, devsetup status
```

---

## 🏗️ Architecture Deep Dive

### Project Structure

```
dev-setup/
├── cmd/devsetup/              # CLI entry point
│   └── main.go               # Cobra commands (install, setup, verify, status, update, doctor)
├── internal/                  # Internal packages (not importable)
│   ├── config/               # Configuration loading and validation
│   │   ├── tools_config.go  # Tools.yaml data structures + dependency resolution
│   │   ├── setup_config.go  # Setup.yaml data structures
│   │   ├── state.go         # State tracking (installed tools, configured tasks)
│   │   ├── loader.go        # YAML parsers (legacy)
│   │   └── *_test.go        # Config tests
│   ├── installer/            # Tool installation orchestration
│   │   ├── tool_installer.go # Idempotent installation with parallel execution
│   │   ├── installer.go     # Legacy stage-based installer (being phased out)
│   │   ├── parallel.go      # Parallel task execution engine
│   │   └── *_test.go
│   ├── setup/                # Post-install configuration
│   │   └── setup_executor.go # Remote-first/local-fallback setup execution
│   ├── verify/               # Verification engine
│   │   └── verifier.go      # Accurate verification without false positives
│   ├── status/               # Status reporting
│   │   └── reporter.go      # Installation and configuration status display
│   ├── updater/              # Self-update functionality
│   │   ├── updater.go       # GitHub releases integration
│   │   └── updater_test.go
│   └── ui/                   # Terminal UI
│       ├── interface.go     # UI interface definition
│       └── progress.go      # Progress UI implementation
├── configs/                   # Configuration files (embedded in binary)
│   ├── embed.go             # Go embed directive
│   ├── tools.yaml           # Tool installation declarations
│   └── setup.yaml           # Post-install setup tasks
├── external/                  # Git submodules for external dependencies
│   ├── claude-standard-env/ # Claude CLI + API key setup
│   ├── git-config/          # Git configuration
│   ├── flutter-wrapper/     # Flutter version management
│   ├── zsh-syntax-highlighting/
│   └── zsh-autosuggestions/
├── .github/workflows/         # CI/CD pipelines
│   ├── ci.yml               # Tests, linting, builds
│   └── release.yml          # Automated releases
├── bootstrap.sh               # One-line installer (installs to ~/.local/bin)
├── Makefile                   # Build automation
├── go.mod / go.sum            # Go dependencies
├── README.md                  # User documentation
├── CONTRIBUTING.md            # Developer guide
└── CLAUDE.md                  # This file
```

### Core Design Principles

1. **Idempotent Installation**
   - Every tool has a `check` command that verifies existence
   - Installation only happens if check fails
   - Safe to re-run install command multiple times
   - State tracking in `~/.local/share/devsetup/state.json`
   - Re-verification on each run (not just state check)

2. **Dependency Resolution**
   - Tools declare dependencies via `depends_on` field
   - Topological sort (Kahn's algorithm) determines install order
   - Circular dependency detection
   - Required vs optional tool handling

3. **Parallel Task Execution**
   - Tools in same `parallel_group` run concurrently
   - Different groups run sequentially (respects dependencies)
   - Goroutines + WaitGroup for coordination
   - Error aggregation (first error fails group if required)

4. **Remote-First with Local Fallback**
   - Setup tasks try remote scripts first (latest version)
   - Falls back to local git submodules if remote fails
   - Works offline with local copies
   - Best of both worlds: latest + reliability

5. **Git Submodules for External Dependencies**
   - Version-locked external tools (claude-standard-env, git-config, flutter-wrapper)
   - Proper version control for third-party scripts
   - Easy updates via `git submodule update --remote`
   - Single source of truth for dependency versions

6. **Accurate Verification**
   - Runs actual check commands (not just state comparison)
   - No false positives from stale state
   - Verifies configuration files contain expected content
   - Checks environment variables are set
   - TOML value validation (planned)

7. **Self-Updating**
   - Checks GitHub releases for new versions
   - Downloads and atomically replaces binary
   - Preserves backup of old version
   - Installed to ~/.local/bin permanently

### Key Packages Reference

| Package | Purpose |
|---------|---------|
| `cmd/devsetup` | CLI entry point with Cobra commands (install, setup, verify, status, update, doctor) |
| `internal/config` | Configuration loading (tools.yaml, setup.yaml) + state tracking |
| `internal/installer` | Tool installation with idempotency and parallel execution |
| `internal/setup` | Post-install configuration with remote-first/local-fallback |
| `internal/verify` | Accurate verification without false positives |
| `internal/status` | Status reporting with accurate progress |
| `internal/updater` | Self-update via GitHub releases |
| `internal/ui` | Terminal UI (progress bars, colors) with interface abstraction |

### Embedded Configuration System (Critical for Binary Distribution)

**Problem**: Downloaded binary needs config files but they're not on user's filesystem.

**Solution**: Configs are embedded directly into the binary at build time using Go's `embed` package.

**How It Works**:
```go
// configs/embed.go embeds all YAML files
//go:embed *.yaml
var ConfigFS embed.FS

// main.go sets the embedded FS
func init() {
    config.SetEmbeddedFS(configs.ConfigFS)
}

// loader.go tries filesystem first, falls back to embedded
func LoadStageConfig(path string) (*StageConfig, error) {
    // Try filesystem first (for development)
    if data, err := os.ReadFile(path); err == nil {
        return parseYAML(data)
    }

    // Fall back to embedded (for distributed binary)
    return loadFromEmbedded(path)
}
```

**Why This Matters**:
- Binary works standalone without external files
- Development uses filesystem files (faster iteration)
- Production uses embedded files (always consistent)
- Users don't need to download config files separately

**When Adding New Configs**:
- Add YAML file to `configs/` directory
- Embedding happens automatically via `//go:embed *.yaml`
- Both filesystem and embedded paths work identically
- No code changes needed in loader

### State & Path Locations

```
~/.local/
├── bin/                          # User binaries (on PATH)
│   ├── devsetup                 # Main binary (installed by bootstrap.sh or make install)
│   └── flutterw -> ...          # Symlink to Flutter wrapper (created by setup)
├── share/
│   └── dev-setup/                # State directory
│       └── state.json            # Installation and configuration state
│                                 # Format: { "installed": {...}, "configured": {...} }
~/.config/
└── starship.toml                 # Starship prompt config (edited by setup)

~/.zshrc                           # Shell config (edited by setup to add PATH, load plugins)

# External dependencies (git submodules in repo, not in ~/.local)
dev-setup/external/
├── claude-standard-env/          # Claude CLI + API key setup
├── git-config/                   # Git global configuration
├── flutter-wrapper/              # Flutter version manager (flutterw)
├── zsh-syntax-highlighting/      # Zsh plugin
└── zsh-autosuggestions/          # Zsh plugin
```

**Key Locations:**
- **Binary**: `~/.local/bin/devsetup` (added to PATH by bootstrap.sh)
- **State**: `~/.local/share/devsetup/state.json` (tracks installed tools + configured tasks)
- **Config**: Embedded in binary (configs/tools.yaml, configs/setup.yaml)
- **External tools**: Git submodules in `external/` directory (version-locked)

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

### Testing with Mock UI

The installer uses a `UI` interface for all user feedback, enabling comprehensive testing:

```go
// Real implementation
type UI interface {
    StartTask(name string)
    CompleteTask(name string)
    FailTask(name string, err error)
    Info(format string, args ...interface{})
    Warning(format string, args ...interface{})
    Error(format string, args ...interface{})
}

// Mock for testing
type mockUI struct {
    mu sync.Mutex
    tasks []string
    errors []error
}

func (m *mockUI) StartTask(name string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.tasks = append(m.tasks, name)
}
```

**Writing Tests**:
```go
func TestParallelExecutor_Execute(t *testing.T) {
    // Create mock UI
    ui := &mockUI{}
    executor := NewParallelExecutor(4, 30*time.Second, ui)

    tasks := []config.Task{
        {Name: "Task 1", Command: "echo test"},
    }

    // Execute tasks
    err := executor.Execute(tasks)

    // Verify UI interactions
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }

    if len(ui.tasks) != 1 {
        t.Errorf("Expected 1 task, got %d", len(ui.tasks))
    }
}
```

**Why This Pattern**:
- Decouples business logic from UI presentation
- Enables testing without terminal interaction
- Allows capturing and verifying UI calls
- Thread-safe mock implementation with mutex

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

### 1. Add to tools.yaml

```yaml
# configs/tools.yaml
tools:
  - name: new-tool
    check: command -v new-tool              # Idempotency check
    install:
      command: brew install new-tool
      parallel_group: homebrew-cli          # Run with other Homebrew tools
      timeout: 120s
    depends_on: [homebrew]                  # Dependencies (optional)
    required: false                         # Optional tool (won't fail install)
```

**Key Fields:**
- `check`: Command to verify tool exists (for idempotency)
- `install.command`: Installation command
- `install.parallel_group`: Group for parallel execution
- `depends_on`: List of dependencies (installed first)
- `required`: If true, failure stops installation

### 2. Add post-install configuration (if needed)

```yaml
# configs/setup.yaml
setup_tasks:
  - name: configure-new-tool
    zshrc_lines:
      - comment: "# Initialize new-tool"
        content: 'eval "$(new-tool init)"'
    verify:
      - command: "grep -q 'new-tool init' ~/.zshrc"
```

### Adding Git Submodules for External Tools

```bash
# Add new external dependency as submodule
git submodule add https://github.com/user/repo external/repo-name
git submodule update --init --recursive
```

```yaml
# configs/setup.yaml
- name: setup-external-tool
  strategy: remote_first
  remote:
    command: 'bash -c "$(curl -fsSL https://raw.githubusercontent.com/user/repo/main/install.sh)"'
    timeout: 180s
  local:
    command: bash ./external/repo-name/install.sh
    timeout: 180s
  interactive: true
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

# Check test coverage by package
go test -v -cover ./internal/config
go test -v -cover ./internal/installer
go test -v -cover ./internal/updater
```

### Debugging & Development

```bash
# Dry run mode (see what would happen)
./devsetup install --dry-run
./devsetup setup --dry-run

# Run with embedded configs (test binary distribution)
./devsetup install

# Run with filesystem configs (test config changes)
# Automatically uses filesystem if files exist in configs/

# Test individual commands
go run ./cmd/devsetup install
go run ./cmd/devsetup status
go run ./cmd/devsetup verify

# Inspect embedded files
go list -f '{{.EmbedFiles}}' ./configs

# Check state file
cat ~/.local/share/devsetup/state.json | jq .

# Verbose test output
go test -v ./...

# Run diagnostics
./devsetup doctor

# Test binary works as standalone
rm -rf configs/  # Temporarily remove configs
./devsetup --version  # Should still work (uses embedded)
git restore configs/  # Restore configs
```

### Modifying Configuration Files

**tools.yaml** (Installation configuration):

```yaml
tools:
  - name: "tool-name"
    check: "command -v tool-name"        # Idempotency check
    install:
      command: "brew install tool-name"  # Installation command
      parallel_group: "homebrew-cli"     # Parallel execution group
      timeout: 120s                      # Timeout (optional)
    depends_on: [homebrew]               # Dependencies (optional)
    required: false                      # Required vs optional
```

**setup.yaml** (Post-install configuration):

```yaml
setup_tasks:
  - name: "task-name"
    strategy: "remote_first"             # remote_first, local_only, or omit

    # Remote-first strategy
    remote:
      command: 'bash -c "$(curl ...)"'
      timeout: 180s
    local:
      command: "bash ./external/tool/install.sh"
      timeout: 180s

    # OR: .zshrc editing
    zshrc_lines:
      - comment: "# Comment"
        content: "export FOO=bar"

    # OR: Interactive prompt
    prompt:
      message: "Enter API key:"
      env_var: "MY_API_KEY"
      add_to: "$HOME/.zshrc"
      format: 'export MY_API_KEY="{value}"'
      skip_if_set: true

    # Verification
    verify:
      - command: "test -f ~/.config/tool.conf"
      - env_var: "MY_API_KEY"
      - file_exists: "$HOME/.config/tool.conf"
      - file_contains:
          path: "$HOME/.zshrc"
          text: "export FOO"

    optional: true                       # Optional task
    interactive: true                    # Requires user interaction
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
- **Embedded configs**: Binary works standalone without external files

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
5. **Install → Setup → Verify workflow** - clean separation of concerns
6. **Idempotent installation** - check before install, never reinstall
7. **Dependency resolution** - topological sort ensures correct order
8. **Parallel execution** - maximize speed with goroutines within groups
9. **Remote-first with local fallback** - latest version with offline capability
10. **Git submodules** - version-locked external dependencies
11. **Accurate verification** - no false positives from stale state
12. **Binary in ~/.local/bin** - user-level installation, on PATH permanently
13. **Self-updating** - binary updates via GitHub releases
14. **Embedded configs** - binary works standalone (tools.yaml, setup.yaml embedded)
15. **Go best practices** - proper error handling, clean code, interface-based testing
16. **CI/CD automated** - tests, builds, releases all automated
