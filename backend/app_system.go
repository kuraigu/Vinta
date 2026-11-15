// Package backend provides core application services and state management for Vinta.
package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"

	"Vinta/backend/services"
)

const (
	autostartRunKeyPath             = `Software\Microsoft\Windows\CurrentVersion\Run`
	autostartStartupApprovedKeyPath = `Software\Microsoft\Windows\CurrentVersion\Explorer\StartupApproved\Run`
	autostartValueName              = "Vinta"
)

func resolveExecutablePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(exePath); err == nil && resolved != "" {
		exePath = resolved
	}
	return filepath.Clean(exePath), nil
}

func parseRunCommandExecutable(cmd string) string {
	trimmed := strings.TrimSpace(cmd)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, `"`) {
		endIdx := strings.Index(trimmed[1:], `"`)
		if endIdx <= 0 {
			return ""
		}
		return filepath.Clean(trimmed[1 : endIdx+1])
	}
	if spaceIdx := strings.IndexAny(trimmed, " \t"); spaceIdx > 0 {
		trimmed = trimmed[:spaceIdx]
	}
	return filepath.Clean(trimmed)
}

func isStartupApprovedEnabled() (bool, bool) {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		autostartStartupApprovedKeyPath,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return false, true
	}
	defer func() { _ = key.Close() }()

	data, _, err := key.GetBinaryValue(autostartValueName)
	if err != nil || len(data) == 0 {
		return false, true
	}

	return true, data[0] == 0x02 || data[0] == 0x06
}

// SetAutostart registers or unregisters the app in Windows Startup Registry.
func (a *App) SetAutostart(enable bool) error {
	if isTestModeSync() {
		return nil
	}

	if enable {
		exePath, err := resolveExecutablePath()
		if err != nil {
			return err
		}

		key, _, err := registry.CreateKey(
			registry.CURRENT_USER,
			autostartRunKeyPath,
			registry.SET_VALUE,
		)
		if err != nil {
			return err
		}
		err = key.SetStringValue(autostartValueName, fmt.Sprintf(`"%s"`, exePath))
		_ = key.Close()
		if err != nil {
			return err
		}

		approvedKey, _, err := registry.CreateKey(
			registry.CURRENT_USER,
			autostartStartupApprovedKeyPath,
			registry.SET_VALUE,
		)
		if err != nil {
			return err
		}
		defer func() { _ = approvedKey.Close() }()

		// Windows Task Manager / Settings use a 12-byte StartupApproved entry where byte 0 == 0x02 means enabled.
		enabledBlob := []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		return approvedKey.SetBinaryValue(autostartValueName, enabledBlob)
	}

	if key, err := registry.OpenKey(
		registry.CURRENT_USER,
		autostartRunKeyPath,
		registry.SET_VALUE,
	); err == nil {
		err = key.DeleteValue(autostartValueName)
		_ = key.Close()
		if err != nil && !errors.Is(err, registry.ErrNotExist) {
			return err
		}
	} else if !errors.Is(err, registry.ErrNotExist) {
		return err
	}

	if approvedKey, err := registry.OpenKey(
		registry.CURRENT_USER,
		autostartStartupApprovedKeyPath,
		registry.SET_VALUE,
	); err == nil {
		_ = approvedKey.DeleteValue(autostartValueName)
		_ = approvedKey.Close()
	}

	return nil
}

// GetAutostart checks whether autostart is currently enabled and valid in Windows Registry.
func (a *App) GetAutostart() bool {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		autostartRunKeyPath,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return false
	}
	defer func() { _ = key.Close() }()

	val, _, err := key.GetStringValue(autostartValueName)
	if err != nil || val == "" {
		return false
	}

	regExe := parseRunCommandExecutable(val)
	if regExe == "" {
		return false
	}
	info, err := os.Stat(regExe)
	if err != nil || info.IsDir() {
		return false
	}

	_, approved := isStartupApprovedEnabled()
	return approved
}

// SyncAutostartOnBoot repairs missing or broken Run registry paths when autostart is enabled
// in settings, while respecting explicit user disables in Windows StartupApproved and leaving
// any existing valid executable path untouched.
func (a *App) SyncAutostartOnBoot() {
	if isTestModeSync() {
		return
	}

	settingsJSON, err := a.GetSettings()
	if err != nil || settingsJSON == "" || settingsJSON == "{}" {
		return
	}

	var m map[string]any
	if json.Unmarshal([]byte(settingsJSON), &m) != nil {
		return
	}

	enabled, ok := m["vinta_autostart"].(bool)
	if !ok || !enabled {
		return
	}

	if exists, approved := isStartupApprovedEnabled(); exists && !approved {
		return
	}

	currentExe, err := resolveExecutablePath()
	if err != nil {
		return
	}

	tempDir := filepath.Clean(os.TempDir())
	if tempDir != "" && strings.HasPrefix(strings.ToLower(currentExe), strings.ToLower(tempDir+string(filepath.Separator))) {
		return
	}

	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		autostartRunKeyPath,
		registry.QUERY_VALUE,
	)
	if err != nil {
		_ = a.SetAutostart(true)
		return
	}
	val, _, valErr := key.GetStringValue(autostartValueName)
	_ = key.Close()

	regExe := parseRunCommandExecutable(val)
	if valErr != nil || regExe == "" {
		_ = a.SetAutostart(true)
		return
	}

	if info, statErr := os.Stat(regExe); statErr != nil || info.IsDir() {
		_ = a.SetAutostart(true)
	}
}

