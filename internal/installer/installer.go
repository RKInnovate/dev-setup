// File: internal/installer/installer.go
// Purpose: Main installer orchestrator that coordinates stage execution and installation flow
// Problem: Need high-level orchestration of multi-stage installation with proper error handling
// Role: Coordinates config loading, parallel execution, and user feedback for complete installation
// Usage: Create Installer instance, call RunStage() for each stage file
// Design choices: Uses composition (embeds ParallelExecutor); supports dry-run mode; tracks state
// Assumptions: Stage config files exist and are valid; system has required permissions

package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rkinnovate/dev-setup/internal/config"
)

// Installer orchestrates the complete installation process
// What: High-level installer that manages stage execution, state tracking, and error recovery
// Why: Provides clean API for running multi-stage installation with proper error handling
type Installer struct {
	ui       UI
	executor *ParallelExecutor
	dryRun   bool
	stateDir string
}

// InstallState tracks installation progress and state
// What: Persistent state for tracking what's been installed
// Why: Allows resuming failed installations and verification
type InstallState struct {
	Version        string
	LastStage      string
	CompletedTasks []string
	StartTime      time.Time
	LastUpdate     time.Time
}

// NewInstaller creates a new Installer instance
// What: Constructor for Installer with default configuration
// Why: Centralizes installer creation with sensible defaults (8 concurrent, 30min timeout)
// Params: ui - UI for user feedback, dryRun - if true, show what would be done without doing it
// Returns: Configured Installer instance
// Example: installer := NewInstaller(progressUI, false)
func NewInstaller(ui UI, dryRun bool) *Installer {
	// Default: 8 concurrent tasks, 30 minute timeout per stage
	executor := NewParallelExecutor(8, 30*time.Minute, ui)

	// State directory for tracking installation progress
	homeDir, _ := os.UserHomeDir()
	stateDir := filepath.Join(homeDir, ".local", "share", "dev-setup")

	return &Installer{
		ui:       ui,
		executor: executor,
		dryRun:   dryRun,
		stateDir: stateDir,
	}
}

// RunStage executes a single installation stage
// What: Loads stage config, executes tasks via parallel executor, updates state
// Why: Main entry point for stage execution with complete error handling
// Params: stageConfigPath - path to stage YAML file (e.g. "configs/stage1.yaml")
// Returns: Error if stage failed, nil if successful
// Example: err := installer.RunStage("configs/stage1.yaml")
// Edge cases: Creates state directory if missing; handles partial failures; updates state on success
func (i *Installer) RunStage(stageConfigPath string) error {
	// Load stage configuration
	i.ui.Info("Loading stage configuration: %s", stageConfigPath)
	stageCfg, err := config.LoadStageConfig(stageConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load stage config: %w", err)
	}

	i.ui.Info("Stage: %s (%d tasks)", stageCfg.Name, len(stageCfg.Tasks))

	// Dry run mode - show what would be done
	if i.dryRun {
		return i.dryRunStage(stageCfg)
	}

	// Ensure state directory exists
	if err := os.MkdirAll(i.stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Load previous state if exists
	state, err := i.loadState()
	if err != nil {
		i.ui.Warning("Could not load previous state: %v", err)
		state = &InstallState{
			StartTime: time.Now(),
		}
	}

	// Execute stage tasks
	stageStart := time.Now()
	if err := i.executor.Execute(stageCfg.Tasks); err != nil {
		// Save state even on failure for resume capability
		state.LastStage = stageConfigPath
		state.LastUpdate = time.Now()
		i.saveState(state)

		return fmt.Errorf("stage execution failed: %w", err)
	}

	// Stage completed successfully
	stageDuration := time.Since(stageStart)
	i.ui.Info("")
	i.ui.Info("⏱  Stage completed in %v", stageDuration.Round(time.Second))

	// Update state
	state.LastStage = stageConfigPath
	state.LastUpdate = time.Now()
	if err := i.saveState(state); err != nil {
		i.ui.Warning("Failed to save state: %v", err)
	}

	// Execute post-stage actions
	if stageCfg.PostStage.Message != "" {
		i.ui.Info("")
		i.ui.Info(stageCfg.PostStage.Message)
	}

	return nil
}

// dryRunStage shows what would be installed without actually installing
// What: Prints task list and commands that would be executed
// Why: Allows users to preview installation before committing
// Params: stageCfg - stage configuration to preview
// Returns: Always returns nil (dry run doesn't fail)
func (i *Installer) dryRunStage(stageCfg *config.StageConfig) error {
	i.ui.Info("")
	i.ui.Info("DRY RUN - Would execute %d tasks:", len(stageCfg.Tasks))
	i.ui.Info("")

	// Group tasks by parallel group
	groups := make(map[string][]config.Task)
	for _, task := range stageCfg.Tasks {
		groups[task.ParallelGroup] = append(groups[task.ParallelGroup], task)
	}

	// Show sequential tasks
	if seqTasks, ok := groups[""]; ok {
		i.ui.Info("Sequential tasks:")
		for _, task := range seqTasks {
			required := ""
			if task.Required {
				required = " (required)"
			}
			i.ui.Info("  • %s%s", task.Name, required)
			i.ui.Info("    $ %s", task.Command)
		}
		i.ui.Info("")
	}

	// Show parallel groups
	for groupName, tasks := range groups {
		if groupName == "" {
			continue // Already showed sequential
		}

		i.ui.Info("Parallel group '%s':", groupName)
		for _, task := range tasks {
			required := ""
			if task.Required {
				required = " (required)"
			}
			i.ui.Info("  • %s%s", task.Name, required)
			i.ui.Info("    $ %s", task.Command)
		}
		i.ui.Info("")
	}

	return nil
}

// loadState loads installation state from disk
// What: Reads InstallState from state file in ~/.local/share/dev-setup
// Why: Enables resume capability and verification of what's installed
// Returns: InstallState pointer and error if any
func (i *Installer) loadState() (*InstallState, error) {
	statePath := filepath.Join(i.stateDir, "state.json")

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// No state file yet, return empty state
			return &InstallState{
				StartTime: time.Now(),
			}, nil
		}
		return nil, err
	}

	// TODO: Unmarshal JSON state
	// For now, return empty state
	state := &InstallState{
		StartTime: time.Now(),
	}

	_ = data // Suppress unused warning

	return state, nil
}

