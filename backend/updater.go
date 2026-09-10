// Package backend provides in-app update checking, release notes parsing, and post-update notifications.
package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ReleaseAsset represents a downloadable release file asset.
type ReleaseAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"downloadUrl"`
	Size        int64  `json:"size"`
}

// ReleaseInfo represents structured release details from GitHub Releases.
type ReleaseInfo struct {
	Version      string `json:"version"`
	TagName      string `json:"tagName"`
	Name         string `json:"name"`
	Body         string `json:"body"`
	PublishedAt  string `json:"publishedAt"`
	HTMLURL      string `json:"htmlUrl"`
	InstallerURL string `json:"installerUrl"`
	PortableURL  string `json:"portableUrl"`
	IsPrerelease bool   `json:"isPrerelease"`
}

// UpdateCheckResult represents the outcome of an update verification query.
type UpdateCheckResult struct {
	Status          string      `json:"status"` // "up-to-date", "update-available", "dev-mode", "error"
	UpdateAvailable bool        `json:"updateAvailable"`
	CurrentVersion  string      `json:"currentVersion"`
	LatestVersion   string      `json:"latestVersion"`
	Release         ReleaseInfo `json:"release"`
	CheckedAt       string      `json:"checkedAt"`
	Message         string      `json:"message"`
}

type cachedUpdateData struct {
	Timestamp int64             `json:"timestamp"`
	Result    UpdateCheckResult `json:"result"`
}

type versionStateData struct {
	LastVersion string `json:"lastVersion"`
}

type ghReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type ghReleaseResponse struct {
	TagName     string           `json:"tag_name"`
	Name        string           `json:"name"`
	Body        string           `json:"body"`
	HTMLURL     string           `json:"html_url"`
	Prerelease  bool             `json:"prerelease"`
	Draft       bool             `json:"draft"`
	PublishedAt string           `json:"published_at"`
	Assets      []ghReleaseAsset `json:"assets"`
}

// CheckPostUpdateInstalled checks if the app was recently upgraded to a new version and updates state.
func (a *App) CheckPostUpdateInstalled() string {
	stateDir := filepath.Join(GetAppDataDir(), "cache")
	_ = os.MkdirAll(stateDir, 0755)
	stateFile := filepath.Join(stateDir, "version_state.json")

	current := AppVersion
	if current == "" || current == "dev" {
		current = "0.1.0-dev"
	}

	var state versionStateData
	if data, err := os.ReadFile(stateFile); err == nil {
		_ = json.Unmarshal(data, &state)
	}

	// First run on a fresh install: record current version without sending an "upgraded" message
	if state.LastVersion == "" {
		state.LastVersion = current
		if stateBytes, err := json.Marshal(state); err == nil {
			_ = os.WriteFile(stateFile, stateBytes, 0600)
		}
		return ""
	}

	// If the version has changed (and isn't identical to the last recorded version)
	if state.LastVersion != current {
		oldVersion := state.LastVersion
		state.LastVersion = current
		if stateBytes, err := json.Marshal(state); err == nil {
			_ = os.WriteFile(stateFile, stateBytes, 0600)
		}

		// Only notify if moving from an older version to a newer version
		if isNewerVersion(oldVersion, current) {
			return current
		}
	}

	return ""
}

// resolveUpdaterToken resolves an authentication token for GitHub API requests.
// Resolution order:
// -UpdaterToken injected during compilation via -ldflags (from GitHub Secrets).
// -VINTA_GITHUB_TOKEN or GITHUB_TOKEN environment variables.
// -GitHub CLI authentication token via 'gh auth token' during local development.
func resolveUpdaterToken() string {
	if strings.TrimSpace(UpdaterToken) != "" {
		return strings.TrimSpace(UpdaterToken)
	}
	if tok := strings.TrimSpace(os.Getenv("VINTA_GITHUB_TOKEN")); tok != "" {
		return tok
	}
	if tok := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); tok != "" {
		return tok
	}
	if BuildType == "Development" {
		cmd := exec.Command("gh", "auth", "token")
		if out, err := cmd.Output(); err == nil {
			trimmed := strings.TrimSpace(string(out))
			if trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

// CheckForUpdates queries GitHub Releases API to check for newer versions.
func (a *App) CheckForUpdates(force bool) (*UpdateCheckResult, error) {
	currentVersion := AppVersion
	if currentVersion == "" || currentVersion == "dev" {
		currentVersion = "0.1.0-dev"
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	// 1. In Development Builds: Disable update checking
	if BuildType == "Development" && !force {
		return &UpdateCheckResult{
			Status:          "dev-mode",
			UpdateAvailable: false,
			CurrentVersion:  currentVersion,
			LatestVersion:   currentVersion,
			CheckedAt:       now,
			Message:         "Update checking is disabled for development builds.",
		}, nil
	}

	cacheFile := filepath.Join(GetAppDataDir(), "cache", "update_check.json")

	// 2. Check local 1-hour cache unless forced
	if !force {
		if data, err := os.ReadFile(cacheFile); err == nil {
			var cached cachedUpdateData
			if err := json.Unmarshal(data, &cached); err == nil {
				// Cache valid for 1 hour (3600 seconds)
				if time.Now().Unix()-cached.Timestamp < 3600 {
					return &cached.Result, nil
				}
			}
		}
	}

	// 3. Query GitHub Releases API (queries release list to accommodate pre-release channels)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/kuraigu/Vinta/releases?per_page=5", nil)
	if err != nil {
		return &UpdateCheckResult{
			Status:          "error",
			UpdateAvailable: false,
			CurrentVersion:  currentVersion,
			CheckedAt:       now,
			Message:         fmt.Sprintf("Failed to construct update request: %v", err),
		}, nil
	}

	req.Header.Set("User-Agent", "Vinta-Update-Checker/1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	if token := resolveUpdaterToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return &UpdateCheckResult{
			Status:          "error",
			UpdateAvailable: false,
			CurrentVersion:  currentVersion,
			CheckedAt:       now,
			Message:         "Network error while checking for updates. Please check your internet connection.",
		}, nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		result := &UpdateCheckResult{
			Status:          "up-to-date",
			UpdateAvailable: false,
			CurrentVersion:  currentVersion,
			LatestVersion:   currentVersion,
			CheckedAt:       now,
			Message:         "Vinta is currently up to date.",
		}
		a.saveUpdateCache(cacheFile, result)
		return result, nil
	}

	if resp.StatusCode != http.StatusOK {
		return &UpdateCheckResult{
			Status:          "error",
			UpdateAvailable: false,
			CurrentVersion:  currentVersion,
			CheckedAt:       now,
			Message:         fmt.Sprintf("GitHub API returned status code %d", resp.StatusCode),
		}, nil
	}

	var releases []ghReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return &UpdateCheckResult{
			Status:          "error",
			UpdateAvailable: false,
			CurrentVersion:  currentVersion,
			CheckedAt:       now,
			Message:         "Failed to parse release information.",
		}, nil
	}

	var ghRelease *ghReleaseResponse
	for i := range releases {
		if !releases[i].Draft {
			ghRelease = &releases[i]
			break
		}
	}

	if ghRelease == nil {
		result := &UpdateCheckResult{
			Status:          "up-to-date",
			UpdateAvailable: false,
			CurrentVersion:  currentVersion,
			LatestVersion:   currentVersion,
			CheckedAt:       now,
			Message:         "Vinta is currently up to date.",
		}
		a.saveUpdateCache(cacheFile, result)
		return result, nil
	}

	var installerURL, portableURL string
	for _, asset := range ghRelease.Assets {
		lowerName := strings.ToLower(asset.Name)
		if strings.HasSuffix(lowerName, ".exe") && (strings.Contains(lowerName, "setup") || strings.Contains(lowerName, "installer")) {
			installerURL = asset.BrowserDownloadURL
		} else if strings.HasSuffix(lowerName, ".zip") && strings.Contains(lowerName, "portable") {
			portableURL = asset.BrowserDownloadURL
		}
	}

	cleanCurrent := strings.TrimPrefix(currentVersion, "v")
	cleanLatest := strings.TrimPrefix(ghRelease.TagName, "v")
	hasUpdate := isNewerVersion(cleanCurrent, cleanLatest)

	status := "up-to-date"
	msg := "Vinta is currently up to date."
	if hasUpdate {
		status = "update-available"
		msg = fmt.Sprintf("A new version (%s) of Vinta is available!", ghRelease.TagName)
	}

	result := &UpdateCheckResult{
		Status:          status,
		UpdateAvailable: hasUpdate,
		CurrentVersion:  currentVersion,
		LatestVersion:   ghRelease.TagName,
		Release: ReleaseInfo{
			Version:      ghRelease.TagName,
			TagName:      ghRelease.TagName,
			Name:         ghRelease.Name,
			Body:         ghRelease.Body,
			PublishedAt:  ghRelease.PublishedAt,
			HTMLURL:      ghRelease.HTMLURL,
			InstallerURL: installerURL,
			PortableURL:  portableURL,
			IsPrerelease: ghRelease.Prerelease,
		},
		CheckedAt: now,
		Message:   msg,
	}

	a.saveUpdateCache(cacheFile, result)
	return result, nil
}

func (a *App) saveUpdateCache(cacheFile string, result *UpdateCheckResult) {
	_ = os.MkdirAll(filepath.Dir(cacheFile), 0755)
	cached := cachedUpdateData{
		Timestamp: time.Now().Unix(),
		Result:    *result,
	}
	if data, err := json.Marshal(cached); err == nil {
		_ = os.WriteFile(cacheFile, data, 0600)
	}
}

// StartBackgroundUpdateChecker launches a background ticker that checks for updates periodically.
func (a *App) StartBackgroundUpdateChecker() {
	go func() {
		// Wait 10 seconds after app startup before first check
		time.Sleep(10 * time.Second)

		// Perform initial check
		if res, err := a.CheckForUpdates(false); err == nil && res.UpdateAvailable {
			runtime.EventsEmit(a.ctx, "update:available", res)
		}

		// Check every 6 hours
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			if res, err := a.CheckForUpdates(false); err == nil && res.UpdateAvailable {
				runtime.EventsEmit(a.ctx, "update:available", res)
			}
		}
	}()
}

// isNewerVersion returns true if candidate is newer than current.
func isNewerVersion(current, candidate string) bool {
	currMajor, currMinor, currPatch, currPre := parseSemVer(current)
	candMajor, candMinor, candPatch, candPre := parseSemVer(candidate)

	if candMajor != currMajor {
		return candMajor > currMajor
	}
	if candMinor != currMinor {
		return candMinor > currMinor
	}
	if candPatch != currPatch {
		return candPatch > currPatch
	}

	// If base versions are equal:
	// A release version (no prerelease suffix) is newer than a prerelease version (e.g. 0.1.0 > 0.1.0-alpha)
	if currPre != "" && candPre == "" {
		return true
	}
	if currPre == "" && candPre != "" {
		return false
	}
	if currPre != "" && candPre != "" {
		return candPre > currPre
	}

	return false
}

func parseSemVer(v string) (int, int, int, string) {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, "-", 2)
	base := parts[0]
	pre := ""
	if len(parts) > 1 {
		pre = parts[1]
	}

	nums := strings.Split(base, ".")
	major, minor, patch := 0, 0, 0
	if len(nums) > 0 {
		major, _ = strconv.Atoi(nums[0])
	}
	if len(nums) > 1 {
		minor, _ = strconv.Atoi(nums[1])
	}
	if len(nums) > 2 {
		patch, _ = strconv.Atoi(nums[2])
	}

	// Also extract numbers from pre (e.g., dev, alpha.1)
	cleanPre := regexp.MustCompile(`[^a-zA-Z0-9.]`).ReplaceAllString(pre, "")
	return major, minor, patch, cleanPre
}

// OpenURL opens the specified URL in the user's default web browser.
func (a *App) OpenURL(targetURL string) error {
	if targetURL == "" {
		return nil
	}
	return browser.OpenURL(targetURL)
}
