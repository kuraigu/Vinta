// Package backend provides core application services and state management for Vinta.
package backend

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var themeIDRegex = regexp.MustCompile(`^[a-z0-9-_]+$`)

// VintaThemeManifest represents the schema definition of a .vintatheme manifest.json.
type VintaThemeManifest struct {
	SchemaVersion int                    `json:"schemaVersion"`
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Author        string                 `json:"author"`
	Version       string                 `json:"version"`
	Description   string                 `json:"description,omitempty"`
	Entry         string                 `json:"entry"`
	Preview       string                 `json:"preview,omitempty"`
	AccentColor   string                 `json:"accentColor,omitempty"`
	Mode          string                 `json:"mode,omitempty"`
	Meta          map[string]interface{} `json:"meta,omitempty"`
}

// ThemeInfo encapsulates manifest data and filesystem origin for frontend consumers.
type ThemeInfo struct {
	Manifest  VintaThemeManifest `json:"manifest"`
	IsBuiltIn bool               `json:"isBuiltIn"`
	Path      string             `json:"path"`
}

// ThemeService manages user-installed theme packages in %APPDATA%/FreeMatrix Studios/Vinta/themes.
type ThemeService struct {
	themesDir string
}

// NewThemeService initializes the theme directory under %APPDATA%/FreeMatrix Studios/Vinta/themes.
func NewThemeService() *ThemeService {
	themesDir := filepath.Join(GetAppDataDir(), "themes")
	_ = os.MkdirAll(themesDir, 0755)
	return &ThemeService{
		themesDir: themesDir,
	}
}

// ListThemes scans the local user themes directory for installed custom themes.
func (s *ThemeService) ListThemes() ([]ThemeInfo, error) {
	var themes []ThemeInfo

	_ = os.MkdirAll(s.themesDir, 0755)
	userEntries, err := os.ReadDir(s.themesDir)
	if err == nil {
		for _, entry := range userEntries {
			fullPath := filepath.Join(s.themesDir, entry.Name())
			if entry.IsDir() {
				manifestPath := filepath.Join(fullPath, "manifest.json")
				data, readErr := os.ReadFile(manifestPath)
				if readErr != nil {
					continue
				}
				var manifest VintaThemeManifest
				if jsonErr := json.Unmarshal(data, &manifest); jsonErr != nil {
					continue
				}
				themes = append(themes, ThemeInfo{
					Manifest:  manifest,
					IsBuiltIn: false,
					Path:      fullPath,
				})
			} else if strings.HasSuffix(strings.ToLower(entry.Name()), ".vintatheme") || strings.HasSuffix(strings.ToLower(entry.Name()), ".zip") {
				manifest, parseErr := readZipManifest(fullPath)
				if parseErr != nil {
					continue
				}
				themes = append(themes, ThemeInfo{
					Manifest:  manifest,
					IsBuiltIn: false,
					Path:      fullPath,
				})
			}
		}
	}

	return themes, nil
}

