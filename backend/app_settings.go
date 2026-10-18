// Package backend provides core application services and state management for Vinta.
package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

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

// ResetSettings deletes settings.json to restore factory configuration defaults.
func (a *App) ResetSettings() error {
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

// ApplyImportedSettings validates and saves sanitized settings JSON to %APPDATA%/FreeMatrix Studios/Vinta/settings.json.
func (a *App) ApplyImportedSettings(sanitizedJSON string) error {
	if strings.TrimSpace(sanitizedJSON) == "" {
		return fmt.Errorf("settings payload cannot be empty")
	}

	if !json.Valid([]byte(sanitizedJSON)) {
		return fmt.Errorf("invalid JSON payload")
	}

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

// GetHardwareAccelerationConfig reads the hardware acceleration preference from configuration storage.
func GetHardwareAccelerationConfig() bool {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return true
	}
	filePath := filepath.Join(configDir, "FreeMatrix Studios", "Vinta", "hw_accel.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return true
	}
	return strings.TrimSpace(string(data)) != "false"
}

// GetHardwareAcceleration returns whether GPU acceleration is enabled.
func (a *App) GetHardwareAcceleration() bool {
	return GetHardwareAccelerationConfig()
}

// SetHardwareAcceleration persists the hardware acceleration preference to disk.
func (a *App) SetHardwareAcceleration(enabled bool) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	dirPath := filepath.Join(configDir, "FreeMatrix Studios", "Vinta")
	_ = os.MkdirAll(dirPath, 0755)
	filePath := filepath.Join(dirPath, "hw_accel.txt")
	val := "true"
	if !enabled {
		val = "false"
	}
	return os.WriteFile(filePath, []byte(val), 0644)
}
