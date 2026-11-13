// Package backend provides core application services and state management for Vinta.
//
// app_window.go manages Win32 window state, multi-monitor layout caching, and Z-order
// rules so Vinta behaves like a native desktop overlay. It keeps the window topmost
// during normal use, steps behind Windows Snipping Tool and native menus when active,
// and moves Vinta to a free display (or hides it on single-monitor setups) when a
// fullscreen game or presentation takes over the current screen. It also opts out of
// Windows EcoQoS power throttling and trims WebView2 working-set memory on demand.
package backend

import (
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/sys/windows"

	"Vinta/backend/services"
)

var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	procGetWindowLong            = user32.NewProc("GetWindowLongW")
	procSetWindowLong            = user32.NewProc("SetWindowLongW")
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
	procIsWindow                 = user32.NewProc("IsWindow")
	procSetWindowRgn             = user32.NewProc("SetWindowRgn")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procWindowFromPoint          = user32.NewProc("WindowFromPoint")
	procEnumChildWindows         = user32.NewProc("EnumChildWindows")
	procGetAsyncKeyState         = user32.NewProc("GetAsyncKeyState")
	procSetLayeredWindowAttrs    = user32.NewProc("SetLayeredWindowAttributes")
	procGetSystemMetrics         = user32.NewProc("GetSystemMetrics")
	procGetMessage               = user32.NewProc("GetMessageW")
	procDispatchMessage          = user32.NewProc("DispatchMessageW")
	procTranslateMessage         = user32.NewProc("TranslateMessage")
	procUnhookWinEvent           = user32.NewProc("UnhookWinEvent")

	shell32                          = syscall.NewLazyDLL("shell32.dll")
	procSHQueryUserNotificationState = shell32.NewProc("SHQueryUserNotificationState")

	ole32              = syscall.NewLazyDLL("ole32.dll")
	procCoInitializeEx = ole32.NewProc("CoInitializeEx")
	procCoUninitialize = ole32.NewProc("CoUninitialize")

	gdi32                             = syscall.NewLazyDLL("gdi32.dll")
	procD3DKMTCheckExclusiveOwnership = gdi32.NewProc("D3DKMTCheckExclusiveOwnership")

	psapi               = syscall.NewLazyDLL("psapi.dll")
	procEmptyWorkingSet = psapi.NewProc("EmptyWorkingSet")

	kernel32                      = syscall.NewLazyDLL("kernel32.dll")
	procGetCurrentProcess         = kernel32.NewProc("GetCurrentProcess")
	procSetProcessInformation     = kernel32.NewProc("SetProcessInformation")
	procOpenProcess               = kernel32.NewProc("OpenProcess")
	procCloseHandle               = kernel32.NewProc("CloseHandle")
	procQueryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageNameW")

	winmm               = syscall.NewLazyDLL("winmm.dll")
	procTimeBeginPeriod = winmm.NewProc("timeBeginPeriod")

	dwmapi                    = syscall.NewLazyDLL("dwmapi.dll")
	procDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
)

const (
	//nolint:revive // Windows API constants
	DWMWA_NCRENDERING_POLICY = 2
	//nolint:revive // Windows API constants
	DWMNCRP_DISABLED = 1
	//nolint:revive // Windows API constants
	DWMWA_TRANSITIONS_FORCEDISABLED = 3
	//nolint:revive // Windows API constants
	WS_SYSMENU = 0x00080000
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
	WS_EX_TRANSPARENT = 0x00000020
	rgnOr             = 2

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
	GA_ROOT = 2
	//nolint:revive // Windows API constants
	GWL_STYLE = -16
	//nolint:revive // Windows API constants
	WS_POPUP = 0x80000000
	//nolint:revive // Windows API constants
	WS_CAPTION = 0x00C00000
	//nolint:revive // Windows API constants
	WS_MAXIMIZE = 0x01000000
	//nolint:revive // Windows API constants
	WS_THICKFRAME = 0x00040000
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
	SWP_HIDEWINDOW = 0x0080
	//nolint:revive // Windows API constants
	SWP_NOCOPYBITS = 0x0100
	//nolint:revive // Windows API constants
	SWP_NOOWNERZORDER = 0x0200
	//nolint:revive // Windows API constants
	SWP_NOSENDCHANGING = 0x0400
	//nolint:revive // Windows API constants
	SWP_DEFERERASE = 0x2000

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

type rect32 struct{ Left, Top, Right, Bottom int32 }
type point32 struct{ X, Y int32 }

type monitorInfoW struct {
	CbSize    uint32
	RcMonitor rect32
	RcWork    rect32
	DwFlags   uint32
}

// MonitorInfo holds physical display bounds and orientation flags.
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

// DisplayLayout represents the combined virtual desktop and all connected monitors.
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

	cachedWindowHwnd uintptr
	cachedHwndMu     sync.Mutex

	// Caches PID to exe basename for 30s to avoid repeated OpenProcess calls during 150ms polling.
	pidExeCacheMu   sync.RWMutex
	pidExeCache     = make(map[uint32]string, 64)
	pidExeLastClear = time.Now()

	// Blocks stale frontend SetWidgetBounds calls from reverting backend monitor relocations.
	relocateGraceUntil time.Time
	relocateTargetX    int

	enumSnipMu         sync.Mutex
	enumSnipTargetHwnd uintptr
	cbEnumSnipOnce     sync.Once
	cbEnumSnipPtr      uintptr

	enumMonitorsMu     sync.Mutex
	enumMonitorsList   []MonitorInfo
	enumMonitorsID     int
	cbEnumMonitorsOnce sync.Once
	cbEnumMonitorsPtr  uintptr

	lastTrimTime          time.Time
	trimMu                sync.Mutex
	enumChildWebviewOnce  sync.Once
	enumChildWebviewPtr   uintptr
	enumChildWebviewMu    sync.Mutex
	discoveredWebviewPIDs []uint32
)

