// Package backend provides the core application logic for Vinta.
package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"golang.org/x/sys/windows/registry"

	"Vinta/backend/services"
)

var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	procGetWindowLong            = user32.NewProc("GetWindowLongW")
	procSetWindowLong            = user32.NewProc("SetWindowLongW")
	procSetWindowLongPtr         = user32.NewProc("SetWindowLongPtrW")
	procFindWindow               = user32.NewProc("FindWindowW")
	procGetDesktopWindow         = user32.NewProc("GetDesktopWindow")
	procGetWindow                = user32.NewProc("GetWindow")
	procGetWindowThreadProcessID = user32.NewProc("GetWindowThreadProcessId")
	procGetClassName             = user32.NewProc("GetClassNameW")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procSetWinEventHook          = user32.NewProc("SetWinEventHook")
	procEnumDisplayMonitors      = user32.NewProc("EnumDisplayMonitors")
	procGetMonitorInfoW          = user32.NewProc("GetMonitorInfoW")
	procGetWindowRect            = user32.NewProc("GetWindowRect")
	procMonitorFromWindow        = user32.NewProc("MonitorFromWindow")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procIsIconic                 = user32.NewProc("IsIconic")
	procGetAncestor              = user32.NewProc("GetAncestor")
	procGetCursorPos             = user32.NewProc("GetCursorPos")
	procScreenToClient           = user32.NewProc("ScreenToClient")

	shell32                          = syscall.NewLazyDLL("shell32.dll")
	procSHQueryUserNotificationState = shell32.NewProc("SHQueryUserNotificationState")

	ole32              = syscall.NewLazyDLL("ole32.dll")
	procCoInitializeEx = ole32.NewProc("CoInitializeEx")
	procCoUninitialize = ole32.NewProc("CoUninitialize")

	gdi32                             = syscall.NewLazyDLL("gdi32.dll")
	procSetWindowRgn                  = user32.NewProc("SetWindowRgn")
	procD3DKMTCheckExclusiveOwnership = gdi32.NewProc("D3DKMTCheckExclusiveOwnership")

	psapi                         = syscall.NewLazyDLL("psapi.dll")
	procEmptyWorkingSet           = psapi.NewProc("EmptyWorkingSet")
	kernel32                      = syscall.NewLazyDLL("kernel32.dll")
	procGetCurrentProcess         = kernel32.NewProc("GetCurrentProcess")
	procSetProcessInformation     = kernel32.NewProc("SetProcessInformation")
	procOpenProcess               = kernel32.NewProc("OpenProcess")
	procCloseHandle               = kernel32.NewProc("CloseHandle")
	procQueryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageNameW")

	winmm               = syscall.NewLazyDLL("winmm.dll")
	procTimeBeginPeriod = winmm.NewProc("timeBeginPeriod")
)

const (
	//nolint:revive // Windows API constants
	GWL_EXSTYLE = -20
	//nolint:revive // Windows API constants
	GWLP_WNDPROC = -4
	//nolint:revive // Windows API constants
	WM_ACTIVATE = 0x0006
	//nolint:revive // Windows API constants
	WM_MOUSEACTIVATE = 0x0021
	//nolint:revive // Windows API constants
	WM_NCACTIVATE = 0x0086
	//nolint:revive // Windows API constants
	WM_PARENTNOTIFY = 0x0210
	//nolint:revive // Windows API constants
	MA_ACTIVATE = 1
	//nolint:revive // Windows API constants
	MA_NOACTIVATE = 3
	//nolint:revive // Windows API constants
	WS_EX_TOPMOST = 0x00000008
	//nolint:revive // Windows API constants
	WS_EX_LAYERED = 0x00080000
	//nolint:revive // Windows API constants
	WS_EX_APPWINDOW = 0x00040000
	//nolint:revive // Windows API constants
	WS_EX_TOOLWINDOW = 0x00000080
	//nolint:revive // Windows API constants
	WS_EX_NOACTIVATE = 0x08000000
	//nolint:revive // Windows API constants
	WS_EX_TRANSPARENT    = 0x00000020
	OutsidePadding       = 8
	HiddenOutsidePadding = 2
	rgnOr                = 2

	//nolint:revive // Windows API constants
	GW_CHILD = 5
	//nolint:revive // Windows API constants
	GW_HWNDNEXT = 2
	//nolint:revive // Windows API constants
	EVENT_SYSTEM_FOREGROUND = 0x0003
	//nolint:revive // Windows API constants
	EVENT_SYSTEM_MOVESIZEEND = 0x000B
	//nolint:revive // Windows API constants
	WINEVENT_OUTOFCONTEXT = 0x0000

	//nolint:revive // Windows API constants
	MONITOR_DEFAULTTONEAREST = 2
	//nolint:revive // Windows API constants
	GWL_STYLE = -16
	//nolint:revive // Windows API constants
	WS_CAPTION = 0x00C00000
	//nolint:revive // Windows API constants
	SW_HIDE = 0
	//nolint:revive // Windows API constants
	SW_SHOWNA = 8

	//nolint:revive // Windows API constants
	SWP_NOSIZE = 0x0001
	//nolint:revive // Windows API constants
	SWP_NOMOVE = 0x0002
	//nolint:revive // Windows API constants
	SWP_NOZORDER = 0x0004
	//nolint:revive // Windows API constants
	SWP_NOACTIVATE = 0x0010
	//nolint:revive // Windows API constants
	SWP_FRAMECHANGED = 0x0020
	//nolint:revive // Windows API constants
	SWP_SHOWWINDOW = 0x0040
	//nolint:revive // Windows API constants
	SWP_NOCOPYBITS = 0x0100
	//nolint:revive // Windows API constants
	SWP_NOOWNERZORDER = 0x0200
	//nolint:revive // Windows API constants
	SWP_NOSENDCHANGING = 0x0400
	//nolint:revive // Windows API constants
	SWP_DEFERERASE = 0x2000

	// QUERY_USER_NOTIFICATION_STATE enum
	//nolint:revive // Windows API constants
	QUNS_NOT_PRESENT = 1
	//nolint:revive // Windows API constants
	QUNS_BUSY = 2
	//nolint:revive // Windows API constants
	QUNS_RUNNING_D3D_FULL_SCREEN = 3
	//nolint:revive // Windows API constants
	QUNS_PRESENTATION_MODE = 4
	//nolint:revive // Windows API constants
	QUNS_ACCEPTS_NOTIFICATIONS = 5
	//nolint:revive // Windows API constants
	QUNS_QUIET_TIME = 6
	//nolint:revive // Windows API constants
	QUNS_APP = 7

	//nolint:revive // Windows API constants
	SM_CXSCREEN = 0
	//nolint:revive // Windows API constants
	SM_CYSCREEN = 1
	//nolint:revive // Windows API constants
	SM_XVIRTUALSCREEN = 76
	//nolint:revive // Windows API constants
	SM_YVIRTUALSCREEN = 77
	//nolint:revive // Windows API constants
	SM_CXVIRTUALSCREEN = 78
	//nolint:revive // Windows API constants
	SM_CYVIRTUALSCREEN = 79

	// Windows Power Throttling / EcoQoS constants (Windows 10 1709+ / Windows 11)
	//nolint:revive // Windows API constants
	processPowerThrottling = 4
	//nolint:revive // Windows API constants
	processPowerThrottlingCurrentVersion = 1
	//nolint:revive // Windows API constants
	processPowerThrottlingExecutionSpeed = 0x1
	//nolint:revive // Windows API constants
	processPowerThrottlingIgnoreTimerResolution = 0x4
)

type processPowerThrottlingState struct {
	Version     uint32
	ControlMask uint32
	StateMask   uint32
}

type rect32 struct {
	Left, Top, Right, Bottom int32
}

type point32 struct {
	X, Y int32
}

// HitRect defines a physical client rectangle for hit testing.
type HitRect struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type pillHitBounds struct {
	left, top, right, bottom int32
}

type monitorInfoW struct {
	CbSize    uint32
	RcMonitor rect32
	RcWork    rect32
	DwFlags   uint32
}

