// File: cmd/devsetup/main.go
// Purpose: CLI entry point for dev-setup tool - orchestrates developer environment setup
// Problem: Manual dev environment setup takes days; this tool reduces it to 30 minutes
// Role: Main command-line interface using Cobra for subcommands (install, setup, verify, status, update)
// Usage: Run `devsetup install` to install tools, `devsetup setup` to configure
// Design choices: Install/setup separation; idempotency; parallel execution; state tracking
// Assumptions: macOS host with internet access; Homebrew will be installed if missing

package main

import (
	"fmt"
	"os"

	"github.com/rkinnovate/dev-setup/configs"
	"github.com/rkinnovate/dev-setup/internal/config"
	"github.com/rkinnovate/dev-setup/internal/installer"
	"github.com/rkinnovate/dev-setup/internal/setup"
	"github.com/rkinnovate/dev-setup/internal/status"
	"github.com/rkinnovate/dev-setup/internal/ui"
	"github.com/rkinnovate/dev-setup/internal/updater"
	"github.com/rkinnovate/dev-setup/internal/verify"
	"github.com/spf13/cobra"
)

func init() {
	// Set the embedded filesystem in the config package
	config.SetEmbeddedFS(configs.ConfigFS)
}

// version is set during build via -ldflags
var version = "0.2.0-dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "devsetup",
	Short: "Setup macOS development environment in 30 minutes",
	Long: `devsetup: Zero to productive developer in 5 minutes

Complete development environment setup with:
- Idempotent tool installation (check before install)
- Post-install configuration (interactive where needed)
- Accurate verification (no false positives)
- Version-locked dependencies via git submodules
- Single source of truth (tools.yaml + setup.yaml)

Commands:
  install  Install all tools from tools.yaml
  setup    Configure installed tools (interactive)
  verify   Verify installation and configuration
  status   Show current environment status
  update   Update devsetup binary`,
	Version: version,
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install all tools",
	Long: `Install all tools defined in tools.yaml.

Features:
- Idempotent: Checks if tool exists before installing
- Parallel: Tools in same parallel_group install concurrently
- Dependencies: Respects depends_on relationships
- State tracking: Saves installation state to ~/.local/share/devsetup/state.json

After installation completes, run 'devsetup setup' to configure tools.`,
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Initialize UI
		progressUI := ui.NewProgressUI()
		progressUI.PrintBanner()

		// Load configurations
		toolsConfig, err := config.LoadToolsConfig("configs/tools.yaml")
		if err != nil {
			progressUI.Error("❌ Failed to load tools config: %v", err)
			os.Exit(1)
		}

		// Load state
		state, err := config.LoadState()
		if err != nil {
			progressUI.Error("❌ Failed to load state: %v", err)
			os.Exit(1)
		}

		// Create installer
		toolInstaller := installer.NewToolInstaller(toolsConfig, state, progressUI, dryRun, version)

		// Install all tools
		if err := toolInstaller.InstallAll(); err != nil {
			progressUI.Error("❌ Installation failed: %v", err)
			progressUI.Info("Run 'devsetup doctor' to diagnose issues")
			os.Exit(1)
		}

		progressUI.Info("Next step: Run 'devsetup setup' to configure tools")
	},
}

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure installed tools",
	Long: `Configure installed tools defined in setup.yaml.

Features:
- Interactive: Prompts for API keys and configuration
- Remote-first: Tries remote scripts, falls back to local submodules
- File operations: Edits .zshrc, starship.toml, etc.
- Verification: Checks configuration succeeded
- State tracking: Saves setup state

This command may prompt you for:
- API keys (Claude, Gemini)
- Git configuration (name, email)
- Shell preferences

Run 'devsetup setup --help' for options.`,
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Initialize UI
		progressUI := ui.NewProgressUI()

		// Load configurations
		setupConfig, err := config.LoadSetupConfig("configs/setup.yaml")
		if err != nil {
			progressUI.Error("❌ Failed to load setup config: %v", err)
			os.Exit(1)
		}

		// Load state
		state, err := config.LoadState()
		if err != nil {
			progressUI.Error("❌ Failed to load state: %v", err)
			os.Exit(1)
		}

		// Create setup executor
		setupExecutor := setup.NewSetupExecutor(setupConfig, state, progressUI, dryRun)

		// Execute all setup tasks
		if err := setupExecutor.SetupAll(); err != nil {
			progressUI.Error("❌ Setup failed: %v", err)
			os.Exit(1)
		}

		progressUI.Info("Next step: Run 'devsetup verify' to check everything works")
	},
}

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify environment matches configuration",
	Long: `Verify that installed tools and configuration match expectations.

Checks:
- Tool binaries exist and are accessible
- Configuration files have expected content
- Environment variables are set
- TOML values match expected values

This command provides accurate verification without false positives.

Exit codes:
  0 - All checks passed
  1 - One or more checks failed`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize UI
		progressUI := ui.NewProgressUI()

		// Load configurations
		toolsConfig, err := config.LoadToolsConfig("configs/tools.yaml")
		if err != nil {
			progressUI.Error("❌ Failed to load tools config: %v", err)
			os.Exit(1)
		}

		setupConfig, err := config.LoadSetupConfig("configs/setup.yaml")
		if err != nil {
			progressUI.Error("❌ Failed to load setup config: %v", err)
			os.Exit(1)
		}

		// Load state
		state, err := config.LoadState()
		if err != nil {
			progressUI.Error("❌ Failed to load state: %v", err)
			os.Exit(1)
		}

		// Create verifier
		verifier := verify.NewVerifier(toolsConfig, setupConfig, state, progressUI)

		// Verify all
		result, err := verifier.VerifyAll()
		if err != nil {
			progressUI.Info("")
			progressUI.Info("Summary:")
			progressUI.Info("  Tools: %d OK, %d failed", result.ToolsOK, result.ToolsFailed)
			progressUI.Info("  Setup: %d OK, %d failed", result.SetupOK, result.SetupFailed)
			os.Exit(1)
		}

		progressUI.Info("")
		progressUI.Success("Environment verified successfully!")
	},
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current environment status",
	Long: `Display current installation and configuration status.

Shows:
- Installed tools with versions and paths
- Configured tasks
- Overall completion percentage
- Next steps to complete setup

This command reads from state.json and provides accurate status reporting.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize UI
		progressUI := ui.NewProgressUI()

		// Load configurations
		toolsConfig, err := config.LoadToolsConfig("configs/tools.yaml")
		if err != nil {
			progressUI.Error("❌ Failed to load tools config: %v", err)
			os.Exit(1)
		}

		setupConfig, err := config.LoadSetupConfig("configs/setup.yaml")
		if err != nil {
			progressUI.Error("❌ Failed to load setup config: %v", err)
			os.Exit(1)
		}

		// Load state
		state, err := config.LoadState()
		if err != nil {
			progressUI.Error("❌ Failed to load state: %v", err)
			os.Exit(1)
		}

		// Create reporter
		reporter := status.NewReporter(toolsConfig, setupConfig, state, progressUI)

		// Show status
		reporter.ShowStatus()
	},
}

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update devsetup binary",
	Long: `Check for and install the latest version of devsetup.

