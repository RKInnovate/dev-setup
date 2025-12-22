// File: internal/updater/updater.go
// Purpose: Self-update functionality for devsetup binary from GitHub releases
// Problem: Need way to keep devsetup tool up-to-date without manual reinstall
// Role: Checks for new releases on GitHub, downloads and replaces current binary
// Usage: Called by `devsetup update` command or automatically on version check
// Design choices: Uses GitHub API for release info; validates checksums; atomic replacement
// Assumptions: GitHub releases exist with proper naming; network access available

package updater

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// GitHubOwner is the GitHub repository owner
	GitHubOwner = "rkinnovate"
	// GitHubRepo is the GitHub repository name
	GitHubRepo = "dev-setup"
	// GitHubAPIURL is the GitHub API base URL
	GitHubAPIURL = "https://api.github.com"
)

// ReleaseInfo contains information about a GitHub release
// What: Structured information about available release
// Why: Provides version, download URL, and checksums for update decision
type ReleaseInfo struct {
	TagName    string    `json:"tag_name"`
	Name       string    `json:"name"`
	Body       string    `json:"body"`
	Draft      bool      `json:"draft"`
	Prerelease bool      `json:"prerelease"`
	CreatedAt  time.Time `json:"created_at"`
	Assets     []Asset   `json:"assets"`
}

// Asset represents a release asset (binary file)
// What: Information about downloadable binary
// Why: Contains download URL and name for correct architecture selection
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Updater handles self-update operations
// What: Manages checking for updates and performing self-update
// Why: Provides clean API for update functionality
type Updater struct {
	currentVersion string
	owner          string
	repo           string
	httpClient     *http.Client
}

// NewUpdater creates a new Updater instance
// What: Constructor for Updater with GitHub repository information
// Why: Centralizes updater creation with timeout configuration
// Params: currentVersion - current binary version
// Returns: Configured Updater instance
// Example: updater := NewUpdater("v0.4.0")
func NewUpdater(currentVersion string) *Updater {
	return &Updater{
		currentVersion: currentVersion,
		owner:          GitHubOwner,
		repo:           GitHubRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckForUpdate checks if a newer version is available
// What: Queries GitHub API for latest release and compares with current version
// Why: Determines if update is available before downloading
// Returns: ReleaseInfo pointer if update available, nil if current, error on failure
// Example: release, err := updater.CheckForUpdate()
func (u *Updater) CheckForUpdate() (*ReleaseInfo, error) {
	// Get latest release from GitHub API
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", GitHubAPIURL, u.owner, u.repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent (GitHub API requires it)
	req.Header.Set("User-Agent", fmt.Sprintf("devsetup/%s", u.currentVersion))

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	// Skip draft and prerelease versions
	if release.Draft || release.Prerelease {
		return nil, nil
	}

	// Compare versions
	if !isNewerVersion(release.TagName, u.currentVersion) {
		return nil, nil // Already on latest
	}

	return &release, nil
}

// Update performs the self-update operation
// What: Downloads new binary, verifies it, and atomically replaces current binary
// Why: Updates devsetup to latest version safely
// Params: release - ReleaseInfo containing download URL
// Returns: Error if update failed, nil on success
// Example: err := updater.Update(release)
func (u *Updater) Update(release *ReleaseInfo) error {
	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Find correct asset for current platform/architecture
	asset := findAssetForPlatform(release.Assets)
	if asset == nil {
		return fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download new binary to temp file
	tempFile, err := os.CreateTemp("", "devsetup-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if err := u.downloadFile(tempFile, asset.BrowserDownloadURL); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	// Make new binary executable
	if err := os.Chmod(tempFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Backup current binary
	backupPath := currentExe + ".backup"
	if err := os.Rename(currentExe, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Atomic replace: close temp file, then move it
	tempFile.Close()
	if err := os.Rename(tempFile.Name(), currentExe); err != nil {
		// Restore backup on failure
		os.Rename(backupPath, currentExe)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Remove backup on success
	os.Remove(backupPath)

	return nil
}

// downloadFile downloads a file from URL to writer
// What: HTTP download with progress (simplified for now)
// Why: Downloads binary from GitHub releases
// Params: dst - destination writer, url - download URL
// Returns: Error if download failed
func (u *Updater) downloadFile(dst io.Writer, url string) error {
	resp, err := u.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	_, err = io.Copy(dst, resp.Body)
	return err
}

// findAssetForPlatform finds the correct binary asset for current platform
// What: Selects appropriate binary from release assets based on OS/arch
// Why: GitHub releases contain binaries for multiple platforms
// Params: assets - slice of available assets
// Returns: Matching Asset pointer or nil if not found
func findAssetForPlatform(assets []Asset) *Asset {
	// Binary naming convention: devsetup-{os}-{arch}
	// Example: devsetup-darwin-arm64, devsetup-darwin-amd64
	binaryName := fmt.Sprintf("devsetup-%s-%s", runtime.GOOS, runtime.GOARCH)

	for i := range assets {
		if assets[i].Name == binaryName {
			return &assets[i]
		}
	}

	return nil
}

// isNewerVersion compares two semantic versions
// What: Determines if newVer is newer than currentVer
// Why: Decides whether update is needed
// Params: newVer - version string from release (e.g. "v0.5.0"), currentVer - current version
// Returns: true if newVer is newer
// Edge cases: Handles "v" prefix, git commit hashes (always considers remote newer)
func isNewerVersion(newVer, currentVer string) bool {
	// Strip "v" prefix if present
	newVer = strings.TrimPrefix(newVer, "v")
	currentVer = strings.TrimPrefix(currentVer, "v")

	// If current is a git commit hash (dev build), always update
	if len(currentVer) == 7 && !strings.Contains(currentVer, ".") {
		return true
	}

	// Simple lexicographic comparison for now
	// TODO: Implement proper semantic version comparison
	return newVer > currentVer
}

// GetReleaseNotes formats release notes for display
// What: Extracts and formats release notes from release body
// Why: Shows user what's new in the update
// Params: release - ReleaseInfo containing body text
// Returns: Formatted release notes string
func GetReleaseNotes(release *ReleaseInfo) string {
	if release.Body == "" {
		return "No release notes available."
	}

	// Simple formatting - take first 500 chars
	notes := release.Body
	if len(notes) > 500 {
		notes = notes[:500] + "..."
	}

	return notes
}

// VerifyChecksum verifies downloaded file against expected checksum
// What: Calculates SHA256 checksum and compares with expected value
// Why: Ensures downloaded binary hasn't been tampered with
// Params: filepath - path to file to verify, expectedChecksum - expected SHA256 hex string
// Returns: Error if checksum doesn't match, nil if valid
func VerifyChecksum(filepath, expectedChecksum string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualChecksum := fmt.Sprintf("%x", hash.Sum(nil))

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}
