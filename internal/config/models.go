// File: internal/config/models.go
// Purpose: Defines data models for configuration files (stage configs, version locks)
// Problem: Need structured representation of YAML/TOML configs for type-safe access
// Role: Provides Go structs that map to configuration file formats
// Usage: Used by config loader to parse YAML/TOML into typed structures
// Design choices: Uses struct tags for YAML/TOML parsing; embedded structs for shared fields
// Assumptions: Configuration files follow documented schema; YAML for stages, TOML for versions

package config

import (
	"time"
)

// StageConfig represents a single installation stage configuration
// What: Complete definition of an installation stage including tasks and ordering
// Why: Allows declarative definition of what to install and how to install it
type StageConfig struct {
	// Name is the human-readable stage name
	Name string `yaml:"name"`

	// Timeout is maximum time allowed for entire stage
	Timeout time.Duration `yaml:"timeout"`

	// Parallel indicates if tasks can run concurrently
	Parallel bool `yaml:"parallel"`

	// Tasks are the individual operations to perform in this stage
	Tasks []Task `yaml:"tasks"`

	// PostStage contains actions to take after stage completes
	PostStage PostStageAction `yaml:"post_stage"`
}

// Task represents a single operation within a stage
// What: Individual unit of work (e.g. install Homebrew, clone repo)
// Why: Granular operations that can be tracked, parallelized, and retried
type Task struct {
	// Name is the human-readable task name
	Name string `yaml:"name"`

	// Command is the shell command to execute
	Command string `yaml:"command"`

	// ParallelGroup identifies tasks that can run together (empty = sequential)
	ParallelGroup string `yaml:"parallel_group"`

	// Required indicates if stage should fail if this task fails
	Required bool `yaml:"required"`

	// Timeout is maximum time allowed for this task
	Timeout time.Duration `yaml:"timeout"`

	// RetryCount is number of times to retry on failure
	RetryCount int `yaml:"retry_count"`

	// Condition is optional shell command to check if task should run
	// (empty string = always run, non-zero exit = skip task)
	Condition string `yaml:"condition"`
}

// PostStageAction defines what happens after a stage completes
// What: Post-stage actions like showing messages or triggering next stage
// Why: Provides user feedback and orchestrates multi-stage execution
type PostStageAction struct {
	// Message to show user after stage completes
	Message string `yaml:"message"`

	// NextStage is path to next stage config file
	NextStage string `yaml:"next_stage"`

	// Blocking indicates if next stage should block (true) or run in background (false)
	Blocking bool `yaml:"blocking"`
}

// VersionsLock represents the complete versions.lock file
// What: Single source of truth for all tool versions across the environment
// Why: Ensures identical versions installed on all developer machines
type VersionsLock struct {
	// Metadata contains version lock file information
	Metadata VersionsMetadata `toml:"metadata"`

	// Homebrew contains all Homebrew-managed packages
	Homebrew HomebrewConfig `toml:"homebrew"`

	// Tools contains non-Homebrew tools (uv, etc)
	Tools map[string]ToolConfig `toml:"tools"`

	// GitRepos contains git repositories to clone
	GitRepos map[string]GitRepoConfig `toml:"git_repos"`

	// Shell contains shell configuration details
	Shell ShellConfig `toml:"shell"`
}

// VersionsMetadata contains metadata about the versions.lock file
// What: Tracks schema version, platform, update time
// Why: Enables validation and migration of version lock files
type VersionsMetadata struct {
	// SchemaVersion is the versions.lock schema version
	SchemaVersion string `toml:"schema_version"`

	// Platform is the target platform (darwin, linux)
	Platform string `toml:"platform"`

	// MinMacOSVersion is minimum macOS version required
	MinMacOSVersion string `toml:"min_macos_version"`

	// Updated is when this file was last updated
	Updated time.Time `toml:"updated"`
}

// HomebrewConfig contains all Homebrew package definitions
// What: Defines formulas and casks to install via Homebrew
// Why: Separates formulas (CLI tools) from casks (GUI apps)
type HomebrewConfig struct {
	// Formulas are CLI tools/libraries
	Formulas map[string]HomebrewFormula `toml:"formulas"`

	// Casks are GUI applications
	Casks map[string]HomebrewCask `toml:"casks"`
}

// HomebrewFormula represents a Homebrew formula (CLI tool)
// What: Version and tap information for a formula
// Why: Allows version pinning and custom tap usage
type HomebrewFormula struct {
	// Version is the exact version to install
	Version string `toml:"version"`

	// Tap is the Homebrew tap (default: homebrew/core)
	Tap string `toml:"tap"`

	// Options are additional install flags (e.g. --HEAD, --with-feature)
	Options []string `toml:"options"`
}

// HomebrewCask represents a Homebrew cask (GUI app)
// What: Version and tap information for a cask
// Why: Ensures consistent GUI app versions across machines
type HomebrewCask struct {
	// Version is the exact version to install
	Version string `toml:"version"`

	// Tap is the Homebrew tap (default: homebrew/cask)
	Tap string `toml:"tap"`
}

// ToolConfig represents a non-Homebrew tool
// What: Configuration for tools installed via curl/script
// Why: Some tools (uv) use custom installers, not Homebrew
type ToolConfig struct {
	// Version is the exact version to install
	Version string `toml:"version"`

	// Installer is URL to installation script
	Installer string `toml:"installer"`

	// Env are environment variables to set during install
	Env map[string]string `toml:"env"`
}

// GitRepoConfig represents a git repository to clone
// What: URL, version, and destination for git clone operations
// Why: Many dev tools (Flutter wrapper, Zsh plugins) distributed via git
type GitRepoConfig struct {
	// URL is the git repository URL (https or git protocol)
	URL string `toml:"url"`

	// Commit is the exact commit SHA to checkout (preferred)
	Commit string `toml:"commit"`

	// Tag is the git tag to checkout (if no commit specified)
	Tag string `toml:"tag"`

	// Branch is the branch to checkout (if no commit or tag)
	Branch string `toml:"branch"`

	// Path is where to clone the repository
	Path string `toml:"path"`

	// Shallow indicates if should do shallow clone (--depth=1)
	Shallow bool `toml:"shallow"`

	// Stage indicates which stage should install this repo (1, 2, or 3)
	Stage int `toml:"stage"`
}

// ShellConfig contains shell configuration settings
// What: Defines which shell to configure and paths to config files
// Why: Automates shell setup (PATH, plugins, etc)
type ShellConfig struct {
	// Default is the default shell (zsh, bash)
	Default string `toml:"default"`

	// ProfilePath is path to shell profile file (~/.zprofile)
	ProfilePath string `toml:"profile_path"`

	// RCPath is path to shell RC file (~/.zshrc)
	RCPath string `toml:"rc_path"`
}

// Brewfile represents a parsed Brewfile
// What: Structured representation of Homebrew Bundle Brewfile
// Why: Allows programmatic access to Brewfile contents
type Brewfile struct {
	// Taps are Homebrew taps to add
	Taps []string

	// Brews are formulas to install
	Brews []BrewfileFormula

	// Casks are casks to install
	Casks []BrewfileCask
}

// BrewfileFormula represents a formula in Brewfile
// What: Formula name and optional arguments
// Why: Maps to brew "name", args: [...] syntax in Brewfile
type BrewfileFormula struct {
	// Name is the formula name
	Name string

	// Args are optional install arguments
	Args []string
}

// BrewfileCask represents a cask in Brewfile
// What: Cask name
// Why: Maps to cask "name" syntax in Brewfile
type BrewfileCask struct {
	// Name is the cask name
	Name string
}