// OpenWindowsSettings launches Windows Settings flyouts and pages corresponding to the target category.
func (a *App) OpenWindowsSettings(target string) error {
	var uri string
	switch target {
	case "wifi":
		uri = "ms-settings:network-wifi"
	case "network":
		uri = "ms-availablenetworks:"
	case "bluetooth":
		uri = "ms-settings:bluetooth"
	case "battery":
		uri = "ms-settings:batterysaver"
	case "airplane":
		uri = "ms-settings:network-airplanemode"
	default:
		uri = "ms-settings:"
	}
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", uri).Start()
}

// GetBatteryStatus returns current system power status from the services layer.
func (a *App) GetBatteryStatus() services.BatteryStatus {
	return services.GetBatteryStatus()
}

// SetBatterySaver updates the system battery saver mode threshold.
func (a *App) SetBatterySaver(enabled bool) bool {
	return services.SetBatterySaver(enabled)
}

// GetWirelessStatus returns current wireless connectivity and Wi-Fi signal status from the services layer.
func (a *App) GetWirelessStatus() services.WirelessStatus {
	return services.GetWirelessStatus()
}

// SetWifiEnabled toggles the system Wi-Fi radio.
func (a *App) SetWifiEnabled(enabled bool) bool {
	return services.SetWifiEnabled(enabled)
}

// SetBluetoothEnabled toggles the system Bluetooth radio.
func (a *App) SetBluetoothEnabled(enabled bool) bool {
	return services.SetBluetoothEnabled(enabled)
}

// SetAirplaneMode toggles the system Airplane mode.
func (a *App) SetAirplaneMode(enabled bool) bool {
	return services.SetAirplaneMode(enabled)
}

// GetCacheSize computes the total byte size of disk caches and temporary sync databases.
func (a *App) GetCacheSize() (int64, error) {
	var totalSize int64

	cacheDir := filepath.Join(GetAppDataDir(), "cache")
	if info, err := os.Stat(cacheDir); err == nil && info.IsDir() {
		_ = filepath.Walk(cacheDir, func(_ string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
	}

	tempDir := os.TempDir()
	matches, _ := filepath.Glob(filepath.Join(tempDir, "wpn_vinta_sync.db*"))
	for _, match := range matches {
		if info, err := os.Stat(match); err == nil && !info.IsDir() {
			totalSize += info.Size()
		}
	}

	return totalSize, nil
}

// ClearCache deletes persistent disk cache files, cleans temp sync files, and flushes in-memory caches.
func (a *App) ClearCache() error {
	cacheDir := filepath.Join(GetAppDataDir(), "cache")

	if entries, err := os.ReadDir(cacheDir); err == nil {
		for _, entry := range entries {
			targetPath := filepath.Join(cacheDir, entry.Name())
			_ = os.RemoveAll(targetPath)
		}
	} else if os.IsNotExist(err) {
		_ = os.MkdirAll(cacheDir, 0755)
	}

	_ = os.MkdirAll(filepath.Join(cacheDir, "icons"), 0755)

	tempDir := os.TempDir()
	matches, _ := filepath.Glob(filepath.Join(tempDir, "wpn_vinta_sync.db*"))
	for _, match := range matches {
		_ = os.Remove(match)
	}

	FlushIconMemoryCache()
	services.ClearNativeMediaArtCache()

	return nil
}

// GetSystemTelemetry retrieves instantaneous CPU, RAM, and GPU hardware metrics.
func (a *App) GetSystemTelemetry() services.SystemTelemetry {
	return services.GetCurrentTelemetry()
}

// StartSystemTelemetry initiates periodic background telemetry emission to the frontend.
func (a *App) StartSystemTelemetry() bool {
	services.StartTelemetryStream()
	return true
}

// StopSystemTelemetry decrements active telemetry listeners and pauses the sampling loop when unobserved.
func (a *App) StopSystemTelemetry() bool {
	services.StopTelemetryStream()
	return true
}
