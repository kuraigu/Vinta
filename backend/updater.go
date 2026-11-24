// Package backend provides in-app update checking, resumable downloading, release notes handling, and post-update notifications.
package backend

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/browser"
	"github.com/wailsapp/wails/v3/pkg/application"
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
	ChecksumsURL string `json:"checksumsUrl"`
	IsPrerelease bool   `json:"isPrerelease"`
	IsPortable   bool   `json:"isPortable"`
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
	IsPortable      bool        `json:"isPortable"`
	IsCustomRepo    bool        `json:"isCustomRepo"`
}

// DownloadProgress represents the real-time progress of an in-flight update download.
type DownloadProgress struct {
	Status           string  `json:"status"` // "downloading", "paused", "verifying", "ready", "error", "idle"
	DownloadedBytes  int64   `json:"downloadedBytes"`
	TotalBytes       int64   `json:"totalBytes"`
	Percentage       float64 `json:"percentage"`
	SpeedBytesPerSec int64   `json:"speedBytesPerSec"`
	Error            string  `json:"error,omitempty"`
}

type cachedUpdateData struct {
	Timestamp int64             `json:"timestamp"`
	Result    UpdateCheckResult `json:"result"`
}

type versionStateData struct {
	LastVersion string `json:"lastVersion"`
}

type ghReleaseAsset struct {
	ID                 int64  `json:"id"`
	URL                string `json:"url"`
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

	// First run on a fresh install: record current version without sending an upgraded message
	if state.LastVersion == "" {
		state.LastVersion = current
		if stateBytes, err := json.Marshal(state); err == nil {
			_ = os.WriteFile(stateFile, stateBytes, 0600)
		}
		return ""
	}

	// If the version has changed and is not identical to the last recorded version
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
// 1. UpdaterToken injected during compilation via -ldflags (from GitHub Secrets).
// 2. VINTA_UPDATER_TOKEN, UPDATER_TOKEN, VINTA_GITHUB_TOKEN, or GITHUB_TOKEN environment variables.
// 3. User configuration file settings.json.
// resolveUpdaterToken resolves an authentication token for GitHub API requests.
// Resolution order:
// 1. UpdaterToken injected during compilation via -ldflags (from GitHub Secrets).
// 2. VINTA_UPDATER_TOKEN, UPDATER_TOKEN, VINTA_GITHUB_TOKEN, or GITHUB_TOKEN environment variables.
// 3. Local .env file (for local development).
// 4. User configuration file settings.json.
func resolveUpdaterToken() string {
	if t := strings.TrimSpace(UpdaterToken); t != "" {
		return t
	}

	envKeys := []string{"VINTA_UPDATER_TOKEN", "UPDATER_TOKEN", "VINTA_GITHUB_TOKEN", "GITHUB_TOKEN"}
	for _, env := range envKeys {
		if tok := strings.TrimSpace(os.Getenv(env)); tok != "" {
			return tok
		}
	}

	// Check local .env file across common working directory depths (root, build/, build/bin/)
	candidatePaths := []string{".env", "../.env", "../../.env", "../../../.env"}
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidatePaths = append(candidatePaths,
			filepath.Join(exeDir, ".env"),
			filepath.Join(exeDir, "..", ".env"),
			filepath.Join(exeDir, "..", "..", ".env"),
		)
	}

	for _, envPath := range candidatePaths {
		data, err := os.ReadFile(envPath)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			line = strings.TrimPrefix(line, "export ")
			for _, key := range envKeys {
				prefix := key + "="
				if strings.HasPrefix(line, prefix) {
					val := strings.TrimSpace(strings.TrimPrefix(line, prefix))
					val = strings.Trim(val, `"'`)
					if val != "" {
						return val
					}
				}
			}
		}
	}

	settingsPath := filepath.Join(GetAppDataDir(), "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		var s struct {
			UpdaterToken string `json:"updaterToken"`
			GitHubToken  string `json:"githubToken"`
		}
		if json.Unmarshal(data, &s) == nil {
			if t := strings.TrimSpace(s.UpdaterToken); t != "" {
				return t
			}
			if t := strings.TrimSpace(s.GitHubToken); t != "" {
				return t
			}
		}
	}
	return ""
}

