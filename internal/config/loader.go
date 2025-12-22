// File: internal/config/loader.go
// Purpose: Loads and parses configuration files (YAML stages, TOML versions.lock, Brewfile)
// Problem: Need to read various config formats and convert to Go structs
// Role: Centralized configuration loading with validation and error handling
// Usage: Call LoadStageConfig() or LoadVersionsLock() with file paths
// Design choices: Uses gopkg.in/yaml.v3 for YAML, BurntSushi/toml for TOML; validates after parsing
// Assumptions: Config files exist and are readable; YAML for stages, TOML for version locks

package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// LoadStageConfig loads a stage configuration file from disk or embedded filesystem
// What: Reads YAML file and parses into StageConfig struct
// Why: Stages are defined declaratively in YAML files
// Params: path - filesystem path to stage YAML file
// Returns: Parsed StageConfig struct and error if any
// Example: cfg, err := LoadStageConfig("configs/stage1.yaml")
func LoadStageConfig(path string) (*StageConfig, error) {
	// Try to read from filesystem first (for development)
	data, err := os.ReadFile(path)
	if err != nil {
		// If file not found on disk, try embedded filesystem
		data, err = readEmbeddedFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read stage config %s (tried filesystem and embedded): %w", path, err)
		}
	}

	// Parse YAML
	var cfg StageConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse stage config %s: %w", path, err)
	}

	// Validate configuration
	if err := validateStageConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid stage config %s: %w", path, err)
	}

	return &cfg, nil
}

// LoadVersionsLock loads the versions.lock file
// What: Reads TOML file and parses into VersionsLock struct
// Why: Version locks are defined in TOML format for readability
// Params: path - filesystem path to versions.lock file (default: "versions.lock")
// Returns: Parsed VersionsLock struct and error if any
// Example: lock, err := LoadVersionsLock("versions.lock")
func LoadVersionsLock(path string) (*VersionsLock, error) {
	// Use default path if not specified
	if path == "" {
		path = "versions.lock"
	}

	// Parse TOML file
	var lock VersionsLock
	if _, err := toml.DecodeFile(path, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse versions.lock %s: %w", path, err)
	}

	// Validate configuration
	if err := validateVersionsLock(&lock); err != nil {
		return nil, fmt.Errorf("invalid versions.lock %s: %w", path, err)
	}

	return &lock, nil
}

