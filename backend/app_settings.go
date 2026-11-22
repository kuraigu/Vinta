// Package backend provides core application services and state management for Vinta.
package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const maxSettingsBackups = 3

var (
	bootHardwareAccel     = true
	bootHardwareAccelOnce sync.Once
)

// SettingsBackupInfo describes a rotating local settings snapshot on disk.
type SettingsBackupInfo struct {
	Filename  string `json:"filename"`
	CreatedAt string `json:"createdAt"`
	SizeBytes int64  `json:"sizeBytes"`
}

// GetSettings reads settings.json from %APPDATA%/FreeMatrix Studios/Vinta.
func (a *App) GetSettings() (string, error) {
	filePath := filepath.Join(GetAppDataDir(), "settings.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "{}", nil
	}
	return string(data), nil
}

// SaveSettings writes settings.json to %APPDATA%/FreeMatrix Studios/Vinta.
func (a *App) SaveSettings(settingsJSON string) error {
	dir := GetAppDataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	filePath := filepath.Join(dir, "settings.json")
	return os.WriteFile(filePath, []byte(settingsJSON), 0644)
}

// ResetSettings creates a safety backup and deletes settings.json to restore factory configuration defaults.
func (a *App) ResetSettings() error {
	_, _ = a.CreateSettingsBackup()
	filePath := filepath.Join(GetAppDataDir(), "settings.json")
	_ = os.Remove(filePath)
	return nil
}

// ExportSettingsFileDialog opens a native save dialog and exports settings with schema metadata.
func (a *App) ExportSettingsFileDialog(defaultName string) (string, error) {
	wailsApp := application.Get()
	if wailsApp == nil {
		return "", fmt.Errorf("application instance not available")
	}

	if defaultName == "" {
		defaultName = "vinta-settings.json"
	}

	dialog := wailsApp.Dialog.SaveFile()
	dialog.SetMessage("Export Vinta Settings")
	dialog.SetFilename(defaultName)
	dialog.AddFilter("JSON Files (*.json)", "*.json")

	targetPath, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	if targetPath == "" {
		return "", nil
	}

	currentJSON, _ := a.GetSettings()
	var rawSettings any
	if err := json.Unmarshal([]byte(currentJSON), &rawSettings); err != nil {
		rawSettings = map[string]any{}
	}

	exportPayload := map[string]any{
		"$schemaVersion": 1,
		"exportedAt":     time.Now().UTC().Format(time.RFC3339),
		"settings":       rawSettings,
	}

	data, err := json.MarshalIndent(exportPayload, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return "", err
	}

	return targetPath, nil
}

// ImportSettingsFileDialog opens a native file dialog and loads the raw content of the selected file.
func (a *App) ImportSettingsFileDialog() (string, error) {
	wailsApp := application.Get()
	if wailsApp == nil {
		return "", fmt.Errorf("application instance not available")
	}

	dialog := wailsApp.Dialog.OpenFile()
	dialog.SetTitle("Import Vinta Settings")
	dialog.AddFilter("JSON Files (*.json)", "*.json")

	selectedPath, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	if selectedPath == "" {
		return "", nil
	}

	info, err := os.Stat(selectedPath)
	if err != nil {
		return "", err
	}

	const maxSettingsPayload = 256 * 1024
	if info.Size() > maxSettingsPayload {
		return "", fmt.Errorf("file size exceeds allowable limit (256 KB)")
	}

	data, err := os.ReadFile(selectedPath)
	if err != nil {
		return "", err
	}

	if !json.Valid(data) {
		return "", fmt.Errorf("selected file does not contain valid JSON")
	}

	return string(data), nil
}

// ApplyImportedSettings creates a safety backup, validates, and saves sanitized settings JSON to %APPDATA%/FreeMatrix Studios/Vinta/settings.json.
func (a *App) ApplyImportedSettings(sanitizedJSON string) error {
	if strings.TrimSpace(sanitizedJSON) == "" {
		return fmt.Errorf("settings payload cannot be empty")
	}

	if !json.Valid([]byte(sanitizedJSON)) {
		return fmt.Errorf("invalid JSON payload")
	}

	_, _ = a.CreateSettingsBackup()

	dir := GetAppDataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(dir, "settings.json")
	tmpPath := filepath.Join(dir, "settings.json.tmp")

	if err := os.WriteFile(tmpPath, []byte(sanitizedJSON), 0644); err != nil {
		return err
	}

	_ = os.Remove(filePath)
	if err := os.Rename(tmpPath, filePath); err != nil {
		return os.WriteFile(filePath, []byte(sanitizedJSON), 0644)
	}

	return nil
}

func readSettingsMap(dir string) (map[string]any, bool) {
	filePath := filepath.Join(dir, "settings.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return map[string]any{}, false
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil || parsed == nil {
		return map[string]any{}, false
	}
	return parsed, true
}

func listSettingsBackupEntries(backupsDir string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	infos := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "settings-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		if info, statErr := entry.Info(); statErr == nil {
			infos = append(infos, info)
		}
	}
	sort.Slice(infos, func(i, j int) bool {
		if !infos[i].ModTime().Equal(infos[j].ModTime()) {
			return infos[i].ModTime().After(infos[j].ModTime())
		}
		return infos[i].Name() > infos[j].Name()
	})
	return infos, nil
}