func winClass(hwnd uintptr) string {
	if hwnd == 0 {
		return ""
	}
	var buf [256]uint16
	n, _, _ := procGetClassName.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 256)
	if n == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:n])
}

func winTitle(hwnd uintptr) string {
	if hwnd == 0 {
		return ""
	}
	var buf [256]uint16
	n, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 256)
	if n == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:n])
}

func winPID(hwnd uintptr) uint32 {
	var pid uint32
	if hwnd != 0 {
		_, _, _ = procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	}
	return pid
}

func winStyle(hwnd uintptr) (style, exStyle uint32) {
	gwlStyle, gwlExStyle := int32(GWL_STYLE), int32(GWL_EXSTYLE)
	s, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwlStyle))
	ex, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwlExStyle))
	return uint32(s), uint32(ex)
}

func getProcessExe(pid uint32) string {
	if pid == 0 {
		return ""
	}
	pidExeCacheMu.RLock()
	if time.Since(pidExeLastClear) < 30*time.Second {
		if exe, ok := pidExeCache[pid]; ok {
			pidExeCacheMu.RUnlock()
			return exe
		}
	}
	pidExeCacheMu.RUnlock()

	hProc, _, _ := procOpenProcess.Call(0x1000, 0, uintptr(pid))
	if hProc == 0 {
		hProc, _, _ = procOpenProcess.Call(0x0400, 0, uintptr(pid))
	}
	if hProc == 0 {
		return ""
	}

	var buf [1024]uint16
	size := uint32(len(buf))
	res, _, _ := procQueryFullProcessImageName.Call(hProc, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if res == 0 {
		size = uint32(len(buf))
		res, _, _ = procQueryFullProcessImageName.Call(hProc, 1, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	}
	_, _, _ = procCloseHandle.Call(hProc)
	if res == 0 {
		return ""
	}

	baseName := strings.ToLower(filepath.Base(syscall.UTF16ToString(buf[:size])))
	pidExeCacheMu.Lock()
	if time.Since(pidExeLastClear) >= 30*time.Second || len(pidExeCache) > 256 {
		clear(pidExeCache)
		pidExeLastClear = time.Now()
	}
	pidExeCache[pid] = baseName
	pidExeCacheMu.Unlock()
	return baseName
}

func rectCoversBounds(rc rect32, left, top, right, bottom int) bool {
	return int(rc.Left) <= left+4 && int(rc.Top) <= top+4 &&
		int(rc.Right) >= right-4 && int(rc.Bottom) >= bottom-4
}

// isFullscreenStyle returns false for standard maximized windows that still have a titlebar.
func isFullscreenStyle(style uint32) bool {
	if (style&WS_CAPTION == WS_CAPTION) && (style&WS_MAXIMIZE != 0) {
		return false
	}
	return (style&(WS_CAPTION|WS_THICKFRAME)) == 0 || (style&WS_POPUP) != 0
}

func (a *App) effectiveScale() float64 {
	if a.scale > 0 {
		return a.scale
	}
	return 1.0
}

func (a *App) monitorTargetOffset(m MonitorInfo, vLeft, pCenter int) int {
	centerScreenX := m.Left + (m.Width / 2)
	return int(float64(centerScreenX-(vLeft+pCenter)) / a.effectiveScale())
}

// GetDisplayMetrics returns virtual screen bounds and connected monitors, caching results for 1s.
func GetDisplayMetrics() (vLeft, vTop, vWidth, vHeight, pCenter int, monitors []MonitorInfo) {
	vx, _, _ := procGetSystemMetrics.Call(SM_XVIRTUALSCREEN)
	vy, _, _ := procGetSystemMetrics.Call(SM_YVIRTUALSCREEN)
	vw, _, _ := procGetSystemMetrics.Call(SM_CXVIRTUALSCREEN)
	vh, _, _ := procGetSystemMetrics.Call(SM_CYVIRTUALSCREEN)

	currVx, currVy, currVw, currVh := int(int32(vx)), int(int32(vy)), int(vw), int(vh)

	isCacheValid := func() bool {
		return time.Since(lastDisplayFetch) < time.Second &&
			cachedDisplayLayout.VirtualWidth > 0 &&
			cachedDisplayLayout.VirtualWidth == currVw &&
			cachedDisplayLayout.VirtualHeight == currVh &&
			cachedDisplayLayout.VirtualLeft == currVx &&
			cachedDisplayLayout.VirtualTop == currVy
	}

	cachedDisplayMu.RLock()
	if isCacheValid() {
		l := cachedDisplayLayout
		cachedDisplayMu.RUnlock()
		return l.VirtualLeft, l.VirtualTop, l.VirtualWidth, l.VirtualHeight, l.PrimaryCenter, l.Monitors
	}
	cachedDisplayMu.RUnlock()

	cachedDisplayMu.Lock()
	defer cachedDisplayMu.Unlock()

	if isCacheValid() {
		l := cachedDisplayLayout
		return l.VirtualLeft, l.VirtualTop, l.VirtualWidth, l.VirtualHeight, l.PrimaryCenter, l.Monitors
	}

	vLeft, vTop, vWidth, vHeight = currVx, currVy, currVw, currVh
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
		if vHeight = int(ph); vHeight <= 0 {
			vHeight = 1080
		}
	}
	pCenter = -vLeft + (primaryW / 2)

	if procEnumDisplayMonitors.Find() == nil && procGetMonitorInfoW.Find() == nil {
		enumMonitorsMu.Lock()
		enumMonitorsID = 0
		enumMonitorsList = nil
		_, _, _ = procEnumDisplayMonitors.Call(0, 0, getEnumMonitorsCallback(), 0)
		monitors = enumMonitorsList
		enumMonitorsMu.Unlock()
	}

	if len(monitors) == 0 {
		monitors = []MonitorInfo{{
			ID: 1, Name: "Primary Display", Width: primaryW, Height: vHeight,
			IsPrimary: true, IsPortrait: vHeight > primaryW,
		}}
	}

	cachedDisplayLayout = DisplayLayout{
		VirtualLeft: vLeft, VirtualTop: vTop, VirtualWidth: vWidth, VirtualHeight: vHeight,
		PrimaryCenter: pCenter, Monitors: monitors,
	}
	lastDisplayFetch = time.Now()
	return
}

// resolveCurrentMonitor matches a horizontal screen coordinate to its containing monitor.
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
	return MonitorInfo{ID: 1, Name: "Primary Display", Width: 1920, Height: 1080, IsPrimary: true}
}

