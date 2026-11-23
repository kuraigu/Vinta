// Package backend provides core application services and state management for Vinta.
package backend

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf16"
)

// WindowsAppItem describes a discovered Start Menu shortcut or UWP application.
type WindowsAppItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Target    string `json:"target"`
	AppID     string `json:"appId"`
	SourceApp string `json:"sourceApp"`
	Kind      string `json:"kind"`
}

var (
	launcherCacheMu sync.Mutex
	launcherFetchMu sync.Mutex
	launcherCache   []WindowsAppItem
)

func shouldSkipShortcutName(lowerName string) bool {
	skipTerms := []string{
		"uninstall", "uninst", "readme", "release notes", "documentation",
		"help", "manual", "license", "changelog", "diagnostic", "troubleshoot",
		"website", "visit ", "donate", "bug report", "developer command prompt",
		"developer powershell", "native tools", "redistributable", "(x86)",
		"recovery drive",
	}
	for _, term := range skipTerms {
		if strings.Contains(lowerName, term) {
			return true
		}
	}
	return false
}

func shouldSkipUWPApp(name, appID string) bool {
	lowerID := strings.ToLower(strings.TrimSpace(appID))
	if lowerID == "" || strings.HasPrefix(lowerID, "http://") || strings.HasPrefix(lowerID, "https://") {
		return true
	}
	allowedWinPrefixes := []string{
		"microsoft.windowsterminal",
		"microsoft.windowscalculator",
		"microsoft.windowsnotepad",
		"microsoft.windowscamera",
		"microsoft.windowsalarms",
		"microsoft.windowssoundrecorder",
		"microsoft.windows.photos",
		"microsoft.windowsstore",
		"microsoft.windowsfeedbackhub",
	}
	if strings.HasPrefix(lowerID, "microsoft.windows.") {
		for _, allowed := range allowedWinPrefixes {
			if strings.HasPrefix(lowerID, allowed) {
				return false
			}
		}
		return true
	}
	blockedPrefixes := []string{
		"microsoft.aad.",
		"microsoft.accountscontrol",
		"microsoft.asynctextservice",
		"microsoft.bioenrollment",
		"microsoft.creddialoghost",
		"microsoft.ecapp",
		"microsoft.lockapp",
		"microsoft.microsoftedgedevtoolsclient",
		"microsoft.ppiprojection",
		"microsoft.win32webviewhost",
		"microsoft.windowsappruntime",
		"microsoft.ui.xaml",
		"microsoft.vclibs",
		"microsoft.net.",
		"ncsiuwpapp",
		"windows.",
		"c5e2524a-ea46-4f67-841f-6a9465d9d515",
		"1527c705-839a-4832-9118-54d4bd6a0c89",
		"e2a4f912-2574-4a75-9bb0-0d023378592b",
		"f46d4000-fd22-4db4-ac8e-4e1ddde828fe",
	}
	for _, prefix := range blockedPrefixes {
		if strings.HasPrefix(lowerID, prefix) {
			return true
		}
	}
	return shouldSkipShortcutName(strings.ToLower(name))
}

// resolveLnkLocalTarget extracts the local target path from a Windows .lnk file if present.
func resolveLnkLocalTarget(lnkPath string) string {
	data, err := os.ReadFile(lnkPath)
	if err != nil || len(data) < 0x4C {
		return ""
	}
	if binary.LittleEndian.Uint32(data[0:4]) != 0x0000004C {
		return ""
	}
	flags := binary.LittleEndian.Uint32(data[0x14:0x18])
	offset := 0x4C
	if flags&0x01 != 0 { // HasLinkTargetIDList
		if offset+2 > len(data) {
			return ""
		}
		idListSize := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
		offset += 2 + idListSize
	}
	if flags&0x02 == 0 || offset+0x1C > len(data) { // HasLinkInfo
		return ""
	}
	linkInfoStart := offset
	linkInfoSize := int(binary.LittleEndian.Uint32(data[linkInfoStart : linkInfoStart+4]))
	headerSize := int(binary.LittleEndian.Uint32(data[linkInfoStart+4 : linkInfoStart+8]))
	infoFlags := binary.LittleEndian.Uint32(data[linkInfoStart+8 : linkInfoStart+12])
	if linkInfoStart+linkInfoSize > len(data) || infoFlags&0x01 == 0 {
		return ""
	}
	if headerSize >= 0x24 && linkInfoStart+0x20 <= len(data) {
		uniOff := int(binary.LittleEndian.Uint32(data[linkInfoStart+0x1C : linkInfoStart+0x20]))
		if uniOff > 0 && linkInfoStart+uniOff < len(data) {
			raw := data[linkInfoStart+uniOff:]
			var u16 []uint16
			for i := 0; i+1 < len(raw); i += 2 {
				ch := binary.LittleEndian.Uint16(raw[i : i+2])
				if ch == 0 {
					break
				}
				u16 = append(u16, ch)
			}
			if len(u16) > 0 {
				return string(utf16.Decode(u16))
			}
		}
	}
	ansiOff := int(binary.LittleEndian.Uint32(data[linkInfoStart+0x10 : linkInfoStart+0x14]))
	if ansiOff > 0 && linkInfoStart+ansiOff < len(data) {
		raw := data[linkInfoStart+ansiOff:]
		for i, b := range raw {
			if b == 0 {
				return string(raw[:i])
			}
		}
	}
	return ""
}

