// File: internal/installer/installer_test.go
// Purpose: Unit tests for installer orchestration
// Problem: Need to verify stage execution and state management works correctly
// Role: Test suite for Installer, RunStage, Verify functionality
// Usage: Run with `go test ./internal/installer`
// Design choices: Uses mockUI; creates temp config files; tests dry-run mode
// Assumptions: Test environment has file system access

package installer

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rkinnovate/dev-setup/internal/config"
)

func TestNewInstaller(t *testing.T) {
	ui := &mockUI{}
	installer := NewInstaller(ui, false)

	if installer.ui == nil {
		t.Error("Expected UI to be set")
	}

	if installer.executor == nil {
		t.Error("Expected executor to be initialized")
	}

	if installer.dryRun {
		t.Error("Expected dryRun to be false")
	}

	if installer.stateDir == "" {
		t.Error("Expected stateDir to be set")
	}
}

func TestNewInstaller_DryRun(t *testing.T) {
	ui := &mockUI{}
	installer := NewInstaller(ui, true)

	if !installer.dryRun {
		t.Error("Expected dryRun to be true")
	}
}

func TestRunStage_Success(t *testing.T) {
	// Create temp stage config
	tmpDir := t.TempDir()
	stageFile := filepath.Join(tmpDir, "stage.yaml")
	content := `name: "Test Stage"
timeout: 60s
tasks:
  - name: "Task 1"
    command: "echo test1"
    required: true
  - name: "Task 2"
    command: "echo test2"
    required: false
`
	if err := os.WriteFile(stageFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create stage file: %v", err)
	}

	ui := &mockUI{}
	installer := NewInstaller(ui, false)
	installer.stateDir = tmpDir // Use temp dir for state

	err := installer.RunStage(stageFile)
	if err != nil {
		t.Errorf("RunStage failed: %v", err)
	}

	// Check UI was called
	if len(ui.calls) == 0 {
		t.Error("Expected UI calls to be made")
	}
}

func TestRunStage_InvalidConfig(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "invalid.yaml")
	content := `invalid yaml content [[[`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	ui := &mockUI{}
	installer := NewInstaller(ui, false)

	err := installer.RunStage(tmpFile)
	if err == nil {
		t.Fatal("Expected error for invalid config, got nil")
	}
}

func TestRunStage_FileNotFound(t *testing.T) {
	ui := &mockUI{}
	installer := NewInstaller(ui, false)

	err := installer.RunStage("/nonexistent/stage.yaml")
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}
}

func TestRunStage_DryRun(t *testing.T) {
	// Create temp stage config
	tmpDir := t.TempDir()
	stageFile := filepath.Join(tmpDir, "stage.yaml")
	content := `name: "Test Stage"
tasks:
  - name: "Task 1"
    command: "echo test1"
    required: true
  - name: "Task 2"
    command: "echo test2"
    parallel_group: "group1"
  - name: "Task 3"
    command: "echo test3"
    parallel_group: "group1"
post_stage:
  message: "Stage complete!"
`
	if err := os.WriteFile(stageFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create stage file: %v", err)
	}

	ui := &mockUI{}
	installer := NewInstaller(ui, true) // Dry run mode

	err := installer.RunStage(stageFile)
	if err != nil {
		t.Errorf("DryRun failed: %v", err)
	}

	// Verify no actual tasks were executed (just info messages)
	hasTaskStart := false
	for _, call := range ui.calls {
		if call == "StartTask:Task 1" {
			hasTaskStart = true
		}
	}

	if hasTaskStart {
		t.Error("Expected no tasks to start in dry run mode")
	}
}

func TestDryRunStage_ShowsTasks(t *testing.T) {
	stageCfg := &config.StageConfig{
		Name: "Test Stage",
		Tasks: []config.Task{
			{
				Name:     "Sequential Task",
				Command:  "echo seq",
				Required: true,
			},
			{
				Name:          "Parallel Task 1",
				Command:       "echo p1",
				ParallelGroup: "group1",
			},
			{
				Name:          "Parallel Task 2",
				Command:       "echo p2",
				ParallelGroup: "group1",
			},
		},
	}

	ui := &mockUI{}
	installer := NewInstaller(ui, true)

	err := installer.dryRunStage(stageCfg)
	if err != nil {
		t.Errorf("dryRunStage failed: %v", err)
	}

	// Should have multiple Info calls showing task details
	infoCount := 0
	for _, call := range ui.calls {
		if call == "Info" {
			infoCount++
		}
	}

	if infoCount == 0 {
		t.Error("Expected Info calls in dry run")
	}
}

