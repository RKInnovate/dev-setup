// File: internal/status/reporter.go
// Purpose: Status reporting with accurate progress tracking
// Problem: Need clear visibility into what's installed and configured
// Role: Displays installation and configuration status with progress
// Usage: Create Reporter, call ShowStatus() to display status
// Design choices: Pretty-print with colors; show versions and paths; calculate progress
// Assumptions: State file contains accurate information

package status

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rkinnovate/dev-setup/internal/config"
	"github.com/rkinnovate/dev-setup/internal/ui"
)

// Reporter displays status information
type Reporter struct {
	toolsConfig *config.ToolsConfig
	setupConfig *config.SetupConfig
	state       *config.State
	ui          ui.UI
}

// NewReporter creates a new status reporter
func NewReporter(toolsConfig *config.ToolsConfig, setupConfig *config.SetupConfig, state *config.State, ui ui.UI) *Reporter {
	return &Reporter{
		toolsConfig: toolsConfig,
		setupConfig: setupConfig,
		state:       state,
		ui:          ui,
	}
}

// ShowStatus displays current installation and configuration status
func (r *Reporter) ShowStatus() {
	r.ui.Info("")
	r.ui.Info("╔══════════════════════════════════════════════════════╗")
	r.ui.Info("║           Development Environment Status             ║")
	r.ui.Info("╚══════════════════════════════════════════════════════╝")
	r.ui.Info("")

	// Tools status
	r.showToolsStatus()

	r.ui.Info("")

	// Setup status
	r.showSetupStatus()

	r.ui.Info("")

	// Overall progress
	r.showOverallProgress()

	r.ui.Info("")
}

// showToolsStatus displays installed tools
// What: Shows which tools are installed, checking state first then running actual checks
// Why: Provides accurate status even for manually installed tools
func (r *Reporter) showToolsStatus() {
	totalTools := len(r.toolsConfig.Tools)
	installedCount := 0

	// Count installed tools (state + actual checks)
	for _, tool := range r.toolsConfig.Tools {
		if _, ok := r.state.Installed[tool.Name]; ok {
			installedCount++
		} else if r.isToolActuallyInstalled(tool) {
			installedCount++
		}
	}

	r.ui.Info("📦 Installed Tools (%d/%d complete):", installedCount, totalTools)

	for _, tool := range r.toolsConfig.Tools {
		if toolState, ok := r.state.Installed[tool.Name]; ok {
			// Tool in state - show version info
			r.ui.Success("  ✓ %-20s %s", tool.Name, r.formatToolInfo(toolState))
		} else if r.isToolActuallyInstalled(tool) {
			// Not in state but actually installed - show without version
			r.ui.Success("  ✓ %-20s (installed)", tool.Name)
		} else {
			// Not installed
			r.ui.Error("  ✗ %-20s (not installed)", tool.Name)
		}
	}
}

// showSetupStatus displays configured tasks
// What: Shows which tasks are configured, checking state first then running actual checks
// Why: Provides accurate status even for manually configured tasks
func (r *Reporter) showSetupStatus() {
	totalTasks := len(r.setupConfig.SetupTasks)
	configuredCount := 0

	// Count configured tasks (state + actual checks)
	for _, task := range r.setupConfig.SetupTasks {
		if r.state.Configured[task.Name] {
			configuredCount++
		} else if r.isTaskActuallyConfigured(task) {
			configuredCount++
		}
	}

	r.ui.Info("⚙️  Configuration Status (%d/%d complete):", configuredCount, totalTasks)

	for _, task := range r.setupConfig.SetupTasks {
		if r.state.Configured[task.Name] {
			// In state - configured by devsetup
			r.ui.Success("  ✓ %s", task.Name)
		} else if r.isTaskActuallyConfigured(task) {
			// Not in state but actually configured (manually or externally)
			r.ui.Success("  ✓ %s (verified)", task.Name)
		} else {
			// Not configured
			r.ui.Error("  ✗ %s (not configured)", task.Name)
		}
	}
}

