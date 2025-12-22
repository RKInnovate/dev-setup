// File: internal/ui/progress.go
// Purpose: Provides rich UI and progress indicators for terminal output
// Problem: Plain text output doesn't show installation progress clearly; developers want visual feedback
// Role: Handles all terminal output with colors, progress bars, spinners, and structured formatting
// Usage: Create ProgressUI instance, call StartStage/StartTask/Success/Error methods
// Design choices: Uses ANSI colors for compatibility; supports both interactive and non-interactive terminals
// Assumptions: Terminal supports ANSI escape codes (standard on macOS); UTF-8 encoding

package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Color codes for terminal output
const (
	// colorReset resets all attributes
	colorReset = "\033[0m"
	// colorBold makes text bold
	colorBold = "\033[1m"
	// colorDim makes text dimmed
	colorDim = "\033[2m"

	// Foreground colors
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"

	// Background colors
	bgGreen = "\033[42m"
	bgRed   = "\033[41m"
)

// ProgressUI provides methods for rich terminal output
// What: Manages all user-facing terminal output with colors and formatting
// Why: Provides clear visual feedback during long-running installation processes
type ProgressUI struct {
	writer     io.Writer
	mu         sync.Mutex
	isInteractive bool
	startTime  time.Time
}

// NewProgressUI creates a new ProgressUI instance
// What: Constructor for ProgressUI with stdout as default writer
// Why: Centralizes UI creation and configuration
// Returns: Configured ProgressUI instance
// Example: ui := NewProgressUI()
func NewProgressUI() *ProgressUI {
	return &ProgressUI{
		writer:     os.Stdout,
		isInteractive: isTerminal(os.Stdout),
		startTime:  time.Now(),
	}
}

// PrintBanner prints the devsetup welcome banner
// What: Displays ASCII art banner with tool name and version
// Why: Professional appearance and clear indication that tool is running
func (p *ProgressUI) PrintBanner() {
	p.mu.Lock()
	defer p.mu.Unlock()

	banner := `
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃                                                    ┃
┃   ██████╗ ███████╗██╗   ██╗      ███████╗███████╗  ┃
┃   ██╔══██╗██╔════╝██║   ██║      ██╔════╝██╔════╝  ┃
┃   ██║  ██║█████╗  ██║   ██║█████╗███████╗█████╗    ┃
┃   ██║  ██║██╔══╝  ╚██╗ ██╔╝╚════╝╚════██║██╔══╝    ┃
┃   ██████╔╝███████╗ ╚████╔╝       ███████║███████╗  ┃
┃   ╚═════╝ ╚══════╝  ╚═══╝        ╚══════╝╚══════╝  ┃
┃                                                    ┃
┃   Zero to Productive in 5 Minutes                  ┃
┃   github.com/rkinnovate/dev-setup                  ┃
┃                                                    ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
`
	fmt.Fprint(p.writer, colorCyan+banner+colorReset+"\n")
}

// StartStage indicates a new installation stage is beginning
// What: Prints stage header with name and estimated time
// Why: Shows progress through multi-stage installation
// Params: name - stage name (e.g. "Critical Path"), estimatedTime - human readable time (e.g. "5 minutes")
// Example: ui.StartStage("Stage 1: Critical Path", "5 minutes")
func (p *ProgressUI) StartStage(name, estimatedTime string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	header := fmt.Sprintf("\n╔════════════════════════════════════════════════════════╗\n"+
		"║ %s%-50s%s     ║\n"+
		"║ %sEstimated time: %-38s%s ║\n"+
		"╚════════════════════════════════════════════════════════╝\n",
		colorBold+colorCyan, name, colorReset,
		colorDim, estimatedTime, colorReset)

	fmt.Fprint(p.writer, header)
}

// StartTask indicates a task is starting
// What: Prints task name with spinner/indicator
// Why: Shows which specific operation is currently running
// Params: taskName - human readable task description
// Example: ui.StartTask("Installing Homebrew")
func (p *ProgressUI) StartTask(taskName string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fmt.Fprintf(p.writer, "  %s⚡%s %s...\n", colorYellow, colorReset, taskName)
}

