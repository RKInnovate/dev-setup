# dev-setup v2.0 Refactoring TODO

## ✅ Completed

- [x] Add git submodules structure
  - [x] external/claude-standard-env
  - [x] external/git-config
  - [x] external/flutter-wrapper
  - [x] third-party/zsh-syntax-highlighting
  - [x] third-party/zsh-autosuggestions

- [x] Create new config files
  - [x] configs/tools.yaml (declarative tool installation)
  - [x] configs/setup.yaml (post-install configuration)

- [x] Create config models
  - [x] internal/config/tools_config.go (with dependency resolution)
  - [x] internal/config/setup_config.go (with verification)
  - [x] internal/config/state.go (proper state tracking)

- [x] Research external tools
  - [x] Understand claude-standard-env behavior (interactive API key prompt)
  - [x] Understand git-config behavior (multiple modes, flags)

## 🚧 In Progress

- [ ] Implement install command
  - [ ] Create internal/installer/tool_installer.go
  - [ ] Implement idempotency checks (check before install)
  - [ ] Implement parallel execution for tools in same group
  - [ ] Integrate with state tracking
  - [ ] Handle dependencies correctly

## 📋 TODO

### Core Implementation

- [ ] Implement setup command
  - [ ] Create internal/setup/setup_executor.go
  - [ ] Implement remote-first/local-fallback strategy
  - [ ] Handle interactive prompts (claude-standard-env, git-config)
  - [ ] Implement file operations (.zshrc editing)
  - [ ] Implement TOML editing (starship.toml)
  - [ ] Add env var prompts (GEMINI_API_KEY)
  - [ ] Integrate with state tracking

- [ ] Implement verify command
  - [ ] Create internal/verify/verifier.go
  - [ ] Check tool existence (command -v)
  - [ ] Check tool versions (if specified)
  - [ ] Verify configuration (files exist, contain expected content)
  - [ ] Check environment variables
  - [ ] Validate TOML values

- [ ] Implement status command
  - [ ] Create internal/status/reporter.go
  - [ ] Show installed tools with versions
  - [ ] Show configured tasks
  - [ ] Calculate progress percentages
  - [ ] Pretty-print with colors and icons

### Main Command Refactoring

- [ ] Update cmd/devsetup/main.go
  - [ ] Remove old stages concept
  - [ ] Add new install command
  - [ ] Add new setup command
  - [ ] Update verify command
  - [ ] Update status command
  - [ ] Keep update command (already works)

### Testing & Quality

- [ ] Write tests
  - [ ] Test tools_config.go (dependency resolution)
  - [ ] Test setup_config.go (validation)
  - [ ] Test state.go (save/load)
  - [ ] Test tool_installer.go (idempotency)
  - [ ] Test setup_executor.go (fallback logic)

- [ ] Run linter
  - [ ] Fix any golangci-lint errors
  - [ ] Ensure all files have proper documentation

- [ ] Integration testing
  - [ ] Test install command end-to-end
  - [ ] Test setup command with real submodules
  - [ ] Test verify command accuracy
  - [ ] Test status command display

### Documentation

- [ ] Update CLAUDE.md
  - [ ] Remove stages architecture
  - [ ] Document new install/setup/verify/status commands
  - [ ] Update config file formats
  - [ ] Add submodule update instructions
  - [ ] Document remote-first/local-fallback strategy

- [ ] Update README.md
  - [ ] Update command examples
  - [ ] Update architecture diagram
  - [ ] Update quick start guide

- [ ] Update CONTRIBUTING.md
  - [ ] Document how to add new tools
  - [ ] Document how to add new setup tasks
  - [ ] Explain config file structure

### Cleanup

- [ ] Remove old files
  - [ ] configs/stage1.yaml
  - [ ] configs/stage2.yaml
  - [ ] configs/stage3.yaml
  - [ ] versions.lock (maybe keep for reference?)
  - [ ] Brewfile (maybe keep for reference?)

- [ ] Archive old code
  - [ ] Save old installer implementation for reference
  - [ ] Document migration path from v1 to v2

## 🎯 Success Criteria

- [ ] `devsetup install` installs all tools with proper idempotency
- [ ] `devsetup setup` configures all tools (interactive where needed)
- [ ] `devsetup verify` shows accurate status (no false positives)
- [ ] `devsetup status` shows correct progress percentages
- [ ] All tests pass
- [ ] All linters pass
- [ ] Documentation is up-to-date
- [ ] Submodules work correctly (both remote and local)

## 📝 Notes

### Architecture Changes
- Old: Stages (1/2/3) with blocking/background execution
- New: Install (idempotent, parallel) → Setup (interactive, configures) → Verify (accurate)

### Key Improvements
- Proper idempotency (check before install)
- Accurate state tracking (state.json)
- Better UX (clear phases, progress reporting)
- Version-controlled external tools (git submodules)
- Remote-first with local fallback (works offline after clone)

### Dependencies
- tools.yaml defines dependency order
- GetInstallOrder() uses topological sort
- Setup tasks also have depends_on

### Testing Strategy
1. Unit tests for config loading/validation
2. Unit tests for dependency resolution
3. Integration tests with mock commands
4. Manual testing with real tools