// showOverallProgress displays overall completion percentage
// What: Shows overall progress based on actual verification, not just state
// Why: Provides accurate progress percentage
func (r *Reporter) showOverallProgress() {
	totalTools := len(r.toolsConfig.Tools)
	totalTasks := len(r.setupConfig.SetupTasks)
	total := totalTools + totalTasks

	// Count actual installed tools (state + verification)
	installedCount := 0
	for _, tool := range r.toolsConfig.Tools {
		if _, ok := r.state.Installed[tool.Name]; ok {
			installedCount++
		} else if r.isToolActuallyInstalled(tool) {
			installedCount++
		}
	}

	// Count actual configured tasks (state + verification)
	configuredCount := 0
	for _, task := range r.setupConfig.SetupTasks {
		if r.state.Configured[task.Name] {
			configuredCount++
		} else if r.isTaskActuallyConfigured(task) {
			configuredCount++
		}
	}

	completed := installedCount + configuredCount

	percentage := 0
	if total > 0 {
		percentage = (completed * 100) / total
	}

	r.ui.Info("📈 Overall Progress: %d%% complete (%d/%d tasks)", percentage, completed, total)

	if completed < total {
		r.ui.Info("")
		r.ui.Info("💡 Next steps:")
		if installedCount < totalTools {
			r.ui.Info("   • Run 'devsetup install' to install missing tools")
		}
		if configuredCount < totalTasks {
			r.ui.Info("   • Run 'devsetup setup' to configure remaining items")
		}
		r.ui.Info("   • Run 'devsetup verify' to check everything works")
	} else {
		r.ui.Info("")
		r.ui.Success("🎉 Environment fully configured!")
	}
}

// formatToolInfo formats tool state information
func (r *Reporter) formatToolInfo(toolState config.ToolState) string {
	version := toolState.Version
	if len(version) > 30 {
		version = version[:27] + "..."
	}
	return fmt.Sprintf("%-30s", version)
}

// expandPath expands ~ and environment variables in a path
// What: Converts ~/ to $HOME/ and expands $VAR and ${VAR} syntax
// Why: Config files use ~ but Go doesn't expand it
// Params: path - path that may contain ~ or env vars
// Returns: Expanded absolute path
func expandPath(path string) string {
	// Expand environment variables first
	path = os.ExpandEnv(path)

	// Expand tilde
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	} else if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			path = home
		}
	}

	return path
}

// isToolActuallyInstalled runs the tool's check command to verify it exists
// What: Executes the check command to see if tool is installed
// Why: Provides fallback verification when state file is missing/inaccurate
// Params: tool - Tool configuration with check command
// Returns: true if check command succeeds
func (r *Reporter) isToolActuallyInstalled(tool config.Tool) bool {
	if tool.Check == "" {
		return false
	}

	cmd := exec.Command("sh", "-c", tool.Check)
	return cmd.Run() == nil
}

// isTaskActuallyConfigured runs verification checks to see if task is configured
// What: Runs all verification checks defined for a setup task
// Why: Provides accurate status even when state file is missing/inaccurate
// Params: task - SetupTask with verification checks
// Returns: true if all verification checks pass
func (r *Reporter) isTaskActuallyConfigured(task config.SetupTask) bool {
	// If no verification checks, can't verify
	if len(task.Verify) == 0 {
		return false
	}

	// All checks must pass
	for _, check := range task.Verify {
		if !r.runVerifyCheck(check) {
			return false
		}
	}

	return true
}

// runVerifyCheck runs a single verification check
// What: Executes one verification check (command, env var, file exists, file contains)
// Why: Shared verification logic for setup tasks
// Params: check - VerifyCheck configuration
// Returns: true if check passes
func (r *Reporter) runVerifyCheck(check config.VerifyCheck) bool {
	if check.Command != "" {
		cmd := exec.Command("sh", "-c", check.Command)
		return cmd.Run() == nil
	}

	if check.EnvVar != "" {
		return os.Getenv(check.EnvVar) != ""
	}

	if check.FileExists != "" {
		path := expandPath(check.FileExists)
		_, err := os.Stat(path)
		return err == nil
	}

	if check.FileContains != nil {
		path := expandPath(check.FileContains.Path)
		content, err := os.ReadFile(path)
		if err != nil {
			return false
		}
		return strings.Contains(string(content), check.FileContains.Text)
	}

	// TODO: Implement TomlValue check

	return true
}
