# Contributing to dev-setup

Thank you for your interest in contributing to dev-setup! This document provides guidelines and instructions for developers who want to contribute to the project.

## Table of Contents

- [Development Setup](#development-setup)
- [Architecture Overview](#architecture-overview)
- [Adding New Tools](#adding-new-tools)
- [Modifying Stage Configurations](#modifying-stage-configurations)
- [Testing Guidelines](#testing-guidelines)
- [Code Style](#code-style)
- [Commit Guidelines](#commit-guidelines)
- [Release Process](#release-process)
- [Troubleshooting](#troubleshooting)

## Development Setup

### Prerequisites

- Go 1.21 or higher
- macOS (darwin) - primary development platform
- Git
- Basic knowledge of YAML and TOML

### Getting Started

1. Clone the repository:
```bash
git clone https://github.com/rkinnovate/dev-setup.git
cd dev-setup
```

2. Install dependencies:
```bash
go mod download
```

3. Build the binary:
```bash
make build
```

4. Run tests:
```bash
make test
```

5. Run linter:
```bash
make lint
```

### Project Structure

```
dev-setup/
├── cmd/devsetup/          # CLI entry point
│   └── main.go           # Cobra commands and main logic
├── internal/             # Internal packages (not importable)
│   ├── config/          # Configuration loading and validation
│   │   ├── models.go   # Data models (structs)
│   │   └── loader.go   # YAML/TOML parsers
│   ├── installer/       # Installation orchestration
│   │   ├── installer.go    # High-level stage management
│   │   ├── parallel.go     # Parallel task execution
│   │   └── *_test.go      # Test files
│   ├── updater/         # Self-update functionality
│   │   ├── updater.go
│   │   └── updater_test.go
│   └── ui/              # User interface (progress bars, colors)
│       └── progress.go
├── configs/             # Stage configuration files
│   ├── stage1.yaml     # Critical path (5 min)
│   ├── stage2.yaml     # Full stack (10 min)
│   └── stage3.yaml     # Polish (15 min)
├── Brewfile            # Homebrew package declarations
├── versions.lock       # Version pinning (TOML)
├── bootstrap.sh        # One-line installer script
└── Makefile           # Build automation

```

## Architecture Overview

### Three-Stage Progressive Setup

dev-setup uses a three-stage approach to minimize time-to-productivity:

1. **Stage 1 (5 min, blocking)**: Critical tools - developer can code immediately
   - Git, Node, Python, Editor
   - Required tools for basic development

2. **Stage 2 (10 min, background)**: Full development stack
   - Flutter wrapper, Zsh plugins, Starship
   - Additional dev tools and utilities

3. **Stage 3 (15 min, background)**: Polish and optional tools
   - Fonts, AI CLIs, Docker, optional utilities
   - Nice-to-have tools

### Parallel Execution Engine

Tasks are executed concurrently to maximize performance:
- **Parallel Groups**: Tasks with the same `parallel_group` run concurrently
- **Sequential Tasks**: Tasks without `parallel_group` run one at a time
- **Concurrency Limit**: Maximum 8 concurrent tasks (configurable)
- **Semaphore Pattern**: Used to control concurrency

### Version Locking

Two files ensure environment consistency:
- **Brewfile.lock.json**: Locks Homebrew package versions
- **versions.lock**: TOML file locking git repo commits and tool versions

## Adding New Tools

### Adding a Homebrew Package

1. Add to `Brewfile`:
```ruby
brew "package-name"
```

2. Lock the version:
```bash
brew bundle install
brew bundle lock --lockfile=Brewfile.lock.json
```

3. Add to appropriate stage config (e.g., `configs/stage2.yaml`):
```yaml
tasks:
  - name: "Install package-name"
    command: "brew install package-name"
    parallel_group: "homebrew-group"
    required: false
    timeout: 120s
```

### Adding a Git Repository

1. Add to `versions.lock`:
```toml
[git_repos.repo-name]
url = "https://github.com/user/repo.git"
commit = "abc123def456"  # Specific commit SHA
path = "~/dev/repo-name"
shallow = true
stage = 2
```

2. Add installation task to stage config:
```yaml
tasks:
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

### Adding a Tool with Custom Installer

1. Add to `versions.lock`:
```toml
[tools.tool-name]
version = "1.2.3"
installer = "https://example.com/install.sh"

[tools.tool-name.env]
TOOL_VERSION = "1.2.3"
TOOL_INSTALL_DIR = "/usr/local/bin"
```

2. Add installation task:
```yaml
tasks:
  - name: "Install tool-name"
    command: |
      curl -fsSL https://example.com/install.sh | TOOL_VERSION=1.2.3 bash
    required: false
    timeout: 120s
    condition: "! command -v tool-name"
```

## Modifying Stage Configurations

### Stage Config Format

Stage configs are YAML files with the following structure:

```yaml
name: "Stage Name"
timeout: 30m        # Max time for entire stage
parallel: true      # Allow parallel execution

tasks:
  - name: "Task Name"
    command: "shell command to execute"
    parallel_group: "group-name"  # Tasks in same group run concurrently
    required: true                 # Fail stage if this fails
    timeout: 60s                   # Max time for this task
    retry_count: 2                 # Retry on failure
    condition: "test command"      # Run only if condition passes

post_stage:
  message: "Stage complete! Next steps..."
  next_stage: "configs/stage2.yaml"
  blocking: false
```

### Task Configuration Options

- **name**: Human-readable task name (required)
- **command**: Shell command to execute (required)
- **parallel_group**: Tasks with same group run in parallel (empty = sequential)
- **required**: If true, stage fails when task fails (default: false)
- **timeout**: Maximum duration for task (default: stage timeout)
- **retry_count**: Number of retries on failure (default: 0)
- **condition**: Shell command that must succeed for task to run (empty = always run)

### Best Practices

1. **Group related tasks**: Put similar operations in the same parallel group
   ```yaml
   # Good: All Homebrew installs in one group
   - name: "Install git"
     parallel_group: "homebrew"
   - name: "Install node"
     parallel_group: "homebrew"
   ```

2. **Use conditions to avoid redundant work**:
   ```yaml
   - name: "Install tool"
     command: "curl ... | bash"
     condition: "! command -v tool"  # Skip if already installed
   ```

3. **Set appropriate timeouts**: Heavy operations need longer timeouts
   ```yaml
   - name: "Install Docker"
     timeout: 5m  # Large download, needs more time
   ```

4. **Make non-critical tasks optional**:
   ```yaml
   - name: "Install nice-to-have"
     required: false  # Don't fail if this fails
   ```

## Testing Guidelines

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Run tests for specific package
go test ./internal/config -v

# Run specific test
go test ./internal/installer -run TestParallelExecutor_ExecuteParallel -v
```

### Writing Tests

All new functionality must include tests. Follow these guidelines:

1. **Test file naming**: `*_test.go` alongside source files
2. **Test function naming**: `Test<FunctionName>_<Scenario>`
3. **Table-driven tests**: Use for multiple input/output combinations
4. **Mock external dependencies**: Use mock implementations for UI, HTTP, etc.

Example test structure:

```go
func TestLoadStageConfig_Valid(t *testing.T) {
    // Setup: Create test data
    tmpFile := filepath.Join(t.TempDir(), "stage.yaml")
    content := `name: "Test Stage"
tasks:
  - name: "Task 1"
    command: "echo test"
`
    os.WriteFile(tmpFile, []byte(content), 0644)

    // Execute: Run function under test
    cfg, err := LoadStageConfig(tmpFile)

    // Assert: Verify expectations
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
    if cfg.Name != "Test Stage" {
        t.Errorf("Expected name 'Test Stage', got '%s'", cfg.Name)
    }
}
```

### Test Coverage Requirements

- Aim for >80% coverage on all packages
- Critical paths (installer, config loader) should have >90% coverage
- Test edge cases: empty inputs, invalid data, timeouts, failures

## Code Style

### Go Code Style

- Follow standard Go conventions (use `gofmt`, `golint`)
- Use meaningful variable names (no single-letter variables except in loops)
- Add comments for exported functions and types
- Keep functions small and focused (< 50 lines preferred)

### Documentation Comments

Every file, function, and type must have documentation comments:

```go
// LoadStageConfig loads a stage configuration file from disk
// What: Reads YAML file and parses into StageConfig struct
// Why: Stages are defined declaratively in YAML files
// Params: path - filesystem path to stage YAML file
// Returns: Parsed StageConfig struct and error if any
// Example: cfg, err := LoadStageConfig("configs/stage1.yaml")
func LoadStageConfig(path string) (*StageConfig, error) {
    // Implementation
}
```

### Error Handling

- Always check errors, never ignore them
- Wrap errors with context: `fmt.Errorf("failed to load config: %w", err)`
- Return early on errors (avoid deep nesting)

## Commit Guidelines

### Commit Message Format

Follow Conventional Commits specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Build process, tooling changes

**Examples:**

```
feat(installer): add retry logic for failed tasks

Implements exponential backoff retry for tasks that fail.
Configurable via retry_count field in task config.

Closes #42
```

```
fix(updater): handle redirect during GitHub API calls

GitHub API can return redirects, causing update check to fail.
Now properly follows redirects and extracts final URL.
```

### Commit Best Practices

- Keep commits small and focused on one change
- Write descriptive commit messages (explain "why", not just "what")
- Reference issue numbers in commits
- Test before committing

## Release Process

### Versioning

We use Semantic Versioning (SemVer):
- **Major**: Breaking changes (e.g., v1.0.0 → v2.0.0)
- **Minor**: New features, backward compatible (e.g., v1.0.0 → v1.1.0)
- **Patch**: Bug fixes, backward compatible (e.g., v1.0.0 → v1.0.1)

### Creating a Release

1. **Update version** in necessary files
2. **Run tests** to ensure everything passes:
   ```bash
   make test
   make lint
   make build-all
   ```

3. **Create git tag**:
   ```bash
   git tag -a v0.5.0 -m "Release v0.5.0: Add parallel execution"
   git push origin v0.5.0
   ```

4. **GitHub Actions automatically**:
   - Builds binaries for darwin-arm64 and darwin-amd64
   - Generates SHA256 checksums
   - Creates GitHub release with artifacts
   - Publishes release notes

5. **Verify release**: Check https://github.com/rkinnovate/dev-setup/releases

### What Gets Released

- Compiled binaries: `devsetup-darwin-arm64`, `devsetup-darwin-amd64`
- Checksum files: `*.sha256`
- Release notes (auto-generated from commits)

### Update Process for Users

Users can update with:
```bash
devsetup update
```

This command:
1. Checks GitHub releases for newer version
2. Downloads appropriate binary for architecture
3. Verifies checksum
4. Atomically replaces current binary
5. Creates backup of old version

## Troubleshooting

### Build Issues

**Problem**: `go build` fails with "package not found"
```bash
# Solution: Download dependencies
go mod download
go mod tidy
```

**Problem**: Tests fail with import errors
```bash
# Solution: Ensure you're in project root
cd /path/to/dev-setup
go mod download
```

### Test Issues

**Problem**: Tests hang or timeout
```bash
# Solution: Increase timeout or check for deadlocks
go test -timeout 5m ./...
```

**Problem**: Parallel tests fail intermittently
```bash
# Solution: Run with race detector
go test -race ./...
```

### Development Tips

1. **Use dry-run mode** to test changes without side effects:
   ```bash
   ./devsetup install --dry-run
   ```

2. **Test specific stages**:
   ```bash
   ./devsetup install --fast  # Stage 1 only
   ```

3. **Check logs** for debugging:
   ```bash
   ./devsetup doctor  # Run diagnostics
   ```

4. **Validate configs** before committing:
   ```bash
   # Validate YAML
   yq eval '.' configs/stage1.yaml

   # Validate TOML
   python3 -c "import toml; toml.load(open('versions.lock'))"
   ```

## Questions?

- Open an issue: https://github.com/rkinnovate/dev-setup/issues
- Read the docs: [README.md](README.md)
- Check existing issues and PRs for similar questions

Thank you for contributing to dev-setup! 🚀