// CompleteTask marks a task as successfully completed
// What: Prints green checkmark with task name
// Why: Visual confirmation of successful completion
// Params: taskName - task that completed
// Example: ui.CompleteTask("Installing Homebrew")
func (p *ProgressUI) CompleteTask(taskName string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fmt.Fprintf(p.writer, "  %s✓%s %s\n", colorGreen, colorReset, taskName)
}

// FailTask marks a task as failed
// What: Prints red X with task name and error
// Why: Clear indication of failure for debugging
// Params: taskName - task that failed, err - error that occurred
// Example: ui.FailTask("Installing Homebrew", err)
func (p *ProgressUI) FailTask(taskName string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fmt.Fprintf(p.writer, "  %s✗%s %s: %v\n", colorRed, colorReset, taskName, err)
}

// Success prints a success message in green
// What: Prints formatted success message with checkmark
// Why: Highlights successful operations
// Params: format - printf-style format string, args - format arguments
// Example: ui.Success("Installation complete!")
func (p *ProgressUI) Success(format string, args ...interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(p.writer, "%s%s%s\n", colorGreen, message, colorReset)
}

// Error prints an error message in red
// What: Prints formatted error message with X symbol
// Why: Highlights errors for immediate attention
// Params: format - printf-style format string, args - format arguments
// Example: ui.Error("Installation failed: %v", err)
func (p *ProgressUI) Error(format string, args ...interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(p.writer, "%s%s%s\n", colorRed, message, colorReset)
}

// Warning prints a warning message in yellow
// What: Prints formatted warning message with warning symbol
// Why: Highlights non-critical issues that need attention
// Params: format - printf-style format string, args - format arguments
// Example: ui.Warning("Optional tool not available: %s", tool)
func (p *ProgressUI) Warning(format string, args ...interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(p.writer, "%s%s%s\n", colorYellow, message, colorReset)
}

// Info prints an informational message in default color
// What: Prints formatted info message
// Why: Provides context and instructions to user
// Params: format - printf-style format string, args - format arguments
// Example: ui.Info("Run 'devsetup verify' to check installation")
func (p *ProgressUI) Info(format string, args ...interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(p.writer, "%s\n", message)
}

// PrintProgress prints a progress bar
// What: Displays visual progress bar with percentage
// Why: Shows completion progress for long operations
// Params: current - current progress value, total - total expected value, label - operation description
// Example: ui.PrintProgress(7, 10, "Installing packages")
func (p *ProgressUI) PrintProgress(current, total int, label string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	percentage := float64(current) / float64(total) * 100
	barWidth := 40
	filledWidth := int(float64(barWidth) * float64(current) / float64(total))

	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)

	fmt.Fprintf(p.writer, "\r  [%s] %3.0f%% %s", bar, percentage, label)

	if current == total {
		fmt.Fprint(p.writer, "\n")
	}
}

// PrintElapsedTime prints time elapsed since UI creation
// What: Shows total elapsed time for operation
// Why: Helps track performance and estimate future runs
// Example: ui.PrintElapsedTime()
func (p *ProgressUI) PrintElapsedTime() {
	p.mu.Lock()
	defer p.mu.Unlock()

	elapsed := time.Since(p.startTime)
	fmt.Fprintf(p.writer, "\n%s⏱  Total time: %v%s\n", colorDim, elapsed.Round(time.Second), colorReset)
}

// isTerminal checks if output is an interactive terminal
// What: Determines if stdout is connected to a terminal (not redirected)
// Why: Disables interactive features (colors, progress bars) when output is piped
// Params: w - writer to check (usually os.Stdout)
// Returns: true if interactive terminal, false if piped/redirected
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		fileInfo, err := f.Stat()
		if err != nil {
			return false
		}
		// Check if it's a character device (terminal)
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	}
	return false
}