// GetThemeCSS returns the CSS content for the specified custom theme ID from disk.
func (s *ThemeService) GetThemeCSS(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("theme id cannot be empty")
	}

	themeDir := filepath.Join(s.themesDir, id)
	if dirInfo, statErr := os.Stat(themeDir); statErr == nil && dirInfo.IsDir() {
		manifestPath := filepath.Join(themeDir, "manifest.json")
		data, readErr := os.ReadFile(manifestPath)
		if readErr == nil {
			var manifest VintaThemeManifest
			if jsonErr := json.Unmarshal(data, &manifest); jsonErr == nil {
				entry := manifest.Entry
				if entry == "" {
					entry = "style.css"
				}
				cleanedEntry := filepath.Clean(entry)
				if !strings.HasPrefix(cleanedEntry, "..") && !filepath.IsAbs(cleanedEntry) {
					targetCSS := filepath.Join(themeDir, cleanedEntry)
					if strings.HasPrefix(targetCSS, filepath.Clean(themeDir)+string(filepath.Separator)) {
						cssBytes, cssErr := os.ReadFile(targetCSS)
						if cssErr == nil {
							return string(cssBytes), nil
						}
					}
				}
			}
		}
	}

	userEntries, readErr := os.ReadDir(s.themesDir)
	if readErr == nil {
		for _, entry := range userEntries {
			if entry.IsDir() {
				continue
			}
			lowerName := strings.ToLower(entry.Name())
			if strings.HasSuffix(lowerName, ".vintatheme") || strings.HasSuffix(lowerName, ".zip") {
				fullZipPath := filepath.Join(s.themesDir, entry.Name())
				manifest, parseErr := readZipManifest(fullZipPath)
				if parseErr == nil && manifest.ID == id {
					entryName := manifest.Entry
					if entryName == "" {
						entryName = "style.css"
					}
					cssStr, zipCSSErr := readZipFileContent(fullZipPath, entryName)
					if zipCSSErr == nil {
						return cssStr, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("theme with id %q not found in %s", id, s.themesDir)
}

// ImportThemeDialog opens a native file dialog to import a .vintatheme or .zip package.
func (s *ThemeService) ImportThemeDialog() (*ThemeInfo, error) {
	wailsApp := application.Get()
	if wailsApp == nil {
		return nil, fmt.Errorf("application instance not available")
	}

	dialog := wailsApp.Dialog.OpenFile()
	dialog.SetTitle("Import Vinta Theme")
	dialog.AddFilter("Vinta Theme Packages (*.vintatheme;*.zip)", "*.vintatheme;*.zip")

	selectedPath, err := dialog.PromptForSingleSelection()
	if err != nil {
		return nil, err
	}
	if selectedPath == "" {
		return nil, nil
	}

	manifest, err := readZipManifest(selectedPath)
	if err != nil {
		return nil, fmt.Errorf("invalid theme package: %w", err)
	}

	if !themeIDRegex.MatchString(manifest.ID) {
		return nil, fmt.Errorf("invalid theme ID %q in manifest: must be lowercase alphanumeric with hyphens or underscores", manifest.ID)
	}

	entryName := manifest.Entry
	if entryName == "" {
		entryName = "style.css"
	}
	if _, err := readZipFileContent(selectedPath, entryName); err != nil {
		return nil, fmt.Errorf("theme entry stylesheet %q not found in archive", entryName)
	}

	destDir := filepath.Join(s.themesDir, manifest.ID)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := extractZipArchive(selectedPath, destDir); err != nil {
		return nil, fmt.Errorf("failed to unpack theme archive: %w", err)
	}

	return &ThemeInfo{
		Manifest:  manifest,
		IsBuiltIn: false,
		Path:      destDir,
	}, nil
}

// OpenThemesFolder launches Windows Explorer targeting %APPDATA%/FreeMatrix Studios/Vinta/themes.
func (s *ThemeService) OpenThemesFolder() error {
	_ = os.MkdirAll(s.themesDir, 0755)
	return exec.Command("explorer.exe", s.themesDir).Start()
}

func readZipManifest(archivePath string) (VintaThemeManifest, error) {
	var manifest VintaThemeManifest
	content, err := readZipFileContent(archivePath, "manifest.json")
	if err != nil {
		return manifest, err
	}
	if err := json.Unmarshal([]byte(content), &manifest); err != nil {
		return manifest, err
	}
	if manifest.ID == "" || manifest.Name == "" {
		return manifest, fmt.Errorf("manifest must contain non-empty id and name")
	}
	return manifest, nil
}

func readZipFileContent(archivePath, relativePath string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = r.Close() }()

	targetClean := filepath.Clean(relativePath)
	for _, f := range r.File {
		cleanName := filepath.Clean(f.Name)
		if cleanName == targetClean || cleanName == strings.ReplaceAll(targetClean, "\\", "/") {
			rc, openErr := f.Open()
			if openErr != nil {
				return "", openErr
			}
			bytes, readErr := io.ReadAll(rc)
			_ = rc.Close()
			if readErr != nil {
				return "", readErr
			}
			return string(bytes), nil
		}
	}
	return "", fmt.Errorf("file %q not found in archive", relativePath)
}

func extractZipArchive(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	cleanDest := filepath.Clean(destDir)

	for _, f := range r.File {
		cleanName := filepath.Clean(f.Name)
		if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) {
			return fmt.Errorf("illegal file path in archive: %s", f.Name)
		}

		targetPath := filepath.Join(cleanDest, cleanName)
		if !strings.HasPrefix(targetPath, cleanDest+string(filepath.Separator)) && targetPath != cleanDest {
			return fmt.Errorf("archive entry escapes destination: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			_ = outFile.Close()
			return err
		}

		_, copyErr := io.Copy(outFile, rc)
		_ = rc.Close()
		_ = outFile.Close()
		if copyErr != nil {
			return copyErr
		}
	}

	return nil
}