// saveState saves installation state to disk
// What: Writes InstallState to state file for persistence
// Why: Tracks progress for resume and verification capabilities
// Params: state - current InstallState to save
// Returns: Error if save failed, nil if successful
func (i *Installer) saveState(state *InstallState) error {
	statePath := filepath.Join(i.stateDir, "state.json")

	// TODO: Marshal state to JSON
	// For now, just create empty file
	return os.WriteFile(statePath, []byte("{}"), 0644)
}

// Verify checks if installed tools match expected versions
// What: Compares installed versions against versions.lock
// Why: Ensures environment consistency across machines
// Returns: VerifyResult with list of mismatches
// Example: result := installer.Verify()
func (i *Installer) Verify() (*VerifyResult, error) {
	i.ui.Info("Loading versions.lock...")

	// Load versions lock
	versionsLock, err := config.LoadVersionsLock("versions.lock")
	if err != nil {
		return nil, fmt.Errorf("failed to load versions.lock: %w", err)
	}

	result := &VerifyResult{
		Checks: []VersionCheck{},
	}

	// Verify Homebrew formulas
	i.ui.Info("Checking Homebrew formulas...")
	for name, formula := range versionsLock.Homebrew.Formulas {
		check := i.verifyHomebrewFormula(name, formula.Version)
		result.Checks = append(result.Checks, check)

		if !check.Matches {
			result.Mismatches++
		}
	}

	// Verify Homebrew casks
	i.ui.Info("Checking Homebrew casks...")
	for name, cask := range versionsLock.Homebrew.Casks {
		check := i.verifyHomebrewCask(name, cask.Version)
		result.Checks = append(result.Checks, check)

		if !check.Matches {
			result.Mismatches++
		}
	}

	// Verify git repos
	i.ui.Info("Checking git repositories...")
	for name, repo := range versionsLock.GitRepos {
		check := i.verifyGitRepo(name, repo)
		result.Checks = append(result.Checks, check)

		if !check.Matches {
			result.Mismatches++
		}
	}

	return result, nil
}

// verifyHomebrewFormula checks if a formula matches expected version
// What: Runs `brew ls --versions <formula>` and compares with expected version
// Why: Core version verification for Homebrew formulas
// Params: name - formula name, expectedVersion - version from versions.lock
// Returns: VersionCheck with result
func (i *Installer) verifyHomebrewFormula(name, expectedVersion string) VersionCheck {
	// TODO: Implement actual version checking via brew command
	return VersionCheck{
		Name:            name,
		Type:            "homebrew-formula",
		ExpectedVersion: expectedVersion,
		ActualVersion:   expectedVersion, // Placeholder
		Matches:         true,            // Placeholder
	}
}

// verifyHomebrewCask checks if a cask matches expected version
// What: Runs `brew list --cask <cask>` and checks version
// Why: Core version verification for Homebrew casks
// Params: name - cask name, expectedVersion - version from versions.lock
// Returns: VersionCheck with result
func (i *Installer) verifyHomebrewCask(name, expectedVersion string) VersionCheck {
	// TODO: Implement actual version checking via brew command
	return VersionCheck{
		Name:            name,
		Type:            "homebrew-cask",
		ExpectedVersion: expectedVersion,
		ActualVersion:   expectedVersion, // Placeholder
		Matches:         true,            // Placeholder
	}
}

// verifyGitRepo checks if a git repo is at expected commit
// What: Runs `git -C <path> rev-parse HEAD` and compares with expected commit
// Why: Core version verification for git repositories
// Params: name - repo name, repo - repo config from versions.lock
// Returns: VersionCheck with result
func (i *Installer) verifyGitRepo(name string, repo config.GitRepoConfig) VersionCheck {
	// TODO: Implement actual git commit checking
	return VersionCheck{
		Name:            name,
		Type:            "git-repo",
		ExpectedVersion: repo.Commit,
		ActualVersion:   repo.Commit, // Placeholder
		Matches:         true,         // Placeholder
	}
}

// VerifyResult contains results of environment verification
// What: Aggregated results from checking all tools against versions.lock
// Why: Provides structured output for verification reporting
type VerifyResult struct {
	Checks     []VersionCheck
	Mismatches int
}

// VersionCheck represents a single version verification check
// What: Result of checking one tool's version
// Why: Detailed information for reporting and fixing mismatches
type VersionCheck struct {
	Name            string
	Type            string // homebrew-formula, homebrew-cask, git-repo, etc.
	ExpectedVersion string
	ActualVersion   string
	Matches         bool
	Error           error
}