// getMyWindowHwnd finds and caches the HWND for Vinta when Wails native handle lookup is unavailable.
func getMyWindowHwnd() uintptr {
	cachedHwndMu.Lock()
	if cachedWindowHwnd != 0 {
		if ret, _, _ := procIsWindow.Call(cachedWindowHwnd); ret != 0 {
			h := cachedWindowHwnd
			cachedHwndMu.Unlock()
			return h
		}
		cachedWindowHwnd = 0
	}
	cachedHwndMu.Unlock()

	titlePtr, _ := syscall.UTF16PtrFromString("Vinta")
	if hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr))); hwnd != 0 {
		cachedHwndMu.Lock()
		cachedWindowHwnd = hwnd
		cachedHwndMu.Unlock()
		return hwnd
	}

	myPid := uint32(os.Getpid())
	desktop, _, _ := procGetDesktopWindow.Call()
	for child, _, _ := procGetWindow.Call(desktop, GW_CHILD); child != 0; child, _, _ = procGetWindow.Call(child, GW_HWNDNEXT) {
		if winPID(child) == myPid && winTitle(child) == "Vinta" {
			cachedHwndMu.Lock()
			cachedWindowHwnd = child
			cachedHwndMu.Unlock()
			return child
		}
	}
	return 0
}

func (a *App) getHwnd() uintptr {
	if a.window != nil {
		if h := uintptr(a.window.NativeWindow()); h != 0 {
			return h
		}
	}
	return getMyWindowHwnd()
}

func isDesktopOrShell(hwnd uintptr) bool {
	switch winClass(hwnd) {
	case "WorkerW", "Progman", "Shell_TrayWnd", "Shell_SecondaryTrayWnd", "Windows.UI.Core.CoreWindow":
		return true
	}
	return false
}

func isScreenClippingHost(hwnd uintptr) bool {
	if hwnd == 0 {
		return false
	}
	cls := winClass(hwnd)
	if cls == "ScreenClippingWindow" || strings.Contains(cls, "ScreenClipping") {
		return true
	}
	_, exStyle := winStyle(hwnd)
	return (exStyle&WS_EX_TOPMOST) != 0 && getProcessExe(winPID(hwnd)) == "screenclippinghost.exe"
}