var ghReleaseDownloadRegex = regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+)/releases/download/([^/]+)/([^/?#]+)`)

// resolveAssetURL converts a GitHub browser download URL into an authenticated API asset URL when an updater token is present.
func resolveAssetURL(rawURL string) string {
	if rawURL == "" || strings.Contains(rawURL, "api.github.com") {
		return rawURL
	}
	token := resolveUpdaterToken()
	if token == "" {
		return rawURL
	}
	matches := ghReleaseDownloadRegex.FindStringSubmatch(rawURL)
	if len(matches) < 5 {
		return rawURL
	}
	owner, repo, tag, filename := matches[1], matches[2], matches[3], matches[4]
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, tag)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return rawURL
	}
	req.Header.Set("User-Agent", "Vinta-Updater/1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return rawURL
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return rawURL
	}

	var rel ghReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return rawURL
	}

	for _, asset := range rel.Assets {
		if strings.EqualFold(asset.Name, filename) && asset.URL != "" {
			return asset.URL
		}
	}
	return rawURL
}

// IsPortableMode returns false as standalone portable packaging has been deprecated in favor of the NSIS installer.
func (a *App) IsPortableMode() bool {
	return false
}

// GetUpdateRepo returns the repository owner/name used for update queries.
func (a *App) GetUpdateRepo() string {
	a.updaterMu.Lock()
	defer a.updaterMu.Unlock()
	if a.customUpdateRepo != "" {
		return a.customUpdateRepo
	}
	if envRepo := strings.TrimSpace(os.Getenv("VINTA_UPDATE_REPO")); envRepo != "" {
		return envRepo
	}
	return "kuraigu/Vinta"
}

// SetCustomUpdateRepo overrides the repository used for update queries.
func (a *App) SetCustomUpdateRepo(repo string) error {
	a.updaterMu.Lock()
	defer a.updaterMu.Unlock()
	a.customUpdateRepo = strings.TrimSpace(repo)
	return nil
}

