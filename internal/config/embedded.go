// File: internal/config/embedded.go
// Purpose: Handles embedded config files for standalone binary distribution
// Problem: Downloaded binary needs config files to work
// Role: Provides access to embedded filesystem set by main package
// Usage: Automatically used by LoadStageConfig as fallback
// Design choices: Global variable set by main.go init(); clean API
// Assumptions: Main package calls SetEmbeddedFS before using config functions

package config

import (
	"embed"
	"fmt"
)

// Global variable to hold the embedded filesystem
// Set by main package via SetEmbeddedFS()
var embeddedFS embed.FS

// SetEmbeddedFS sets the embedded filesystem for config loading
// What: Stores reference to embedded FS for use by loader functions
// Why: Allows main package to provide embedded configs to this package
// Params: fs - embedded filesystem containing config files
// Example: config.SetEmbeddedFS(embeddedConfigs)
func SetEmbeddedFS(fs embed.FS) {
	embeddedFS = fs
}

// readEmbeddedFile reads a file from the embedded filesystem
// What: Attempts to read config file from embedded FS
// Why: Allows binary to work without external config files
// Params: path - path to config file (e.g., "configs/stage1.yaml")
// Returns: File contents as bytes and error if not found
func readEmbeddedFile(path string) ([]byte, error) {
	// The configs package embeds *.yaml files directly
	// So we just need the filename, not the full path
	// Extract filename from path (e.g., "configs/stage1.yaml" -> "stage1.yaml")
	filename := path
	if len(path) > 8 && path[:8] == "configs/" {
		filename = path[8:]
	}

	data, err := embeddedFS.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("file not found in embedded configs: %w", err)
	}

	return data, nil
}