// isSnippingTool detects screen clipping overlays and native/XAML context menus that need to sit above Vinta.
func isSnippingTool(hwnd uintptr) bool {
	if hwnd == 0 {
		return false
	}
	check := func(h uintptr) bool {
		if isScreenClippingHost(h) {
			return true
		}
		cls := winClass(h)
		if cls == "#32768" {
			return true
		}
		style, exStyle := winStyle(h)
		if (exStyle&WS_EX_TOPMOST) == 0 && (style&WS_POPUP) == 0 && (style&WS_CAPTION) != 0 {
			return false
		}
		switch getProcessExe(winPID(h)) {
		case "screenclippinghost.exe", "screensketch.exe", "snippingtoolapp.exe", "snippingtool.exe":
			return true
		case "explorer.exe", "shellexperiencehost.exe":
			if (cls == "Xaml_WindowedPopupClass" || cls == "Microsoft.UI.Content.PopupWindowSiteBridge") && (exStyle&WS_EX_TOPMOST) != 0 {
				return true
			}
		}
		title := strings.ToLower(winTitle(h))
		return strings.Contains(title, "snipping tool") || strings.Contains(title, "screen snipping") ||
			strings.Contains(title, "snip & sketch") || strings.Contains(title, "screen clipping")
	}

	if check(hwnd) {
		return true
	}
	root, _, _ := procGetAncestor.Call(hwnd, GA_ROOT)
	return root != 0 && root != hwnd && check(root)
}

func getWindowAppName(hwnd uintptr) string {
	if title := strings.TrimSpace(winTitle(hwnd)); title != "" {
		return title
	}
	if base := getProcessExe(winPID(hwnd)); base != "" {
		return strings.TrimSpace(strings.TrimSuffix(base, filepath.Ext(base)))
	}
	return ""
}

func isHwndFullscreenOnMonitor(hwnd uintptr, m MonitorInfo, vintaHwnd uintptr) bool {
	if hwnd == 0 || hwnd == vintaHwnd {
		return false
	}
	if root, _, _ := procGetAncestor.Call(hwnd, GA_ROOT); root != 0 {
		hwnd = root
	}
	if hwnd == vintaHwnd || isDesktopOrShell(hwnd) || isSnippingTool(hwnd) {
		return false
	}
	vis, _, _ := procIsWindowVisible.Call(hwnd)
	iconic, _, _ := procIsIconic.Call(hwnd)
	if vis == 0 || iconic != 0 {
		return false
	}
	var rc rect32
	if res, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rc))); res == 0 {
		return false
	}
	if !rectCoversBounds(rc, m.Left, m.Top, m.Left+m.Width, m.Top+m.Height) {
		return false
	}
	style, _ := winStyle(hwnd)
	return isFullscreenStyle(style)
}

// isMonitorCoveredByFullscreen checks the foreground window and 3 sample points on monitor m
// so a centered dialog or notification does not hide an underlying fullscreen game.
func isMonitorCoveredByFullscreen(m MonitorInfo, vintaHwnd uintptr) bool {
	if curFg, _, _ := procGetForegroundWindow.Call(); isHwndFullscreenOnMonitor(curFg, m, vintaHwnd) {
		return true
	}
	pts := [3][2]int{
		{m.Left + m.Width/2, m.Top + m.Height/2},
		{m.Left + m.Width/4, m.Top + m.Height/4},
		{m.Left + (m.Width*3)/4, m.Top + (m.Height*3)/4},
	}
	for _, pt := range pts {
		packed := (uintptr(uint32(pt[0])) & 0xFFFFFFFF) | ((uintptr(uint32(pt[1])) & 0xFFFFFFFF) << 32)
		if top, _, _ := procWindowFromPoint.Call(packed); isHwndFullscreenOnMonitor(top, m, vintaHwnd) {
			return true
		}
	}
	return false
}

// windowCoversMonitor checks if fgHwnd is on Vinta's current monitor and spans its full bounds.
func windowCoversMonitor(fgHwnd uintptr, vintaMon MonitorInfo) (bool, monitorInfoW) {
	var mi monitorInfoW
	if fgHwnd == 0 || isDesktopOrShell(fgHwnd) || isSnippingTool(fgHwnd) {
		return false, mi
	}
	hMonFg, _, _ := procMonitorFromWindow.Call(fgHwnd, MONITOR_DEFAULTTONEAREST)
	if hMonFg == 0 {
		return false, mi
	}
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	if res, _, _ := procGetMonitorInfoW.Call(hMonFg, uintptr(unsafe.Pointer(&mi))); res == 0 {
		return false, mi
	}
	if int(mi.RcMonitor.Left) != vintaMon.Left || int(mi.RcMonitor.Top) != vintaMon.Top {
		return false, mi
	}
	var rc rect32
	if res, _, _ := procGetWindowRect.Call(fgHwnd, uintptr(unsafe.Pointer(&rc))); res == 0 {
		return false, mi
	}
	return rectCoversBounds(rc, int(mi.RcMonitor.Left), int(mi.RcMonitor.Top), int(mi.RcMonitor.Right), int(mi.RcMonitor.Bottom)), mi
}

func hasExclusiveDisplayOwnership() bool {
	if procD3DKMTCheckExclusiveOwnership.Find() != nil {
		return false
	}
	res, _, _ := procD3DKMTCheckExclusiveOwnership.Call()
	return res != 0
}