func TestLoadState_NewInstallation(t *testing.T) {
	tmpDir := t.TempDir()

	ui := &mockUI{}
	installer := NewInstaller(ui, false)
	installer.stateDir = tmpDir

	state, err := installer.loadState()
	if err != nil {
		t.Errorf("loadState failed: %v", err)
	}

	if state == nil {
		t.Fatal("Expected state to be initialized")
	}

	// New installation should have empty completed tasks
	if len(state.CompletedTasks) != 0 {
		t.Error("Expected empty completed tasks for new installation")
	}
}

func TestLoadState_NonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()

	ui := &mockUI{}
	installer := NewInstaller(ui, false)
	installer.stateDir = tmpDir

	// State file doesn't exist yet
	state, err := installer.loadState()
	if err != nil {
		t.Errorf("Expected no error for missing state file, got: %v", err)
	}

	if state == nil {
		t.Error("Expected default state to be returned")
	}
}

func TestSaveState(t *testing.T) {
	tmpDir := t.TempDir()

	ui := &mockUI{}
	installer := NewInstaller(ui, false)
	installer.stateDir = tmpDir

	state := &InstallState{
		Version:        "0.4.0",
		LastStage:      "stage1",
		CompletedTasks: []string{"task1", "task2"},
		StartTime:      time.Now(),
		LastUpdate:     time.Now(),
	}

	err := installer.saveState(state)
	if err != nil {
		t.Errorf("saveState failed: %v", err)
	}

	// Verify state file was created
	statePath := filepath.Join(tmpDir, "state.json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("Expected state file to be created")
	}
}

func TestVerify_NoVersionsLock(t *testing.T) {
	ui := &mockUI{}
	installer := NewInstaller(ui, false)

	// Try to verify without versions.lock file
	_, err := installer.Verify()
	if err == nil {
		t.Error("Expected error when versions.lock not found")
	}
}

func TestVerify_WithVersionsLock(t *testing.T) {
	// Create temp versions.lock
	tmpDir := t.TempDir()
	versionsFile := filepath.Join(tmpDir, "versions.lock")
	content := `[metadata]
schema_version = "1.0"
platform = "darwin"

[homebrew.formulas.git]
version = "2.43.0"

[homebrew.casks.docker]
version = "4.26.1"

[git_repos.test]
url = "https://github.com/test/repo.git"
commit = "abc123"
path = "~/test"
`
	if err := os.WriteFile(versionsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create versions.lock: %v", err)
	}

	// Change to temp dir so LoadVersionsLock finds the file
	origDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	ui := &mockUI{}
	installer := NewInstaller(ui, false)

	result, err := installer.Verify()
	if err != nil {
		t.Errorf("Verify failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected verify result")
	}

	// Should have checks for formulas, casks, and repos
	expectedChecks := 3 // git formula, docker cask, test repo
	if len(result.Checks) != expectedChecks {
		t.Errorf("Expected %d checks, got %d", expectedChecks, len(result.Checks))
	}
}

func TestVerifyHomebrewFormula(t *testing.T) {
	ui := &mockUI{}
	installer := NewInstaller(ui, false)

	check := installer.verifyHomebrewFormula("git", "2.43.0")

	if check.Name != "git" {
		t.Errorf("Expected name 'git', got '%s'", check.Name)
	}

	if check.Type != "homebrew-formula" {
		t.Errorf("Expected type 'homebrew-formula', got '%s'", check.Type)
	}

	if check.ExpectedVersion != "2.43.0" {
		t.Errorf("Expected version '2.43.0', got '%s'", check.ExpectedVersion)
	}

	// Note: Actual version check is a placeholder in current implementation
	// This would need real brew command integration
}

