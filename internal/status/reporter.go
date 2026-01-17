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
func (r *Reporter) showToolsStatus() {
	totalTools := len(r.toolsConfig.Tools)
	installedCount := len(r.state.Installed)

	r.ui.Info("📦 Installed Tools (%d/%d complete):", installedCount, totalTools)

	for _, tool := range r.toolsConfig.Tools {
		if toolState, ok := r.state.Installed[tool.Name]; ok {
			r.ui.Success("  ✓ %-20s %s", tool.Name, r.formatToolInfo(toolState))
		} else {
			r.ui.Error("  ✗ %-20s (not installed)", tool.Name)
		}
	}
}

// showSetupStatus displays configured tasks
func (r *Reporter) showSetupStatus() {
	totalTasks := len(r.setupConfig.SetupTasks)
	configuredCount := 0
	for _, configured := range r.state.Configured {
		if configured {
			configuredCount++
		}
	}

	r.ui.Info("⚙️  Configuration Status (%d/%d complete):", configuredCount, totalTasks)

	for _, task := range r.setupConfig.SetupTasks {
		if r.state.Configured[task.Name] {
			r.ui.Success("  ✓ %s", task.Name)
		} else {
			r.ui.Error("  ✗ %s (not configured)", task.Name)
		}
	}
}

// showOverallProgress displays overall completion percentage
func (r *Reporter) showOverallProgress() {
	totalTools := len(r.toolsConfig.Tools)
	totalTasks := len(r.setupConfig.SetupTasks)
	total := totalTools + totalTasks

	installedCount := len(r.state.Installed)
	configuredCount := 0
	for _, configured := range r.state.Configured {
		if configured {
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