// GetInstalledWindowsApps discovers Start Menu .lnk shortcuts and UWP apps lazily, caching in memory.
func (a *App) GetInstalledWindowsApps(forceRefresh bool) ([]WindowsAppItem, error) {
	launcherCacheMu.Lock()
	if !forceRefresh && len(launcherCache) > 0 {
		cached := make([]WindowsAppItem, len(launcherCache))
		copy(cached, launcherCache)
		launcherCacheMu.Unlock()
		return cached, nil
	}
	launcherCacheMu.Unlock()

	launcherFetchMu.Lock()
	defer launcherFetchMu.Unlock()

	// Double-check cache after acquiring fetch lock in case a concurrent pre-warm just completed.
	launcherCacheMu.Lock()
	if !forceRefresh && len(launcherCache) > 0 {
		cached := make([]WindowsAppItem, len(launcherCache))
		copy(cached, launcherCache)
		launcherCacheMu.Unlock()
		return cached, nil
	}
	launcherCacheMu.Unlock()

	seenNames := make(map[string]bool)
	items := make([]WindowsAppItem, 0, 64)

	var dirs []string
	if programData := strings.TrimSpace(os.Getenv("ProgramData")); programData != "" {
		dirs = append(dirs, filepath.Join(programData, "Microsoft", "Windows", "Start Menu", "Programs"))
	}
	if appData := strings.TrimSpace(os.Getenv("APPDATA")); appData != "" {
		dirs = append(dirs, filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs"))
	}

	for _, root := range dirs {
		if _, err := os.Stat(root); err != nil {
			continue
		}
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d == nil || d.IsDir() {
				return nil
			}
			if !strings.EqualFold(filepath.Ext(d.Name()), ".lnk") {
				return nil
			}
			base := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
			lower := strings.ToLower(strings.TrimSpace(base))
			if lower == "" || shouldSkipShortcutName(lower) || seenNames[lower] {
				return nil
			}
			if localTarget := resolveLnkLocalTarget(path); localTarget != "" {
				if _, statErr := os.Stat(localTarget); os.IsNotExist(statErr) {
					return nil
				}
				exeLower := strings.ToLower(filepath.Base(localTarget))
				if strings.Contains(exeLower, "unins") {
					return nil
				}
			}
			seenNames[lower] = true
			items = append(items, WindowsAppItem{
				ID:        "lnk:" + lower,
				Name:      base,
				Target:    path,
				AppID:     "",
				SourceApp: base,
				Kind:      "shortcut",
			})
			return nil
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	cmd := exec.CommandContext(
		ctx,
		"powershell",
		"-NoProfile",
		"-NonInteractive",
		"-Command",
		"Get-StartApps | Select-Object Name,AppID | ConvertTo-Json -Compress",
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := cmd.Output(); err == nil && len(out) > 0 {
		var uwpApps []struct {
			Name  string `json:"Name"`
			AppID string `json:"AppID"`
		}
		if json.Unmarshal(out, &uwpApps) == nil {
			for _, u := range uwpApps {
				name := strings.TrimSpace(u.Name)
				appID := strings.TrimSpace(u.AppID)
				if name == "" || appID == "" {
					continue
				}
				lower := strings.ToLower(name)
				if seenNames[lower] || shouldSkipUWPApp(name, appID) {
					continue
				}
				// Focus on UWP / packaged apps (containing '!') or remaining StartApps entries
				target := appID
				if strings.Contains(appID, "!") || !strings.Contains(appID, "\\") {
					target = `shell:AppsFolder\` + appID
				}
				seenNames[lower] = true
				kind := "shortcut"
				if strings.Contains(appID, "!") {
					kind = "uwp"
				}
				items = append(items, WindowsAppItem{
					ID:        "uwp:" + appID,
					Name:      name,
					Target:    target,
					AppID:     appID,
					SourceApp: name,
					Kind:      kind,
				})
			}
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	launcherCacheMu.Lock()
	launcherCache = make([]WindowsAppItem, len(items))
	copy(launcherCache, items)
	launcherCacheMu.Unlock()

	return items, nil
}

// LaunchWindowsApp starts a discovered Start Menu shortcut or shell:AppsFolder target.
func (a *App) LaunchWindowsApp(target string) error {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return fmt.Errorf("empty launch target")
	}
	cmd := exec.Command("cmd", "/c", "start", "", trimmed)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}
