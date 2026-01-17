// File: internal/verify/verifier.go
// Purpose: Verification of installed tools and configured tasks
// Problem: Need accurate verification without false positives
// Role: Checks actual tool existence, versions, and configuration
// Usage: Create Verifier, call VerifyAll() to check everything
// Design choices: Real checks via shell commands; state comparison
// Assumptions: Tools and config files are in expected locations

package verify

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rkinnovate/dev-setup/internal/config"
	"github.com/rkinnovate/dev-setup/internal/ui"
)

// Verifier checks tool installation and configuration status
type Verifier struct {
	toolsConfig *config.ToolsConfig
	setupConfig *config.SetupConfig
	state       *config.State
	ui          ui.UI
}

// VerifyResult contains verification results
type VerifyResult struct {
	ToolsOK     int
	ToolsFailed int
	SetupOK     int
	SetupFailed int
	Errors      []string
}

// NewVerifier creates a new verifier
func NewVerifier(toolsConfig *config.ToolsConfig, setupConfig *config.SetupConfig, state *config.State, ui ui.UI) *Verifier {
	return &Verifier{
		toolsConfig: toolsConfig,
		setupConfig: setupConfig,
		state:       state,
		ui:          ui,
	}
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

// VerifyAll verifies all tools and setup tasks
func (v *Verifier) VerifyAll() (*VerifyResult, error) {
	v.ui.Info("🔍 Verifying environment...")
	v.ui.Info("")

	result := &VerifyResult{}

	// Verify tools
	v.ui.Info("📦 Checking installed tools...")
	for _, tool := range v.toolsConfig.Tools {
		if v.verifyTool(tool) {
			result.ToolsOK++
			v.ui.Success("  ✓ %s", tool.Name)
		} else {
			result.ToolsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("Tool not installed: %s", tool.Name))
			v.ui.Error("  ✗ %s (not installed)", tool.Name)
		}
	}

	v.ui.Info("")
	v.ui.Info("⚙️  Checking configured tasks...")
	for _, task := range v.setupConfig.SetupTasks {
		if v.verifySetupTask(task) {
			result.SetupOK++
			v.ui.Success("  ✓ %s", task.Name)
		} else {
			result.SetupFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("Task not configured: %s", task.Name))
			v.ui.Error("  ✗ %s (not configured)", task.Name)
		}
	}

	v.ui.Info("")

	// Summary
	total := result.ToolsOK + result.ToolsFailed + result.SetupOK + result.SetupFailed
	passed := result.ToolsOK + result.SetupOK

	if len(result.Errors) == 0 {
		v.ui.Success("✅ Verification PASSED (%d/%d checks)", passed, total)
		return result, nil
	}

	v.ui.Error("❌ Verification FAILED (%d/%d checks)", passed, total)
	v.ui.Info("")
	v.ui.Info("Run 'devsetup install' or 'devsetup setup' to fix issues")

	return result, fmt.Errorf("verification failed with %d errors", len(result.Errors))
}

// verifyTool checks if a tool is installed
func (v *Verifier) verifyTool(tool config.Tool) bool {
	if tool.Check == "" {
		return true // No check specified
	}

	cmd := exec.Command("sh", "-c", tool.Check)
	return cmd.Run() == nil
}

// verifySetupTask checks if a setup task is configured
func (v *Verifier) verifySetupTask(task config.SetupTask) bool {
	if len(task.Verify) == 0 {
		// No verification specified, check state
		return config.IsTaskConfigured(v.state, task.Name)
	}

	// Run all verification checks
	for _, check := range task.Verify {
		if !v.runVerifyCheck(check) {
			return false
		}
	}

	return true
}

// runVerifyCheck runs a single verification check
func (v *Verifier) runVerifyCheck(check config.VerifyCheck) bool {
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