// LoadBrewfile loads and parses a Brewfile
// What: Reads Brewfile and extracts taps, formulas, and casks
// Why: Brewfiles use Ruby-like DSL that needs custom parsing
// Params: path - filesystem path to Brewfile (default: "Brewfile")
// Returns: Parsed Brewfile struct and error if any
// Example: brewfile, err := LoadBrewfile("Brewfile")
func LoadBrewfile(path string) (*Brewfile, error) {
	// Use default path if not specified
	if path == "" {
		path = "Brewfile"
	}

	// Open file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Brewfile %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	brewfile := &Brewfile{
		Taps:  []string{},
		Brews: []BrewfileFormula{},
		Casks: []BrewfileCask{},
	}

	// Parse line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse tap directives
		if strings.HasPrefix(line, "tap ") {
			tap := extractQuotedString(line)
			if tap != "" {
				brewfile.Taps = append(brewfile.Taps, tap)
			}
		}

		// Parse brew directives
		if strings.HasPrefix(line, "brew ") {
			name := extractQuotedString(line)
			if name != "" {
				brewfile.Brews = append(brewfile.Brews, BrewfileFormula{
					Name: name,
					Args: extractArgs(line),
				})
			}
		}

		// Parse cask directives
		if strings.HasPrefix(line, "cask ") {
			name := extractQuotedString(line)
			if name != "" {
				brewfile.Casks = append(brewfile.Casks, BrewfileCask{
					Name: name,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading Brewfile %s: %w", path, err)
	}

	return brewfile, nil
}

// validateStageConfig validates a StageConfig for correctness
// What: Checks that stage config has required fields and valid values
// Why: Catches configuration errors early before execution
// Params: cfg - StageConfig to validate
// Returns: Error if validation fails, nil if valid
func validateStageConfig(cfg *StageConfig) error {
	if cfg.Name == "" {
		return fmt.Errorf("stage name is required")
	}

	if len(cfg.Tasks) == 0 {
		return fmt.Errorf("stage must have at least one task")
	}

	// Validate each task
	for i, task := range cfg.Tasks {
		if task.Name == "" {
			return fmt.Errorf("task %d: name is required", i)
		}
		if task.Command == "" {
			return fmt.Errorf("task %d (%s): command is required", i, task.Name)
		}
	}

	return nil
}

// validateVersionsLock validates a VersionsLock for correctness
// What: Checks that versions.lock has required fields and valid values
// Why: Catches configuration errors early before installation
// Params: lock - VersionsLock to validate
// Returns: Error if validation fails, nil if valid
func validateVersionsLock(lock *VersionsLock) error {
	// Validate metadata
	if lock.Metadata.SchemaVersion == "" {
		return fmt.Errorf("metadata.schema_version is required")
	}

	if lock.Metadata.Platform == "" {
		return fmt.Errorf("metadata.platform is required")
	}

	// Validate platform
	if lock.Metadata.Platform != "darwin" && lock.Metadata.Platform != "linux" {
		return fmt.Errorf("unsupported platform: %s (must be darwin or linux)", lock.Metadata.Platform)
	}

	// Validate Homebrew formulas
	for name, formula := range lock.Homebrew.Formulas {
		if formula.Version == "" {
			return fmt.Errorf("homebrew formula %s: version is required", name)
		}
	}

	// Validate Homebrew casks
	for name, cask := range lock.Homebrew.Casks {
		if cask.Version == "" {
			return fmt.Errorf("homebrew cask %s: version is required", name)
		}
	}

	// Validate git repos
	for name, repo := range lock.GitRepos {
		if repo.URL == "" {
			return fmt.Errorf("git repo %s: url is required", name)
		}
		if repo.Commit == "" && repo.Tag == "" && repo.Branch == "" {
			return fmt.Errorf("git repo %s: must specify commit, tag, or branch", name)
		}
		if repo.Path == "" {
			return fmt.Errorf("git repo %s: path is required", name)
		}
	}

	return nil
}

// extractQuotedString extracts a quoted string from a line
// What: Finds and returns content between first pair of quotes
// Why: Brewfile uses quoted strings for package names
// Params: line - input line containing quoted string
// Returns: Extracted string without quotes, empty if not found
// Example: extractQuotedString('brew "git"') returns "git"
func extractQuotedString(line string) string {
	start := strings.Index(line, "\"")
	if start == -1 {
		return ""
	}

	end := strings.Index(line[start+1:], "\"")
	if end == -1 {
		return ""
	}

	return line[start+1 : start+1+end]
}

// extractArgs extracts arguments from a brew/cask line
// What: Parses args: [...] section from Brewfile line
// Why: Some packages need additional install flags
// Params: line - input line potentially containing args
// Returns: Slice of argument strings, empty if no args
// Example: extractArgs('brew "git", args: ["--HEAD"]') returns ["--HEAD"]
func extractArgs(line string) []string {
	argsStart := strings.Index(line, "args: [")
	if argsStart == -1 {
		return []string{}
	}

	argsEnd := strings.Index(line[argsStart:], "]")
	if argsEnd == -1 {
		return []string{}
	}

	argsStr := line[argsStart+7 : argsStart+argsEnd]
	args := strings.Split(argsStr, ",")

	// Clean up args (remove quotes and whitespace)
	cleaned := []string{}
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		arg = strings.Trim(arg, "\"")
		if arg != "" {
			cleaned = append(cleaned, arg)
		}
	}

	return cleaned
}
