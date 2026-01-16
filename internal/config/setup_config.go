// File: internal/config/setup_config.go
// Purpose: Data models for setup.yaml configuration
// Problem: Need structured representation of post-install configuration tasks
// Role: Provides Go structs for setup configuration with verification and prompts
// Usage: Loaded by setup command to configure installed tools
// Design choices: Supports interactive prompts, file edits, env vars, TOML edits
// Assumptions: Tools already installed; user available for interactive prompts

package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// SetupConfig represents the complete setup.yaml file
// What: List of configuration tasks to run after tools are installed
// Why: Tools need configuration (API keys, dotfiles, etc) after installation
type SetupConfig struct {
	// SetupTasks are the list of configuration tasks
	SetupTasks []SetupTask `yaml:"setup_tasks"`
}

// SetupTask represents a single configuration task
// What: Individual setup operation with verification and optional interactivity
// Why: Each configuration step needs execution strategy and validation
type SetupTask struct {
	// Name is the unique identifier for this task
	Name string `yaml:"name"`

	// Description is human-readable description
	Description string `yaml:"description"`

	// Strategy determines how to run this task
	Strategy string `yaml:"strategy"` // remote_first, local_only

	// Remote is the remote execution command (for remote_first)
	Remote *CommandConfig `yaml:"remote"`

	// Local is the local execution command (fallback or local_only)
	Local *CommandConfig `yaml:"local"`

	// Install contains commands for local_only strategy
	Install []string `yaml:"install"`

	// Steps for complex multi-step configurations
	Steps []SetupStep `yaml:"steps"`

	// ZshrcLines for adding lines to .zshrc
	ZshrcLines []ZshrcLine `yaml:"zshrc_lines"`

	// Prompt for interactive user input
	Prompt *PromptConfig `yaml:"prompt"`

	// Verify contains verification checks
	Verify []VerifyCheck `yaml:"verify"`

	// DependsOn lists tasks that must complete first
	DependsOn []string `yaml:"depends_on"`

	// Interactive indicates if this task requires user interaction
	Interactive bool `yaml:"interactive"`

	// Optional indicates if this task can be skipped on failure
	Optional bool `yaml:"optional"`
}

// CommandConfig contains command execution details
// What: Shell command with timeout
// Why: Need consistent command execution with timeout support
type CommandConfig struct {
	// Command is the shell command to execute
	Command string `yaml:"command"`

	// Timeout is maximum time allowed
	Timeout time.Duration `yaml:"timeout"`
}

// SetupStep represents a single step in a multi-step task
// What: Individual operation within a setup task
// Why: Some tasks require multiple sequential operations
type SetupStep struct {
	// Command to execute
	Command string `yaml:"command"`

	// Creates indicates file/dir this command creates (for idempotency)
	Creates string `yaml:"creates"`

	// Description of what this step does
	Description string `yaml:"description"`

	// EditToml for editing TOML configuration files
	EditToml *TomlEdit `yaml:"edit_toml"`
}

// TomlEdit represents a TOML file edit operation
// What: Modify specific key in TOML file
// Why: Common operation for tool configuration (e.g., starship.toml)
type TomlEdit struct {
	// File path to TOML file
	File string `yaml:"file"`

	// Section name in TOML (e.g., "package")
	Section string `yaml:"section"`

	// Key to set
	Key string `yaml:"key"`

	// Value to set
	Value interface{} `yaml:"value"`

	// Description of this edit
	Description string `yaml:"description"`
}

// ZshrcLine represents a line to add to .zshrc
// What: Comment and content line to add to shell RC file
// Why: Common operation for shell configuration
type ZshrcLine struct {
	// Comment to add above the line
	Comment string `yaml:"comment"`

	// Content to add
	Content string `yaml:"content"`
}

// PromptConfig defines interactive user prompt
// What: Configuration for prompting user for input (e.g., API keys)
// Why: Some tools need user-provided configuration (API keys, tokens)
type PromptConfig struct {
	// Message to show user
	Message string `yaml:"message"`

	// EnvVar name to set
	EnvVar string `yaml:"env_var"`

	// AddTo specifies file to add export statement
	AddTo string `yaml:"add_to"`

	// Format string for export statement (use {value} placeholder)
	Format string `yaml:"format"`

	// SkipIfSet skips prompt if env var already set
	SkipIfSet bool `yaml:"skip_if_set"`
}

// VerifyCheck represents a verification check
// What: Single verification operation to confirm setup worked
// Why: Need to validate configuration actually succeeded
type VerifyCheck struct {
	// Command to run (exit 0 = success)
	Command string `yaml:"command"`

	// EnvVar to check is set
	EnvVar string `yaml:"env_var"`

	// FileExists checks if file exists
	FileExists string `yaml:"file_exists"`

	// FileContains checks if file contains text
	FileContains *FileContainsCheck `yaml:"file_contains"`

	// TomlValue checks TOML value
	TomlValue *TomlValueCheck `yaml:"toml_value"`

	// Description of what this check verifies
	Description string `yaml:"description"`
}

// FileContainsCheck checks if file contains specific text
// What: Verify file contains expected content
// Why: Common check for dotfile modifications
type FileContainsCheck struct {
	// Path to file
	Path string `yaml:"path"`

	// Text that must be present
	Text string `yaml:"text"`

	// Description of check
	Description string `yaml:"description"`
}

// TomlValueCheck checks TOML file has expected value
// What: Verify TOML key has expected value
// Why: Validate TOML configuration edits
type TomlValueCheck struct {
	// File path to TOML file
	File string `yaml:"file"`

	// Section name
	Section string `yaml:"section"`

	// Key name
	Key string `yaml:"key"`

	// Equals is the expected value
	Equals interface{} `yaml:"equals"`

	// Description of check
	Description string `yaml:"description"`
}

// LoadSetupConfig loads and parses setup.yaml
// What: Reads setup.yaml from filesystem or embedded, parses into SetupConfig
// Why: Main entry point for loading setup task definitions
// Params: path - path to setup.yaml (e.g., "configs/setup.yaml")
// Returns: Parsed SetupConfig and error if any
// Example: cfg, err := LoadSetupConfig("configs/setup.yaml")
// Edge cases: Falls back to embedded if file not found on disk
func LoadSetupConfig(path string) (*SetupConfig, error) {
	// Try filesystem first (development)
	data, err := os.ReadFile(path)
	if err != nil {
		// Fall back to embedded (production)
		data, err = readEmbeddedFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read setup config: %w", err)
		}
	}

	var config SetupConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse setup config: %w", err)
	}

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid setup config: %w", err)
	}

	return &config, nil
}

// Validate checks if the setup configuration is valid
// What: Validates task names are unique, dependencies exist, strategies valid
// Why: Catch configuration errors early before setup starts
// Returns: Error describing validation failure, nil if valid
func (sc *SetupConfig) Validate() error {
	names := make(map[string]bool)
	for _, task := range sc.SetupTasks {
		// Check unique names
		if names[task.Name] {
			return fmt.Errorf("duplicate task name: %s", task.Name)
		}
		names[task.Name] = true

		// Validate strategy
		if task.Strategy != "" && task.Strategy != "remote_first" && task.Strategy != "local_only" {
			return fmt.Errorf("invalid strategy for task %s: %s", task.Name, task.Strategy)
		}

		// Validate dependencies exist
		for _, dep := range task.DependsOn {
			found := false
			for _, t := range sc.SetupTasks {
				if t.Name == dep {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("task %s depends on unknown task: %s", task.Name, dep)
			}
		}
	}

	return nil
}
