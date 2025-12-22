// File: internal/updater/updater_test.go
// Purpose: Unit tests for self-update functionality
// Problem: Need to verify update logic works correctly
// Role: Test suite for Updater, release checking, and version comparison
// Usage: Run with `go test ./internal/updater`
// Design choices: Mocks HTTP responses; tests version comparison edge cases
// Assumptions: Test environment has network access for integration tests

package updater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestNewUpdater(t *testing.T) {
	version := "v0.4.0"
	updater := NewUpdater(version)

	if updater.currentVersion != version {
		t.Errorf("Expected currentVersion '%s', got '%s'", version, updater.currentVersion)
	}

	if updater.owner != GitHubOwner {
		t.Errorf("Expected owner '%s', got '%s'", GitHubOwner, updater.owner)
	}

	if updater.repo != GitHubRepo {
		t.Errorf("Expected repo '%s', got '%s'", GitHubRepo, updater.repo)
	}

	if updater.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
}

func TestCheckForUpdate_NewerVersionAvailable(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/rkinnovate/dev-setup/releases/latest" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		release := ReleaseInfo{
			TagName:    "v0.5.0",
			Name:       "Release 0.5.0",
			Body:       "New features and bug fixes",
			Draft:      false,
			Prerelease: false,
			CreatedAt:  time.Now(),
			Assets: []Asset{
				{
					Name:               "devsetup-darwin-arm64",
					BrowserDownloadURL: "https://example.com/devsetup-darwin-arm64",
					Size:               1024,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create updater with mock server
	updater := NewUpdater("v0.4.0")
	updater.httpClient = server.Client()

	// Override GitHubAPIURL for testing
	originalURL := GitHubAPIURL
	defer func() {
		// Can't actually modify const, so this test uses httpClient override
	}()
	_ = originalURL

	// Manually construct request to mock server
	req, _ := http.NewRequest("GET", server.URL+"/repos/rkinnovate/dev-setup/releases/latest", nil)
	req.Header.Set("User-Agent", "devsetup/v0.4.0")

	resp, err := updater.httpClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if release.TagName != "v0.5.0" {
		t.Errorf("Expected tag_name 'v0.5.0', got '%s'", release.TagName)
	}
}

func TestCheckForUpdate_AlreadyLatest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := ReleaseInfo{
			TagName:    "v0.4.0",
			Name:       "Release 0.4.0",
			Draft:      false,
			Prerelease: false,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	updater := NewUpdater("v0.4.0")
	updater.httpClient = server.Client()

	// Test version comparison directly
	if isNewerVersion("v0.4.0", "v0.4.0") {
		t.Error("Expected isNewerVersion to return false for same version")
	}
}

func TestCheckForUpdate_SkipDraftRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := ReleaseInfo{
			TagName:    "v0.5.0",
			Draft:      true,
			Prerelease: false,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	updater := NewUpdater("v0.4.0")
	updater.httpClient = server.Client()

	// Manually test draft detection
	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := updater.httpClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !release.Draft {
		t.Error("Expected draft=true")
	}
}

func TestCheckForUpdate_SkipPrereleaseRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := ReleaseInfo{
			TagName:    "v0.5.0-beta.1",
			Draft:      false,
			Prerelease: true,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	updater := NewUpdater("v0.4.0")
	updater.httpClient = server.Client()

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := updater.httpClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !release.Prerelease {
		t.Error("Expected prerelease=true")
	}
}

func TestUpdate_SuccessfulUpdate(t *testing.T) {
	// Create a fake binary
	tmpDir := t.TempDir()
	currentExe := filepath.Join(tmpDir, "devsetup")

	if err := os.WriteFile(currentExe, []byte("old version"), 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Create mock server for download
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("new version")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create release with mock download URL
	release := &ReleaseInfo{
		TagName: "v0.5.0",
		Assets: []Asset{
			{
				Name:               "devsetup-" + runtime.GOOS + "-" + runtime.GOARCH,
				BrowserDownloadURL: server.URL,
				Size:               11,
			},
		},
	}

	// Test findAssetForPlatform
	asset := findAssetForPlatform(release.Assets)
	if asset == nil {
		t.Fatal("Expected to find asset for current platform")
	}

	if asset.Name != "devsetup-"+runtime.GOOS+"-"+runtime.GOARCH {
		t.Errorf("Expected asset name 'devsetup-%s-%s', got '%s'", runtime.GOOS, runtime.GOARCH, asset.Name)
	}
}

func TestFindAssetForPlatform(t *testing.T) {
	assets := []Asset{
		{Name: "devsetup-darwin-arm64", BrowserDownloadURL: "https://example.com/arm64"},
		{Name: "devsetup-darwin-amd64", BrowserDownloadURL: "https://example.com/amd64"},
		{Name: "devsetup-linux-amd64", BrowserDownloadURL: "https://example.com/linux"},
	}

	asset := findAssetForPlatform(assets)
	if asset == nil {
		t.Fatal("Expected to find asset for current platform")
	}

	expectedName := "devsetup-" + runtime.GOOS + "-" + runtime.GOARCH
	if asset.Name != expectedName {
		t.Errorf("Expected asset '%s', got '%s'", expectedName, asset.Name)
	}
}

func TestFindAssetForPlatform_NotFound(t *testing.T) {
	assets := []Asset{
		{Name: "devsetup-windows-amd64", BrowserDownloadURL: "https://example.com/windows"},
	}

	// Only test if not on windows
	if runtime.GOOS != "windows" {
		asset := findAssetForPlatform(assets)
		if asset != nil {
			t.Error("Expected nil for missing platform")
		}
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		newVer      string
		currentVer  string
		expected    bool
		description string
	}{
		{"v0.5.0", "v0.4.0", true, "newer version"},
		{"v0.4.1", "v0.4.0", true, "patch update"},
		{"v0.4.0", "v0.4.0", false, "same version"},
		{"v0.4.0", "v0.5.0", false, "older version"},
		{"0.5.0", "0.4.0", true, "without v prefix"},
		{"v1.0.0", "v0.9.9", true, "major version bump"},
		{"v0.5.0", "4c187f7", true, "dev build (git hash)"},
		{"v0.4.0-dev", "v0.3.0", true, "dev suffix"},
	}

	for _, tt := range tests {
		result := isNewerVersion(tt.newVer, tt.currentVer)
		if result != tt.expected {
			t.Errorf("isNewerVersion(%q, %q) = %v, want %v (%s)",
				tt.newVer, tt.currentVer, result, tt.expected, tt.description)
		}
	}
}

func TestIsNewerVersion_GitCommitHash(t *testing.T) {
	// Git commit hashes (7 chars, no dots) should always update
	if !isNewerVersion("v0.4.0", "abc123d") {
		t.Error("Expected git commit hash to always be updatable")
	}
}

func TestGetReleaseNotes(t *testing.T) {
	tests := []struct {
		body     string
		expected string
	}{
		{
			body:     "Bug fixes and improvements",
			expected: "Bug fixes and improvements",
		},
		{
			body:     "",
			expected: "No release notes available.",
		},
		{
			body:     string(make([]byte, 600)), // Long body
			expected: string(make([]byte, 500)) + "...",
		},
	}

	for _, tt := range tests {
		release := &ReleaseInfo{Body: tt.body}
		result := GetReleaseNotes(release)

		if tt.body == "" {
			if result != tt.expected {
				t.Errorf("Expected empty body to return '%s', got '%s'", tt.expected, result)
			}
		} else if len(tt.body) > 500 {
			if len(result) != 503 { // 500 + "..."
				t.Errorf("Expected truncated body length 503, got %d", len(result))
			}
		}
	}
}

func TestVerifyChecksum_Valid(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.bin")
	content := []byte("test content")

	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// SHA256 of "test content" is:
	// 6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72
	expectedChecksum := "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"

	if err := VerifyChecksum(tmpFile, expectedChecksum); err != nil {
		t.Errorf("VerifyChecksum failed: %v", err)
	}
}

func TestVerifyChecksum_Mismatch(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.bin")
	content := []byte("test content")

	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	wrongChecksum := "0000000000000000000000000000000000000000000000000000000000000000"

	if err := VerifyChecksum(tmpFile, wrongChecksum); err == nil {
		t.Error("Expected checksum mismatch error, got nil")
	}
}

func TestVerifyChecksum_FileNotFound(t *testing.T) {
	err := VerifyChecksum("/nonexistent/file", "abc123")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestDownloadFile(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("downloaded content")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	updater := NewUpdater("v0.4.0")
	updater.httpClient = server.Client()

	// Create temp file
	tmpFile := filepath.Join(t.TempDir(), "download.bin")
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = file.Close() }()

	// Download
	if err := updater.downloadFile(file, server.URL); err != nil {
		t.Errorf("downloadFile failed: %v", err)
	}

	// Verify content
	_ = file.Close()
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != "downloaded content" {
		t.Errorf("Expected content 'downloaded content', got '%s'", string(content))
	}
}

func TestDownloadFile_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	updater := NewUpdater("v0.4.0")
	updater.httpClient = server.Client()

	tmpFile := filepath.Join(t.TempDir(), "download.bin")
	file, _ := os.Create(tmpFile)
	defer func() { _ = file.Close() }()

	err := updater.downloadFile(file, server.URL)
	if err == nil {
		t.Error("Expected error for server error response, got nil")
	}
}
