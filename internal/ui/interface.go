// File: internal/ui/interface.go
// Purpose: Defines UI interface for terminal output abstraction
// Problem: Need consistent UI contract for installer, setup, verify commands
// Role: Interface definition that ProgressUI implements
// Usage: Accept UI interface in installer/setup/verify for testability
// Design choices: Interface allows mock implementations for testing
// Assumptions: All UI operations are synchronous

package ui

// UI defines the interface for user interface operations
// What: Contract for terminal output operations with progress tracking
// Why: Allows different UI implementations (real, mock for tests)
type UI interface {
	// Banner and headers
	PrintBanner()
	StartStage(name, estimatedTime string)

	// Task progress
	StartTask(taskName string)
	CompleteTask(taskName string)
	FailTask(taskName string, err error)

	// Messages
	Success(format string, args ...interface{})
	Error(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Info(format string, args ...interface{})

	// Progress indicators
	PrintProgress(current, total int, label string)
	PrintElapsedTime()
}

// Compile-time check that ProgressUI implements UI interface
var _ UI = (*ProgressUI)(nil)