// MonitorInfo describes a connected physical display's boundaries and properties.
type MonitorInfo struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Left       int    `json:"left"`
	Top        int    `json:"top"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	IsPrimary  bool   `json:"isPrimary"`
	IsPortrait bool   `json:"isPortrait"`
}

// DisplayLayout captures multi-monitor geometry and the composite virtual desktop space.
type DisplayLayout struct {
	VirtualLeft    int           `json:"virtualLeft"`
	VirtualTop     int           `json:"virtualTop"`
	VirtualWidth   int           `json:"virtualWidth"`
	VirtualHeight  int           `json:"virtualHeight"`
	PrimaryCenter  int           `json:"primaryCenter"`
	CurrentMonitor MonitorInfo   `json:"currentMonitor"`
	Monitors       []MonitorInfo `json:"monitors"`
}

var (
	cachedDisplayLayout DisplayLayout
	cachedDisplayMu     sync.RWMutex
	lastDisplayFetch    time.Time
)

// GetDisplayMetrics returns the virtual screen boundaries and monitor list for window placement.
func GetDisplayMetrics() (vLeft, vTop, vWidth, vHeight, pCenter int, monitors []MonitorInfo) {
	procGetSystemMetrics := user32.NewProc("GetSystemMetrics")
	vx, _, _ := procGetSystemMetrics.Call(SM_XVIRTUALSCREEN)
	vy, _, _ := procGetSystemMetrics.Call(SM_YVIRTUALSCREEN)
	vw, _, _ := procGetSystemMetrics.Call(SM_CXVIRTUALSCREEN)
	vh, _, _ := procGetSystemMetrics.Call(SM_CYVIRTUALSCREEN)

	currVw := int(vw)
	currVh := int(vh)

	cachedDisplayMu.RLock()
	if time.Since(lastDisplayFetch) < 10*time.Second && cachedDisplayLayout.VirtualWidth > 0 &&
		cachedDisplayLayout.VirtualWidth == currVw && cachedDisplayLayout.VirtualHeight == currVh {
		layout := cachedDisplayLayout
		cachedDisplayMu.RUnlock()
		return layout.VirtualLeft, layout.VirtualTop, layout.VirtualWidth, layout.VirtualHeight, layout.PrimaryCenter, layout.Monitors
	}
	cachedDisplayMu.RUnlock()

	cachedDisplayMu.Lock()
	defer cachedDisplayMu.Unlock()

	if time.Since(lastDisplayFetch) < 10*time.Second && cachedDisplayLayout.VirtualWidth > 0 &&
		cachedDisplayLayout.VirtualWidth == currVw && cachedDisplayLayout.VirtualHeight == currVh {
		layout := cachedDisplayLayout
		return layout.VirtualLeft, layout.VirtualTop, layout.VirtualWidth, layout.VirtualHeight, layout.PrimaryCenter, layout.Monitors
	}

	vLeft = int(int32(vx))
	vTop = int(int32(vy))
	vWidth = currVw
	vHeight = currVh

	pw, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	ph, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)
	primaryW := int(pw)
	if primaryW <= 0 {
		primaryW = 1920
	}
	if vWidth <= 0 {
		vWidth = primaryW
	}
	if vHeight <= 0 {
		vHeight = int(ph)
		if vHeight <= 0 {
			vHeight = 1080
		}
	}

	pCenter = (0 - vLeft) + (primaryW / 2)

	if procEnumDisplayMonitors.Find() == nil && procGetMonitorInfoW.Find() == nil {
		enumMonitorsMu.Lock()
		enumMonitorsID = 0
		enumMonitorsList = nil
		cb := getEnumMonitorsCallback()
		_, _, _ = procEnumDisplayMonitors.Call(0, 0, cb, 0)
		monitors = enumMonitorsList
		enumMonitorsMu.Unlock()
	}

	if len(monitors) == 0 {
		monitors = append(monitors, MonitorInfo{
			ID:         1,
			Name:       "Primary Display",
			Left:       0,
			Top:        0,
			Width:      primaryW,
			Height:     vHeight,
			IsPrimary:  true,
			IsPortrait: vHeight > primaryW,
		})
	}

	cachedDisplayLayout = DisplayLayout{
		VirtualLeft:   vLeft,
		VirtualTop:    vTop,
		VirtualWidth:  vWidth,
		VirtualHeight: vHeight,
		PrimaryCenter: pCenter,
		Monitors:      monitors,
	}
	lastDisplayFetch = time.Now()
	return
}

// resolveCurrentMonitor finds the display monitor corresponding to screenCenterX.
func resolveCurrentMonitor(screenCenterX int, monitors []MonitorInfo) MonitorInfo {
	for _, mon := range monitors {
		if screenCenterX >= mon.Left && screenCenterX < mon.Left+mon.Width {
			return mon
		}
	}
	for _, mon := range monitors {
		if mon.IsPrimary {
			return mon
		}
	}
	if len(monitors) > 0 {
		return monitors[0]
	}
	return MonitorInfo{
		ID:         1,
		Name:       "Primary Display",
		Left:       0,
		Top:        0,
		Width:      1920,
		Height:     1080,
		IsPrimary:  true,
		IsPortrait: false,
	}
}

var (
	cachedWindowHwnd uintptr
	cachedHwndMu     sync.Mutex
	procIsWindow     = user32.NewProc("IsWindow")
)

func getMyWindowHwnd() uintptr {
	cachedHwndMu.Lock()
	if cachedWindowHwnd != 0 {
		ret, _, _ := procIsWindow.Call(cachedWindowHwnd)
		if ret != 0 {
			h := cachedWindowHwnd
			cachedHwndMu.Unlock()
			return h
		}
		cachedWindowHwnd = 0
	}
	cachedHwndMu.Unlock()

	titlePtr, _ := syscall.UTF16PtrFromString("Vinta")
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd != 0 {
		cachedHwndMu.Lock()
		cachedWindowHwnd = hwnd
		cachedHwndMu.Unlock()
		return hwnd
	}

	myPid := uint32(os.Getpid())
	desktop, _, _ := procGetDesktopWindow.Call()
	child, _, _ := procGetWindow.Call(desktop, uintptr(GW_CHILD))
	for child != 0 {
		var pid uint32
		_, _, _ = procGetWindowThreadProcessID.Call(child, uintptr(unsafe.Pointer(&pid)))
		if pid == myPid {
			var textBuf [256]uint16
			_, _, _ = procGetWindowTextW.Call(child, uintptr(unsafe.Pointer(&textBuf[0])), 256)
			windowTitle := syscall.UTF16ToString(textBuf[:])

			// Strictly accept only the window explicitly titled Vinta.
			// Internal Wails synchronization windows, IME windows, and helper windows are excluded.
			if windowTitle == "Vinta" {
				cachedHwndMu.Lock()
				cachedWindowHwnd = child
				cachedHwndMu.Unlock()
				return child
			}
		}
		child, _, _ = procGetWindow.Call(child, uintptr(GW_HWNDNEXT))
	}
	return 0
}

// windowCoversMonitor determines whether the foreground window covers the monitor
// where Vinta is currently hosted.
func windowCoversMonitor(fgHwnd uintptr, vintaMon MonitorInfo) (bool, monitorInfoW) {
	var mi monitorInfoW
	if fgHwnd == 0 {
		return false, mi
	}

	if isSnippingTool(fgHwnd) || isScreenClippingHost(fgHwnd) {
		return false, mi
	}

	var classBuf [256]uint16
	_, _, _ = procGetClassName.Call(fgHwnd, uintptr(unsafe.Pointer(&classBuf[0])), 256)
	className := syscall.UTF16ToString(classBuf[:])
	if className == "WorkerW" || className == "Progman" ||
		className == "Shell_TrayWnd" || className == "Shell_SecondaryTrayWnd" ||
		className == "Windows.UI.Core.CoreWindow" {
		return false, mi
	}

	hMonFg, _, _ := procMonitorFromWindow.Call(fgHwnd, uintptr(MONITOR_DEFAULTTONEAREST))
	if hMonFg == 0 {
		return false, mi
	}

	mi.CbSize = uint32(unsafe.Sizeof(mi))
	res, _, _ := procGetMonitorInfoW.Call(hMonFg, uintptr(unsafe.Pointer(&mi)))
	if res == 0 {
		return false, mi
	}

	// The foreground window must be located on the same physical monitor as Vinta
	if int(mi.RcMonitor.Left) != vintaMon.Left || int(mi.RcMonitor.Top) != vintaMon.Top {
		return false, mi
	}

	var rc rect32
	res, _, _ = procGetWindowRect.Call(fgHwnd, uintptr(unsafe.Pointer(&rc)))
	if res == 0 {
		return false, mi
	}

	covers := rc.Left <= mi.RcMonitor.Left+4 &&
		rc.Top <= mi.RcMonitor.Top+4 &&
		rc.Right >= mi.RcMonitor.Right-4 &&
		rc.Bottom >= mi.RcMonitor.Bottom-4

	return covers, mi
}

// hasExclusiveDisplayOwnership queries the Windows graphics kernel thunk to determine
// whether any graphics context currently holds exclusive hardware ownership over a display output.
func hasExclusiveDisplayOwnership() bool {
	if procD3DKMTCheckExclusiveOwnership.Find() != nil {
		return false
	}
	res, _, _ := procD3DKMTCheckExclusiveOwnership.Call()
	return res != 0
}

// logFullscreenDebug records forensic diagnostics into %APPDATA%/FreeMatrix Studios/Vinta/fullscreen_debug.log.
func logFullscreenDebug(format string, args ...interface{}) {
	dir := GetAppDataDir()
	logPath := filepath.Join(dir, "fullscreen_debug.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	msg := fmt.Sprintf("[%s] ", time.Now().Format("15:04:05.000")) + fmt.Sprintf(format, args...) + "\n"
	_, _ = f.WriteString(msg)
}

// isSnippingTool checks whether the given window belongs to Windows Snipping Tool, ScreenClippingHost, or related clipping hosts.
func isSnippingTool(hwnd uintptr) bool {
	if hwnd == 0 {
		return false
	}

	checkHwnd := func(h uintptr) bool {
		var pid uint32
		_, _, _ = procGetWindowThreadProcessID.Call(h, uintptr(unsafe.Pointer(&pid)))
		if pid != 0 {
			hProc, _, _ := procOpenProcess.Call(0x1000, 0, uintptr(pid)) // PROCESS_QUERY_LIMITED_INFORMATION
			if hProc == 0 {
				hProc, _, _ = procOpenProcess.Call(0x0400, 0, uintptr(pid)) // PROCESS_QUERY_INFORMATION
			}
			if hProc != 0 {
				var buf [1024]uint16
				size := uint32(len(buf))
				res, _, _ := procQueryFullProcessImageName.Call(hProc, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
				if res == 0 {
					size = uint32(len(buf))
					res, _, _ = procQueryFullProcessImageName.Call(hProc, 1, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
				}
				_, _, _ = procCloseHandle.Call(hProc)
				if res != 0 {
					exePath := syscall.UTF16ToString(buf[:size])
					baseName := strings.ToLower(filepath.Base(exePath))
					if baseName == "screenclippinghost.exe" || baseName == "snippingtool.exe" ||
						baseName == "screensketch.exe" || baseName == "snippingtoolapp.exe" {
						return true
					}
				}
			}
		}

		var titleBuf [256]uint16
		_, _, _ = procGetWindowTextW.Call(h, uintptr(unsafe.Pointer(&titleBuf[0])), 256)
		title := strings.ToLower(syscall.UTF16ToString(titleBuf[:]))
		if strings.Contains(title, "snipping tool") || strings.Contains(title, "screen snipping") ||
			strings.Contains(title, "snip & sketch") || strings.Contains(title, "screen clipping") {
			return true
		}

		var classBuf [256]uint16
		_, _, _ = procGetClassName.Call(h, uintptr(unsafe.Pointer(&classBuf[0])), 256)
		className := syscall.UTF16ToString(classBuf[:])
		if className == "ScreenClippingWindow" || strings.Contains(className, "ScreenClipping") {
			return true
		}

		return false
	}

	if checkHwnd(hwnd) {
		return true
	}
	root, _, _ := procGetAncestor.Call(hwnd, 2) // GA_ROOT = 2
	if root != 0 && root != hwnd {
		return checkHwnd(root)
	}
	return false
}

// isScreenClippingHost checks whether the window belongs specifically to the active screen clipping host session.
func isScreenClippingHost(hwnd uintptr) bool {
	if hwnd == 0 {
		return false
	}

	var pid uint32
	_, _, _ = procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid != 0 {
		hProc, _, _ := procOpenProcess.Call(0x1000, 0, uintptr(pid))
		if hProc == 0 {
			hProc, _, _ = procOpenProcess.Call(0x0400, 0, uintptr(pid))
		}
		if hProc != 0 {
			var buf [1024]uint16
			size := uint32(len(buf))
			res, _, _ := procQueryFullProcessImageName.Call(hProc, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
			if res == 0 {
				size = uint32(len(buf))
				res, _, _ = procQueryFullProcessImageName.Call(hProc, 1, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
			}
			_, _, _ = procCloseHandle.Call(hProc)
			if res != 0 {
				exePath := syscall.UTF16ToString(buf[:size])
				baseName := strings.ToLower(filepath.Base(exePath))
				if baseName == "screenclippinghost.exe" {
					return true
				}
			}
		}
	}

	var classBuf [256]uint16
	_, _, _ = procGetClassName.Call(hwnd, uintptr(unsafe.Pointer(&classBuf[0])), 256)
	className := syscall.UTF16ToString(classBuf[:])
	return className == "ScreenClippingWindow" || strings.Contains(className, "ScreenClipping")
}

// getWindowAppName extracts the friendly application name or window title of a given window.
func getWindowAppName(hwnd uintptr) string {
	if hwnd == 0 {
		return ""
	}

	var titleBuf [256]uint16
	res, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&titleBuf[0])), 256)
	if res > 0 {
		title := strings.TrimSpace(syscall.UTF16ToString(titleBuf[:res]))
		if title != "" {
			return title
		}
	}

	var pid uint32
	_, _, _ = procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid != 0 {
		hProc, _, _ := procOpenProcess.Call(0x1000, 0, uintptr(pid))
		if hProc == 0 {
			hProc, _, _ = procOpenProcess.Call(0x0400, 0, uintptr(pid))
		}
		if hProc != 0 {
			var buf [1024]uint16
			size := uint32(len(buf))
			resQuery, _, _ := procQueryFullProcessImageName.Call(hProc, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
			if resQuery == 0 {
				size = uint32(len(buf))
				resQuery, _, _ = procQueryFullProcessImageName.Call(hProc, 1, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
			}
			_, _, _ = procCloseHandle.Call(hProc)
			if resQuery != 0 {
				exePath := syscall.UTF16ToString(buf[:size])
				baseName := filepath.Base(exePath)
				ext := filepath.Ext(baseName)
				if ext != "" {
					baseName = strings.TrimSuffix(baseName, ext)
				}
				return strings.TrimSpace(baseName)
			}
		}
	}

	return ""
}

// getWindowProcessName resolves the underlying executable basename of a given HWND.
func getWindowProcessName(hwnd uintptr) string {
	if hwnd == 0 {
		return ""
	}
	var pid uint32
	_, _, _ = procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return ""
	}
	hProc, _, _ := procOpenProcess.Call(0x1000, 0, uintptr(pid)) // PROCESS_QUERY_LIMITED_INFORMATION
	if hProc == 0 {
		hProc, _, _ = procOpenProcess.Call(0x0400, 0, uintptr(pid)) // PROCESS_QUERY_INFORMATION
	}
	if hProc == 0 {
		return ""
	}
	defer func() { _, _, _ = procCloseHandle.Call(hProc) }()

	var buf [1024]uint16
	size := uint32(len(buf))
	resQuery, _, _ := procQueryFullProcessImageName.Call(hProc, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if resQuery == 0 {
		size = uint32(len(buf))
		resQuery, _, _ = procQueryFullProcessImageName.Call(hProc, 1, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	}
	if resQuery != 0 {
		exePath := syscall.UTF16ToString(buf[:size])
		baseName := filepath.Base(exePath)
		ext := filepath.Ext(baseName)
		if ext != "" {
			baseName = strings.TrimSuffix(baseName, ext)
		}
		return strings.TrimSpace(baseName)
	}
	return ""
}

// isWhitelistedForDemotion evaluates whether the application is on the whitelist of legacy
// exclusive fullscreen games that require borderless popup demotion to coexist with top-level overlays.
func isWhitelistedForDemotion(procName string) bool {
	p := strings.ToLower(strings.TrimSpace(procName))
	return p == "stardew valley" || p == "stardewvalley" || p == "stardewmoddingapi"
}

// setWindowLongPtr updates a window attribute with 64-bit compatibility.
func setWindowLongPtr(hwnd uintptr, nIndex int32, dwNewLong uintptr) uintptr {
	if procSetWindowLongPtr.Find() == nil {
		res, _, _ := procSetWindowLongPtr.Call(hwnd, uintptr(nIndex), dwNewLong)
		return res
	}
	res, _, _ := procSetWindowLong.Call(hwnd, uintptr(nIndex), dwNewLong)
	return res
}

var (
	enumSnipMu         sync.Mutex
	enumSnipTargetHwnd uintptr
	cbEnumSnipOnce     sync.Once
	cbEnumSnipPtr      uintptr

	enumMonitorsMu     sync.Mutex
	enumMonitorsList   []MonitorInfo
	enumMonitorsID     int
	cbEnumMonitorsOnce sync.Once
	cbEnumMonitorsPtr  uintptr
)

func getEnumSnipCallback() uintptr {
	cbEnumSnipOnce.Do(func() {
		cbEnumSnipPtr = syscall.NewCallback(func(h, _ uintptr) uintptr {
			vis, _, _ := procIsWindowVisible.Call(h)
			if vis == 0 {
				return 1
			}

			iconic, _, _ := procIsIconic.Call(h)
			if iconic != 0 {
				return 1
			}

			var rc struct {
				Left, Top, Right, Bottom int32
			}
			res, _, _ := procGetWindowRect.Call(h, uintptr(unsafe.Pointer(&rc)))
			if res == 0 || rc.Left <= -10000 || rc.Right <= rc.Left || rc.Bottom <= rc.Top {
				return 1
			}

			if isScreenClippingHost(h) {
				enumSnipTargetHwnd = h
				return 0
			}
			return 1
		})
	})
	return cbEnumSnipPtr
}

func getEnumMonitorsCallback() uintptr {
	cbEnumMonitorsOnce.Do(func() {
		cbEnumMonitorsPtr = syscall.NewCallback(func(hMonitor, _, _, _ uintptr) uintptr {
			var mi monitorInfoW
			mi.CbSize = uint32(unsafe.Sizeof(mi))
			ret, _, _ := procGetMonitorInfoW.Call(hMonitor, uintptr(unsafe.Pointer(&mi)))
			if ret != 0 {
				enumMonitorsID++
				w := int(mi.RcMonitor.Right - mi.RcMonitor.Left)
				h := int(mi.RcMonitor.Bottom - mi.RcMonitor.Top)
				isPrimary := (mi.DwFlags & 1) != 0
				isPortrait := h > w
				enumMonitorsList = append(enumMonitorsList, MonitorInfo{
					ID:         enumMonitorsID,
					Name:       fmt.Sprintf("Display %d", enumMonitorsID),
					Left:       int(mi.RcMonitor.Left),
					Top:        int(mi.RcMonitor.Top),
					Width:      w,
					Height:     h,
					IsPrimary:  isPrimary,
					IsPortrait: isPortrait,
				})
			}
			return 1
		})
	})
	return cbEnumMonitorsPtr
}

// setupForegroundEventHook registers WinEvent hooks on a dedicated OS-locked thread
// with a Win32 message pump to maintain topmost Z-order on window changes and
// auto-hide Vinta when exclusive fullscreen games or presentations are active.
func (a *App) setupForegroundEventHook(vintaHwnd uintptr) {
	if vintaHwnd == 0 {
		return
	}

	procSetWindowPos := user32.NewProc("SetWindowPos")
	procGetMessage := user32.NewProc("GetMessageW")
	procDispatchMessage := user32.NewProc("DispatchMessageW")
	procTranslateMessage := user32.NewProc("TranslateMessage")
	procUnhookWinEvent := user32.NewProc("UnhookWinEvent")
	procGetForegroundWindow := user32.NewProc("GetForegroundWindow")

	myPid := uint32(os.Getpid())

	isVintaForeground := func(fg uintptr) bool {
		if fg == 0 {
			return false
		}
		if fg == vintaHwnd {
			return true
		}
		root, _, _ := procGetAncestor.Call(fg, 2) // GA_ROOT = 2
		if root == vintaHwnd {
			return true
		}
		var pid uint32
		_, _, _ = procGetWindowThreadProcessID.Call(fg, uintptr(unsafe.Pointer(&pid)))
		return pid == myPid
	}

	findActiveSnippingWindow := func(curFg uintptr) uintptr {
		if curFg != 0 && isSnippingTool(curFg) {
			return curFg
		}

		enumSnipMu.Lock()
		defer enumSnipMu.Unlock()
		enumSnipTargetHwnd = 0
		cb := getEnumSnipCallback()
		_, _, _ = procEnumWindows.Call(cb, 0)
		return enumSnipTargetHwnd
	}

	var lastElevatedSnipHwnd uintptr

	reassertTopmost := func(curFg uintptr) {
		snipHwnd := findActiveSnippingWindow(curFg)
		if snipHwnd != 0 {
			// Elevate Snipping Tool to HWND_TOPMOST so it is guaranteed to sit directly on top of Vinta's layer
			exStyle, _, _ := procGetWindowLong.Call(snipHwnd, uintptr(0xFFFFFFEC))
			if exStyle&0x00000008 == 0 { // Not yet WS_EX_TOPMOST
				_, _, _ = procSetWindowPos.Call(
					snipHwnd,
					uintptr(^uintptr(0)), // HWND_TOPMOST
					0, 0, 0, 0,
					0x0001|0x0002|0x0010, // SWP_NOSIZE | SWP_NOMOVE | SWP_NOACTIVATE
				)
				lastElevatedSnipHwnd = snipHwnd
			} else {
				// Reassert Snipping Tool at the very head of the topmost band
				_, _, _ = procSetWindowPos.Call(
					snipHwnd,
					uintptr(^uintptr(0)), // HWND_TOPMOST
					0, 0, 0, 0,
					0x0001|0x0002|0x0010, // SWP_NOSIZE | SWP_NOMOVE | SWP_NOACTIVATE
				)
			}
			// While Snipping Tool is active, Vinta stays visible directly beneath Snipping Tool.
			// Do NOT call SetWindowPos(vintaHwnd, HWND_TOPMOST) here, as that would bring Vinta back over Snipping Tool.
			return
		}

		if lastElevatedSnipHwnd != 0 {
			_, _, _ = procSetWindowPos.Call(
				lastElevatedSnipHwnd,
				uintptr(^uintptr(1)), // HWND_NOTOPMOST
				0, 0, 0, 0,
				0x0001|0x0002|0x0010,
			)
			lastElevatedSnipHwnd = 0
		}

		// Normal operation: Reassert Vinta as HWND_TOPMOST above standard desktop windows
		_, _, _ = procSetWindowPos.Call(
			vintaHwnd,
			uintptr(^uintptr(0)), // HWND_TOPMOST
			0, 0, 0, 0,
			0x0001|0x0002|0x0010, // SWP_NOSIZE | SWP_NOMOVE | SWP_NOACTIVATE
		)
	}

	var (
		autoHiddenByFullscreen bool
		autoHiddenMu           sync.Mutex
		lastDemotedHwnd        uintptr
	)

	evaluateForegroundState := func() {
		if !services.IsFrontendReady() {
			return
		}
		curFg, _, _ := procGetForegroundWindow.Call()
		if curFg == 0 || isVintaForeground(curFg) {
			reassertTopmost(curFg)
			return
		}

		appName := getWindowAppName(curFg)
		procName := getWindowProcessName(curFg)

		snipHwnd := findActiveSnippingWindow(curFg)
		if snipHwnd != 0 {
			autoHiddenMu.Lock()
			if autoHiddenByFullscreen {
				autoHiddenByFullscreen = false
				services.SetAppHidden(false)
				if a.window != nil {
					a.window.Show()
				}
				if services.IsFrontendReady() {
					if app := application.Get(); app != nil {
						app.Event.Emit("game-mode-exit")
					}
				}
			}
			autoHiddenMu.Unlock()

			reassertTopmost(curFg)
			return
		}

		vLeft, _, _, _, pCenter, monitors := GetDisplayMetrics()
		scale := a.scale
		if scale <= 0 {
			scale = 1.0
		}
		vintaMon := a.GetCurrentMonitor()
		covers, fgMonInfo := windowCoversMonitor(curFg, vintaMon)

		// Tier 2: Whitelist Demotion Mode for legacy exclusive games (such as Stardew Valley)
		if covers && isWhitelistedForDemotion(procName) {
			if curFg != lastDemotedHwnd {
				lastDemotedHwnd = curFg
				gwlStyle := int32(-16)                                    // GWL_STYLE
				newStyle := uintptr(0x80000000 | 0x10000000 | 0x04000000) // WS_POPUP | WS_VISIBLE | WS_CLIPSIBLINGS
				setWindowLongPtr(curFg, gwlStyle, newStyle)

				monW := uintptr(fgMonInfo.RcMonitor.Right - fgMonInfo.RcMonitor.Left)
				monH := uintptr(fgMonInfo.RcMonitor.Bottom - fgMonInfo.RcMonitor.Top)
				_, _, _ = procSetWindowPos.Call(
					curFg,
					uintptr(^uintptr(1)), // HWND_NOTOPMOST
					uintptr(fgMonInfo.RcMonitor.Left),
					uintptr(fgMonInfo.RcMonitor.Top),
					monW,
					monH,
					0x0020|0x0010, // SWP_FRAMECHANGED | SWP_NOACTIVATE
				)
			}
			reassertTopmost(curFg)
			return
		}

		shouldHide := false
		reason := ""

		var qState uint32
		goruntime.LockOSThread()
		coInitRes, _, _ := procCoInitializeEx.Call(0, 0)
		ret, _, _ := procSHQueryUserNotificationState.Call(uintptr(unsafe.Pointer(&qState)))
		if coInitRes == 0 || coInitRes == 1 {
			_, _, _ = procCoUninitialize.Call()
		}
		goruntime.UnlockOSThread()

		if ret == 0 && qState == QUNS_PRESENTATION_MODE {
			shouldHide = true
			reason = "presentation_mode"
		}

		hasEx := hasExclusiveDisplayOwnership()

		if !shouldHide && covers {
			if hasEx {
				shouldHide = true
				reason = "exclusive_fullscreen"
			}
		}

		logFullscreenDebug("fg=0x%x app='%s' proc='%s' qState=%d covers=%v hasEx=%v shouldHide=%v reason=%s",
			curFg, appName, procName, qState, covers, hasEx, shouldHide, reason)

		if shouldHide {
			if len(monitors) > 1 {
				var targetMon MonitorInfo
				foundTarget := false
				for _, m := range monitors {
					if m.ID != vintaMon.ID {
						targetMon = m
						foundTarget = true
						break
					}
				}

				if foundTarget {
					a.fullscreenRelocatedMu.Lock()
					if !a.isFullscreenRelocated {
						a.isFullscreenRelocated = true
						a.originalFullscreenOffset = a.offsetX
						a.originalFullscreenMonID = vintaMon.ID

						targetCenterScreenX := targetMon.Left + (targetMon.Width / 2)
						scaledOffset := targetCenterScreenX - (vLeft + pCenter)
						targetOffsetX := int(float64(scaledOffset) / scale)

						a.offsetX = targetOffsetX
						a.updateWindowRegion()

						if a.window != nil {
							a.window.Show()
						}

						if services.IsFrontendReady() {
							if app := application.Get(); app != nil {
								app.Event.Emit("game-mode-relocate", map[string]interface{}{
									"appName":         appName,
									"targetMonitor":   targetMon,
									"targetOffset":    targetOffsetX,
									"originalMonitor": vintaMon,
									"reason":          reason,
								})
							}
						}
					}
					a.fullscreenRelocatedMu.Unlock()
					return
				}
			}

			autoHiddenMu.Lock()
			if !autoHiddenByFullscreen {
				autoHiddenByFullscreen = true
				services.SetAppHidden(true)
				if a.window != nil {
					a.window.Hide()
				}
				if services.IsFrontendReady() {
					if app := application.Get(); app != nil {
						app.Event.Emit("game-mode-enter", map[string]interface{}{
							"reason":  reason,
							"appName": appName,
						})
					}
				}
			}
			autoHiddenMu.Unlock()
			return
		}

		a.fullscreenRelocatedMu.Lock()
		if a.isFullscreenRelocated {
			a.isFullscreenRelocated = false
			restoreOffset := a.originalFullscreenOffset
			origMonID := a.originalFullscreenMonID
			a.offsetX = restoreOffset
			a.fullscreenRelocatedMu.Unlock()

			a.updateWindowRegion()

			origMonName := "Original Display"
			for _, m := range monitors {
				if m.ID == origMonID {
					origMonName = m.Name
					break
				}
			}

			if services.IsFrontendReady() {
				if app := application.Get(); app != nil {
					app.Event.Emit("game-mode-return", map[string]interface{}{
						"restoreOffset":       restoreOffset,
						"originalMonitorName": origMonName,
					})
				}
			}
		} else {
			a.fullscreenRelocatedMu.Unlock()
		}

		autoHiddenMu.Lock()
		if autoHiddenByFullscreen {
			autoHiddenByFullscreen = false
			services.SetAppHidden(false)
			if a.window != nil {
				a.window.Show()
			}
			if services.IsFrontendReady() {
				if app := application.Get(); app != nil {
					app.Event.Emit("game-mode-exit")
				}
			}
		}
		autoHiddenMu.Unlock()

		reassertTopmost(curFg)
	}

	go func() {
		ticker := time.NewTicker(150 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			evaluateForegroundState()
		}
	}()

	go func() {
		goruntime.LockOSThread()

		cb := syscall.NewCallback(func(_ uintptr, event uint32, fgHwnd uintptr, _ int32, _ int32, _ uint32, _ uint32) uintptr {
			if event == EVENT_SYSTEM_FOREGROUND && fgHwnd != 0 && !isVintaForeground(fgHwnd) {
				time.AfterFunc(40*time.Millisecond, func() {
					evaluateForegroundState()
				})
				time.AfterFunc(200*time.Millisecond, func() {
					evaluateForegroundState()
				})
				// Fullscreen games frequently initialize graphics after window creation.
				// A third check catches transitions into exclusive swapchains.
				time.AfterFunc(1500*time.Millisecond, func() {
					evaluateForegroundState()
				})
			}
			return 0
		})

		hHook, _, _ := procSetWinEventHook.Call(
			uintptr(EVENT_SYSTEM_FOREGROUND),
			uintptr(EVENT_SYSTEM_FOREGROUND),
			0,
			cb,
			0,
			0,
			uintptr(WINEVENT_OUTOFCONTEXT),
		)

		type MSG struct {
			Hwnd    uintptr
			Message uint32
			WParam  uintptr
			LParam  uintptr
			Time    uint32
			Pt      [2]int32
		}

		var msg MSG
		for {
			ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
			if ret == 0 || ret == ^uintptr(0) {
				break
			}
			_, _, _ = procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			_, _, _ = procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
		}

		if hHook != 0 {
			_, _, _ = procUnhookWinEvent.Call(hHook)
		}
	}()
}

// screenWidth and screenHeight cache the screen dimensions so SetWidgetBounds can center regions correctly.
var screenWidth, screenHeight int

// App struct
type App struct {
	ctx                      context.Context
	wailsApp                 *application.App
	window                   *application.WebviewWindow
	devWindow                *application.WebviewWindow
	icon                     []byte
	pillW                    int
	pillH                    int
	pillR                    int
	restingW                 int
	restingH                 int
	restingR                 int
	offsetX                  int
	devW                     int
	devH                     int
	devExtra                 int
	satellitesMu             sync.RWMutex
	satellites               map[string]SatelliteBounds
	tooltipsMu               sync.RWMutex
	tooltips                 map[string]TooltipBounds
	scale                    float64
	isHidden                 bool
	autoHideBorderless       bool
	autoHideMu               sync.RWMutex
	updaterMu                sync.Mutex
	downloadCancel           context.CancelFunc
	downloadPaused           bool
	activeDownloadURL        string
	activeChecksumURL        string
	pendingInstallPath       string
	customUpdateRepo         string
	fullscreenRelocatedMu    sync.Mutex
	isFullscreenRelocated    bool
	originalFullscreenOffset int
	originalFullscreenMonID  int
	pillHitMu                sync.RWMutex
	pillHitRects             []pillHitBounds
	isForceInteractive       bool
	isCurrentlyInteractive   bool
	initialShowOnce          sync.Once
}

// NewApp creates a new App application struct
func NewApp(icon []byte) *App {
	smwVal, _, _ := user32.NewProc("GetSystemMetrics").Call(0)
	smhVal, _, _ := user32.NewProc("GetSystemMetrics").Call(1)
	sw := int(smwVal)
	sh := int(smhVal)
	if sw <= 0 {
		sw = 1920
	}
	if sh <= 0 {
		sh = 1080
	}
	screenWidth = sw
	screenHeight = sh

	return &App{
		icon:               icon,
		pillW:              340,
		pillH:              34,
		pillR:              24,
		scale:              1.0,
		autoHideBorderless: false,
		tooltips:           make(map[string]TooltipBounds),
	}
}

// getHwnd retrieves the window handle, preferring the native Wails v3 window handle.
func (a *App) getHwnd() uintptr {
	if a.window != nil {
		if h := uintptr(a.window.NativeWindow()); h != 0 {
			return h
		}
	}
	return getMyWindowHwnd()
}

// Setup initializes the Wails v3 application instance, window hooks, and native system tray.
func (a *App) Setup(wailsApp *application.App, win *application.WebviewWindow) {
	defer SetupPanicRecovery()
	a.wailsApp = wailsApp
	a.window = win
	if wailsApp != nil {
		a.ctx = wailsApp.Context()
		wailsApp.Event.On("cancel-fullscreen-relocate", func(_ *application.CustomEvent) {
			a.CancelFullscreenRelocate()
		})
	}

	if procTimeBeginPeriod.Find() == nil {
		_, _, _ = procTimeBeginPeriod.Call(1)
	}

	disableProcessPowerThrottling()

	smwVal, _, _ := user32.NewProc("GetSystemMetrics").Call(0)
	smhVal, _, _ := user32.NewProc("GetSystemMetrics").Call(1)
	if smwVal > 0 {
		screenWidth = int(smwVal)
	} else if screenWidth <= 0 {
		screenWidth = 1920
	}
	if smhVal > 0 {
		screenHeight = int(smhVal)
	} else if screenHeight <= 0 {
		screenHeight = 1080
	}

	var initOnce sync.Once
	applyInit := func(hwnd uintptr) {
		if hwnd == 0 {
			return
		}
		initOnce.Do(func() {
			cachedHwndMu.Lock()
			cachedWindowHwnd = hwnd
			cachedHwndMu.Unlock()

			gwl := int32(GWL_EXSTYLE)
			style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwl))
			style = (style | WS_EX_TOOLWINDOW | WS_EX_TOPMOST | WS_EX_NOACTIVATE | WS_EX_LAYERED | WS_EX_TRANSPARENT) &^ WS_EX_APPWINDOW
			_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)

			// Remove any clipping region so the canvas is fully transparent and unclipped
			_, _, _ = procSetWindowRgn.Call(hwnd, 0, 1)

			a.setupForegroundEventHook(hwnd)

			// Size stationary overlay window across virtual desktop space once at startup
			vLeft, vTop, vWidth, vHeight, _, monitors := GetDisplayMetrics()
			maxTopOffset := 0
			for _, mon := range monitors {
				if offset := mon.Top - vTop; offset > maxTopOffset {
					maxTopOffset = offset
				}
			}
			overlayHeight := maxTopOffset + 680
			if vHeight > 0 && overlayHeight >= vHeight {
				overlayHeight = vHeight - 1
			}
			if vWidth > 0 {
				_, _, _ = procSetWindowPos.Call(
					hwnd,
					uintptr(^uintptr(0)), // HWND_TOPMOST
					uintptr(int32(vLeft)), uintptr(int32(vTop)),
					uintptr(vWidth), uintptr(overlayHeight),
					0x0010|0x0020, // SWP_NOACTIVATE | SWP_FRAMECHANGED
				)
			}

			// Seed initial resting pill hit rect so the island is immediately interactive on launch
			primaryCenter := (vWidth / 2)
			for _, mon := range monitors {
				if mon.IsPrimary {
					primaryCenter = (mon.Left - vLeft) + (mon.Width / 2)
					break
				}
			}
			initW := int32(360)
			initH := int32(50)
			initLeft := int32(primaryCenter) - (initW / 2)
			initTop := int32(0)
			a.pillHitMu.Lock()
			if len(a.pillHitRects) == 0 {
				a.pillHitRects = []pillHitBounds{{
					left:   initLeft,
					top:    initTop,
					right:  initLeft + initW,
					bottom: initTop + initH,
				}}
			}
			a.pillHitMu.Unlock()

			go services.StartServicesListener()

			a.startHitTestingLoop(hwnd)
		})
	}

	if win != nil {
		win.OnWindowEvent(events.Common.WindowShow, func(_ *application.WindowEvent) {
			if h := uintptr(win.NativeWindow()); h != 0 {
				applyInit(h)
			}
		})
		win.OnWindowEvent(events.Windows.WindowShow, func(_ *application.WindowEvent) {
			if h := uintptr(win.NativeWindow()); h != 0 {
				applyInit(h)
			}
		})
		win.OnWindowEvent(events.Common.WindowRuntimeReady, func(_ *application.WindowEvent) {
			if h := uintptr(win.NativeWindow()); h != 0 {
				applyInit(h)
			}
			services.SetFrontendReady(true)
		})
		win.OnWindowEvent(events.Common.WindowClosing, func(_ *application.WindowEvent) {
			services.SetFrontendReady(false)
			services.StopServicesListener()
		})

		go func() {
			time.Sleep(3000 * time.Millisecond)
			a.initialShowOnce.Do(func() {
				if a.window != nil {
					a.window.Show()
				}
			})
		}()
	}

	go func() {
		for i := 0; i < 400; i++ {
			if win != nil {
				if h := uintptr(win.NativeWindow()); h != 0 {
					applyInit(h)
					return
				}
			}
			if h := getMyWindowHwnd(); h != 0 {
				applyInit(h)
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()

	a.StartBackgroundUpdateChecker()

	// Initialize global shortcut manager
	go func() {
		time.Sleep(600 * time.Millisecond)
		toggleHk := "Alt+Shift+V"
		quickNotesHk := "Alt+Shift+N"
		modeSwitchHk := "Alt+Shift+D"

		if settingsJSON, err := a.GetSettings(); err == nil && settingsJSON != "" && settingsJSON != "{}" {
			var m map[string]any
			if json.Unmarshal([]byte(settingsJSON), &m) == nil {
				if hk, ok := m["vinta_global_hotkey"].(string); ok && strings.TrimSpace(hk) != "" {
					toggleHk = strings.TrimSpace(hk)
				}
				if hk, ok := m["vinta_hotkey_quick_notes"].(string); ok && strings.TrimSpace(hk) != "" {
					quickNotesHk = strings.TrimSpace(hk)
				}
				if hk, ok := m["vinta_hotkey_mode_switch"].(string); ok && strings.TrimSpace(hk) != "" {
					modeSwitchHk = strings.TrimSpace(hk)
				}
			}
		}
		_, _ = a.RegisterNamedGlobalShortcut("toggle", toggleHk)
		_, _ = a.RegisterNamedGlobalShortcut("quick_notes", quickNotesHk)
		_, _ = a.RegisterNamedGlobalShortcut("mode_switch", modeSwitchHk)
	}()

	if wailsApp != nil {
		systray := wailsApp.SystemTray.New()
		systray.SetIcon(a.icon)
		systray.SetTooltip("Vinta Dynamic Island")

		menu := wailsApp.NewMenu()
		menu.Add("Show Vinta").OnClick(func(_ *application.Context) {
			if app := application.Get(); app != nil {
				app.Event.Emit("show-pill")
			}
		})
		menu.AddSeparator()
		menu.Add("Toggle Theme").OnClick(func(_ *application.Context) {
			if app := application.Get(); app != nil {
				app.Event.Emit("toggle-theme")
			}
		})
		menu.AddSeparator()
		menu.Add("Quit Vinta").OnClick(func(_ *application.Context) {
			a.Quit()
		})

		systray.SetMenu(menu)

		wailsApp.OnShutdown(func() {
			services.SetFrontendReady(false)
			services.StopServicesListener()
		})
	}
}

// Startup satisfies legacy startup calls if invoked.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Shutdown stops native background listeners.
func (a *App) Shutdown(_ context.Context) {
	services.StopServicesListener()
}

// SetWidgetBounds is called by the frontend whenever the widget dimensions change.
// It updates the native Win32 click-through region so only the visible widget is clickable.
func (a *App) SetWidgetBounds(width int, height int, radius int, scale float64, offsetX int) {
	if radius < 0 {
		if radius == -999 {
			if a.isHidden || height <= 24 {
				radius = 6
			} else if height > 52 {
				radius = 30
			} else if a.restingR > 0 {
				radius = a.restingR
			} else {
				radius = 20
			}
		} else {
			radius = -radius
		}
	}
	a.pillW = width
	a.pillH = height
	a.pillR = radius
	if (width >= 130 || height >= 24) && a.isHidden {
		a.isHidden = false
		services.SetAppHidden(false)
	}
	if !a.isHidden && (width >= 130 || height >= 24) && height <= 52 {
		a.restingW = width
		a.restingH = height
		a.restingR = radius
	}
	a.offsetX = offsetX
	if scale > 0 {
		a.scale = scale
	}
	a.updateWindowRegion()
}

// SetHitRects updates the list of physical pixel rectangles that receive mouse input.
func (a *App) SetHitRects(rects []HitRect) {
	a.pillHitMu.Lock()
	a.pillHitRects = make([]pillHitBounds, 0, len(rects))
	for _, r := range rects {
		if r.Width > 0 && r.Height > 0 {
			a.pillHitRects = append(a.pillHitRects, pillHitBounds{
				left:   int32(r.Left),
				top:    int32(r.Top),
				right:  int32(r.Left + r.Width),
				bottom: int32(r.Top + r.Height),
			})
		}
	}
	a.pillHitMu.Unlock()

	if len(rects) > 0 {
		a.initialShowOnce.Do(func() {
			if a.window != nil {
				a.window.Show()
			}
		})
	}
}

// SetPillBounds updates the physical pixel bounding box of the active island for hit testing.
func (a *App) SetPillBounds(left, top, width, height int) {
	if width <= 0 || height <= 0 {
		a.SetHitRects(nil)
	} else {
		a.SetHitRects([]HitRect{{Left: left, Top: top, Width: width, Height: height}})
	}
}

// SetForceInteractive forces the window to be interactive during dragging or modal flows.
func (a *App) SetForceInteractive(active bool) {
	a.pillHitMu.Lock()
	a.isForceInteractive = active
	a.pillHitMu.Unlock()

	hwnd := a.getHwnd()
	if hwnd == 0 {
		return
	}
	gwl := int32(GWL_EXSTYLE)
	style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwl))
	if active {
		a.isCurrentlyInteractive = true
		style &^= WS_EX_TRANSPARENT
		_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)
	}
}

// startHitTestingLoop runs a 60Hz background loop to dynamically toggle WS_EX_TRANSPARENT
// based on cursor proximity to the active pill bounding box, active satellites, or force-interactive mode.
func (a *App) startHitTestingLoop(hwnd uintptr) {
	go func() {
		defer SetupPanicRecovery()
		ticker := time.NewTicker(16 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			if hwnd == 0 {
				continue
			}
			isWin, _, _ := procIsWindow.Call(hwnd)
			if isWin == 0 {
				return
			}

			a.pillHitMu.RLock()
			force := a.isForceInteractive
			rects := make([]pillHitBounds, len(a.pillHitRects))
			copy(rects, a.pillHitRects)
			a.pillHitMu.RUnlock()

			isHit := false
			if force {
				isHit = true
			} else if len(rects) > 0 {
				var pt point32
				ret, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
				if ret != 0 {
					retScreen, _, _ := procScreenToClient.Call(hwnd, uintptr(unsafe.Pointer(&pt)))
					if retScreen != 0 {
						// 6-pixel buffer around the hit bounding boxes for smooth hover transitions across satellite gaps
						const buf int32 = 6
						for _, rect := range rects {
							if pt.X >= rect.left-buf && pt.X <= rect.right+buf &&
								pt.Y >= rect.top-buf && pt.Y <= rect.bottom+buf {
								isHit = true
								break
							}
						}
					}
				}
			}

			gwl := int32(GWL_EXSTYLE)
			style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwl))
			hasTransparent := (style & WS_EX_TRANSPARENT) != 0

			needsUpdate := false
			if isHit && hasTransparent {
				style &^= WS_EX_TRANSPARENT
				needsUpdate = true
			} else if !isHit && !hasTransparent {
				style |= WS_EX_TRANSPARENT
				needsUpdate = true
			}

			if needsUpdate {
				a.isCurrentlyInteractive = isHit
				_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)
			}
		}
	}()
}

// SetClipboardMonitoring satisfies existing frontend bindings as clipboard monitoring is retired.
func (a *App) SetClipboardMonitoring(copyEnabled bool, pasteEnabled bool) {
	_, _ = copyEnabled, pasteEnabled
}

// SetAutostart registers or unregisters the app in Windows Startup Registry
func (a *App) SetAutostart(enable bool) error {
	// Dev/test mode guard: do not write autostart registry entries when running dev/test binaries
	if isTestModeSync() {
		return nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	key, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.SET_VALUE,
	)
	if err != nil {
		return err
	}
	defer func() { _ = key.Close() }()

	if enable {
		return key.SetStringValue("Vinta", fmt.Sprintf(`"%s"`, exePath))
	}
	_ = key.DeleteValue("Vinta")
	return nil
}

// GetAutostart checks whether autostart is currently enabled in Windows Registry
func (a *App) GetAutostart() bool {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return false
	}
	defer func() { _ = key.Close() }()

	val, _, err := key.GetStringValue("Vinta")
	return err == nil && val != ""
}

// SendMediaControl forwards playback commands to VintaServices.exe
func (a *App) SendMediaControl(action string) {
	services.SendMediaControl(action)
}

// SelectMediaSession asks the native media service to switch active playback to sessionID.
func (a *App) SelectMediaSession(sessionID string) bool {
	return services.SelectMediaSession(sessionID)
}

// RequestMediaUpdate asks the backend and native WinRT bridge to immediately emit
// a fresh media state snapshot to the frontend without suppressing the album art.
func (a *App) RequestMediaUpdate() {
	services.RequestMediaUpdate()
}

// RequeryMediaArt requests the native WinRT service to re-extract the thumbnail
// stream for the current track and immediately push it to the frontend.
func (a *App) RequeryMediaArt() {
	services.RequeryMediaArt()
}

// FocusMediaApp brings the top-level window of the specified media application to the foreground.
func (a *App) FocusMediaApp(sourceApp string) bool {
	return services.FocusMediaApp(sourceApp)
}

// FocusNotificationApp brings the source window of a notification to the foreground,
// matching by process name or App User Model ID (for UWP/Store apps).
func (a *App) FocusNotificationApp(sourceApp string, aumid string) bool {
	return services.FocusNotificationApp(sourceApp, aumid)
}

// GetDisplayLayout returns the multi-monitor geometry and composite virtual desktop layout.
func (a *App) GetDisplayLayout() DisplayLayout {
	vLeft, vTop, vWidth, vHeight, pCenter, monitors := GetDisplayMetrics()
	scale := a.scale
	if scale <= 0 {
		scale = 1.0
	}
	screenCenterX := vLeft + pCenter + int(float64(a.offsetX)*scale)
	currentMon := resolveCurrentMonitor(screenCenterX, monitors)
	return DisplayLayout{
		VirtualLeft:    vLeft,
		VirtualTop:     vTop,
		VirtualWidth:   vWidth,
		VirtualHeight:  vHeight,
		PrimaryCenter:  pCenter,
		CurrentMonitor: currentMon,
		Monitors:       monitors,
	}
}

// GetCurrentMonitor returns the active monitor where Vinta is currently located.
func (a *App) GetCurrentMonitor() MonitorInfo {
	vLeft, _, _, _, pCenter, monitors := GetDisplayMetrics()
	scale := a.scale
	if scale <= 0 {
		scale = 1.0
	}
	screenCenterX := vLeft + pCenter + int(float64(a.offsetX)*scale)
	return resolveCurrentMonitor(screenCenterX, monitors)
}

// SetDevWindow registers the developer tools window for lifecycle management.
func (a *App) SetDevWindow(win *application.WebviewWindow) {
	a.devWindow = win
}

// Quit terminates the application and all child windows.
func (a *App) Quit() {
	services.StopServicesListener()

	// Defer window teardown slightly so the active RPC call resolves cleanly in WebView2 before destruction
	go func() {
		time.Sleep(50 * time.Millisecond)

		if a.devWindow != nil {
			a.devWindow.Close()
		}

		if a.wailsApp != nil {
			a.wailsApp.Quit()
		} else if app := application.Get(); app != nil {
			app.Quit()
		}

		if a.window != nil {
			a.window.Close()
		}

		time.Sleep(100 * time.Millisecond)
		os.Exit(0)
	}()
}

// CancelFullscreenRelocate clears any temporary relocation state when the user explicitly moves Vinta.
func (a *App) CancelFullscreenRelocate() {
	a.fullscreenRelocatedMu.Lock()
	a.isFullscreenRelocated = false
	a.fullscreenRelocatedMu.Unlock()
}

// FocusWindow requests OS input focus so typing and keyboard events work immediately on click/focus
func (a *App) FocusWindow() {
	hwnd := a.getHwnd()
	if hwnd != 0 {
		procGetForegroundWindow := user32.NewProc("GetForegroundWindow")
		curFg, _, _ := procGetForegroundWindow.Call()
		if curFg != hwnd {
			procGetAncestor := user32.NewProc("GetAncestor")
			root, _, _ := procGetAncestor.Call(curFg, 2) // GA_ROOT = 2
			if root != hwnd {
				procSetForegroundWindow := user32.NewProc("SetForegroundWindow")
				_, _, _ = procSetForegroundWindow.Call(hwnd)
			}
		}
	}
}

// SetDragCapture controls Win32 OS-level mouse capture during drag gestures.
// Deprecated/No-op: WebView2 child window requires uninterrupted mouse messages to fire DOM pointer capture.
// Calling Win32 SetCapture on the parent HWND starves WebView2 of WM_MOUSEMOVE and WM_LBUTTONUP.
func (a *App) SetDragCapture(enable bool) {
	_ = enable
}

type glideController struct {
	mu         sync.Mutex
	stopCh     chan struct{}
	generation uint64
}

var activeGlideController glideController

// StartInertiaGlide resets active glide generation tokens.
func (a *App) StartInertiaGlide(startOffset float64, velocity float64, minOffset float64, maxOffset float64) {
	_, _, _, _ = startOffset, velocity, minOffset, maxOffset
	activeGlideController.mu.Lock()
	if activeGlideController.stopCh != nil {
		close(activeGlideController.stopCh)
		activeGlideController.stopCh = nil
	}
	activeGlideController.generation++
	myGen := activeGlideController.generation
	activeGlideController.mu.Unlock()

	if app := application.Get(); app != nil {
		app.Event.Emit("inertia-started", myGen)
	}
}

// StopInertiaGlide immediately halts any in-progress inertia glide and returns current offset.
func (a *App) StopInertiaGlide() float64 {
	activeGlideController.mu.Lock()
	if activeGlideController.stopCh != nil {
		close(activeGlideController.stopCh)
		activeGlideController.stopCh = nil
	}
	activeGlideController.generation++
	offset := float64(a.offsetX)
	activeGlideController.mu.Unlock()
	return offset
}

// SetIgnoreMouseEvents toggles click-through for the entire window and its children using Win32 WS_EX_TRANSPARENT.
// When enabled, all mouse clicks pass straight through to underlying applications and the clipping region is lifted.
func (a *App) SetIgnoreMouseEvents(ignore bool) {
	if a.window != nil {
		a.window.SetIgnoreMouseEvents(ignore)
	}

	hwnd := a.getHwnd()
	if hwnd == 0 {
		return
	}

	gwl := int32(GWL_EXSTYLE)
	style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwl))
	if ignore {
		style |= WS_EX_TRANSPARENT
	} else {
		style &^= WS_EX_TRANSPARENT
	}
	_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)
}

// TrimMemory requests Windows to reclaim unused physical RAM pages from the Vinta process working set.
func (a *App) TrimMemory() {
	hProcess, _, _ := procGetCurrentProcess.Call()
	if hProcess != 0 && procEmptyWorkingSet.Find() == nil {
		_, _, _ = procEmptyWorkingSet.Call(hProcess)
	}
}

// disableProcessPowerThrottling opts the application out of Windows 11 Efficiency Mode (EcoQoS) and timer throttling.
func disableProcessPowerThrottling() {
	if procSetProcessInformation.Find() != nil {
		return
	}
	hProcess, _, _ := procGetCurrentProcess.Call()
	if hProcess == 0 {
		return
	}

	state := processPowerThrottlingState{
		Version:     processPowerThrottlingCurrentVersion,
		ControlMask: processPowerThrottlingExecutionSpeed | processPowerThrottlingIgnoreTimerResolution,
		StateMask:   0,
	}

	_, _, _ = procSetProcessInformation.Call(
		hProcess,
		uintptr(processPowerThrottling),
		uintptr(unsafe.Pointer(&state)),
		uintptr(unsafe.Sizeof(state)),
	)
}

// SetIsHidden informs native background services whether Vinta is currently hidden to throttle high-frequency IPC.
func (a *App) SetIsHidden(hidden bool) {
	a.isHidden = hidden
	services.SetAppHidden(hidden)
	a.updateWindowRegion()
}

// SetAutoHideBorderless toggles automatic hiding when a borderless fullscreen window is detected on Vinta's monitor.
func (a *App) SetAutoHideBorderless(enabled bool) {
	a.autoHideMu.Lock()
	a.autoHideBorderless = enabled
	a.autoHideMu.Unlock()
}

// SetDevPillBounds is called by the frontend to update the size of the developer menu.
func (a *App) SetDevPillBounds(width int, height int, extraH int, scale float64) {
	a.devW = width
	a.devH = height
	a.devExtra = extraH
	if scale > 0 {
		a.scale = scale
	}
	a.updateWindowRegion()
}

// TooltipBounds holds the screen-relative bounding box and corner radius for an active tooltip.
type TooltipBounds struct {
	ID     string  `json:"id"`
	X      int     `json:"x"`
	Y      int     `json:"y"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Radius int     `json:"radius"`
	Scale  float64 `json:"scale"`
	Active bool    `json:"active"`
}

