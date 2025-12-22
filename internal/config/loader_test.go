// File: internal/config/loader_test.go
// Purpose: Unit tests for configuration loader
// Problem: Need to verify config parsing and validation works correctly
// Role: Test suite for LoadStageConfig, LoadVersionsLock, LoadBrewfile
// Usage: Run with `go test ./internal/config`
// Design choices: Uses temp files for testing; tests both valid and invalid configs
// Assumptions: Test environment has file system access

package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadStageConfig_Valid(t *testing.T) {
	// Create temp YAML file with valid stage config
	tmpFile := filepath.Join(t.TempDir(), "stage.yaml")
	content := `name: "Test Stage"
timeout: 300s
parallel: true
tasks:
  - name: "Task 1"
    command: "echo test"
    required: true
    timeout: 30s
  - name: "Task 2"
    command: "echo test2"
    parallel_group: "group1"
post_stage:
  message: "Stage complete"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Load config
	cfg, err := LoadStageConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadStageConfig failed: %v", err)
	}

	// Verify parsed values
	if cfg.Name != "Test Stage" {
		t.Errorf("Expected name 'Test Stage', got '%s'", cfg.Name)
	}

	if cfg.Timeout != 300*time.Second {
		t.Errorf("Expected timeout 300s, got %v", cfg.Timeout)
	}

	if !cfg.Parallel {
		t.Error("Expected parallel=true")
	}

	if len(cfg.Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(cfg.Tasks))
	}

	if cfg.Tasks[0].Name != "Task 1" {
		t.Errorf("Expected task name 'Task 1', got '%s'", cfg.Tasks[0].Name)
	}

	if !cfg.Tasks[0].Required {
		t.Error("Expected first task to be required")
	}

	if cfg.Tasks[1].ParallelGroup != "group1" {
		t.Errorf("Expected parallel_group 'group1', got '%s'", cfg.Tasks[1].ParallelGroup)
	}

	if cfg.PostStage.Message != "Stage complete" {
		t.Errorf("Expected post_stage message 'Stage complete', got '%s'", cfg.PostStage.Message)
	}
}

func TestLoadStageConfig_MissingName(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "stage.yaml")
	content := `tasks:
  - name: "Task 1"
    command: "echo test"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err := LoadStageConfig(tmpFile)
	if err == nil {
		t.Fatal("Expected error for missing stage name, got nil")
	}
}

func TestLoadStageConfig_NoTasks(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "stage.yaml")
	content := `name: "Empty Stage"
tasks: []
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err := LoadStageConfig(tmpFile)
	if err == nil {
		t.Fatal("Expected error for empty tasks, got nil")
	}
}

func TestLoadStageConfig_InvalidTaskName(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "stage.yaml")
	content := `name: "Test Stage"
tasks:
  - command: "echo test"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err := LoadStageConfig(tmpFile)
	if err == nil {
		t.Fatal("Expected error for missing task name, got nil")
	}
}

func TestLoadStageConfig_InvalidTaskCommand(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "stage.yaml")
	content := `name: "Test Stage"
tasks:
  - name: "Task 1"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err := LoadStageConfig(tmpFile)
	if err == nil {
		t.Fatal("Expected error for missing task command, got nil")
	}
}

func TestLoadStageConfig_FileNotFound(t *testing.T) {
	_, err := LoadStageConfig("/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}
}

func TestLoadVersionsLock_Valid(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "versions.lock")
	content := `[metadata]
schema_version = "1.0"
platform = "darwin"
min_macos_version = "13.0"
updated = 2024-01-15T10:30:00Z

[homebrew.formulas.git]
version = "2.43.0"
tap = "homebrew/core"

[homebrew.casks.docker]
version = "4.26.1"
tap = "homebrew/cask"

[tools.uv]
version = "0.1.9"
installer = "https://astral.sh/uv/install.sh"

[git_repos.flutter-wrapper]
url = "https://github.com/rkinnovate/flutter-wrapper.git"
commit = "abc123def456"
path = "~/dev/flutter-wrapper"
shallow = true
stage = 2
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Load config
	lock, err := LoadVersionsLock(tmpFile)
	if err != nil {
		t.Fatalf("LoadVersionsLock failed: %v", err)
	}

	// Verify metadata
	if lock.Metadata.SchemaVersion != "1.0" {
		t.Errorf("Expected schema_version '1.0', got '%s'", lock.Metadata.SchemaVersion)
	}

	if lock.Metadata.Platform != "darwin" {
		t.Errorf("Expected platform 'darwin', got '%s'", lock.Metadata.Platform)
	}

	// Verify Homebrew formula
	if git, ok := lock.Homebrew.Formulas["git"]; !ok {
		t.Error("Expected git formula not found")
	} else {
		if git.Version != "2.43.0" {
			t.Errorf("Expected git version '2.43.0', got '%s'", git.Version)
		}
		if git.Tap != "homebrew/core" {
			t.Errorf("Expected git tap 'homebrew/core', got '%s'", git.Tap)
		}
	}

	// Verify Homebrew cask
	if docker, ok := lock.Homebrew.Casks["docker"]; !ok {
		t.Error("Expected docker cask not found")
	} else {
		if docker.Version != "4.26.1" {
			t.Errorf("Expected docker version '4.26.1', got '%s'", docker.Version)
		}
	}

	// Verify tool
	if uv, ok := lock.Tools["uv"]; !ok {
		t.Error("Expected uv tool not found")
	} else {
		if uv.Version != "0.1.9" {
			t.Errorf("Expected uv version '0.1.9', got '%s'", uv.Version)
		}
	}

	// Verify git repo
	if fw, ok := lock.GitRepos["flutter-wrapper"]; !ok {
		t.Error("Expected flutter-wrapper repo not found")
	} else {
		if fw.Commit != "abc123def456" {
			t.Errorf("Expected commit 'abc123def456', got '%s'", fw.Commit)
		}
		if fw.Stage != 2 {
			t.Errorf("Expected stage 2, got %d", fw.Stage)
		}
	}
}