// CheckForUpdates queries GitHub Releases API to check for newer versions.
func (a *App) CheckForUpdates(force bool) (*UpdateCheckResult, error) {
	currentVersion := AppVersion
	if currentVersion == "" || currentVersion == "dev" {
		currentVersion = "0.1.0-dev"
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	isPortable := a.IsPortableMode()
	repo := a.GetUpdateRepo()
	isCustom := (repo != "kuraigu/Vinta")

	// Block update checking on non-official builds unless explicitly configured for a custom repository
	isOfficial := BuildType == "Official" || BuildType == "Release"
	if !isOfficial && !isCustom {
		return &UpdateCheckResult{
			Status:          "dev-mode",
			UpdateAvailable: false,
			CurrentVersion:  currentVersion,
			LatestVersion:   currentVersion,
			CheckedAt:       now,
			Message:         "Update checking is disabled for non-official releases.",
			IsCustomRepo:    isCustom,
		}, nil
	}

	channel := "stable"
	settingsPath := filepath.Join(GetAppDataDir(), "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		var settings map[string]interface{}
		if err := json.Unmarshal(data, &settings); err == nil {
			if ch, ok := settings["vinta_update_channel"].(string); ok && ch != "" {
				channel = ch
			}
		}
	}

	cacheFile := filepath.Join(GetAppDataDir(), "cache", fmt.Sprintf("update_check_%s.json", channel))

	// Check local 1-hour cache unless forced
	if !force && !isCustom {
		if data, err := os.ReadFile(cacheFile); err == nil {
			var cached cachedUpdateData
			if err := json.Unmarshal(data, &cached); err == nil {
				if time.Now().Unix()-cached.Timestamp < 3600 {
					cleanCur := cleanVersion(currentVersion)
					cleanCachedCur := cleanVersion(cached.Result.CurrentVersion)
					cleanCachedLatest := cleanVersion(cached.Result.LatestVersion)

					// Only return cached result if it was checked for the current running version,
					// and it does not falsely claim the current version is an update.
					if cleanCur == cleanCachedCur && (!cached.Result.UpdateAvailable || (cleanCur != cleanCachedLatest && isNewerVersion(cleanCur, cleanCachedLatest))) {
						return &cached.Result, nil
					}
				}
			}
		}
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=30", repo)
	if customURL := strings.TrimSpace(os.Getenv("VINTA_UPDATE_URL")); customURL != "" {
		apiURL = customURL
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return &UpdateCheckResult{
			Status:          "error",
			UpdateAvailable: false,
			CurrentVersion:  currentVersion,
			CheckedAt:       now,
			Message:         fmt.Sprintf("Failed to construct update request: %v", err),
			IsPortable:      isPortable,
			IsCustomRepo:    isCustom,
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
			IsPortable:      isPortable,
			IsCustomRepo:    isCustom,
		}, nil
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		// Successful response, proceed to decode release payload below
	case http.StatusNotFound:
		result := &UpdateCheckResult{
			Status:          "up-to-date",
			UpdateAvailable: false,
			CurrentVersion:  currentVersion,
			LatestVersion:   currentVersion,
			CheckedAt:       now,
			Message:         "Vinta is currently up to date.",
			IsPortable:      isPortable,
			IsCustomRepo:    isCustom,
		}
		a.saveUpdateCache(cacheFile, result)
		return result, nil
	case http.StatusForbidden:
		return &UpdateCheckResult{
			Status:          "error",
			UpdateAvailable: false,
			CurrentVersion:  currentVersion,
			CheckedAt:       now,
			Message:         "GitHub API rate limit reached. Please try again later or supply a GitHub token.",
			IsPortable:      isPortable,
			IsCustomRepo:    isCustom,
		}, nil
	default:
		return &UpdateCheckResult{
			Status:          "error",
			UpdateAvailable: false,
			CurrentVersion:  currentVersion,
			CheckedAt:       now,
			Message:         fmt.Sprintf("GitHub API returned status code %d", resp.StatusCode),
			IsPortable:      isPortable,
			IsCustomRepo:    isCustom,
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
			IsPortable:      isPortable,
			IsCustomRepo:    isCustom,
		}, nil
	}

	var ghRelease *ghReleaseResponse
	for i := range releases {
		rel := &releases[i]
		if channel == "stable" && !isCustom {
			if rel.Draft || rel.Prerelease {
				continue
			}
		} else if channel == "prerelease" && !isCustom {
			if rel.Draft {
				continue
			}
		}
		// For testing builds or custom repos, include drafts and prereleases
		ghRelease = rel
		break
	}

	if ghRelease == nil {
		result := &UpdateCheckResult{
			Status:          "up-to-date",
			UpdateAvailable: false,
			CurrentVersion:  currentVersion,
			LatestVersion:   currentVersion,
			CheckedAt:       now,
			Message:         "Vinta is currently up to date.",
			IsPortable:      isPortable,
			IsCustomRepo:    isCustom,
		}
		a.saveUpdateCache(cacheFile, result)
		return result, nil
	}

	var installerURL, checksumsURL string
	token := resolveUpdaterToken()
	for _, asset := range ghRelease.Assets {
		lowerName := strings.ToLower(asset.Name)
		assetURL := asset.BrowserDownloadURL
		if token != "" && asset.URL != "" {
			assetURL = asset.URL
		}
		if strings.HasSuffix(lowerName, ".exe") && (strings.Contains(lowerName, "setup") || strings.Contains(lowerName, "installer")) {
			installerURL = assetURL
		} else if lowerName == "sha256sums.txt" || strings.HasSuffix(lowerName, ".sha256") {
			checksumsURL = assetURL
		}
	}

	cleanCurrent := cleanVersion(currentVersion)
	cleanLatest := cleanVersion(ghRelease.TagName)
	hasUpdate := cleanCurrent != cleanLatest && isNewerVersion(cleanCurrent, cleanLatest)

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
			PortableURL:  "",
			ChecksumsURL: checksumsURL,
			IsPrerelease: ghRelease.Prerelease,
			IsPortable:   isPortable,
		},
		CheckedAt:    now,
		Message:      msg,
		IsPortable:   isPortable,
		IsCustomRepo: isCustom,
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
	isOfficial := BuildType == "Official" || BuildType == "Release"
	if !isOfficial && os.Getenv("VINTA_UPDATE_URL") == "" && os.Getenv("VINTA_UPDATE_REPO") == "" {
		return
	}

	go func() {
		stopCtx := a.lifecycleContext()
		// Wait 10 seconds after app startup before first check
		startupTimer := time.NewTimer(10 * time.Second)
		select {
		case <-stopCtx.Done():
			startupTimer.Stop()
			return
		case <-startupTimer.C:
		}

		shouldCheck := func() bool {
			settingsPath := filepath.Join(GetAppDataDir(), "settings.json")
			if data, err := os.ReadFile(settingsPath); err == nil {
				var settings map[string]interface{}
				if err := json.Unmarshal(data, &settings); err == nil {
					if val, ok := settings["autoCheckUpdates"].(bool); ok && !val {
						return false
					}
				}
			}
			return true
		}

		if shouldCheck() {
			if res, err := a.CheckForUpdates(false); err == nil && res.UpdateAvailable {
				if app := application.Get(); app != nil {
					app.Event.Emit("update:available", res)
				}
			}
		}

		// Check every 6 hours
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-stopCtx.Done():
				return
			case <-ticker.C:
				if shouldCheck() {
					if res, err := a.CheckForUpdates(false); err == nil && res.UpdateAvailable {
						if app := application.Get(); app != nil {
							app.Event.Emit("update:available", res)
						}
					}
				}
			}
		}
	}()
}