// SetTooltipBounds registers an active floating tooltip region to permit Win32 painting beyond the main pill.
func (a *App) SetTooltipBounds(bounds TooltipBounds) {
	a.tooltipsMu.Lock()
	if a.tooltips == nil {
		a.tooltips = make(map[string]TooltipBounds)
	}
	if bounds.Active && bounds.Width > 0 && bounds.Height > 0 {
		a.tooltips[bounds.ID] = bounds
	} else {
		delete(a.tooltips, bounds.ID)
	}
	a.tooltipsMu.Unlock()
	a.updateWindowRegion()
}

// ClearTooltipBounds removes an active floating tooltip region from the Win32 window clipping mask.
func (a *App) ClearTooltipBounds(id string) {
	a.tooltipsMu.Lock()
	if a.tooltips != nil {
		if id == "*" || id == "all" || id == "" {
			a.tooltips = make(map[string]TooltipBounds)
		} else {
			delete(a.tooltips, id)
		}
	}
	a.tooltipsMu.Unlock()
	a.updateWindowRegion()
}

// ClearAllTooltips removes all active floating tooltip regions from the Win32 window clipping mask.
func (a *App) ClearAllTooltips() {
	a.tooltipsMu.Lock()
	a.tooltips = make(map[string]TooltipBounds)
	a.tooltipsMu.Unlock()
	a.updateWindowRegion()
}