func TestLoadVersionsLock_MissingMetadata(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "versions.lock")
	content := `[homebrew.formulas.git]
version = "2.43.0"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err := LoadVersionsLock(tmpFile)
	if err == nil {
		t.Fatal("Expected error for missing metadata, got nil")
	}
}

func TestLoadVersionsLock_InvalidPlatform(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "versions.lock")
	content := `[metadata]
schema_version = "1.0"
platform = "windows"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err := LoadVersionsLock(tmpFile)
	if err == nil {
		t.Fatal("Expected error for invalid platform, got nil")
	}
}

func TestLoadVersionsLock_FormulaWithoutVersion(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "versions.lock")
	content := `[metadata]
schema_version = "1.0"
platform = "darwin"

[homebrew.formulas.git]
tap = "homebrew/core"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err := LoadVersionsLock(tmpFile)
	if err == nil {
		t.Fatal("Expected error for formula without version, got nil")
	}
}

func TestLoadVersionsLock_GitRepoWithoutURL(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "versions.lock")
	content := `[metadata]
schema_version = "1.0"
platform = "darwin"

[git_repos.test]
commit = "abc123"
path = "~/test"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err := LoadVersionsLock(tmpFile)
	if err == nil {
		t.Fatal("Expected error for git repo without URL, got nil")
	}
}

func TestLoadVersionsLock_GitRepoWithoutVersion(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "versions.lock")
	content := `[metadata]
schema_version = "1.0"
platform = "darwin"

[git_repos.test]
url = "https://github.com/test/repo.git"
path = "~/test"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err := LoadVersionsLock(tmpFile)
	if err == nil {
		t.Fatal("Expected error for git repo without commit/tag/branch, got nil")
	}
}

func TestLoadBrewfile_Valid(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "Brewfile")
	content := `# Test Brewfile
tap "homebrew/core"
tap "homebrew/cask"

brew "git"
brew "node", args: ["--HEAD"]
brew "python@3.11"

cask "docker"
cask "visual-studio-code"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Load Brewfile
	brewfile, err := LoadBrewfile(tmpFile)
	if err != nil {
		t.Fatalf("LoadBrewfile failed: %v", err)
	}

	// Verify taps
	if len(brewfile.Taps) != 2 {
		t.Errorf("Expected 2 taps, got %d", len(brewfile.Taps))
	}
	if brewfile.Taps[0] != "homebrew/core" {
		t.Errorf("Expected first tap 'homebrew/core', got '%s'", brewfile.Taps[0])
	}

	// Verify brews
	if len(brewfile.Brews) != 3 {
		t.Errorf("Expected 3 brews, got %d", len(brewfile.Brews))
	}
	if brewfile.Brews[0].Name != "git" {
		t.Errorf("Expected first brew 'git', got '%s'", brewfile.Brews[0].Name)
	}
	if brewfile.Brews[1].Name != "node" {
		t.Errorf("Expected second brew 'node', got '%s'", brewfile.Brews[1].Name)
	}
	if len(brewfile.Brews[1].Args) != 1 || brewfile.Brews[1].Args[0] != "--HEAD" {
		t.Errorf("Expected node args [--HEAD], got %v", brewfile.Brews[1].Args)
	}

	// Verify casks
	if len(brewfile.Casks) != 2 {
		t.Errorf("Expected 2 casks, got %d", len(brewfile.Casks))
	}
	if brewfile.Casks[0].Name != "docker" {
		t.Errorf("Expected first cask 'docker', got '%s'", brewfile.Casks[0].Name)
	}
}

func TestLoadBrewfile_EmptyFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "Brewfile")
	if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	brewfile, err := LoadBrewfile(tmpFile)
	if err != nil {
		t.Fatalf("LoadBrewfile failed: %v", err)
	}

	if len(brewfile.Taps) != 0 || len(brewfile.Brews) != 0 || len(brewfile.Casks) != 0 {
		t.Error("Expected empty brewfile")
	}
}

func TestLoadBrewfile_FileNotFound(t *testing.T) {
	_, err := LoadBrewfile("/nonexistent/Brewfile")
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}
}

func TestExtractQuotedString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`brew "git"`, "git"},
		{`tap "homebrew/core"`, "homebrew/core"},
		{`cask "docker"`, "docker"},
		{`brew "python@3.11"`, "python@3.11"},
		{`no quotes here`, ""},
		{`"unclosed`, ""},
	}

	for _, tt := range tests {
		result := extractQuotedString(tt.input)
		if result != tt.expected {
			t.Errorf("extractQuotedString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestExtractArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{`brew "git", args: ["--HEAD"]`, []string{"--HEAD"}},
		{`brew "node", args: ["--with-npm", "--HEAD"]`, []string{"--with-npm", "--HEAD"}},
		{`brew "python"`, []string{}},
		{`brew "git", args: []`, []string{}},
	}

	for _, tt := range tests {
		result := extractArgs(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("extractArgs(%q) returned %d args, want %d", tt.input, len(result), len(tt.expected))
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("extractArgs(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
			}
		}
	}
}