This command:
- Checks GitHub releases for newer versions
- Downloads the appropriate binary for your architecture
- Verifies SHA256 checksum
- Atomically replaces current binary
- Creates backup of old version

Use --check to only check for updates without installing.`,
	Run: func(cmd *cobra.Command, args []string) {
		checkOnly, _ := cmd.Flags().GetBool("check")

		// Initialize UI
		progressUI := ui.NewProgressUI()

		// Create updater
		upd := updater.NewUpdater(version)

		if checkOnly {
			// Check for updates only
			release, err := upd.CheckForUpdate()
			if err != nil {
				progressUI.Error("❌ Failed to check for updates: %v", err)
				os.Exit(1)
			}

			if release != nil {
				progressUI.Info("🎉 New version available: %s", release.TagName)
				progressUI.Info("Run 'devsetup update' to install")
			} else {
				progressUI.Success("✅ You're running the latest version (%s)", version)
			}
			return
		}

		// Check for updates first
		release, err := upd.CheckForUpdate()
		if err != nil {
			progressUI.Error("❌ Failed to check for updates: %v", err)
			os.Exit(1)
		}

		if release == nil {
			progressUI.Success("✅ You're already running the latest version (%s)", version)
			return
		}

		progressUI.Info("📦 Updating to version %s...", release.TagName)

		// Perform update
		if err := upd.Update(release); err != nil {
			progressUI.Error("❌ Update failed: %v", err)
			os.Exit(1)
		}

		progressUI.Success("✅ Update complete!")
		progressUI.Info("Please restart your terminal or run 'devsetup --version' to verify")
	},
}

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics",
	Long: `Run diagnostic checks to identify environment issues.

Checks:
- Homebrew installation and health
- Required tools accessibility
- Configuration file validity
- State file integrity
- Common path issues

This command helps troubleshoot installation problems.`,
	Run: func(cmd *cobra.Command, args []string) {
		progressUI := ui.NewProgressUI()
		progressUI.Info("🔧 Running diagnostics...")
		progressUI.Info("")
		progressUI.Warning("⚠️  Doctor command not yet fully implemented")
		progressUI.Info("For now, try:")
		progressUI.Info("  • devsetup verify - Check installation status")
		progressUI.Info("  • devsetup status - Show what's installed")
		progressUI.Info("  • brew doctor - Check Homebrew health")
	},
}

func main() {
	// Add flags
	installCmd.Flags().Bool("dry-run", false, "Show what would be installed without installing")
	setupCmd.Flags().Bool("dry-run", false, "Show what would be configured without configuring")
	updateCmd.Flags().Bool("check", false, "Check for updates without installing")

	// Add commands
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(doctorCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