func getEnumSnipCallback() uintptr {
	cbEnumSnipOnce.Do(func() {
		cbEnumSnipPtr = syscall.NewCallback(func(h, _ uintptr) uintptr {
			vis, _, _ := procIsWindowVisible.Call(h)
			iconic, _, _ := procIsIconic.Call(h)
			if vis == 0 || iconic != 0 {
				return 1
			}
			var rc rect32
			res, _, _ := procGetWindowRect.Call(h, uintptr(unsafe.Pointer(&rc)))
			// Keep iterating after a match so enumSnipTargetHwnd lands on the lowest Z-order snipping overlay.
			if res != 0 && rc.Left > -10000 && rc.Right > rc.Left && rc.Bottom > rc.Top && isSnippingTool(h) {
				enumSnipTargetHwnd = h
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
			if ret, _, _ := procGetMonitorInfoW.Call(hMonitor, uintptr(unsafe.Pointer(&mi))); ret != 0 {
				enumMonitorsID++
				w := int(mi.RcMonitor.Right - mi.RcMonitor.Left)
				h := int(mi.RcMonitor.Bottom - mi.RcMonitor.Top)
				enumMonitorsList = append(enumMonitorsList, MonitorInfo{
					ID: enumMonitorsID, Name: fmt.Sprintf("Display %d", enumMonitorsID),
					Left: int(mi.RcMonitor.Left), Top: int(mi.RcMonitor.Top),
					Width: w, Height: h, IsPrimary: (mi.DwFlags & 1) != 0, IsPortrait: h > w,
				})
			}
			return 1
		})
	})
	return cbEnumMonitorsPtr
}

// setupForegroundEventHook watches foreground changes via WinEventHook and a 150ms ticker
// to manage topmost Z-order, Snipping Tool subordination, and fullscreen relocation.
func (a *App) setupForegroundEventHook(vintaHwnd uintptr) {
	if vintaHwnd == 0 {
		return
	}
	myPid := uint32(os.Getpid())

	isVintaForeground := func(fg uintptr) bool {
		if fg == 0 {
			return false
		}
		if fg == vintaHwnd {
			return true
		}
		root, _, _ := procGetAncestor.Call(fg, GA_ROOT)
		return root == vintaHwnd || winPID(fg) == myPid
	}

	findActiveSnippingWindow := func(curFg uintptr) uintptr {
		enumSnipMu.Lock()
		enumSnipTargetHwnd = 0
		_, _, _ = procEnumWindows.Call(getEnumSnipCallback(), 0)
		found := enumSnipTargetHwnd
		enumSnipMu.Unlock()
		if found != 0 {
			return found
		}
		if curFg != 0 && isSnippingTool(curFg) {
			return curFg
		}
		return 0
	}

	var (
		autoHiddenByFullscreen bool
		autoHiddenMu           sync.Mutex
		wasSubordinatedToSnip  bool
	)

	exitFullscreenHide := func() {
		autoHiddenMu.Lock()
		if autoHiddenByFullscreen {
			autoHiddenByFullscreen = false
			services.SetAppHidden(false)
			_, _, _ = procSetWindowPos.Call(
				vintaHwnd, ^uintptr(0), 0, 0, 0, 0,
				SWP_NOSIZE|SWP_NOMOVE|SWP_NOACTIVATE|SWP_SHOWWINDOW,
			)
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
	}

	reassertTopmost := func(curFg, snipHwnd uintptr, snipKnown bool) {
		if a.IsOutboundDragging() {
			return
		}
		// Skip SetWindowPos while LMB is held to avoid breaking OS drag-and-drop visuals.
		if lmb, _, _ := procGetAsyncKeyState.Call(0x01); (lmb & 0x8000) != 0 {
			return
		}
		if !snipKnown {
			snipHwnd = findActiveSnippingWindow(curFg)
		}
		if snipHwnd != 0 {
			wasSubordinatedToSnip = true
			_, _, _ = procSetWindowPos.Call(vintaHwnd, snipHwnd, 0, 0, 0, 0, 0x0001|0x0002|0x0010|0x0400|0x4000)
			return
		}
		autoHiddenMu.Lock()
		hidden := autoHiddenByFullscreen
		autoHiddenMu.Unlock()
		if hidden {
			return
		}
		_, exStyle := winStyle(vintaHwnd)
		if !wasSubordinatedToSnip && (exStyle&WS_EX_TOPMOST) != 0 {
			return
		}
		wasSubordinatedToSnip = false
		_, _, _ = procSetWindowPos.Call(vintaHwnd, ^uintptr(0), 0, 0, 0, 0, SWP_NOSIZE|SWP_NOMOVE|SWP_NOACTIVATE)
	}

	evaluateForegroundState := func() {
		if !services.IsFrontendReady() || a.IsOutboundDragging() {
			return
		}
		if lmb, _, _ := procGetAsyncKeyState.Call(0x01); (lmb & 0x8000) != 0 {
			return
		}
		curFg, _, _ := procGetForegroundWindow.Call()
		if curFg == 0 || isVintaForeground(curFg) {
			reassertTopmost(curFg, 0, false)
			return
		}

		if snipHwnd := findActiveSnippingWindow(curFg); snipHwnd != 0 {
			exitFullscreenHide()
			reassertTopmost(curFg, snipHwnd, true)
			return
		}

		vLeft, _, _, _, pCenter, monitors := GetDisplayMetrics()
		vintaMon := a.GetCurrentMonitor()

		a.fullscreenRelocatedMu.Lock()
		a.isFullscreenRelocated = false
		a.fullscreenTriggerHwnd = 0
		a.fullscreenTriggerMonID = 0
		a.fullscreenRelocatedMu.Unlock()

		// Only trigger fullscreen hide/relocation if the foreground app actually covers Vinta's current monitor.
		covers, _ := windowCoversMonitor(curFg, vintaMon)
		shouldHide, reason, appName := false, "", ""

		if covers {
			style, _ := winStyle(curFg)
			isMaximizedStandard := (style&WS_CAPTION == WS_CAPTION) && (style&WS_MAXIMIZE != 0)
			if !isMaximizedStandard {
				isFramelessOrPopup := isFullscreenStyle(style)
				appName = getWindowAppName(curFg)
				isUnsupportedGame := strings.Contains(strings.ToLower(appName), "stardew")

				a.autoHideMu.RLock()
				hideBorderless := a.autoHideBorderless
				a.autoHideMu.RUnlock()

				var qState uint32
				goruntime.LockOSThread()
				coRes, _, _ := procCoInitializeEx.Call(0, 0)
				ret, _, _ := procSHQueryUserNotificationState.Call(uintptr(unsafe.Pointer(&qState)))
				if coRes == 0 || coRes == 1 {
					_, _, _ = procCoUninitialize.Call()
				}
				goruntime.UnlockOSThread()

				switch {
				case ret == 0 && qState == QUNS_PRESENTATION_MODE:
					shouldHide, reason = true, "presentation_mode"
				case (hasExclusiveDisplayOwnership() || (ret == 0 && qState == QUNS_RUNNING_D3D_FULL_SCREEN)) && isFramelessOrPopup:
					shouldHide, reason = true, "exclusive_fullscreen"
				case isUnsupportedGame:
					shouldHide, reason = true, "unsupported_app"
				case isFramelessOrPopup && hideBorderless:
					shouldHide, reason = true, "borderless_fullscreen"
				}
			}
		}

		if shouldHide {
			for _, m := range monitors {
				if m.ID != vintaMon.ID && !isMonitorCoveredByFullscreen(m, vintaHwnd) {
					targetOffsetX := a.monitorTargetOffset(m, vLeft, pCenter)

					a.fullscreenRelocatedMu.Lock()
					relocateGraceUntil = time.Now().Add(1500 * time.Millisecond)
					relocateTargetX = targetOffsetX
					a.offsetX = targetOffsetX
					a.fullscreenRelocatedMu.Unlock()

					a.updateWindowRegion()
					if a.window != nil {
						a.window.Show()
					}
					if app := application.Get(); app != nil && services.IsFrontendReady() {
						app.Event.Emit("game-mode-relocate", map[string]interface{}{
							"appName":         appName,
							"targetMonitor":   m,
							"targetOffset":    targetOffsetX,
							"originalMonitor": vintaMon,
							"reason":          reason,
						})
					}
					return
				}
			}

			autoHiddenMu.Lock()
			if !autoHiddenByFullscreen {
				autoHiddenByFullscreen = true
				services.SetAppHidden(true)
				_, _, _ = procSetWindowPos.Call(
					vintaHwnd, ^uintptr(1), 0, 0, 0, 0,
					SWP_NOSIZE|SWP_NOMOVE|SWP_NOACTIVATE|SWP_HIDEWINDOW,
				)
				if a.window != nil {
					a.window.Hide()
				}
				if app := application.Get(); app != nil && services.IsFrontendReady() {
					app.Event.Emit("game-mode-enter", map[string]interface{}{
						"reason": reason, "appName": appName,
					})
				}
			}
			autoHiddenMu.Unlock()
			return
		}

		exitFullscreenHide()
		reassertTopmost(curFg, 0, true)

		a.autoHideMu.RLock()
		followDisplay := a.followActiveDisplay
		a.autoHideMu.RUnlock()

		if followDisplay && len(monitors) > 1 {
			if hMonFg, _, _ := procMonitorFromWindow.Call(curFg, MONITOR_DEFAULTTONEAREST); hMonFg != 0 {
				var fgMi monitorInfoW
				fgMi.CbSize = uint32(unsafe.Sizeof(fgMi))
				if res, _, _ := procGetMonitorInfoW.Call(hMonFg, uintptr(unsafe.Pointer(&fgMi))); res != 0 {
					for _, m := range monitors {
						if m.Left == int(fgMi.RcMonitor.Left) && m.Top == int(fgMi.RcMonitor.Top) {
							if m.ID != vintaMon.ID && !isMonitorCoveredByFullscreen(m, vintaHwnd) {
								if targetOffsetX := a.monitorTargetOffset(m, vLeft, pCenter); targetOffsetX != a.offsetX {
									a.offsetX = targetOffsetX
									a.updateWindowRegion()
									if app := application.Get(); app != nil && services.IsFrontendReady() {
										app.Event.Emit("active-monitor-follow", map[string]interface{}{
											"targetMonitor": m, "targetOffset": targetOffsetX,
										})
									}
								}
							}
							break
						}
					}
				}
			}
		}
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
		cb := syscall.NewCallback(func(_ uintptr, event uint32, fgHwnd uintptr, _, _ int32, _, _ uint32) uintptr {
			if event == EVENT_SYSTEM_FOREGROUND && fgHwnd != 0 && !isVintaForeground(fgHwnd) {
				time.AfterFunc(40*time.Millisecond, evaluateForegroundState)
				time.AfterFunc(250*time.Millisecond, evaluateForegroundState)
			}
			return 0
		})
		hHook, _, _ := procSetWinEventHook.Call(
			EVENT_SYSTEM_FOREGROUND, EVENT_SYSTEM_FOREGROUND, 0, cb, 0, 0, WINEVENT_OUTOFCONTEXT,
		)
		var msg struct {
			Hwnd    uintptr
			Message uint32
			WParam  uintptr
			LParam  uintptr
			Time    uint32
			Pt      [2]int32
		}
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

// GetDisplayLayout returns the virtual desktop geometry and current monitor state.
func (a *App) GetDisplayLayout() DisplayLayout {
	vLeft, vTop, vWidth, vHeight, pCenter, monitors := GetDisplayMetrics()
	a.fullscreenRelocatedMu.Lock()
	offsetX := a.offsetX
	a.fullscreenRelocatedMu.Unlock()

	screenCenterX := vLeft + pCenter + int(float64(offsetX)*a.effectiveScale())
	return DisplayLayout{
		VirtualLeft: vLeft, VirtualTop: vTop, VirtualWidth: vWidth, VirtualHeight: vHeight,
		PrimaryCenter: pCenter, CurrentMonitor: resolveCurrentMonitor(screenCenterX, monitors), Monitors: monitors,
	}
}

// GetCurrentMonitor returns the display where Vinta is currently positioned.
func (a *App) GetCurrentMonitor() MonitorInfo {
	return a.GetDisplayLayout().CurrentMonitor
}

// CancelFullscreenRelocate resets relocation state and clears the offset grace timer.
func (a *App) CancelFullscreenRelocate() {
	a.fullscreenRelocatedMu.Lock()
	a.isFullscreenRelocated = false
	relocateGraceUntil = time.Time{}
	a.fullscreenRelocatedMu.Unlock()
}

// FocusWindow gives OS input focus to Vinta if it is not already focused.
func (a *App) FocusWindow() {
	if hwnd := a.getHwnd(); hwnd != 0 {
		curFg, _, _ := procGetForegroundWindow.Call()
		root, _, _ := procGetAncestor.Call(curFg, GA_ROOT)
		if curFg != hwnd && root != hwnd {
			_, _, _ = procSetForegroundWindow.Call(hwnd)
		}
	}
}

// SetDragCapture is a no-op stub kept for Wails binding compatibility.
func (a *App) SetDragCapture(enable bool) {
	_ = enable
}

type glideController struct {
	mu         sync.Mutex
	stopCh     chan struct{}
	generation uint64
}

var activeGlideController glideController

// StartInertiaGlide increments the glide generation token and notifies the frontend.
func (a *App) StartInertiaGlide(startOffset, velocity, minOffset, maxOffset float64) {
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

// StopInertiaGlide cancels any active glide and returns the current horizontal offset.
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

// SetIgnoreMouseEvents toggles WS_EX_TRANSPARENT click-through on the native window.
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

func getEnumChildWebviewCallback() uintptr {
	enumChildWebviewOnce.Do(func() {
		enumChildWebviewPtr = syscall.NewCallback(func(h, _ uintptr) uintptr {
			if pid := winPID(h); pid != 0 {
				discoveredWebviewPIDs = append(discoveredWebviewPIDs, pid)
			}
			return 1
		})
	})
	return enumChildWebviewPtr
}

// TrimMemory flushes unused working-set pages from Vinta and its child WebView2 processes.
func (a *App) TrimMemory() {
	trimMu.Lock()
	if time.Since(lastTrimTime) < 500*time.Millisecond {
		trimMu.Unlock()
		return
	}
	lastTrimTime = time.Now()
	trimMu.Unlock()

	if procEmptyWorkingSet.Find() != nil {
		return
	}

	currentPID := uint32(os.Getpid())
	targetPIDs := make(map[uint32]bool)

	enumChildWebviewMu.Lock()
	discoveredWebviewPIDs = discoveredWebviewPIDs[:0]
	cb := getEnumChildWebviewCallback()
	if hwnd := a.getHwnd(); hwnd != 0 && procEnumChildWindows.Find() == nil {
		_, _, _ = procEnumChildWindows.Call(hwnd, cb, 0)
	}
	if a.devWindow != nil && procEnumChildWindows.Find() == nil {
		if devHwnd := uintptr(a.devWindow.NativeWindow()); devHwnd != 0 {
			_, _, _ = procEnumChildWindows.Call(devHwnd, cb, 0)
		}
	}
	for _, pid := range discoveredWebviewPIDs {
		if pid != currentPID {
			targetPIDs[pid] = true
		}
	}
	enumChildWebviewMu.Unlock()

	if snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0); err == nil {
		defer func() { _ = windows.CloseHandle(snapshot) }()

		var pe windows.ProcessEntry32
		pe.Size = uint32(unsafe.Sizeof(pe))

		type procItem struct{ pid, parentPID uint32 }
		allProcs := make([]procItem, 0, 64)
		for err = windows.Process32First(snapshot, &pe); err == nil; err = windows.Process32Next(snapshot, &pe) {
			if strings.Contains(strings.ToLower(windows.UTF16ToString(pe.ExeFile[:])), "msedgewebview2") {
				allProcs = append(allProcs, procItem{pid: pe.ProcessID, parentPID: pe.ParentProcessID})
			}
		}

		for _, p := range allProcs {
			if p.parentPID == currentPID {
				targetPIDs[p.pid] = true
			}
		}
		for added := true; added; {
			added = false
			for _, p := range allProcs {
				if targetPIDs[p.parentPID] && !targetPIDs[p.pid] {
					targetPIDs[p.pid] = true
					added = true
				}
			}
		}

		for pid := range targetPIDs {
			if pid != currentPID {
				if hChild, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid); err == nil {
					_, _, _ = procEmptyWorkingSet.Call(uintptr(hChild))
					_ = windows.CloseHandle(hChild)
				}
			}
		}
	}

	debug.FreeOSMemory()
	if hProcess, _, _ := procGetCurrentProcess.Call(); hProcess != 0 {
		_, _, _ = procEmptyWorkingSet.Call(hProcess)
	}
}

// disableProcessPowerThrottling opts out of Windows 11 Efficiency Mode and timer coalescing.
func disableProcessPowerThrottling() {
	if procSetProcessInformation.Find() != nil {
		return
	}
	if hProcess, _, _ := procGetCurrentProcess.Call(); hProcess != 0 {
		state := processPowerThrottlingState{
			Version:     processPowerThrottlingCurrentVersion,
			ControlMask: processPowerThrottlingExecutionSpeed | processPowerThrottlingIgnoreTimerResolution,
		}
		_, _, _ = procSetProcessInformation.Call(
			hProcess, processPowerThrottling, uintptr(unsafe.Pointer(&state)), uintptr(unsafe.Sizeof(state)),
		)
	}
}

// SetIsHidden updates the hidden state and notifies background services to throttle IPC.
func (a *App) SetIsHidden(hidden bool) {
	a.isHidden = hidden
	services.SetAppHidden(hidden)
	a.updateWindowRegion()
}

// SetAutoHideBorderless toggles auto-hide when borderless fullscreen windows cover Vinta's monitor.
func (a *App) SetAutoHideBorderless(enabled bool) {
	a.autoHideMu.Lock()
	a.autoHideBorderless = enabled
	a.autoHideMu.Unlock()
}

// SetFollowActiveDisplay toggles whether Vinta moves to the monitor with the active foreground window.
func (a *App) SetFollowActiveDisplay(enabled bool) {
	a.autoHideMu.Lock()
	a.followActiveDisplay = enabled
	a.autoHideMu.Unlock()
}

// SetWidgetBounds updates the pill dimensions, scale, and horizontal offset from the frontend.
func (a *App) SetWidgetBounds(width, height, radius int, scale float64, offsetX int) {
	if radius < 0 {
		switch {
		case radius != -999:
			radius = -radius
		case a.isHidden || height <= 24:
			radius = 6
		case height > 52:
			radius = 30
		case a.restingR > 0:
			radius = a.restingR
		default:
			radius = 20
		}
	}
	a.pillW, a.pillH, a.pillR = width, height, radius
	if (width >= 130 || height >= 24) && a.isHidden {
		a.isHidden = false
		services.SetAppHidden(false)
	}
	if !a.isHidden && (width >= 130 || height >= 24) && height <= 52 {
		a.restingW, a.restingH, a.restingR = width, height, radius
	}

	a.fullscreenRelocatedMu.Lock()
	if time.Now().Before(relocateGraceUntil) && offsetX != relocateTargetX {
		offsetX = relocateTargetX
	} else if offsetX == relocateTargetX {
		relocateGraceUntil = time.Time{}
	}
	a.offsetX = offsetX
	a.fullscreenRelocatedMu.Unlock()

	if scale > 0 {
		a.scale = scale
	}
	a.updateWindowRegion()
}

// SetDevPillBounds updates the developer menu dimensions and recalculates the window region.
func (a *App) SetDevPillBounds(width, height, extraH int, scale float64) {
	a.devW, a.devH, a.devExtra = width, height, extraH
	if scale > 0 {
		a.scale = scale
	}
	a.updateWindowRegion()
}
