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

type autostartConfig struct {
	runKeyPath      string
	approvedKeyPath string
	valueName       string
}

func defaultAutostartConfig() autostartConfig {
	return autostartConfig{
		runKeyPath:      autostartRunKeyPath,
		approvedKeyPath: autostartStartupApprovedKeyPath,
		valueName:       autostartValueName,
	}
}

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

// isStartupApprovedEnabled inspects Windows Explorer's StartupApproved\Run binary entry.
// Windows stores a 12-byte REG_BINARY value where byte 0 is the approval flag (even values
// 0x02 and 0x06 mean enabled; odd values such as 0x03, 0x05, or 0x07 mean disabled via Task
// Manager or Windows Settings, with bytes 4..11 holding the disabled FILETIME).
// Returns (exists, enabled); missing entries default to (false, true).
func isStartupApprovedEnabled(cfg autostartConfig) (bool, bool) {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		cfg.approvedKeyPath,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return false, true
	}
	defer func() { _ = key.Close() }()

	data, _, err := key.GetBinaryValue(cfg.valueName)
	if err != nil || len(data) == 0 {
		return false, true
	}

	return true, data[0] == 0x02 || data[0] == 0x06
}

func setAutostartWithConfig(cfg autostartConfig, enable bool, exePath string) error {
	if enable {
		key, _, err := registry.CreateKey(
			registry.CURRENT_USER,
			cfg.runKeyPath,
			registry.SET_VALUE,
		)
		if err != nil {
			return err
		}
		err = key.SetStringValue(cfg.valueName, fmt.Sprintf(`"%s"`, exePath))
		_ = key.Close()
		if err != nil {
			return err
		}

		approvedKey, _, err := registry.CreateKey(
			registry.CURRENT_USER,
			cfg.approvedKeyPath,
			registry.SET_VALUE,
		)
		if err != nil {
			return err
		}
		defer func() { _ = approvedKey.Close() }()

		enabledBlob := []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		return approvedKey.SetBinaryValue(cfg.valueName, enabledBlob)
	}

	if key, err := registry.OpenKey(
		registry.CURRENT_USER,
		cfg.runKeyPath,
		registry.SET_VALUE,
	); err == nil {
		err = key.DeleteValue(cfg.valueName)
		_ = key.Close()
		if err != nil && !errors.Is(err, registry.ErrNotExist) {
			return err
		}
	} else if !errors.Is(err, registry.ErrNotExist) {
		return err
	}

	if approvedKey, err := registry.OpenKey(
		registry.CURRENT_USER,
		cfg.approvedKeyPath,
		registry.SET_VALUE,
	); err == nil {
		_ = approvedKey.DeleteValue(cfg.valueName)
		_ = approvedKey.Close()
	}

	return nil
}

func getAutostartWithConfig(cfg autostartConfig) bool {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		cfg.runKeyPath,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return false
	}
	defer func() { _ = key.Close() }()

	val, _, err := key.GetStringValue(cfg.valueName)
	if err != nil || val == "" {
		return false
	}

	regExe := parseRunCommandExecutable(val)
	if regExe == "" {
		return false
	}
	// Check that the registered target exists on disk rather than requiring an exact
	// os.Executable() match so side-by-side official and dev builds sharing Run\Vinta
	// both reflect that Windows startup is active.
	info, err := os.Stat(regExe)
	if err != nil || info.IsDir() {
		return false
	}

	_, approved := isStartupApprovedEnabled(cfg)
	return approved
}

func syncAutostartWithConfig(cfg autostartConfig, enabled bool, currentExe string) error {
	if !enabled || currentExe == "" {
		return nil
	}

	if exists, approved := isStartupApprovedEnabled(cfg); exists && !approved {
		return nil
	}

	tempDir := filepath.Clean(os.TempDir())
	if tempDir != "" && strings.HasPrefix(strings.ToLower(currentExe), strings.ToLower(tempDir+string(filepath.Separator))) {
		return nil
	}

	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		cfg.runKeyPath,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return setAutostartWithConfig(cfg, true, currentExe)
	}
	val, _, valErr := key.GetStringValue(cfg.valueName)
	_ = key.Close()

	regExe := parseRunCommandExecutable(val)
	if valErr != nil || regExe == "" {
		return setAutostartWithConfig(cfg, true, currentExe)
	}

	if info, statErr := os.Stat(regExe); statErr != nil || info.IsDir() {
		return setAutostartWithConfig(cfg, true, currentExe)
	}

	return nil
}

// SetAutostart registers or unregisters the app in Windows Startup Registry.
func (a *App) SetAutostart(enable bool) error {
	if isTestModeSync() {
		return nil
	}

	cfg := defaultAutostartConfig()
	if !enable {
		return setAutostartWithConfig(cfg, false, "")
	}

	exePath, err := resolveExecutablePath()
	if err != nil {
		return err
	}
	return setAutostartWithConfig(cfg, true, exePath)
}

// GetAutostart checks whether autostart is currently enabled and valid in Windows Registry.
func (a *App) GetAutostart() bool {
	return getAutostartWithConfig(defaultAutostartConfig())
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

	currentExe, err := resolveExecutablePath()
	if err != nil {
		return
	}

	_ = syncAutostartWithConfig(defaultAutostartConfig(), enabled, currentExe)
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
