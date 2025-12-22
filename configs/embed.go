// File: configs/embed.go
// Purpose: Embeds config files into binary for standalone distribution
// Problem: Binary needs config files but they're not on user's system
// Role: Provides embedded filesystem with all YAML config files
// Usage: Import configs package and use ConfigFS
// Design choices: Located in configs package to satisfy embed directory constraints
// Assumptions: YAML files exist in this directory at build time

package configs

import "embed"

// ConfigFS contains all embedded YAML config files
//
//go:embed *.yaml
var ConfigFS embed.FS
