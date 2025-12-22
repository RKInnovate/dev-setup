# dev-setup: Zero to Productive in 5 Minutes

**Transform developer onboarding from days to minutes.**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platform-macOS-lightgrey.svg)](https://www.apple.com/macos/)

## 🎯 Goal

Reduce developer environment setup time from **2-5 days** to **30 minutes total** with developers **productive in just 5 minutes**.

### Current Reality (Manual Setup)
```
Day 1: Install tools, configure shell                   (4-6 hours)
Day 2: Setup Node, Python, package managers              (3-4 hours)
Day 3: Clone repos, install deps, configure IDE          (4-6 hours)
Day 4: Git config, SSH keys, authenticate services       (2-3 hours)
Day 5: Troubleshoot "works on my machine" issues         (3-5 hours)
──────────────────────────────────────────────────────────────────
TOTAL: 2-5 DAYS of wasted developer time ❌
```

### With dev-setup (Automated)
```
Minute 0-5:   Critical tools → Developer can code       ✅ PRODUCTIVE
Minute 5-15:  Full dev stack (background)               ✅ COMPLETE
Minute 15-30: Polish & optional tools (background)      ✅ PROFESSIONAL
──────────────────────────────────────────────────────────────────
TOTAL: 30 MINUTES, productive after 5 ✓
```

## ✨ Features

- **⚡ Parallel Installation**: 8+ concurrent tasks, ~8x speedup
- **📦 Incremental Stages**: Stage 1 (5min critical) → Stage 2 (10min full) → Stage 3 (15min polish)
- **🔒 Version Locking**: `Brewfile.lock.json` + `versions.lock` = zero "works on my machine"
- **✅ Verification**: `devsetup verify` checks all tools match expected versions
- **🔄 Idempotent**: Safe to re-run, skips existing tools
- **🎨 Rich UI**: Progress bars, colors, clear status indicators
- **🚀 Background Execution**: Keep working while installation completes

## 🏗️ Architecture

### Three-Stage Progressive Setup

```
┌──────────────────────────────────────────────────────────────┐
│ STAGE 1: CRITICAL PATH (5 min) - BLOCKING                    │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│ ✓ Homebrew                                                   │
│ ✓ Git + SSH                                                  │
│ ✓ Node + pnpm (parallel)                                     │
│ ✓ Python + uv (parallel)                                     │
│ ✓ Zed/VS Code (parallel)                                     │
│ ✓ Shell config                                               │
│                                                              │
│ 👨‍💻 DEVELOPER CAN NOW CODE                                    │
└──────────────────────────────────────────────────────────────┘
                           ↓
┌──────────────────────────────────────────────────────────────┐
│ STAGE 2: FULL STACK (10 min) - BACKGROUND                    │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│ ⚡ All Homebrew packages (8 parallel)                         │
│ ⚡ Zsh plugins (6 parallel clones)                            │
│ ⚡ Flutter wrapper + git config                               │
│ ⚡ Starship prompt configuration                              │
│                                                              │
│ 👨‍💻 Developer continues working...                            │
└──────────────────────────────────────────────────────────────┘
                           ↓
┌──────────────────────────────────────────────────────────────┐
│ STAGE 3: POLISH (15 min) - BACKGROUND                        │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│ ⚡ Nerd Fonts                                                 │
│ ⚡ Optional AI CLIs (Claude, Codex, Gemini)                   │
│ ⚡ Docker, Kubernetes tools                                   │
│                                                              │
│ 🎉 Complete professional environment                         │
└──────────────────────────────────────────────────────────────┘
```

### Tech Stack

- **Language**: Go 1.21+ (single binary, fast, concurrent)
- **CLI Framework**: Cobra (professional CLI with subcommands)
- **Config Format**: YAML (stages) + TOML (versions.lock)
- **Package Manager**: Homebrew (with Brewfile for declarative deps)
- **Parallel Execution**: Goroutines + semaphore pattern (8 concurrent max)

## 📦 What Gets Installed

### Stage 1: Critical (5 min)
- **Homebrew** (if missing)
- **Git** with basic config
- **Node.js + pnpm** (parallel)
- **Python + uv** (parallel)
- **Zed editor** (parallel)
- Shell PATH configuration

### Stage 2: Full Stack (10 min, background)
- **Dev tools**: git-lfs, wget, jq, ripgrep, fd, bat, eza
- **Starship prompt** with Nerd Font symbols
- **Zsh plugins**: completions, syntax highlighting, autosuggestions, history search
- **Flutter wrapper**: `flutterw` command
- **Git config templates**: Review before applying

### Stage 3: Polish (15 min, background)
- **Hack Nerd Font**: For proper icon display
- **AI CLIs**: Claude Code, Codex, Gemini (best-effort)
- **Docker Desktop** (optional, large download)
- **Modern CLI tools**: fzf, tree, htop, tldr

## 🚀 Quick Start

### One-Line Install
```bash
curl -fsSL https://raw.githubusercontent.com/rkinnovate/dev-setup/main/bootstrap.sh | bash
```

### Or Clone and Build
```bash
git clone https://github.com/rkinnovate/dev-setup
cd dev-setup
make install
devsetup install
```

## 📖 Usage

### Commands

```bash
# Install complete environment (3 stages)
devsetup install

# Fast mode (Stage 1 only - 5 minutes)
devsetup install --fast

# Skip optional tools (Stages 1 & 2 only)
devsetup install --skip-optional

# Preview what would be installed
devsetup install --dry-run

# Verify environment matches versions.lock
devsetup verify

# Auto-fix any mismatches
devsetup verify --fix

# Run diagnostics
devsetup doctor

# Check installation status
devsetup status

# Update versions.lock with current versions
devsetup update --capture-versions

# Show version
devsetup --version
```

### After Installation

1. **Restart Terminal**: `source ~/.zshrc` or close/reopen terminal
2. **Verify**: `devsetup verify`
3. **Set Git Identity**:
   ```bash
   git config --global user.name "Your Name"
   git config --global user.email "your@email.com"
   ```
4. **Switch Terminal Font**: Set to "Hack Nerd Font" for proper icons
5. **Authenticate AI CLIs** (if installed):
   ```bash
   claude auth
   codex login
   gemini login
   ```

## 🔐 Environment Consistency

### Version Locking

All dependencies are version-locked for **exact reproducibility**:

- **Homebrew packages**: `Brewfile` + `Brewfile.lock.json`
- **Git repositories**: `versions.lock` (commit SHAs or tags)
- **Custom tools**: `versions.lock` (version pinned)

### Verification

```bash
# Check all tools match expected versions
devsetup verify

# Output:
✓ git@2.43.0
✓ node@20.10.0
✗ pnpm@8.11.0 (expected: 8.10.5)  # Mismatch!

Environment verification FAILED (1 error)
Run 'devsetup verify --fix' to repair
```

### Fixing Mismatches

```bash
devsetup verify --fix
# Automatically reinstalls mismatched tools to correct versions
```

## 📊 Performance

### Time Breakdown

| Stage | Time | Blocking | What Happens |
|-------|------|----------|--------------|
| **Stage 1** | 5 min | ✅ Yes | Install critical tools - developer can code after |
| **Stage 2** | 10 min | ❌ No | Full dev stack installs in background |
| **Stage 3** | 15 min | ❌ No | Polish/optional tools install in background |
| **Total** | **30 min** | **5 min** | **Productive in 5 minutes** ✓ |

### Speedup Techniques

1. **Parallel Execution**: 8 concurrent tasks (8x speedup)
   ```yaml
   # tasks with same parallel_group run simultaneously
   parallel_group: "homebrew-tools"
   ```

2. **Shallow Git Clones**: `--depth=1` (5x faster)
   ```bash
   git clone --depth=1 --single-branch <repo>
   ```

3. **Background Stages**: Developer works while Stages 2 & 3 complete
4. **Idempotency**: Skip already-installed tools (instant on re-run)

## 🛠️ Development

### Project Structure

```
dev-setup/
├── cmd/
│   └── devsetup/
│       └── main.go              # CLI entry point
├── internal/
│   ├── config/
│   │   ├── models.go            # Config structs
│   │   └── loader.go            # YAML/TOML parsing
│   ├── installer/
│   │   ├── parallel.go          # Parallel executor (core engine)
│   │   └── installer.go         # Stage orchestrator
│   ├── ui/
│   │   └── progress.go          # Progress bars, colors
│   └── verify/
│       └── checker.go           # Version verification
├── configs/
│   ├── stage1.yaml              # Critical path (5 min)
│   ├── stage2.yaml              # Full stack (10 min)
│   └── stage3.yaml              # Polish (15 min)
├── Brewfile                     # Homebrew packages
├── Brewfile.lock.json           # Homebrew version lock
├── versions.lock                # Non-Homebrew version lock
├── bootstrap.sh                 # One-line installer
├── Makefile                     # Build automation
└── go.mod                       # Go dependencies
```

### Build Commands

```bash
# Build for current architecture
make build

# Build for all architectures
make build-all

# Install to ~/.local/bin
make install

# Run tests
make test

# Run linter
make lint

# Clean build artifacts
make clean

# Build and run
make run

# Build and dry-run
make dry-run
```

### Adding New Tools

#### To Brewfile (Homebrew packages):
```ruby
# In Brewfile
brew "newtool"

# Then lock version:
brew bundle dump --force
git add Brewfile.lock.json
```

#### To Stage Config (Custom installs):
```yaml
# In configs/stage2.yaml
- name: "Install custom tool"
  command: |
    curl -fsSL https://example.com/install.sh | sh
  parallel_group: "custom-tools"
  required: false
  timeout: 60s
```

#### To versions.lock (Version tracking):
```toml
# In versions.lock
[tools.newtool]
version = "1.2.3"
installer = "https://example.com/install.sh"
verify_command = "newtool --version"
```

## 🔒 Package Manager Policy

**⚠️ STRICTLY ENFORCED ⚠️**

| Language | Required | Forbidden |
|----------|----------|-----------|
| **Python** | `uv` | pip, pipenv, poetry, conda |
| **Node.js** | `pnpm` | npm, yarn, bun |

Lockfiles:
- Python: `uv.lock`
- Node.js: `pnpm-lock.yaml`

## 📝 Configuration Files

### Stage Configuration (YAML)

```yaml
name: "Stage Name"
timeout: 300s
parallel: true

tasks:
  - name: "Task name"
    command: "shell command to run"
    parallel_group: "group-name"  # Empty = sequential
    required: true                 # Fail stage if fails
    timeout: 60s
    retry_count: 2
    condition: "command -v tool"   # Skip if condition fails
```

### Version Lock (TOML)

```toml
[metadata]
schema_version = "1.0"
platform = "darwin"
updated = 2025-12-22T00:00:00Z

[homebrew.formulas]
git = { version = "2.43.0", tap = "homebrew/core" }

[git_repos.flutter-wrapper]
url = "https://github.com/user/repo"
commit = "abc123"
path = "~/.local/share/dev-setup/flutter-wrapper"
stage = 2
```

## 🎓 Key Concepts

### Idempotency

All operations are safe to re-run:
```bash
devsetup install  # First run: installs everything
devsetup install  # Second run: skips existing, instant
```

### Progressive Enhancement

Developer productive immediately, environment completes in background:
```
[5 min]  Stage 1 completes → Developer clones repos, starts coding
[15 min] Stage 2 completes → Full dev stack ready
[30 min] Stage 3 completes → Professional polish complete
```

### Version Locking

Identical versions on all machines:
```bash
# Developer A (Monday):
git@2.43.0, node@20.10.0, pnpm@8.10.5

# Developer B (Friday):
git@2.43.0, node@20.10.0, pnpm@8.10.5  # Exact same!
```

## 🆘 Troubleshooting

### Installation Issues

```bash
# Check what's wrong
devsetup doctor

# Re-run specific stage
devsetup install  # Idempotent, safe to re-run

# Verbose output
export DEBUG=1
devsetup install
```

### Version Mismatches

```bash
# Identify mismatches
devsetup verify

# Fix automatically
devsetup verify --fix
```

### Common Issues

1. **"Homebrew not found"**: Ensure `/opt/homebrew/bin` or `/usr/local/bin` in PATH
2. **"Permission denied"**: Homebrew install requires sudo password
3. **"Network timeout"**: Check internet connection, retry with longer timeout
4. **"Version mismatch"**: Run `devsetup verify --fix`

## 📈 Success Metrics

### Before (Manual)
- ⏱️ Time to first commit: **2-5 days**
- 💰 Cost per developer: **$1,000-2,500**
- 🐛 "Works on my machine" issues: **Common**
- 😫 Developer frustration: **High**

### After (Automated)
- ⏱️ Time to first commit: **5 minutes** ✅
- 💰 Cost per developer: **$31** ✅
- 🐛 "Works on my machine" issues: **Zero** ✅
- 😊 Developer satisfaction: **High** ✅

### ROI
- **Savings**: $1,000-2,500 per developer
- **Team of 10 devs/year**: **$10K-25K saved**
- **Development cost**: ~$10K
- **Payback**: First 4-10 developers

## 🤝 Contributing

See [AGENTS.md](AGENTS.md) for contribution guidelines including:
- Coding standards
- Documentation requirements
- Commit message format (Conventional Commits)
- Testing expectations

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🔗 Links

- **GitHub**: https://github.com/rkinnovate/dev-setup
- **Issues**: https://github.com/rkinnovate/dev-setup/issues
- **Releases**: https://github.com/rkinnovate/dev-setup/releases

---

**Built with ❤️ by RK Innovate**

*Transforming developer onboarding from days to minutes.*