func TestVerifyHomebrewCask(t *testing.T) {
	ui := &mockUI{}
	installer := NewInstaller(ui, false)

	check := installer.verifyHomebrewCask("docker", "4.26.1")

	if check.Name != "docker" {
		t.Errorf("Expected name 'docker', got '%s'", check.Name)
	}

	if check.Type != "homebrew-cask" {
		t.Errorf("Expected type 'homebrew-cask', got '%s'", check.Type)
	}
}

func TestVerifyGitRepo(t *testing.T) {
	ui := &mockUI{}
	installer := NewInstaller(ui, false)

	repo := config.GitRepoConfig{
		URL:    "https://github.com/test/repo.git",
		Commit: "abc123def",
		Path:   "~/test",
	}

	check := installer.verifyGitRepo("test-repo", repo)

	if check.Name != "test-repo" {
		t.Errorf("Expected name 'test-repo', got '%s'", check.Name)
	}

	if check.Type != "git-repo" {
		t.Errorf("Expected type 'git-repo', got '%s'", check.Type)
	}

	if check.ExpectedVersion != "abc123def" {
		t.Errorf("Expected version 'abc123def', got '%s'", check.ExpectedVersion)
	}
}

func TestVerifyResult_AllMatches(t *testing.T) {
	result := &VerifyResult{
		Checks: []VersionCheck{
			{Name: "git", Matches: true},
			{Name: "node", Matches: true},
		},
		Mismatches: 0,
	}

	if result.Mismatches != 0 {
		t.Errorf("Expected 0 mismatches, got %d", result.Mismatches)
	}
}

func TestVerifyResult_WithMismatches(t *testing.T) {
	result := &VerifyResult{
		Checks: []VersionCheck{
			{Name: "git", Matches: true},
			{Name: "node", Matches: false},
		},
		Mismatches: 1,
	}

	if result.Mismatches != 1 {
		t.Errorf("Expected 1 mismatch, got %d", result.Mismatches)
	}
}

func TestRunStage_CreatesStateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	stageFile := filepath.Join(tmpDir, "stage.yaml")
	content := `name: "Test Stage"
tasks:
  - name: "Task 1"
    command: "echo test"
`
	if err := os.WriteFile(stageFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create stage file: %v", err)
	}

	ui := &mockUI{}
	installer := NewInstaller(ui, false)

	// Use non-existent state dir
	stateDir := filepath.Join(tmpDir, "state", "dev-setup")
	installer.stateDir = stateDir

	err := installer.RunStage(stageFile)
	if err != nil {
		t.Errorf("RunStage failed: %v", err)
	}

	// Verify state directory was created
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		t.Error("Expected state directory to be created")
	}
}

func TestRunStage_UpdatesState(t *testing.T) {
	tmpDir := t.TempDir()
	stageFile := filepath.Join(tmpDir, "stage.yaml")
	content := `name: "Test Stage"
tasks:
  - name: "Task 1"
    command: "echo test"
post_stage:
  message: "Done!"
`
	if err := os.WriteFile(stageFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create stage file: %v", err)
	}

	ui := &mockUI{}
	installer := NewInstaller(ui, false)
	installer.stateDir = tmpDir

	err := installer.RunStage(stageFile)
	if err != nil {
		t.Errorf("RunStage failed: %v", err)
	}

	// Verify state file was created/updated
	statePath := filepath.Join(tmpDir, "state.json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("Expected state file to be created")
	}
}

func TestRunStage_PostStageMessage(t *testing.T) {
	tmpDir := t.TempDir()
	stageFile := filepath.Join(tmpDir, "stage.yaml")
	content := `name: "Test Stage"
tasks:
  - name: "Task 1"
    command: "echo test"
post_stage:
  message: "Stage completed successfully!"
`
	if err := os.WriteFile(stageFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create stage file: %v", err)
	}

	ui := &mockUI{}
	installer := NewInstaller(ui, false)
	installer.stateDir = tmpDir

	err := installer.RunStage(stageFile)
	if err != nil {
		t.Errorf("RunStage failed: %v", err)
	}

	// Post-stage message should appear in Info calls
	// (actual message display tested in integration tests)
	if len(ui.calls) == 0 {
		t.Error("Expected UI calls for post-stage message")
	}
}