// DownloadUpdate initiates or resumes a chunked download with real-time progress and checksum verification.
func (a *App) DownloadUpdate(downloadURL, checksumsURL string) error {
	if downloadURL == "" {
		return fmt.Errorf("no download URL provided")
	}

	downloadURL = resolveAssetURL(downloadURL)
	if checksumsURL != "" {
		checksumsURL = resolveAssetURL(checksumsURL)
	}

	a.updaterMu.Lock()
	if a.downloadCancel != nil {
		a.downloadCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.downloadCancel = cancel
	a.downloadPaused = false
	a.activeDownloadURL = downloadURL
	a.activeChecksumURL = checksumsURL
	a.updaterMu.Unlock()

	tempDir := os.TempDir()
	partPath := filepath.Join(tempDir, "Vinta-Setup-Update.exe.part")
	finalPath := filepath.Join(tempDir, "Vinta-Setup-Update.exe")

	var existingBytes int64
	if stat, err := os.Stat(partPath); err == nil {
		existingBytes = stat.Size()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("User-Agent", "Vinta-Updater/1.0")
	if strings.Contains(downloadURL, "api.github.com") {
		req.Header.Set("Accept", "application/octet-stream")
	}
	if token := resolveUpdaterToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if existingBytes > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existingBytes))
	}

	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		a.emitDownloadProgress("error", existingBytes, 0, 0, 0, fmt.Sprintf("Download failed: %v", err))
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	var outFile *os.File
	var totalBytes int64

	switch resp.StatusCode {
	case http.StatusPartialContent:
		// Server supports range requests. Append to the existing partial file
		outFile, err = os.OpenFile(partPath, os.O_APPEND|os.O_WRONLY, 0755)
		if err != nil {
			return err
		}
		totalBytes = existingBytes + resp.ContentLength
	case http.StatusOK:
		// Fresh download or server does not support range headers
		outFile, err = os.OpenFile(partPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		existingBytes = 0
		totalBytes = resp.ContentLength
	default:
		err := fmt.Errorf("server returned status: %d", resp.StatusCode)
		a.emitDownloadProgress("error", existingBytes, 0, 0, 0, err.Error())
		return err
	}
	defer func() { _ = outFile.Close() }()

	buffer := make([]byte, 32*1024)
	downloaded := existingBytes
	lastEmit := time.Now()
	bytesSinceLastEmit := int64(0)

	for {
		select {
		case <-ctx.Done():
			a.updaterMu.Lock()
			paused := a.downloadPaused
			a.updaterMu.Unlock()
			if paused {
				a.emitDownloadProgress("paused", downloaded, totalBytes, calcPercent(downloaded, totalBytes), 0, "")
			}
			return nil
		default:
		}

		n, rErr := resp.Body.Read(buffer)
		if n > 0 {
			if _, wErr := outFile.Write(buffer[:n]); wErr != nil {
				a.emitDownloadProgress("error", downloaded, totalBytes, calcPercent(downloaded, totalBytes), 0, wErr.Error())
				return wErr
			}
			downloaded += int64(n)
			bytesSinceLastEmit += int64(n)

			if time.Since(lastEmit) >= 120*time.Millisecond {
				elapsed := time.Since(lastEmit).Seconds()
				var speed int64
				if elapsed > 0 {
					speed = int64(float64(bytesSinceLastEmit) / elapsed)
				}
				a.emitDownloadProgress("downloading", downloaded, totalBytes, calcPercent(downloaded, totalBytes), speed, "")
				lastEmit = time.Now()
				bytesSinceLastEmit = 0
			}
		}

		if rErr != nil {
			if rErr == io.EOF {
				break
			}
			if ctx.Err() != nil {
				return nil
			}
			a.emitDownloadProgress("error", downloaded, totalBytes, calcPercent(downloaded, totalBytes), 0, rErr.Error())
			return rErr
		}
	}

	_ = outFile.Close()

	// Download is complete, verify checksum
	a.emitDownloadProgress("verifying", downloaded, totalBytes, 100, 0, "")

	if checksumsURL != "" {
		expectedHash, err := a.fetchExpectedChecksum(checksumsURL, filepath.Base(downloadURL))
		if err != nil || expectedHash == "" {
			_ = os.Remove(partPath)
			a.emitDownloadProgress("error", downloaded, totalBytes, 0, 0, "Failed to retrieve release checksum manifest.")
			return fmt.Errorf("failed to retrieve checksum manifest: %w", err)
		}
		computedHash, err := computeFileSHA256(partPath)
		if err != nil {
			_ = os.Remove(partPath)
			a.emitDownloadProgress("error", downloaded, totalBytes, 0, 0, "Failed to compute checksum: "+err.Error())
			return err
		}
		if !strings.EqualFold(computedHash, expectedHash) {
			_ = os.Remove(partPath)
			a.emitDownloadProgress("error", downloaded, totalBytes, 0, 0, "Checksum verification failed: download corrupted.")
			return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, computedHash)
		}
	}

	// Rename .part file to final executable
	_ = os.Remove(finalPath)
	if err := os.Rename(partPath, finalPath); err != nil {
		a.emitDownloadProgress("error", downloaded, totalBytes, 0, 0, "Failed to finalize download: "+err.Error())
		return err
	}

	a.updaterMu.Lock()
	a.pendingInstallPath = finalPath
	a.updaterMu.Unlock()

	a.emitDownloadProgress("ready", downloaded, totalBytes, 100, 0, "")
	return nil
}

