// File: internal/installer/tool_installer.go
// Purpose: Tool installation with idempotency checks and parallel execution
// Problem: Need to install tools efficiently without reinstalling what exists
// Role: Orchestrates tool installation with dependency resolution and state tracking
// Usage: Create ToolInstaller, call InstallAll() to install all tools from config
// Design choices: Check before install; parallel within groups; dependency-ordered; state tracking
// Assumptions: Tools can be checked via shell commands; Homebrew available after first tool

package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/rkinnovate/dev-setup/internal/config"
	"github.com/rkinnovate/dev-setup/internal/ui"
)

// ToolInstaller manages tool installation with idempotency and parallelism
// What: Installs tools from tools.yaml with proper checking and ordering
// Why: Need reliable, fast installation that doesn't redo completed work
type ToolInstaller struct {
	toolsConfig *config.ToolsConfig
	state       *config.State
	ui          ui.UI
	dryRun      bool
	version     string
}

// NewToolInstaller creates a new tool installer
// What: Constructor for ToolInstaller with config and state
// Why: Centralized creation with all dependencies
// Params: toolsConfig - loaded tools configuration, state - current state, ui - UI for feedback, dryRun - if true, don't actually install
// Returns: Configured ToolInstaller instance
// Example: installer := NewToolInstaller(cfg, state, ui, false)
func NewToolInstaller(toolsConfig *config.ToolsConfig, state *config.State, ui ui.UI, dryRun bool, version string) *ToolInstaller {
	return &ToolInstaller{
		toolsConfig: toolsConfig,
		state:       state,
		ui:          ui,
		dryRun:      dryRun,
		version:     version,
	}
}

// InstallAll installs all tools from configuration
// What: Main entry point for tool installation, handles all tools with dependencies
// Why: Single method to install entire tool suite
// Returns: Error if any required tool fails, nil if all succeeded
// Example: err := installer.InstallAll()
// Edge cases: Skips already-installed tools; respects dependencies; parallel within groups
func (ti *ToolInstaller) InstallAll() error {
	ti.ui.Info("📦 Starting tool installation...")
	ti.ui.Info("")

	// Get tools in dependency order
	orderedTools, err := ti.toolsConfig.GetInstallOrder()
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	ti.ui.Info("Installing %d tools...", len(orderedTools))
	ti.ui.Info("")

	// Group tools by parallel group
	toolGroups := ti.groupToolsByParallelGroup(orderedTools)

	// Install each group (sequential between groups, parallel within groups)
	for _, group := range toolGroups {
		if err := ti.installGroup(group); err != nil {
			return fmt.Errorf("installation failed: %w", err)
		}
	}

	ti.ui.Info("")
	ti.ui.Success("✅ Tool installation complete!")
	ti.ui.Info("")

	// Save final state
	if !ti.dryRun {
		ti.state.Version = ti.version
		if err := config.SaveState(ti.state); err != nil {
			ti.ui.Warning("⚠️  Failed to save state: %v", err)
		}
	}

	return nil
}

// groupToolsByParallelGroup groups tools for parallel execution
// What: Groups tools by parallel_group field, preserving dependency order
// Why: Tools in same group can run concurrently; different groups run sequentially
// Params: tools - ordered tools list
// Returns: Slice of tool groups (each group can run in parallel)
func (ti *ToolInstaller) groupToolsByParallelGroup(tools []config.Tool) [][]config.Tool {
	var groups [][]config.Tool
	currentGroup := []config.Tool{}
	lastParallelGroup := ""

	for _, tool := range tools {
		parallelGroup := tool.Install.ParallelGroup

		// If this tool has different parallel group, start new group
		if parallelGroup != lastParallelGroup && len(currentGroup) > 0 {
			groups = append(groups, currentGroup)
			currentGroup = []config.Tool{}
		}

		currentGroup = append(currentGroup, tool)
		lastParallelGroup = parallelGroup
	}

	// Add final group
	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}

	return groups
}

