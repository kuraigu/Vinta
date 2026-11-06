// Package backend provides core application services and state management for Vinta.
package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetLocalesDir returns the local drop-in directory for community translations:
// %APPDATA%/FreeMatrix Studios/Vinta/locales.
func GetLocalesDir() string {
	return filepath.Join(GetAppDataDir(), "locales")
}

// OpenLocalesFolder launches Windows Explorer targeting %APPDATA%/FreeMatrix Studios/Vinta/locales.
func (a *App) OpenLocalesFolder() error {
	localesDir := GetLocalesDir()
	_ = os.MkdirAll(localesDir, 0755)
	return exec.Command("explorer.exe", localesDir).Start()
}

// GetCustomLocales scans %APPDATA%/FreeMatrix Studios/Vinta/locales/ for *.json files
// and returns a map of locale identifier to raw JSON string.
func (a *App) GetCustomLocales() (map[string]string, error) {
	localesDir := GetLocalesDir()
	_ = os.MkdirAll(localesDir, 0755)

	entries, err := os.ReadDir(localesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read locales directory: %w", err)
	}

	result := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".json") {
			localeID := strings.TrimSuffix(name, filepath.Ext(name))
			content, readErr := os.ReadFile(filepath.Join(localesDir, name))
			if readErr == nil && len(content) > 0 {
				result[localeID] = string(content)
			}
		}
	}

	return result, nil
}