// PauseUpdateDownload pauses the active download stream while preserving written bytes.
func (a *App) PauseUpdateDownload() error {
	a.updaterMu.Lock()
	a.downloadPaused = true
	if a.downloadCancel != nil {
		a.downloadCancel()
	}
	a.updaterMu.Unlock()
	return nil
}

// ResumeUpdateDownload resumes a paused download.
func (a *App) ResumeUpdateDownload() error {
	a.updaterMu.Lock()
	url := a.activeDownloadURL
	checksumURL := a.activeChecksumURL
	a.updaterMu.Unlock()

	if url == "" {
		return fmt.Errorf("no active download to resume")
	}

	go func() {
		_ = a.DownloadUpdate(url, checksumURL)
	}()
	return nil
}

// CancelUpdateDownload cancels an update download and removes partial files.
func (a *App) CancelUpdateDownload() error {
	a.updaterMu.Lock()
	a.downloadPaused = false
	if a.downloadCancel != nil {
		a.downloadCancel()
	}
	a.pendingInstallPath = ""
	a.updaterMu.Unlock()

	partPath := filepath.Join(os.TempDir(), "Vinta-Setup-Update.exe.part")
	_ = os.Remove(partPath)

	a.emitDownloadProgress("idle", 0, 0, 0, 0, "")
	return nil
}

