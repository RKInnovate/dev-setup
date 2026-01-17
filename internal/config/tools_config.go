// File: internal/config/tools_config.go
// Purpose: Data models for tools.yaml configuration
// Problem: Need structured representation of tool installation definitions
// Role: Provides Go structs for tools configuration with install commands and checks
// Usage: Loaded by install command to determine what tools to install
// Design choices: Simple model with check/install/depends; supports parallel groups
// Assumptions: Tools can be checked via command existence; Homebrew available after first tool

package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ToolsConfig represents the complete tools.yaml file
// What: List of tools to install with their installation commands
// Why: Declarative tool installation with idempotency and dependencies
type ToolsConfig struct {
	// Tools are the list of tools to install
	Tools []Tool `yaml:"tools"`
}

// Tool represents a single tool installation definition
// What: Individual tool with check command, install command, and metadata
// Why: Each tool needs idempotency check and installation method
type Tool struct {
	// Name is the unique identifier for this tool
	Name string `yaml:"name"`

	// Description is human-readable description
	Description string `yaml:"description"`

	// Check is shell command that returns 0 if tool is already installed
	Check string `yaml:"check"`

	// Install contains installation details
	Install ToolInstall `yaml:"install"`

	// DependsOn lists tools that must be installed first
	DependsOn []string `yaml:"depends_on"`

	// Required indicates if installation should fail if this tool fails
	Required bool `yaml:"required"`
}

// ToolInstall contains installation command details
// What: How to install the tool (command, timeout, parallel group)
// Why: Need flexibility for different installation methods and parallelism
type ToolInstall struct {
	// Command is the shell command to run
	Command string `yaml:"command"`

	// ParallelGroup identifies tools that can install concurrently
	ParallelGroup string `yaml:"parallel_group"`

	// Timeout is maximum time allowed for installation
	Timeout time.Duration `yaml:"timeout"`
}

// LoadToolsConfig loads and parses tools.yaml
// What: Reads tools.yaml from filesystem or embedded, parses into ToolsConfig
// Why: Main entry point for loading tool definitions
// Params: path - path to tools.yaml (e.g., "configs/tools.yaml")
// Returns: Parsed ToolsConfig and error if any
// Example: cfg, err := LoadToolsConfig("configs/tools.yaml")
// Edge cases: Falls back to embedded if file not found on disk
func LoadToolsConfig(path string) (*ToolsConfig, error) {
	// Try filesystem first (development)
	data, err := os.ReadFile(path)
	if err != nil {
		// Fall back to embedded (production)
		data, err = readEmbeddedFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read tools config: %w", err)
		}
	}

	var config ToolsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse tools config: %w", err)
	}

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid tools config: %w", err)
	}

	return &config, nil
}

// Validate checks if the tools configuration is valid
// What: Validates tool names are unique, dependencies exist, no cycles
// Why: Catch configuration errors early before installation starts
// Returns: Error describing validation failure, nil if valid
func (tc *ToolsConfig) Validate() error {
	names := make(map[string]bool)
	for _, tool := range tc.Tools {
		// Check unique names
		if names[tool.Name] {
			return fmt.Errorf("duplicate tool name: %s", tool.Name)
		}
		names[tool.Name] = true

		// Validate dependencies exist
		for _, dep := range tool.DependsOn {
			if !names[dep] {
				// Dependency might be defined later, check all
				found := false
				for _, t := range tc.Tools {
					if t.Name == dep {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("tool %s depends on unknown tool: %s", tool.Name, dep)
				}
			}
		}
	}

	return nil
}

// GetInstallOrder returns tools in dependency order
// What: Topologically sorts tools based on depends_on relationships
// Why: Must install dependencies before dependents
// Returns: Ordered slice of tools, error if circular dependency detected
func (tc *ToolsConfig) GetInstallOrder() ([]Tool, error) {
	// Build dependency graph
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for _, tool := range tc.Tools {
		if _, exists := inDegree[tool.Name]; !exists {
			inDegree[tool.Name] = 0
		}
		for _, dep := range tool.DependsOn {
			graph[dep] = append(graph[dep], tool.Name)
			inDegree[tool.Name]++
		}
	}

	// Topological sort using Kahn's algorithm
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	var ordered []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		ordered = append(ordered, current)

		for _, dependent := range graph[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check for cycles
	if len(ordered) != len(tc.Tools) {
		return nil, fmt.Errorf("circular dependency detected in tools")
	}

	// Convert names back to Tool objects in order
	nameToTool := make(map[string]Tool)
	for _, tool := range tc.Tools {
		nameToTool[tool.Name] = tool
	}

	var result []Tool
	for _, name := range ordered {
		result = append(result, nameToTool[name])
	}

	return result, nil
}
