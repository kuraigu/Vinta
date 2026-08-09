// Package backend provides versioning, application runtime metadata, and contributor parsing for Vinta.
package backend

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// AppVersion defines the current application version string. Do not change this version number manually in source files.
// Version increments must remain manual and be performed intentionally during release packaging procedures.
const AppVersion = "0.1.0"

// Build metadata variables injected during release compilation via -ldflags.
var (
	GitCommit = "dev"
	BuildType = "Development"
	BuildTime = "Unknown"
)

// AppInfo represents structured application build, authenticity, and runtime metadata.
type AppInfo struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	GitCommit       string `json:"gitCommit"`
	BuildType       string `json:"buildType"`
	BuildTime       string `json:"buildTime"`
	IsOfficialBuild bool   `json:"isOfficialBuild"`
	GoVersion       string `json:"goVersion"`
	Arch            string `json:"arch"`
	OS              string `json:"os"`
	WailsVersion    string `json:"wailsVersion"`
	RepoURL         string `json:"repoUrl"`
}

// Contributor represents a project contributor parsed dynamically from docs/CONTRIBUTORS.md.
type Contributor struct {
	Name   string `json:"name"`
	Handle string `json:"handle"`
	Role   string `json:"role"`
	Avatar string `json:"avatar"`
	URL    string `json:"url"`
}

func getBuildTime() string {
	if BuildTime != "" && BuildTime != "Unknown" {
		return BuildTime
	}
	if execPath, err := os.Executable(); err == nil {
		if fi, err := os.Stat(execPath); err == nil {
			return fi.ModTime().Format("20060102150405")
		}
	}
	return "Development"
}

// GetAppInfo returns application versioning, build verification, and system details.
func (a *App) GetAppInfo() AppInfo {
	isOfficial := BuildType == "Official" || BuildType == "Release"
	return AppInfo{
		Name:            "Vinta",
		Version:         AppVersion,
		GitCommit:       GitCommit,
		BuildType:       BuildType,
		BuildTime:       getBuildTime(),
		IsOfficialBuild: isOfficial,
		GoVersion:       runtime.Version(),
		Arch:            runtime.GOARCH,
		OS:              runtime.GOOS,
		WailsVersion:    "v2.13.0",
		RepoURL:         "https://github.com/kuraigu/Vinta",
	}
}

// GetContributors dynamically reads and parses docs/CONTRIBUTORS.md to render the contributor list.
func (a *App) GetContributors() []Contributor {
	var contributors []Contributor
	seenMap := make(map[string]bool)

	// Try reading docs/CONTRIBUTORS.md relative to executable or working directory
	possiblePaths := []string{
		filepath.Join(".", "docs", "CONTRIBUTORS.md"),
		filepath.Join("..", "docs", "CONTRIBUTORS.md"),
	}

	var content string
	for _, p := range possiblePaths {
		if b, err := os.ReadFile(p); err == nil {
			content = string(b)
			break
		}
	}

	if content != "" {
		// 1. Parse All-Contributors HTML table cells from All Contributors list section
		cellRegex := regexp.MustCompile(`(?s)<td[^>]*>.*?<a[ \t]+href="([^"]+)"[^>]*>.*?<img[ \t]+src="([^"]+)"[^>]*alt="([^"]*)"[^>]*>.*?<sub><b>([^<]+)</b></sub>.*?</a>(?:<br\s*/?>)?\s*(.*?)\s*</td>`)
		cellMatches := cellRegex.FindAllStringSubmatch(content, -1)

		for _, match := range cellMatches {
			if len(match) >= 6 {
				url := strings.TrimSpace(match[1])
				avatar := strings.TrimSpace(match[2])
				name := strings.TrimSpace(match[4])
				rawRole := strings.TrimSpace(match[5])
				handle := "@" + filepath.Base(url)

				// Clean HTML tags from role string
				stripHTML := regexp.MustCompile(`<[^>]*>`)
				cleanRole := stripHTML.ReplaceAllString(rawRole, "")
				cleanRole = strings.TrimSpace(cleanRole)

				if cleanRole == "" {
					cleanRole = "Contributor"
				} else {
					roles := strings.Split(cleanRole, ",")
					for i, r := range roles {
						r = strings.TrimSpace(r)
						if len(r) > 0 {
							roles[i] = strings.ToUpper(r[:1]) + r[1:]
						}
					}
					cleanRole = strings.Join(roles, ", ")
				}

				// Safety bounds: cap lengths to prevent memory or layout issues from long input strings
				if len(name) > 60 {
					name = name[:57] + "..."
				}
				if len(handle) > 30 {
					handle = handle[:27] + "..."
				}
				if len(cleanRole) > 40 {
					cleanRole = cleanRole[:37] + "..."
				}

				if !seenMap[handle] {
					seenMap[handle] = true
					contributors = append(contributors, Contributor{
						Name:   name,
						Handle: handle,
						Role:   cleanRole,
						Avatar: avatar,
						URL:    url,
					})
				}
			}
		}
	}

	return contributors
}
