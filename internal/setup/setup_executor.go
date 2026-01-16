// File: internal/setup/setup_executor.go
// Purpose: Post-install configuration with remote-first/local-fallback strategy
// Problem: Tools need configuration after installation (API keys, dotfiles, etc)
// Role: Executes setup tasks with interactive prompts and file operations
// Usage: Create SetupExecutor, call SetupAll() to configure all tools
// Design choices: Remote-first with local fallback; interactive prompts; file editing helpers
// Assumptions: Tools already installed; user present for interactive prompts; network available

package setup

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rkinnovate/dev-setup/internal/config"
	"github.com/rkinnovate/dev-setup/internal/ui"
)

// SetupExecutor manages post-install configuration tasks
// What: Executes setup tasks from setup.yaml with verification
// Why: Need configurable, verifiable post-install setup
type SetupExecutor struct {
	setupConfig *config.SetupConfig
	state       *config.State
	ui          ui.UI
	dryRun      bool
}

// NewSetupExecutor creates a new setup executor
// What: Constructor for SetupExecutor with config and state
// Why: Centralized creation with dependencies
// Params: setupConfig - loaded setup configuration, state - current state, ui - UI for feedback, dryRun - if true, don't actually configure
// Returns: Configured SetupExecutor instance
// Example: executor := NewSetupExecutor(cfg, state, ui, false)
func NewSetupExecutor(setupConfig *config.SetupConfig, state *config.State, ui ui.UI, dryRun bool) *SetupExecutor {
	return &SetupExecutor{
		setupConfig: setupConfig,
		state:       state,
		ui:          ui,
		dryRun:      dryRun,
	}
}

// SetupAll executes all setup tasks from configuration
// What: Main entry point for post-install configuration
// Why: Single method to configure entire environment
// Returns: Error if any required task fails
// Example: err := executor.SetupAll()
func (se *SetupExecutor) SetupAll() error {
	se.ui.Info("⚙️  Starting post-install setup...")
	se.ui.Info("")

	for _, task := range se.setupConfig.SetupTasks {
		// Check if already configured
		if config.IsTaskConfigured(se.state, task.Name) {
			se.ui.Info("✓ %s (already configured)", task.Name)
			continue
		}

		se.ui.StartTask(task.Name)

		if se.dryRun {
			se.ui.Info("  [DRY RUN] Would configure: %s", task.Name)
			se.ui.CompleteTask(task.Name)
			continue
		}

		// Execute the setup task
		if err := se.executeTask(task); err != nil {
			se.ui.FailTask(task.Name, err)

			if !task.Optional {
				return fmt.Errorf("required task %s failed: %w", task.Name, err)
			}

			se.ui.Warning("⚠️  Optional task %s failed: %v", task.Name, err)
			continue
		}

		se.ui.CompleteTask(task.Name)

		// Mark as configured
		config.MarkTaskConfigured(se.state, task.Name)

		// Save state after each task
		if err := config.SaveState(se.state); err != nil {
			se.ui.Warning("⚠️  Failed to save state: %v", err)
		}
	}

	se.ui.Info("")
	se.ui.Success("✅ Setup complete!")
	se.ui.Info("")

	return nil
}

// executeTask executes a single setup task
// What: Runs one setup task based on its strategy
// Why: Different tasks need different execution strategies
// Params: task - SetupTask to execute
// Returns: Error if task fails
func (se *SetupExecutor) executeTask(task config.SetupTask) error {
	switch task.Strategy {
	case "remote_first":
		return se.executeRemoteFirst(task)
	case "local_only":
		return se.executeLocalOnly(task)
	case "":
		// No strategy specified, try to infer from fields
		if len(task.ZshrcLines) > 0 {
			return se.executeZshrcConfig(task)
		}
		if len(task.Steps) > 0 {
			return se.executeSteps(task)
		}
		if task.Prompt != nil {
			return se.executePrompt(task)
		}
		return fmt.Errorf("no execution strategy specified for task %s", task.Name)
	default:
		return fmt.Errorf("unknown strategy: %s", task.Strategy)
	}
}

// executeRemoteFirst tries remote installation first, falls back to local
// What: Remote-first with local fallback execution strategy
// Why: Prefer latest remote version, but work offline with local copy
// Params: task - Task with remote and local commands
// Returns: Error if both remote and local fail
func (se *SetupExecutor) executeRemoteFirst(task config.SetupTask) error {
	// Try remote first
	if task.Remote != nil {
		se.ui.Info("  Trying remote installation...")
		ctx, cancel := se.getContext(task.Remote.Timeout)
		defer cancel()

		if err := se.runCommand(ctx, task.Remote.Command); err == nil {
			se.ui.Success("  ✓ Remote installation succeeded")
			return nil
		} else {
			se.ui.Warning("  ⚠️  Remote failed: %v", err)
		}
	}

	// Fall back to local
	if task.Local != nil {
		se.ui.Info("  Falling back to local submodule...")
		ctx, cancel := se.getContext(task.Local.Timeout)
		defer cancel()

		if err := se.runCommand(ctx, task.Local.Command); err != nil {
			return fmt.Errorf("both remote and local failed: %w", err)
		}

		se.ui.Success("  ✓ Local installation succeeded")
		return nil
	}

	return fmt.Errorf("no remote or local command specified")
}

// executeLocalOnly executes local-only installation commands
// What: Runs commands from local submodule
// Why: Some tasks only work with local files
// Params: task - Task with install commands
// Returns: Error if commands fail
func (se *SetupExecutor) executeLocalOnly(task config.SetupTask) error {
	for _, cmd := range task.Install {
		ctx, cancel := se.getContext(30 * time.Second)
		defer cancel()

		if err := se.runCommand(ctx, cmd); err != nil {
			return fmt.Errorf("command failed: %w", err)
		}
	}
	return nil
}