// ApplyUpdate launches the downloaded installer and gracefully exits Vinta.
func (a *App) ApplyUpdate() error {
	a.updaterMu.Lock()
	installerPath := a.pendingInstallPath
	a.updaterMu.Unlock()

	if installerPath == "" {
		installerPath = filepath.Join(os.TempDir(), "Vinta-Setup-Update.exe")
	}

	if _, err := os.Stat(installerPath); err != nil {
		return fmt.Errorf("installer binary not found: %w", err)
	}

	_, _ = a.CreateSettingsBackup()

	exePath, err := os.Executable()
	if err != nil {
		exePath = "C:\\"
	}
	exeDir := filepath.Dir(exePath)

	// Launch installer passing current destination directory to maintain existing path without quotes
	cmd := exec.Command(installerPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: fmt.Sprintf(`"%s" /D=%s`, installerPath, exeDir),
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch installer: %w", err)
	}

	// Terminate Vinta gracefully so the installer can overwrite binaries without locks
	go func() {
		time.Sleep(500 * time.Millisecond)
		a.Quit()
	}()

	return nil
}

// DownloadAndInstallUpdate downloads the installer executable for the update and executes it.
func (a *App) DownloadAndInstallUpdate(downloadURL string) error {
	a.updaterMu.Lock()
	checksumURL := a.activeChecksumURL
	a.updaterMu.Unlock()
	if err := a.DownloadUpdate(downloadURL, checksumURL); err != nil {
		return err
	}
	return a.ApplyUpdate()
}

// SimulateUpdateAvailable emits a simulated update payload for UI and lifecycle testing.
func (a *App) SimulateUpdateAvailable(version, notes string) error {
	if version == "" {
		version = "v0.4.0-alpha"
	}
	if notes == "" {
		notes = "## What's Changed\n* Added real-time download progress and pause/resume support ([#127](https://github.com/kuraigu/Vinta/pull/127))\n* Added in-app release notes viewer with categorized PR links\n* Enhanced installer with automatic installation path detection"
	}
	res := &UpdateCheckResult{
		Status:          "update-available",
		UpdateAvailable: true,
		CurrentVersion:  AppVersion,
		LatestVersion:   version,
		Release: ReleaseInfo{
			Version:      version,
			TagName:      version,
			Name:         "Release " + version,
			Body:         notes,
			PublishedAt:  time.Now().Format(time.RFC3339),
			HTMLURL:      "https://github.com/kuraigu/Vinta/releases/tag/" + version,
			InstallerURL: "https://mock.vinta.app/installer.exe",
			PortableURL:  "",
			ChecksumsURL: "https://mock.vinta.app/sha256sums.txt",
			IsPrerelease: true,
			IsPortable:   a.IsPortableMode(),
		},
		CheckedAt:    time.Now().Format("2006-01-02 15:04:05"),
		Message:      "Simulated update available for testing.",
		IsPortable:   a.IsPortableMode(),
		IsCustomRepo: true,
	}
	if app := application.Get(); app != nil {
		app.Event.Emit("update:available", res)
	}
	return nil
}

// SimulateDownloadCycle runs a simulated download progress ticker for offline test validation.
func (a *App) SimulateDownloadCycle(failChecksum bool) error {
	go func() {
		total := int64(62914560)
		step := int64(6291456)
		current := int64(0)

		for current < total {
			time.Sleep(250 * time.Millisecond)
			current += step
			if current > total {
				current = total
			}
			a.emitDownloadProgress("downloading", current, total, calcPercent(current, total), 2500000, "")
		}

		a.emitDownloadProgress("verifying", total, total, 100, 0, "")
		time.Sleep(500 * time.Millisecond)

		if failChecksum {
			a.emitDownloadProgress("error", total, total, 0, 0, "Simulated checksum mismatch.")
		} else {
			a.emitDownloadProgress("ready", total, total, 100, 0, "")
		}
	}()
	return nil
}

