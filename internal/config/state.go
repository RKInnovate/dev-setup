// File: internal/config/state.go
// Purpose: State tracking for installed tools and completed setup tasks
// Problem: Need persistent state to know what's installed and configured
// Role: Manages state.json file with tool/task status
// Usage: Read/write state during install/setup/verify commands
// Design choices: JSON format for readability; separate installed vs configured tracking
// Assumptions: State stored in ~/.local/share/devsetup/state.json

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// State represents the complete installation and configuration state
// What: Tracks which tools are installed and which tasks are configured
// Why: Need persistent state for verify/status commands and idempotency
type State struct {
	// Installed maps tool name to installation details
	Installed map[string]ToolState `json:"installed"`

	// Configured maps setup task name to completion status
	Configured map[string]bool `json:"configured"`

	// LastInstall timestamp
	LastInstall time.Time `json:"last_install"`

	// LastSetup timestamp
	LastSetup time.Time `json:"last_setup"`

	// Version of devsetup that created this state
	Version string `json:"version"`
}

// ToolState represents state of an installed tool
// What: Version, path, and installation time of a tool
// Why: Track what's installed for verification and status reporting
type ToolState struct {
	// Version of the installed tool
	Version string `json:"version"`

	// Path to the tool executable
	Path string `json:"path"`

	// InstalledAt timestamp
	InstalledAt time.Time `json:"installed_at"`
}

// GetStateDir returns the directory for state storage
// What: Returns ~/.local/share/devsetup path
// Why: Centralized location for state file
// Returns: Absolute path to state directory
func GetStateDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to /tmp if can't get home dir
		return "/tmp/devsetup"
	}
	return filepath.Join(home, ".local", "share", "devsetup")
}

// GetStatePath returns the full path to state.json
// What: Returns full path to state file
// Why: Single source of truth for state file location
// Returns: Absolute path to state.json
func GetStatePath() string {
	return filepath.Join(GetStateDir(), "state.json")
}

// LoadState loads state from state.json
// What: Reads and parses state.json file
// Why: Need to load existing state to check what's installed
// Returns: State object and error if any
// Example: state, err := LoadState()
// Edge cases: Returns empty state if file doesn't exist (first run)
func LoadState() (*State, error) {
	statePath := GetStatePath()

	// If state file doesn't exist, return empty state (first run)
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return &State{
			Installed:  make(map[string]ToolState),
			Configured: make(map[string]bool),
		}, nil
	}

	// Read state file
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Parse JSON
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	// Ensure maps are initialized
	if state.Installed == nil {
		state.Installed = make(map[string]ToolState)
	}
	if state.Configured == nil {
		state.Configured = make(map[string]bool)
	}

	return &state, nil
}

// SaveState writes state to state.json
// What: Serializes state to JSON and writes to disk
// Why: Persist state changes after install/setup operations
// Params: state - State object to save
// Returns: Error if save fails, nil if successful
// Example: err := SaveState(state)
// Edge cases: Creates state directory if it doesn't exist
func SaveState(state *State) error {
	stateDir := GetStateDir()

	// Ensure state directory exists
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Serialize to JSON (pretty-printed for readability)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize state: %w", err)
	}

	// Write to file
	statePath := GetStatePath()
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// MarkToolInstalled adds or updates a tool in the state
// What: Records that a tool was installed with version and path
// Why: Track installation for status reporting and verification
// Params: state - State to update, name - tool name, version - version string, path - path to executable
// Example: MarkToolInstalled(state, "git", "2.43.0", "/usr/bin/git")
func MarkToolInstalled(state *State, name, version, path string) {
	if state.Installed == nil {
		state.Installed = make(map[string]ToolState)
	}

	state.Installed[name] = ToolState{
		Version:     version,
		Path:        path,
		InstalledAt: time.Now(),
	}
	state.LastInstall = time.Now()
}

// MarkTaskConfigured marks a setup task as completed
// What: Records that a setup task was successfully completed
// Why: Track configuration for status reporting and skip on re-run
// Params: state - State to update, name - task name
// Example: MarkTaskConfigured(state, "claude-standard-env")
func MarkTaskConfigured(state *State, name string) {
	if state.Configured == nil {
		state.Configured = make(map[string]bool)
	}

	state.Configured[name] = true
	state.LastSetup = time.Now()
}

// IsToolInstalled checks if a tool is in the state
// What: Checks if tool name exists in installed map
// Why: Quick check for tool installation status
// Params: state - State to check, name - tool name
// Returns: True if tool is installed, false otherwise
func IsToolInstalled(state *State, name string) bool {
	if state.Installed == nil {
		return false
	}
	_, exists := state.Installed[name]
	return exists
}

// IsTaskConfigured checks if a setup task is completed
// What: Checks if task name exists in configured map
// Why: Quick check for task completion status
// Params: state - State to check, name - task name
// Returns: True if task is configured, false otherwise
func IsTaskConfigured(state *State, name string) bool {
	if state.Configured == nil {
		return false
	}
	return state.Configured[name]
}

// GetInstallProgress calculates installation progress
// What: Computes percentage of tools installed
// Why: For progress reporting in status command
// Params: state - Current state, totalTools - Total number of tools
// Returns: Percentage (0-100) as integer
func GetInstallProgress(state *State, totalTools int) int {
	if totalTools == 0 {
		return 100
	}
	return (len(state.Installed) * 100) / totalTools
}

// GetSetupProgress calculates setup progress
// What: Computes percentage of tasks configured
// Why: For progress reporting in status command
// Params: state - Current state, totalTasks - Total number of setup tasks
// Returns: Percentage (0-100) as integer
func GetSetupProgress(state *State, totalTasks int) int {
	if totalTasks == 0 {
		return 100
	}
	configured := 0
	for _, completed := range state.Configured {
		if completed {
			configured++
		}
	}
	return (configured * 100) / totalTasks
}
