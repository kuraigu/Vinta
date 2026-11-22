// Package backend provides diagnostic gathering and issue reporting for Vinta.
package backend

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// SystemDiagnostics encapsulates sanitized environment and system metadata.
type SystemDiagnostics struct {
	AppVersion     string  `json:"appVersion"`
	GitCommit      string  `json:"gitCommit"`
	BuildType      string  `json:"buildType"`
	OSName         string  `json:"osName"`
	OSVersion      string  `json:"osVersion"`
	WindowsBuild   string  `json:"windowsBuild"`
	Architecture   string  `json:"architecture"`
	DisplayCount   int     `json:"displayCount"`
	PrimaryScale   float64 `json:"primaryScale"`
	WebViewVersion string  `json:"webViewVersion"`
}

// GetSystemDiagnostics gathers non-sensitive runtime diagnostics.
func (a *App) GetSystemDiagnostics() SystemDiagnostics {
	osName := "Windows"
	osVersion := ""
	winBuild := ""

	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE); err == nil {
		if prod, _, err := k.GetStringValue("ProductName"); err == nil && prod != "" {
			osName = prod
		}
		if disp, _, err := k.GetStringValue("DisplayVersion"); err == nil && disp != "" {
			osVersion = disp
		}
		if bld, _, err := k.GetStringValue("CurrentBuild"); err == nil && bld != "" {
			winBuild = bld
		} else if bldNum, _, err := k.GetStringValue("CurrentBuildNumber"); err == nil && bldNum != "" {
			winBuild = bldNum
		}
		_ = k.Close()
	}

	webViewVer := "Unknown"
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`, registry.QUERY_VALUE); err == nil {
		if pv, _, err := k.GetStringValue("pv"); err == nil && pv != "" {
			webViewVer = pv
		}
		_ = k.Close()
	}

	displayCount := 1
	if count, _, _ := user32.NewProc("GetSystemMetrics").Call(80); count > 0 {
		displayCount = int(count)
	}

	primaryScale := a.scale
	if primaryScale <= 0 {
		primaryScale = 1.0
	}

	return SystemDiagnostics{
		AppVersion:     AppVersion,
		GitCommit:      GitCommit,
		BuildType:      BuildType,
		OSName:         osName,
		OSVersion:      osVersion,
		WindowsBuild:   winBuild,
		Architecture:   runtime.GOARCH,
		DisplayCount:   displayCount,
		PrimaryScale:   primaryScale,
		WebViewVersion: webViewVer,
	}
}

// formatDesktopDiagnostics returns a concise markdown list of sanitized runtime diagnostics.
func (a *App) formatDesktopDiagnostics() string {
	diag := a.GetSystemDiagnostics()
	var sb strings.Builder
	fmt.Fprintf(&sb, "- **Vinta Version**: %s (`%s`)\n", diag.AppVersion, diag.GitCommit)
	fmt.Fprintf(&sb, "- **Build Channel**: %s\n", diag.BuildType)
	fmt.Fprintf(&sb, "- **Operating System**: %s (Build %s, Version %s)\n", diag.OSName, diag.WindowsBuild, diag.OSVersion)
	fmt.Fprintf(&sb, "- **Architecture**: %s\n", diag.Architecture)
	fmt.Fprintf(&sb, "- **Displays**: %d screen(s), Primary Scale: %.2fx\n", diag.DisplayCount, diag.PrimaryScale)
	fmt.Fprintf(&sb, "- **WebView2 Runtime**: %s", diag.WebViewVersion)
	return sb.String()
}

func formatIssueTitle(prefix, raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return prefix + " "
	}
	if strings.HasPrefix(strings.ToLower(trimmed), strings.ToLower(prefix)) {
		return trimmed
	}
	return prefix + " " + trimmed
}

// OpenFeedbackIssue launches the default browser with the consolidated bug report template pre-filled.
func (a *App) OpenFeedbackIssue(title, describe, reproduce string) error {
	diag := a.formatDesktopDiagnostics()
	desc := strings.TrimSpace(describe)
	repro := strings.TrimSpace(reproduce)

	issueURL := fmt.Sprintf(
		"https://github.com/kuraigu/Vinta/issues/new?template=bug_report.yml&title=%s&describe=%s&reproduce=%s&desktop=%s",
		url.QueryEscape(formatIssueTitle("[Bug]", title)),
		url.QueryEscape(desc),
		url.QueryEscape(repro),
		url.QueryEscape(diag),
	)
	return a.OpenURL(issueURL)
}

// OpenFeatureRequest launches the default browser with the feature request template pre-filled.
func (a *App) OpenFeatureRequest(title, describe, useCase string) error {
	desc := strings.TrimSpace(describe)
	uc := strings.TrimSpace(useCase)

	issueURL := fmt.Sprintf(
		"https://github.com/kuraigu/Vinta/issues/new?template=feature_request.yml&title=%s&describe=%s&use_case=%s",
		url.QueryEscape(formatIssueTitle("[Enhancement]", title)),
		url.QueryEscape(desc),
		url.QueryEscape(uc),
	)
	return a.OpenURL(issueURL)
}