func (a *App) emitDownloadProgress(status string, downloaded, total int64, percent float64, speed int64, errMsg string) {
	if app := application.Get(); app != nil {
		app.Event.Emit("update:download-progress", DownloadProgress{
			Status:           status,
			DownloadedBytes:  downloaded,
			TotalBytes:       total,
			Percentage:       percent,
			SpeedBytesPerSec: speed,
			Error:            errMsg,
		})
	}
}

func (a *App) fetchExpectedChecksum(checksumsURL, assetFilename string) (string, error) {
	checksumsURL = resolveAssetURL(checksumsURL)
	req, err := http.NewRequest("GET", checksumsURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Vinta-Updater/1.0")
	if strings.Contains(checksumsURL, "api.github.com") {
		req.Header.Set("Accept", "application/octet-stream")
	}
	if token := resolveUpdaterToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksum server returned %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	targetLower := strings.ToLower(assetFilename)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hash := parts[0]
			filename := parts[1]
			if strings.EqualFold(filepath.Base(filename), targetLower) || strings.EqualFold(filepath.Base(filename), "vinta-setup-x64.exe") {
				return hash, nil
			}
		}
	}
	return "", fmt.Errorf("asset %s not found in checksums manifest", assetFilename)
}

func computeFileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func calcPercent(downloaded, total int64) float64 {
	if total <= 0 {
		return 0
	}
	p := (float64(downloaded) / float64(total)) * 100
	if p > 100 {
		p = 100
	}
	return p
}

// cleanVersion trims any leading 'v' or 'V' and whitespace.
func cleanVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimLeft(v, "vV")
	return v
}

// isNewerVersion returns true if candidate is newer than current.
func isNewerVersion(current, candidate string) bool {
	c1 := cleanVersion(current)
	c2 := cleanVersion(candidate)
	if c1 == c2 || c2 == "" {
		return false
	}

	currMajor, currMinor, currPatch, currPre := parseSemVer(c1)
	candMajor, candMinor, candPatch, candPre := parseSemVer(c2)

	if candMajor != currMajor {
		return candMajor > currMajor
	}
	if candMinor != currMinor {
		return candMinor > currMinor
	}
	if candPatch != currPatch {
		return candPatch > currPatch
	}

	// Base versions are equal
	// Release version is newer than pre-release version (0.1.0 > 0.1.0-alpha)
	if currPre != "" && candPre == "" {
		return true
	}
	if currPre == "" && candPre != "" {
		return false
	}
	if currPre != "" && candPre != "" {
		return comparePreRelease(currPre, candPre) < 0
	}

	return false
}

// comparePreRelease compares two pre-release strings. Returns negative if a < b, positive if a > b, 0 if equal.
func comparePreRelease(a, b string) int {
	if a == b {
		return 0
	}
	if strings.HasPrefix(a, "dev") && !strings.HasPrefix(b, "dev") {
		return -1
	}
	if !strings.HasPrefix(a, "dev") && strings.HasPrefix(b, "dev") {
		return 1
	}

	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")
	minLen := len(partsA)
	if len(partsB) < minLen {
		minLen = len(partsB)
	}

	for i := 0; i < minLen; i++ {
		numA, errA := strconv.Atoi(partsA[i])
		numB, errB := strconv.Atoi(partsB[i])

		if errA == nil && errB == nil {
			if numA != numB {
				return numA - numB
			}
		} else {
			if partsA[i] != partsB[i] {
				return strings.Compare(partsA[i], partsB[i])
			}
		}
	}

	return len(partsA) - len(partsB)
}

func parseSemVer(v string) (int, int, int, string) {
	v = cleanVersion(v)
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
