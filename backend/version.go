// Package backend provides versioning, application runtime metadata, and contributor parsing for Vinta.
package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
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
	appData := os.Getenv("APPDATA")
	if appData != "" {
		seqDir := filepath.Join(appData, "FreeMatrix Studios", "Vinta")
		_ = os.MkdirAll(seqDir, 0755)
		seqFile := filepath.Join(seqDir, "build_seq")
		count := 1
		if data, err := os.ReadFile(seqFile); err == nil {
			trimmed := strings.TrimSpace(string(data))
			if num, err := strconv.Atoi(trimmed); err == nil && num > 0 {
				count = num + 1
			}
		}
		_ = os.WriteFile(seqFile, []byte(strconv.Itoa(count)), 0644)
		year2 := time.Now().Format("06")
		return fmt.Sprintf("%s.%d", year2, count)
	}
	return "26.1"
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
				handle := strings.TrimPrefix(filepath.Base(url), "@")

				cleanRole := getHighestRole(rawRole)

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

// getHighestRole selects the highest priority single role according to the hierarchy:
// Founder > Maintainer > Contributor > Tester
func getHighestRole(rawRole string) string {
	stripHTML := regexp.MustCompile(`<[^>]*>`)
	cleaned := stripHTML.ReplaceAllString(rawRole, "")
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return "Contributor"
	}

	roleRank := map[string]int{
		"founder":     100,
		"maintainer":  80,
		"contributor": 50,
		"code":        50,
		"doc":         50,
		"design":      50,
		"review":      50,
		"tester":      20,
	}

	displayNames := map[string]string{
		"founder":     "Founder",
		"maintainer":  "Maintainer",
		"contributor": "Contributor",
		"code":        "Contributor",
		"doc":         "Contributor",
		"design":      "Contributor",
		"review":      "Contributor",
		"tester":      "Tester",
	}

	roles := strings.Split(cleaned, ",")
	highestRank := -1
	highestRole := ""

	for _, r := range roles {
		r = strings.TrimSpace(strings.ToLower(r))
		if r == "" {
			continue
		}
		rank, exists := roleRank[r]
		if !exists {
			rank = 30 // fallback rank for unknown custom roles
		}
		if rank > highestRank {
			highestRank = rank
			if disp, ok := displayNames[r]; ok {
				highestRole = disp
			} else {
				highestRole = strings.ToUpper(r[:1]) + r[1:]
			}
		}
	}

	if highestRole == "" {
		return "Contributor"
	}
	return highestRole
}