// GetHardwareAccelerationConfig reads the hardware acceleration preference from settings.json,
// migrating any legacy hw_accel.txt file on first run and recording the boot-time state.
func GetHardwareAccelerationConfig() bool {
	dir := GetAppDataDir()
	settingsMap, _ := readSettingsMap(dir)

	enabled := true
	rawVal, hasKey := settingsMap["vinta_hardware_accel"]
	if hasKey {
		if b, ok := rawVal.(bool); ok {
			enabled = b
		}
	}

	legacyPath := filepath.Join(dir, "hw_accel.txt")
	if legacyData, err := os.ReadFile(legacyPath); err == nil {
		legacyEnabled := strings.TrimSpace(string(legacyData)) != "false"
		if !hasKey {
			enabled = legacyEnabled
			if !legacyEnabled {
				settingsMap["vinta_hardware_accel"] = false
				if os.MkdirAll(dir, 0755) == nil {
					if out, mErr := json.Marshal(settingsMap); mErr == nil {
						_ = os.WriteFile(filepath.Join(dir, "settings.json"), out, 0644)
					}
				}
			}
		}
		_ = os.Remove(legacyPath)
	}

	bootHardwareAccelOnce.Do(func() {
		bootHardwareAccel = enabled
	})
	return enabled
}

// GetHardwareAcceleration returns whether GPU acceleration is currently configured in settings.json.
func (a *App) GetHardwareAcceleration() bool {
	settingsMap, _ := readSettingsMap(GetAppDataDir())
	if rawVal, ok := settingsMap["vinta_hardware_accel"]; ok {
		if b, isBool := rawVal.(bool); isBool {
			return b
		}
	}
	return true
}

// GetBootHardwareAcceleration returns whether hardware acceleration was enabled when the application booted.
func (a *App) GetBootHardwareAcceleration() bool {
	return bootHardwareAccel
}

// SetHardwareAcceleration persists the hardware acceleration preference in settings.json and removes any legacy hw_accel.txt file.
func (a *App) SetHardwareAcceleration(enabled bool) error {
	dir := GetAppDataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	settingsMap, _ := readSettingsMap(dir)
	if enabled {
		delete(settingsMap, "vinta_hardware_accel")
	} else {
		settingsMap["vinta_hardware_accel"] = false
	}
	out, err := json.Marshal(settingsMap)
	if err != nil {
		return err
	}
	_ = os.Remove(filepath.Join(dir, "hw_accel.txt"))
	return os.WriteFile(filepath.Join(dir, "settings.json"), out, 0644)
}

// WaitForParentExitIfRestarting blocks startup briefly when relaunched via RestartApp
// so the exiting Vinta process releases its SingleInstance lock and WebView2 handles cleanly.
func WaitForParentExitIfRestarting() {
	const prefix = "--wait-pid="
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, prefix) {
			continue
		}
		pidVal, err := strconv.ParseUint(strings.TrimPrefix(arg, prefix), 10, 32)
		if err != nil || pidVal == 0 {
			return
		}
		handle, openErr := syscall.OpenProcess(syscall.SYNCHRONIZE, false, uint32(pidVal))
		if openErr == nil && handle != 0 {
			_, _ = syscall.WaitForSingleObject(handle, 5000)
			_ = syscall.CloseHandle(handle)
			time.Sleep(150 * time.Millisecond)
		}
		return
	}
}

// RestartApp launches a fresh instance of Vinta waiting for this PID to exit, then terminates the current process.
func (a *App) RestartApp() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	args := make([]string, 0, len(os.Args))
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "--wait-pid=") {
			args = append(args, arg)
		}
	}
	args = append(args, fmt.Sprintf("--wait-pid=%d", os.Getpid()))

	cmd := exec.Command(exePath, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	a.Quit()
	return nil
}

// CreateSettingsBackup snapshots settings.json into the rotating backups directory (retaining the newest 3 backups).
func (a *App) CreateSettingsBackup() (string, error) {
	dir := GetAppDataDir()
	srcPath := filepath.Join(dir, "settings.json")
	data, err := os.ReadFile(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	trimmed := strings.TrimSpace(string(data))
	if len(trimmed) <= 2 || trimmed == "{}" {
		return "", nil
	}

	backupsDir := filepath.Join(dir, "backups")
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("settings-%s.json", time.Now().Format("20060102-150405"))
	destPath := filepath.Join(backupsDir, filename)
	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return "", err
	}

	infos, listErr := listSettingsBackupEntries(backupsDir)
	if listErr == nil && len(infos) > maxSettingsBackups {
		for _, old := range infos[maxSettingsBackups:] {
			_ = os.Remove(filepath.Join(backupsDir, old.Name()))
		}
	}

	return filename, nil
}

// GetSettingsBackups returns up to 3 automatic settings snapshots ordered newest first.
func (a *App) GetSettingsBackups() ([]SettingsBackupInfo, error) {
	backupsDir := filepath.Join(GetAppDataDir(), "backups")
	infos, err := listSettingsBackupEntries(backupsDir)
	if err != nil {
		return nil, err
	}
	limit := len(infos)
	if limit > maxSettingsBackups {
		limit = maxSettingsBackups
	}
	result := make([]SettingsBackupInfo, 0, limit)
	for _, info := range infos[:limit] {
		result = append(result, SettingsBackupInfo{
			Filename:  info.Name(),
			CreatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
			SizeBytes: info.Size(),
		})
	}
	return result, nil
}

// RestoreLatestSettingsBackup restores settings.json from the most recent rotating backup and returns its JSON payload.
func (a *App) RestoreLatestSettingsBackup() (string, error) {
	dir := GetAppDataDir()
	backupsDir := filepath.Join(dir, "backups")
	infos, err := listSettingsBackupEntries(backupsDir)
	if err != nil {
		return "", err
	}
	if len(infos) == 0 {
		return "", fmt.Errorf("no settings backup available")
	}

	latestPath := filepath.Join(backupsDir, infos[0].Name())
	data, err := os.ReadFile(latestPath)
	if err != nil {
		return "", err
	}
	if !json.Valid(data) {
		return "", fmt.Errorf("latest backup file contains invalid JSON")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), data, 0644); err != nil {
		return "", err
	}
	return string(data), nil
}