// executeZshrcConfig adds lines to .zshrc
// What: Adds configuration lines to ~/.zshrc if not present
// Why: Common operation for shell setup
// Params: task - Task with zshrc_lines
// Returns: Error if file operations fail
func (se *SetupExecutor) executeZshrcConfig(task config.SetupTask) error {
	zshrcPath := filepath.Join(os.Getenv("HOME"), ".zshrc")

	// Read existing file
	content, err := os.ReadFile(zshrcPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read .zshrc: %w", err)
	}

	existingContent := string(content)
	newLines := []string{}

	// Check each line
	for _, line := range task.ZshrcLines {
		if !strings.Contains(existingContent, line.Content) {
			// Add comment and content
			if line.Comment != "" {
				newLines = append(newLines, line.Comment)
			}
			newLines = append(newLines, line.Content)
		}
	}

	// If nothing to add, we're done
	if len(newLines) == 0 {
		se.ui.Info("  All lines already present in .zshrc")
		return nil
	}

	// Append new lines
	newContent := existingContent
	if !strings.HasSuffix(newContent, "\n") && newContent != "" {
		newContent += "\n"
	}
	newContent += "\n" + strings.Join(newLines, "\n") + "\n"

	// Write back
	if err := os.WriteFile(zshrcPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write .zshrc: %w", err)
	}

	se.ui.Success("  ✓ Added %d lines to .zshrc", len(newLines))
	return nil
}

// executeSteps executes multi-step configuration
// What: Runs multiple steps in sequence
// Why: Some tasks need multiple operations
// Params: task - Task with steps
// Returns: Error if any step fails
func (se *SetupExecutor) executeSteps(task config.SetupTask) error {
	for i, step := range task.Steps {
		se.ui.Info("  Step %d/%d: %s", i+1, len(task.Steps), step.Description)

		// Check if this step creates a file that already exists
		if step.Creates != "" {
			expanded := os.ExpandEnv(step.Creates)
			if _, err := os.Stat(expanded); err == nil {
				se.ui.Info("    Skipped (already exists)")
				continue
			}
		}

		// Handle TOML edit
		if step.EditToml != nil {
			if err := se.editTomlFile(step.EditToml); err != nil {
				return fmt.Errorf("step %d failed: %w", i+1, err)
			}
			continue
		}

		// Run command
		if step.Command != "" {
			ctx, cancel := se.getContext(30 * time.Second)
			defer cancel()

			if err := se.runCommand(ctx, step.Command); err != nil {
				return fmt.Errorf("step %d failed: %w", i+1, err)
			}
		}
	}
	return nil
}

// executePrompt handles interactive user prompts
// What: Prompts user for input (e.g., API keys) and saves to file
// Why: Some tools need user-provided configuration
// Params: task - Task with prompt configuration
// Returns: Error if prompt or file operations fail
func (se *SetupExecutor) executePrompt(task config.SetupTask) error {
	prompt := task.Prompt

	// Check if already set and skip_if_set is true
	if prompt.SkipIfSet && os.Getenv(prompt.EnvVar) != "" {
		se.ui.Info("  %s already set, skipping prompt", prompt.EnvVar)
		return nil
	}

	// Prompt user
	se.ui.Info("")
	se.ui.Info("  %s", prompt.Message)
	se.ui.Info("")

	reader := bufio.NewReader(os.Stdin)
	value, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	value = strings.TrimSpace(value)

	// If empty and optional, skip
	if value == "" {
		se.ui.Info("  Skipped")
		return nil
	}

	// Add to file
	if prompt.AddTo != "" {
		filePath := os.ExpandEnv(prompt.AddTo)
		exportLine := strings.ReplaceAll(prompt.Format, "{value}", value)

		// Read existing file
		content, err := os.ReadFile(filePath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to read %s: %w", filePath, err)
		}

		// Check if already present
		if strings.Contains(string(content), exportLine) {
			se.ui.Info("  Export already present in %s", filePath)
			return nil
		}

		// Append
		newContent := string(content)
		if !strings.HasSuffix(newContent, "\n") && newContent != "" {
			newContent += "\n"
		}
		newContent += "\n" + exportLine + "\n"

		if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filePath, err)
		}

		se.ui.Success("  ✓ Added to %s", filePath)
	}

	return nil
}

// editTomlFile edits a TOML configuration file
// What: Updates a key in a TOML file
// Why: Common operation for tool configuration (e.g., starship.toml)
// Params: edit - TOML edit configuration
// Returns: Error if file operations fail
func (se *SetupExecutor) editTomlFile(edit *config.TomlEdit) error {
	// TODO: Implement proper TOML editing
	// For now, just log what would be done
	se.ui.Info("    Would edit %s: [%s].%s = %v", edit.File, edit.Section, edit.Key, edit.Value)
	se.ui.Warning("    ⚠️  TOML editing not yet implemented - please edit manually")
	return nil
}

// runCommand executes a shell command
// What: Runs shell command with context for timeout
// Why: Common operation across all strategies
// Params: ctx - context for timeout, command - shell command
// Returns: Error if command fails
func (se *SetupExecutor) runCommand(ctx context.Context, command string) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}

// getContext creates a context with timeout
// What: Creates context with timeout or background context
// Why: Consistent timeout handling
// Params: timeout - duration for timeout (0 = no timeout)
// Returns: Context and cancel function
func (se *SetupExecutor) getContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(context.Background(), timeout)
	}
	return context.WithCancel(context.Background())
}