// installGroup installs a group of tools (in parallel if >1 tool)
// What: Installs all tools in a group concurrently
// Why: Maximize installation speed within a group
// Params: tools - slice of tools to install
// Returns: Error if any required tool fails
func (ti *ToolInstaller) installGroup(tools []config.Tool) error {
	if len(tools) == 0 {
		return nil
	}

	// If only one tool, install sequentially
	if len(tools) == 1 {
		return ti.installTool(tools[0])
	}

	// Multiple tools - install in parallel
	ti.ui.Info("⚡ Installing %d tools in parallel...", len(tools))

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstError error

	for _, tool := range tools {
		wg.Add(1)
		go func(t config.Tool) {
			defer wg.Done()

			if err := ti.installTool(t); err != nil {
				mu.Lock()
				if firstError == nil {
					firstError = err
				}
				mu.Unlock()
			}
		}(tool)
	}

	wg.Wait()

	return firstError
}

// installTool installs a single tool with idempotency check
// What: Checks if tool exists, installs if missing, updates state
// Why: Core installation logic with proper checking
// Params: tool - Tool to install
// Returns: Error if installation fails and tool is required
func (ti *ToolInstaller) installTool(tool config.Tool) error {
	// Check if already installed
	if ti.isToolInstalled(tool) {
		ti.ui.Info("✓ %s (already installed)", tool.Name)
		return nil
	}

	ti.ui.StartTask(tool.Name)

	// Dry run mode
	if ti.dryRun {
		ti.ui.Info("  [DRY RUN] Would install: %s", tool.Name)
		ti.ui.CompleteTask(tool.Name)
		return nil
	}

	// Install the tool
	ctx := context.Background()
	if tool.Install.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, tool.Install.Timeout)
		defer cancel()
	}

	if err := ti.runInstallCommand(ctx, tool); err != nil {
		ti.ui.FailTask(tool.Name, err)

		if tool.Required {
			return fmt.Errorf("required tool %s failed: %w", tool.Name, err)
		}

		ti.ui.Warning("⚠️  Optional tool %s failed: %v", tool.Name, err)
		return nil
	}

	ti.ui.CompleteTask(tool.Name)

	// Update state
	version, path := ti.getToolInfo(tool)
	config.MarkToolInstalled(ti.state, tool.Name, version, path)

	return nil
}

// isToolInstalled checks if a tool is already installed
// What: Runs the check command to see if tool exists
// Why: Idempotency - don't reinstall what exists
// Params: tool - Tool to check
// Returns: True if tool is installed, false otherwise
func (ti *ToolInstaller) isToolInstalled(tool config.Tool) bool {
	if tool.Check == "" {
		return false
	}

	// First check state
	if config.IsToolInstalled(ti.state, tool.Name) {
		// Verify it still exists
		cmd := exec.Command("sh", "-c", tool.Check)
		if err := cmd.Run(); err == nil {
			return true
		}
		// Tool was in state but no longer exists, need to reinstall
	}

	// Check via command
	cmd := exec.Command("sh", "-c", tool.Check)
	err := cmd.Run()
	return err == nil
}

// runInstallCommand executes the installation command
// What: Runs the shell command to install the tool
// Why: Actual installation work
// Params: ctx - context for timeout, tool - Tool to install
// Returns: Error if command fails
func (ti *ToolInstaller) runInstallCommand(ctx context.Context, tool config.Tool) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", tool.Install.Command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install command failed: %w", err)
	}

	return nil
}

// getToolInfo extracts version and path of installed tool
// What: Gets version string and binary path for installed tool
// Why: Populate state with installation details
// Params: tool - Installed tool
// Returns: version string and path string
func (ti *ToolInstaller) getToolInfo(tool config.Tool) (string, string) {
	// Try to get version
	version := "unknown"
	versionCommands := []string{
		tool.Name + " --version",
		tool.Name + " -v",
		tool.Name + " version",
	}

	for _, cmd := range versionCommands {
		if output, err := exec.Command("sh", "-c", cmd).Output(); err == nil {
			version = strings.TrimSpace(string(output))
			// Take first line only
			if lines := strings.Split(version, "\n"); len(lines) > 0 {
				version = lines[0]
			}
			break
		}
	}

	// Get path
	path := "unknown"
	if output, err := exec.Command("sh", "-c", "command -v "+tool.Name).Output(); err == nil {
		path = strings.TrimSpace(string(output))
	}

	return version, path
}