// SatelliteBounds represents the position, dimensions, radius, and anchor configuration for a floating satellite dock.
type SatelliteBounds struct {
	ID          string `json:"id"`
	Anchor      string `json:"anchor"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Radius      int    `json:"radius"`
	Radii       [4]int `json:"radii,omitempty"`
	Gap         int    `json:"gap"`
	Gaps        [4]int `json:"gaps,omitempty"`
	OffsetX     int    `json:"offsetX"`
	OffsetY     int    `json:"offsetY"`
	TooltipPad  int    `json:"tooltipPad"`
	TooltipSide string `json:"tooltipSide"`
	Active      bool   `json:"active"`
}

// SetSatelliteBounds registers or updates a floating satellite dock in the window region manager.
func (a *App) SetSatelliteBounds(bounds SatelliteBounds) {
	a.satellitesMu.Lock()
	if a.satellites == nil {
		a.satellites = make(map[string]SatelliteBounds)
	}
	a.satellites[bounds.ID] = bounds
	a.satellitesMu.Unlock()
	a.updateWindowRegion()
}

// RemoveSatellite removes a floating satellite dock by identifier.
func (a *App) RemoveSatellite(id string) {
	a.satellitesMu.Lock()
	if a.satellites != nil {
		delete(a.satellites, id)
	}
	a.satellitesMu.Unlock()
	a.updateWindowRegion()
}

// SetFloatingControlsBounds adapts legacy calls to the generalized satellite dock registry.
func (a *App) SetFloatingControlsBounds(width int, height int, radius int, gap int) {
	a.SetSatelliteBounds(SatelliteBounds{
		ID:     "controls",
		Anchor: "right",
		Width:  width,
		Height: height,
		Radius: radius,
		Gap:    gap,
		Active: width > 0 && height > 0,
	})
}

// updateWindowRegion is a no-op under the unclipped transparent canvas hit-testing architecture.
func (a *App) updateWindowRegion() {
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// IsTestMode returns true if the application was compiled with the devtest tag.
func (a *App) IsTestMode() bool {
	return isTestModeSync()
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

// GetHardwareAccelerationConfig reads stored GPU acceleration setting from user config dir.
func GetHardwareAccelerationConfig() bool {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return true // Default enabled
	}
	filePath := filepath.Join(configDir, "FreeMatrix Studios", "Vinta", "hw_accel.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return true // Default enabled
	}
	return strings.TrimSpace(string(data)) != "false"
}

// GetHardwareAcceleration returns whether GPU acceleration is enabled.
func (a *App) GetHardwareAcceleration() bool {
	return GetHardwareAccelerationConfig()
}

// SetHardwareAcceleration persists the GPU acceleration setting.
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

// NoteFileInfo describes a saved sticky note file.
type NoteFileInfo struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	FileName string `json:"fileName"`
	ModTime  int64  `json:"modTime"`
	Size     int64  `json:"size"`
}

type noteTitleParser struct {
	Title string `json:"title"`
}

func getNotesDir() (string, error) {
	notesDir := filepath.Join(GetAppDataDir(), "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return "", err
	}
	return notesDir, nil
}

// migrateLegacyNotes scans for older .md notes and converts them into .vnote format.
func migrateLegacyNotes(dir string, entries []os.DirEntry) {
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		oldPath := filepath.Join(dir, entry.Name())
		mdContent, err := os.ReadFile(oldPath)
		if err != nil {
			continue
		}

		baseName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		newFileName := baseName + ".vnote"
		newPath := filepath.Join(dir, newFileName)

		if _, statErr := os.Stat(newPath); os.IsNotExist(statErr) {
			payload := map[string]interface{}{
				"id":        baseName,
				"title":     baseName,
				"color":     "yellow",
				"content":   string(mdContent),
				"pinned":    false,
				"modTime":   info.ModTime().UnixMilli(),
				"createdAt": info.ModTime().UnixMilli(),
			}
			if data, err := json.MarshalIndent(payload, "", "  "); err == nil {
				_ = os.WriteFile(newPath, data, 0644)
			}
		}

		_ = os.Remove(oldPath)
	}
}

// ListNotes scans the notes directory, migrates legacy notes, and returns all .vnote metadata.
func (a *App) ListNotes() ([]NoteFileInfo, error) {
	dir, err := getNotesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	migrateLegacyNotes(dir, entries)

	entries, err = os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var notes []NoteFileInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".vnote") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fileName := entry.Name()
		id := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		title := id

		filePath := filepath.Join(dir, fileName)
		if fileBytes, readErr := os.ReadFile(filePath); readErr == nil {
			var parsed noteTitleParser
			if parseErr := json.Unmarshal(fileBytes, &parsed); parseErr == nil && strings.TrimSpace(parsed.Title) != "" {
				title = strings.TrimSpace(parsed.Title)
			}
		}

		notes = append(notes, NoteFileInfo{
			ID:       id,
			Title:    title,
			FileName: fileName,
			ModTime:  info.ModTime().UnixMilli(),
			Size:     info.Size(),
		})
	}
	return notes, nil
}

// ReadNote reads and returns the JSON content of a .vnote file.
func (a *App) ReadNote(fileName string) (string, error) {
	dir, err := getNotesDir()
	if err != nil {
		return "", err
	}
	cleanName := filepath.Base(fileName)
	if !strings.HasSuffix(strings.ToLower(cleanName), ".vnote") {
		cleanName = strings.TrimSuffix(cleanName, filepath.Ext(cleanName)) + ".vnote"
	}
	filePath := filepath.Join(dir, cleanName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveNote writes or updates a .vnote file.
func (a *App) SaveNote(fileName string, content string) error {
	dir, err := getNotesDir()
	if err != nil {
		return err
	}
	cleanName := filepath.Base(fileName)
	if !strings.HasSuffix(strings.ToLower(cleanName), ".vnote") {
		cleanName = strings.TrimSuffix(cleanName, filepath.Ext(cleanName)) + ".vnote"
	}
	filePath := filepath.Join(dir, cleanName)
	return os.WriteFile(filePath, []byte(content), 0644)
}

// RenameNote renames an existing .vnote file using the new title as the filename.
// Returns the new filename on success.
func (a *App) RenameNote(oldFileName string, newTitle string) (string, error) {
	dir, err := getNotesDir()
	if err != nil {
		return "", err
	}
	sanitized := strings.NewReplacer(
		"/", "-", "\\", "-", ":", "-", "*", "-",
		"?", "-", "\"", "-", "<", "-", ">", "-", "|", "-",
	).Replace(strings.TrimSpace(newTitle))
	if sanitized == "" {
		sanitized = "untitled"
	}
	newFileName := sanitized + ".vnote"

	oldClean := filepath.Base(oldFileName)
	if !strings.HasSuffix(strings.ToLower(oldClean), ".vnote") {
		oldClean = strings.TrimSuffix(oldClean, filepath.Ext(oldClean)) + ".vnote"
	}

	oldPath := filepath.Join(dir, oldClean)
	newPath := filepath.Join(dir, newFileName)

	if !strings.EqualFold(oldPath, newPath) {
		if _, statErr := os.Stat(newPath); statErr == nil {
			base := strings.TrimSuffix(newFileName, ".vnote")
			for n := 2; ; n++ {
				candidate := fmt.Sprintf("%s (%d).vnote", base, n)
				if _, e := os.Stat(filepath.Join(dir, candidate)); os.IsNotExist(e) {
					newFileName = candidate
					newPath = filepath.Join(dir, newFileName)
					break
				}
			}
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			return "", err
		}
	}
	return newFileName, nil
}

// DeleteNote deletes a .vnote file from disk.
func (a *App) DeleteNote(fileName string) error {
	dir, err := getNotesDir()
	if err != nil {
		return err
	}
	cleanName := filepath.Base(fileName)
	if !strings.HasSuffix(strings.ToLower(cleanName), ".vnote") {
		cleanName = strings.TrimSuffix(cleanName, filepath.Ext(cleanName)) + ".vnote"
	}
	filePath := filepath.Join(dir, cleanName)
	return os.Remove(filePath)
}

// OpenNotesFolder opens the Vinta notes directory in Windows File Explorer.
func (a *App) OpenNotesFolder() error {
	dir, err := getNotesDir()
	if err != nil {
		return err
	}
	return exec.Command("explorer.exe", dir).Start()
}

// OpenWindowsSettings opens Windows system settings pages or flyouts.
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

// ExportSettingsFileDialog opens a native save dialog and exports the current settings JSON with envelope metadata.
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

// ImportSettingsFileDialog opens a native open dialog and returns the JSON content of the selected file.
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

var (
	globalHotkeyMutex sync.Mutex
	registeredHotkeys = make(map[string]string)
)

// RegisterNamedGlobalShortcut registers an application-wide hotkey accelerator for a specific action.
func (a *App) RegisterNamedGlobalShortcut(name string, accelerator string) (bool, error) {
	wailsApp := application.Get()
	if wailsApp == nil {
		return false, fmt.Errorf("application instance not available")
	}

	globalHotkeyMutex.Lock()
	defer globalHotkeyMutex.Unlock()

	trimmed := strings.TrimSpace(accelerator)
	oldAccel := registeredHotkeys[name]

	if trimmed == "" {
		if oldAccel != "" {
			_ = wailsApp.GlobalShortcut.Unregister(oldAccel)
			delete(registeredHotkeys, name)
		}
		return true, nil
	}

	if oldAccel != "" {
		_ = wailsApp.GlobalShortcut.Unregister(oldAccel)
		delete(registeredHotkeys, name)
	}

	eventName := "hotkey-toggle-island"
	switch name {
	case "quick_notes":
		eventName = "hotkey-quick-notes"
	case "mode_switch":
		eventName = "hotkey-mode-switch"
	default:
		eventName = "hotkey-toggle-island"
	}

	err := wailsApp.GlobalShortcut.Register(trimmed, func() {
		wailsApp.Event.Emit(eventName)
	})
	if err != nil {
		return false, err
	}

	registeredHotkeys[name] = trimmed
	return true, nil
}

// UnregisterNamedGlobalShortcut removes a registered global shortcut by action name.
func (a *App) UnregisterNamedGlobalShortcut(name string) error {
	wailsApp := application.Get()
	if wailsApp == nil {
		return fmt.Errorf("application instance not available")
	}

	globalHotkeyMutex.Lock()
	defer globalHotkeyMutex.Unlock()

	if oldAccel := registeredHotkeys[name]; oldAccel != "" {
		_ = wailsApp.GlobalShortcut.Unregister(oldAccel)
		delete(registeredHotkeys, name)
	}
	return nil
}

// GetNamedGlobalShortcut returns the currently registered accelerator string for a specific action.
func (a *App) GetNamedGlobalShortcut(name string) string {
	globalHotkeyMutex.Lock()
	defer globalHotkeyMutex.Unlock()
	return registeredHotkeys[name]
}

// RegisterGlobalShortcut registers the main toggle shortcut (retained for backward compatibility).
func (a *App) RegisterGlobalShortcut(accelerator string) (bool, error) {
	return a.RegisterNamedGlobalShortcut("toggle", accelerator)
}

// UnregisterGlobalShortcut removes the main toggle shortcut (retained for backward compatibility).
func (a *App) UnregisterGlobalShortcut() error {
	return a.UnregisterNamedGlobalShortcut("toggle")
}

// GetGlobalShortcut returns the main toggle shortcut (retained for backward compatibility).
func (a *App) GetGlobalShortcut() string {
	return a.GetNamedGlobalShortcut("toggle")
}

// GetCalendarData reads calendar.json from %APPDATA%/FreeMatrix Studios/Vinta.
func (a *App) GetCalendarData() (string, error) {
	filePath := filepath.Join(GetAppDataDir(), "calendar.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "{}", nil
	}
	return string(data), nil
}

// SaveCalendarData writes calendar.json to %APPDATA%/FreeMatrix Studios/Vinta.
func (a *App) SaveCalendarData(calendarJSON string) error {
	dir := GetAppDataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	filePath := filepath.Join(dir, "calendar.json")
	return os.WriteFile(filePath, []byte(calendarJSON), 0644)
}

// GetTimerData reads timer.json from %APPDATA%/FreeMatrix Studios/Vinta.
func (a *App) GetTimerData() (string, error) {
	filePath := filepath.Join(GetAppDataDir(), "timer.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "{}", nil
	}
	return string(data), nil
}

// SaveTimerData writes timer.json to %APPDATA%/FreeMatrix Studios/Vinta.
func (a *App) SaveTimerData(timerJSON string) error {
	dir := GetAppDataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	filePath := filepath.Join(dir, "timer.json")
	return os.WriteFile(filePath, []byte(timerJSON), 0644)
}

// GetNotifications reads notifications.json from %APPDATA%/FreeMatrix Studios/Vinta.
func (a *App) GetNotifications() (string, error) {
	filePath := filepath.Join(GetAppDataDir(), "notifications.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "[]", nil
	}
	return string(data), nil
}

// SaveNotifications writes notifications.json to %APPDATA%/FreeMatrix Studios/Vinta.
func (a *App) SaveNotifications(notificationsJSON string) error {
	dir := GetAppDataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	filePath := filepath.Join(dir, "notifications.json")
	return os.WriteFile(filePath, []byte(notificationsJSON), 0644)
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

	// Read and remove entries inside cacheDir rather than deleting the root cache directory itself
	if entries, err := os.ReadDir(cacheDir); err == nil {
		for _, entry := range entries {
			targetPath := filepath.Join(cacheDir, entry.Name())
			_ = os.RemoveAll(targetPath)
		}
	} else if os.IsNotExist(err) {
		_ = os.MkdirAll(cacheDir, 0755)
	}

	// Always guarantee the persistent icon subdirectory exists
	_ = os.MkdirAll(filepath.Join(cacheDir, "icons"), 0755)

	// Clean up temporary notification sync database files in %TEMP%
	tempDir := os.TempDir()
	matches, _ := filepath.Glob(filepath.Join(tempDir, "wpn_vinta_sync.db*"))
	for _, match := range matches {
		_ = os.Remove(match)
	}

	// Flush in-memory caches
	FlushIconMemoryCache()
	services.ClearNativeMediaArtCache()

	return nil
}
