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
	"math"
	"os"
	"path/filepath"
	goruntime "runtime"
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
	shcore                       = syscall.NewLazyDLL("shcore.dll")
	procGetDpiForMonitor         = shcore.NewProc("GetDpiForMonitor")
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
	procGetCursorInfo            = user32.NewProc("GetCursorInfo")
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
	procRegisterClassEx          = user32.NewProc("RegisterClassExW")
	procCreateWindowEx           = user32.NewProc("CreateWindowExW")
	procDefWindowProc            = user32.NewProc("DefWindowProcW")
	procRegisterRawInputDevices  = user32.NewProc("RegisterRawInputDevices")

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
	GW_HWNDPREV = 3
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

// MonitorInfo holds physical display bounds, orientation flags, and Windows DPI scale.
type MonitorInfo struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Left       int     `json:"left"`
	Top        int     `json:"top"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	IsPrimary  bool    `json:"isPrimary"`
	IsPortrait bool    `json:"isPortrait"`
	DpiScale   float64 `json:"dpiScale"`
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
	trimRunning           bool
	trimPending           bool
	trimTimer             *time.Timer
	trimMu                sync.Mutex
	enumChildWebviewOnce  sync.Once
	enumChildWebviewPtr   uintptr
	enumChildWebviewMu    sync.Mutex
	discoveredWebviewPIDs []uint32
	cachedWebviewPIDs     []uint32
	lastWebviewScanTime   time.Time
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

// ComputeOverlayHeight returns the virtual overlay height covering the lowest monitor top edge.
func ComputeOverlayHeight(vTop, vHeight int, monitors []MonitorInfo) int {
	return ComputeOverlayHeightForScale(vTop, vHeight, monitors, 1.0, 1.0)
}

// ComputeOverlayHeightForScale returns the virtual overlay height scaled for the active UI scale
// and display DPI, capped below vHeight so Windows never treats the overlay as fullscreen exclusive.
func ComputeOverlayHeightForScale(vTop, vHeight int, monitors []MonitorInfo, uiScale, windowDpr float64) int {
	maxTopOffset := 0
	for _, mon := range monitors {
		if offset := mon.Top - vTop; offset > maxTopOffset {
			maxTopOffset = offset
		}
	}
	if uiScale < 1.0 {
		uiScale = 1.0
	}
	if windowDpr <= 0 {
		windowDpr = 1.0
	}
	effectiveDpr := windowDpr
	if uiScale <= 1.0 {
		for _, mon := range monitors {
			if mon.DpiScale > effectiveDpr {
				effectiveDpr = mon.DpiScale
			}
		}
	}
	scaledHeight := int(math.Ceil(960.0 * uiScale * effectiveDpr))
	overlayHeight := maxTopOffset + scaledHeight
	if vHeight > 0 && overlayHeight >= vHeight {
		overlayHeight = vHeight - 1
	}
	return overlayHeight
}

// SetUIScale updates the active UI scale and resizes the Win32 overlay height proportionally
// so scaled pages are never clipped by the bottom edge of the WebView2 window.
func (a *App) SetUIScale(scale float64) {
	if scale < 1.0 {
		scale = 1.0
	}

	a.pillHitMu.Lock()
	if math.Abs(a.uiScale-scale) < 0.001 && a.overlayHeight > 0 {
		a.pillHitMu.Unlock()
		return
	}
	a.uiScale = scale
	a.pillHitMu.Unlock()

	a.resizeOverlayHeight(0)
}

func (a *App) ensureOverlayHeight(minHeight int) {
	if minHeight <= 0 {
		return
	}
	a.pillHitMu.RLock()
	currHeight := a.overlayHeight
	a.pillHitMu.RUnlock()
	if currHeight > 0 && minHeight <= currHeight {
		return
	}
	a.resizeOverlayHeight(minHeight)
}

func (a *App) resizeOverlayHeight(minHeight int) {
	vintaHwnd := a.getHwnd()
	if vintaHwnd == 0 {
		return
	}

	vLeft, vTop, vWidth, vHeight, _, monitors := GetDisplayMetrics()
	a.pillHitMu.RLock()
	uiScale := a.uiScale
	currHeight := a.overlayHeight
	a.pillHitMu.RUnlock()

	overlayHeight := ComputeOverlayHeightForScale(vTop, vHeight, monitors, uiScale, a.effectiveScale())
	if minHeight > overlayHeight {
		overlayHeight = minHeight
		if vHeight > 0 && overlayHeight >= vHeight {
			overlayHeight = vHeight - 1
		}
	}
	if overlayHeight == currHeight {
		return
	}

	targetW := vWidth
	if targetW <= 0 {
		targetW = 1920
	}

	a.pillHitMu.Lock()
	a.overlayHeight = overlayHeight
	a.pillHitMu.Unlock()

	const swpNoZOrder = 0x0004
	_, _, _ = procSetWindowPos.Call(
		vintaHwnd, 0,
		uintptr(vLeft), uintptr(vTop), uintptr(targetW), uintptr(overlayHeight),
		swpNoZOrder|SWP_NOACTIVATE,
	)
}

// InvalidateDisplayCache clears the cached display metrics so the next query re-enumerates monitors.
func InvalidateDisplayCache() {
	cachedDisplayMu.Lock()
	lastDisplayFetch = time.Time{}
	cachedDisplayMu.Unlock()
}

// GetDisplayMetrics returns virtual screen bounds and connected monitors, caching results for 1s unless geometry or monitor count changes.
func GetDisplayMetrics() (vLeft, vTop, vWidth, vHeight, pCenter int, monitors []MonitorInfo) {
	const smCMonitors = 80
	vx, _, _ := procGetSystemMetrics.Call(SM_XVIRTUALSCREEN)
	vy, _, _ := procGetSystemMetrics.Call(SM_YVIRTUALSCREEN)
	vw, _, _ := procGetSystemMetrics.Call(SM_CXVIRTUALSCREEN)
	vh, _, _ := procGetSystemMetrics.Call(SM_CYVIRTUALSCREEN)
	mc, _, _ := procGetSystemMetrics.Call(smCMonitors)

	currVx, currVy, currVw, currVh, currMc := int(int32(vx)), int(int32(vy)), int(vw), int(vh), int(mc)

	isCacheValid := func() bool {
		return time.Since(lastDisplayFetch) < time.Second &&
			cachedDisplayLayout.VirtualWidth > 0 &&
			cachedDisplayLayout.VirtualWidth == currVw &&
			cachedDisplayLayout.VirtualHeight == currVh &&
			cachedDisplayLayout.VirtualLeft == currVx &&
			cachedDisplayLayout.VirtualTop == currVy &&
			(currMc <= 0 || len(cachedDisplayLayout.Monitors) == currMc)
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
			IsPrimary: true, IsPortrait: vHeight > primaryW, DpiScale: 1.0,
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
	return MonitorInfo{ID: 1, Name: "Primary Display", Width: 1920, Height: 1080, IsPrimary: true, DpiScale: 1.0}
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

	myPid := uint32(os.Getpid())
	titlePtr, _ := syscall.UTF16PtrFromString("Vinta")
	if hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr))); hwnd != 0 && winPID(hwnd) == myPid {
		cachedHwndMu.Lock()
		cachedWindowHwnd = hwnd
		cachedHwndMu.Unlock()
		return hwnd
	}

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
	if isFullscreenStyle(style) {
		return true
	}
	return strings.Contains(strings.ToLower(getWindowAppName(hwnd)), "stardew")
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

func getMonitorDpiScale(hMonitor uintptr) float64 {
	if hMonitor != 0 && procGetDpiForMonitor.Find() == nil {
		var dpiX, dpiY uint32
		// MDT_EFFECTIVE_DPI = 0
		ret, _, _ := procGetDpiForMonitor.Call(hMonitor, 0, uintptr(unsafe.Pointer(&dpiX)), uintptr(unsafe.Pointer(&dpiY)))
		if ret == 0 && dpiX > 0 {
			return float64(dpiX) / 96.0
		}
	}
	return 1.0
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
					DpiScale: getMonitorDpiScale(hMonitor),
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

	monitorsSignature := func(mons []MonitorInfo) string {
		var sb strings.Builder
		for _, m := range mons {
			fmt.Fprintf(&sb, "%d:%d,%d,%dx%d,%v;", m.ID, m.Left, m.Top, m.Width, m.Height, m.IsPrimary)
		}
		return sb.String()
	}

	type monEvalCache struct {
		fg         uintptr
		style      uint32
		winRect    rect32
		shouldHide bool
		evalTime   time.Time
	}

	initVLeft, initVTop, initVWidth, initVHeight, _, initMonitors := GetDisplayMetrics()
	var (
		autoHiddenByFullscreen bool
		autoHiddenMu           sync.Mutex
		fgEvalMu               sync.Mutex
		wasSubordinatedToSnip  bool
		lastSnipCheckFg        uintptr
		lastSnipEnumTime       time.Time
		// Cache fullscreen evaluations per monitor ID; sharing one slot across monitors causes
		// Monitor B (covers=false) to overwrite Monitor A and bounce Vinta back onto a game.
		evalCacheByMon   = make(map[int]monEvalCache, 4)
		lastVLeft        = initVLeft
		lastVTop         = initVTop
		lastVWidth       = initVWidth
		lastVHeight      = initVHeight
		lastMonitorCount = len(initMonitors)
		lastMonitorsSig  = monitorsSignature(initMonitors)
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
			if curFg != lastSnipCheckFg || isSnippingTool(curFg) || wasSubordinatedToSnip || time.Since(lastSnipEnumTime) >= 3*time.Second {
				lastSnipCheckFg = curFg
				lastSnipEnumTime = time.Now()
				snipHwnd = findActiveSnippingWindow(curFg)
			}
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
		fgTopmost := false
		if curFg != 0 {
			_, fgExStyle := winStyle(curFg)
			fgTopmost = (fgExStyle & WS_EX_TOPMOST) != 0
		}
		coveredByTopmost := false
		for prev, _, _ := procGetWindow.Call(vintaHwnd, GW_HWNDPREV); prev != 0; prev, _, _ = procGetWindow.Call(prev, GW_HWNDPREV) {
			vis, _, _ := procIsWindowVisible.Call(prev)
			if vis == 0 {
				continue
			}
			if isSnippingTool(prev) {
				coveredByTopmost = false
				break
			}
			var rc rect32
			if res, _, _ := procGetWindowRect.Call(prev, uintptr(unsafe.Pointer(&rc))); res != 0 && rc.Right > rc.Left && rc.Bottom > rc.Top {
				coveredByTopmost = true
				break
			}
		}
		if !wasSubordinatedToSnip && !fgTopmost && !coveredByTopmost && (exStyle&WS_EX_TOPMOST) != 0 {
			return
		}
		wasSubordinatedToSnip = false
		_, _, _ = procSetWindowPos.Call(vintaHwnd, ^uintptr(0), 0, 0, 0, 0, SWP_NOSIZE|SWP_NOMOVE|SWP_NOACTIVATE)
	}

	evaluateForegroundState := func() {
		fgEvalMu.Lock()
		defer fgEvalMu.Unlock()
		if !services.IsFrontendReady() {
			return
		}

		vLeft, vTop, vWidth, vHeight, pCenter, monitors := GetDisplayMetrics()
		currSig := monitorsSignature(monitors)
		if vLeft != lastVLeft || vTop != lastVTop || vWidth != lastVWidth || vHeight != lastVHeight || len(monitors) != lastMonitorCount || currSig != lastMonitorsSig {
			lastVLeft, lastVTop, lastVWidth, lastVHeight = vLeft, vTop, vWidth, vHeight
			lastMonitorCount = len(monitors)
			lastMonitorsSig = currSig
			a.pillHitMu.RLock()
			uiScale := a.uiScale
			a.pillHitMu.RUnlock()
			overlayHeight := ComputeOverlayHeightForScale(vTop, vHeight, monitors, uiScale, a.effectiveScale())
			a.pillHitMu.Lock()
			a.overlayHeight = overlayHeight
			a.pillHitMu.Unlock()
			targetW := vWidth
			if targetW <= 0 {
				targetW = 1920
			}
			const swpNoZOrder = 0x0004
			_, _, _ = procSetWindowPos.Call(
				vintaHwnd, 0,
				uintptr(vLeft), uintptr(vTop), uintptr(targetW), uintptr(overlayHeight),
				swpNoZOrder|SWP_NOACTIVATE,
			)

			// If Vinta was on a disconnected monitor, snap back onto an active monitor.
			a.fullscreenRelocatedMu.Lock()
			curOffsetX := a.offsetX
			if len(monitors) <= 1 {
				a.isFullscreenRelocated = false
				a.fullscreenTriggerHwnd = 0
				a.fullscreenTriggerMonID = 0
			}
			a.fullscreenRelocatedMu.Unlock()

			screenCenterX := vLeft + pCenter + int(float64(curOffsetX)*a.effectiveScale())
			onActiveMonitor := false
			for _, mon := range monitors {
				if screenCenterX >= mon.Left && screenCenterX < mon.Left+mon.Width {
					onActiveMonitor = true
					break
				}
			}
			if !onActiveMonitor && len(monitors) > 0 {
				fallbackMon := resolveCurrentMonitor(screenCenterX, monitors)
				targetOffsetX := a.monitorTargetOffset(fallbackMon, vLeft, pCenter)
				a.fullscreenRelocatedMu.Lock()
				relocateGraceUntil = time.Now().Add(1500 * time.Millisecond)
				relocateTargetX = targetOffsetX
				a.offsetX = targetOffsetX
				a.fullscreenRelocatedMu.Unlock()
				if app := application.Get(); app != nil && services.IsFrontendReady() {
					app.Event.Emit("active-monitor-follow", map[string]interface{}{
						"targetMonitor": fallbackMon, "targetOffset": targetOffsetX,
					})
				}
			}

			a.updateWindowRegion()
			if app := application.Get(); app != nil {
				app.Event.Emit("display-layout-changed", nil)
			}
		}

		if a.IsOutboundDragging() {
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

		if curFg != lastSnipCheckFg || isSnippingTool(curFg) || wasSubordinatedToSnip || time.Since(lastSnipEnumTime) >= 3*time.Second {
			lastSnipCheckFg = curFg
			lastSnipEnumTime = time.Now()
			if snipHwnd := findActiveSnippingWindow(curFg); snipHwnd != 0 {
				exitFullscreenHide()
				reassertTopmost(curFg, snipHwnd, true)
				return
			}
		}

		vintaMon := a.GetCurrentMonitor()

		a.autoHideMu.RLock()
		hideBorderless := a.autoHideBorderless
		relocateOnFullscreen := a.relocateOnFullscreen
		returnOnExit := a.returnOnFullscreenExit
		a.autoHideMu.RUnlock()

		shouldHideForMonitor := func(mon MonitorInfo) (bool, string, string) {
			covers, _ := windowCoversMonitor(curFg, mon)
			if !covers {
				delete(evalCacheByMon, mon.ID)
				return false, "", ""
			}
			style, _ := winStyle(curFg)
			isMaximizedStandard := (style&WS_CAPTION == WS_CAPTION) && (style&WS_MAXIMIZE != 0)
			if isMaximizedStandard {
				delete(evalCacheByMon, mon.ID)
				return false, "", ""
			}
			var fgWinRect rect32
			_, _, _ = procGetWindowRect.Call(curFg, uintptr(unsafe.Pointer(&fgWinRect)))
			if cached, ok := evalCacheByMon[mon.ID]; ok &&
				curFg == cached.fg &&
				style == cached.style &&
				fgWinRect == cached.winRect &&
				!cached.shouldHide &&
				time.Since(cached.evalTime) < 2400*time.Millisecond {
				return false, "", ""
			}
			isFramelessOrPopup := isFullscreenStyle(style)
			appName := getWindowAppName(curFg)
			isUnsupportedGame := strings.Contains(strings.ToLower(appName), "stardew")

			var qState uint32
			goruntime.LockOSThread()
			coRes, _, _ := procCoInitializeEx.Call(0, 0)
			ret, _, _ := procSHQueryUserNotificationState.Call(uintptr(unsafe.Pointer(&qState)))
			if coRes == 0 || coRes == 1 {
				_, _, _ = procCoUninitialize.Call()
			}
			goruntime.UnlockOSThread()

			var hideRes bool
			var reasonRes string
			switch {
			case ret == 0 && qState == QUNS_PRESENTATION_MODE:
				hideRes, reasonRes = true, "presentation_mode"
			case isUnsupportedGame:
				hideRes, reasonRes = true, "unsupported_app"
			case (hasExclusiveDisplayOwnership() || (ret == 0 && qState == QUNS_RUNNING_D3D_FULL_SCREEN)) && isFramelessOrPopup:
				hideRes, reasonRes = true, "exclusive_fullscreen"
			case isFramelessOrPopup && hideBorderless:
				hideRes, reasonRes = true, "borderless_fullscreen"
			default:
				hideRes, reasonRes = false, ""
			}
			evalCacheByMon[mon.ID] = monEvalCache{
				fg:         curFg,
				style:      style,
				winRect:    fgWinRect,
				shouldHide: hideRes,
				evalTime:   time.Now(),
			}
			return hideRes, reasonRes, appName
		}

		// Only trigger fullscreen hide/relocation if the foreground app actually covers Vinta's current monitor.
		shouldHide, reason, appName := shouldHideForMonitor(vintaMon)

		if shouldHide {
			if relocateOnFullscreen {
				for _, m := range monitors {
					if m.ID != vintaMon.ID && !isMonitorCoveredByFullscreen(m, vintaHwnd) {
						targetOffsetX := a.monitorTargetOffset(m, vLeft, pCenter)

						a.fullscreenRelocatedMu.Lock()
						if returnOnExit && !a.isFullscreenRelocated {
							a.originalFullscreenOffset = a.offsetX
							a.isFullscreenRelocated = true
						}
						a.fullscreenTriggerHwnd = curFg
						a.fullscreenTriggerMonID = vintaMon.ID
						relocateGraceUntil = time.Now().Add(1500 * time.Millisecond)
						relocateTargetX = targetOffsetX
						a.offsetX = targetOffsetX
						a.fullscreenRelocatedMu.Unlock()

						exitFullscreenHide()
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
			}

			a.fullscreenRelocatedMu.Lock()
			a.fullscreenTriggerHwnd = curFg
			a.fullscreenTriggerMonID = vintaMon.ID
			a.fullscreenRelocatedMu.Unlock()

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

		a.fullscreenRelocatedMu.Lock()
		wasRelocated := a.isFullscreenRelocated
		triggerHwnd := a.fullscreenTriggerHwnd
		triggerMonID := a.fullscreenTriggerMonID
		restoreOffset := a.originalFullscreenOffset
		a.fullscreenRelocatedMu.Unlock()

		triggerAppActive := false
		if triggerHwnd != 0 {
			if ret, _, _ := procIsWindow.Call(triggerHwnd); ret != 0 {
				vis, _, _ := procIsWindowVisible.Call(triggerHwnd)
				iconic, _, _ := procIsIconic.Call(triggerHwnd)
				if vis != 0 || iconic != 0 {
					triggerAppActive = true
					// If the trigger app itself is focused, visible, and no longer fullscreen (e.g. F11 exited), re-enable.
					if curFg == triggerHwnd && vis != 0 && iconic == 0 {
						for _, m := range monitors {
							if m.ID == triggerMonID {
								if trigHide, _, _ := shouldHideForMonitor(m); !trigHide {
									triggerAppActive = false
								}
								break
							}
						}
					}
				}
			}
			if !triggerAppActive {
				a.fullscreenRelocatedMu.Lock()
				a.fullscreenTriggerHwnd = 0
				a.fullscreenTriggerMonID = 0
				a.fullscreenRelocatedMu.Unlock()
			}
		}

		if wasRelocated {
			if !returnOnExit {
				a.fullscreenRelocatedMu.Lock()
				a.isFullscreenRelocated = false
				a.fullscreenRelocatedMu.Unlock()
			} else if !triggerAppActive {
				a.fullscreenRelocatedMu.Lock()
				a.isFullscreenRelocated = false
				relocateGraceUntil = time.Now().Add(1500 * time.Millisecond)
				relocateTargetX = restoreOffset
				a.offsetX = restoreOffset
				a.fullscreenRelocatedMu.Unlock()

				a.updateWindowRegion()
				if app := application.Get(); app != nil && services.IsFrontendReady() {
					app.Event.Emit("game-mode-exit-relocate", map[string]interface{}{
						"restoreOffset": restoreOffset,
					})
				}
				return
			}
		}
	}

	go func() {
		stopCtx := a.lifecycleContext()
		ticker := time.NewTicker(600 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopCtx.Done():
				return
			case <-ticker.C:
				evaluateForegroundState()
			}
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

		// Hidden top-level listener window for WM_DISPLAYCHANGE, WM_SETTINGCHANGE, and
		// event-driven mouse Raw Input (WM_INPUT via RIDEV_INPUTSINK) so WS_EX_TRANSPARENT
		// reconciles in <1ms without 60Hz polling or global mouse hooks.
		const (
			wmDisplayChange = 0x007E
			wmSettingChange = 0x001A
			wmInput         = 0x00FF
			spiSetWorkArea  = 0x002F
			wsPopup         = 0x80000000
			ridevInputSink  = 0x00000100
		)
		classNamePtr, _ := windows.UTF16PtrFromString("VintaDisplayEventListener")
		wndProc := syscall.NewCallback(func(h uintptr, msg uint32, wParam, lParam uintptr) uintptr {
			switch msg {
			case wmInput:
				a.reconcileHitTest(vintaHwnd)
			case wmDisplayChange:
				InvalidateDisplayCache()
				time.AfterFunc(120*time.Millisecond, evaluateForegroundState)
			case wmSettingChange:
				if wParam == spiSetWorkArea {
					InvalidateDisplayCache()
					time.AfterFunc(120*time.Millisecond, evaluateForegroundState)
				}
			}
			r, _, _ := procDefWindowProc.Call(h, uintptr(msg), wParam, lParam)
			return r
		})
		type wndClassExW struct {
			CbSize        uint32
			Style         uint32
			LpfnWndProc   uintptr
			CbClsExtra    int32
			CbWndExtra    int32
			HInstance     uintptr
			HIcon         uintptr
			HCursor       uintptr
			HbrBackground uintptr
			LpszMenuName  *uint16
			LpszClassName *uint16
			HIconSm       uintptr
		}
		wc := wndClassExW{
			CbSize:        uint32(unsafe.Sizeof(wndClassExW{})),
			LpfnWndProc:   wndProc,
			LpszClassName: classNamePtr,
		}
		atom, _, _ := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
		if atom != 0 {
			listenerHwnd, _, _ := procCreateWindowEx.Call(
				0, uintptr(unsafe.Pointer(classNamePtr)), 0, wsPopup,
				0, 0, 0, 0, 0, 0, 0, 0,
			)
			if listenerHwnd != 0 && procRegisterRawInputDevices.Find() == nil {
				type rawInputDevice struct {
					UsagePage uint16
					Usage     uint16
					Flags     uint32
					Target    uintptr
				}
				rid := rawInputDevice{
					UsagePage: 0x01,
					Usage:     0x02,
					Flags:     ridevInputSink,
					Target:    listenerHwnd,
				}
				if ok, _, _ := procRegisterRawInputDevices.Call(
					uintptr(unsafe.Pointer(&rid)), 1, unsafe.Sizeof(rid),
				); ok != 0 {
					a.pillHitMu.Lock()
					a.rawInputRegistered = true
					a.pillHitMu.Unlock()
				}
			}
		}
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
	a.fullscreenTriggerHwnd = 0
	a.fullscreenTriggerMonID = 0
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
	a.fullscreenRelocatedMu.Lock()
	offset := float64(a.offsetX)
	a.fullscreenRelocatedMu.Unlock()
	activeGlideController.mu.Unlock()
	return offset
}

// SetIgnoreMouseEvents is retained for binding compatibility and does nothing. The hit-testing
// loop is the single owner of WS_EX_TRANSPARENT; use SetForceInteractive to keep the window clickable.
func (a *App) SetIgnoreMouseEvents(_ bool) {}

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

// TrimMemory flushes unused working-set pages from Vinta and its child WebView2 processes asynchronously.
// Requests arriving during the 500ms cooldown or while a trim is running are coalesced into a single timer
// so they never drop silently or spawn recursive timer loops.
func (a *App) TrimMemory() {
	trimMu.Lock()
	if trimRunning {
		trimPending = true
		trimMu.Unlock()
		return
	}
	elapsed := time.Since(lastTrimTime)
	if elapsed < 500*time.Millisecond {
		if !trimPending {
			trimPending = true
			delay := 500*time.Millisecond - elapsed
			if trimTimer != nil {
				trimTimer.Stop()
			}
			trimTimer = time.AfterFunc(delay, a.TrimMemory)
		}
		trimMu.Unlock()
		return
	}
	if trimTimer != nil {
		trimTimer.Stop()
		trimTimer = nil
	}
	trimRunning = true
	trimPending = false
	lastTrimTime = time.Now()
	trimMu.Unlock()

	go a.executeTrimMemory()
}

func (a *App) executeTrimMemory() {
	defer func() {
		trimMu.Lock()
		trimRunning = false
		rerun := trimPending
		trimPending = false
		if rerun {
			if trimTimer != nil {
				trimTimer.Stop()
			}
			trimTimer = time.AfterFunc(500*time.Millisecond, a.TrimMemory)
		}
		trimMu.Unlock()
	}()

	if procEmptyWorkingSet.Find() != nil {
		return
	}

	currentPID := uint32(os.Getpid())

	enumChildWebviewMu.Lock()
	// Reuse discovered WebView2 process IDs for 30s so TrimMemory avoids walking every OS
	// process via CreateToolhelp32Snapshot on every state transition.
	if len(cachedWebviewPIDs) == 0 || time.Since(lastWebviewScanTime) >= 30*time.Second {
		targetPIDs := make(map[uint32]bool)
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

		if snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0); err == nil {
			var pe windows.ProcessEntry32
			pe.Size = uint32(unsafe.Sizeof(pe))

			type procItem struct{ pid, parentPID uint32 }
			allProcs := make([]procItem, 0, 64)
			for err = windows.Process32First(snapshot, &pe); err == nil; err = windows.Process32Next(snapshot, &pe) {
				if strings.Contains(strings.ToLower(windows.UTF16ToString(pe.ExeFile[:])), "msedgewebview2") {
					allProcs = append(allProcs, procItem{pid: pe.ProcessID, parentPID: pe.ParentProcessID})
				}
			}
			_ = windows.CloseHandle(snapshot)

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
		}

		cachedWebviewPIDs = cachedWebviewPIDs[:0]
		for pid := range targetPIDs {
			if pid != currentPID {
				cachedWebviewPIDs = append(cachedWebviewPIDs, pid)
			}
		}
		lastWebviewScanTime = time.Now()
	}
	pidsToTrim := make([]uint32, len(cachedWebviewPIDs))
	copy(pidsToTrim, cachedWebviewPIDs)
	enumChildWebviewMu.Unlock()

	for _, pid := range pidsToTrim {
		if hChild, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid); err == nil {
			_, _, _ = procEmptyWorkingSet.Call(uintptr(hChild))
			_ = windows.CloseHandle(hChild)
		}
	}

	// Flush OS working-set pages without forcing synchronous Go GC (debug.FreeOSMemory).
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

// SetRelocateOnFullscreen toggles whether Vinta moves to a free monitor instead of hiding during fullscreen apps.
func (a *App) SetRelocateOnFullscreen(enabled bool) {
	a.autoHideMu.Lock()
	a.relocateOnFullscreen = enabled
	a.autoHideMu.Unlock()
}

// SetReturnOnFullscreenExit toggles whether Vinta returns to its previous display after a fullscreen app exits.
func (a *App) SetReturnOnFullscreenExit(enabled bool) {
	a.autoHideMu.Lock()
	a.returnOnFullscreenExit = enabled
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

// IsPrivacySensitiveDisplayActive reports whether Windows is in presentation/busy mode
// or a known screen-sharing toolbar (Teams/Zoom) is currently visible.
func (a *App) IsPrivacySensitiveDisplayActive() bool {
	var qState uint32
	goruntime.LockOSThread()
	coRes, _, _ := procCoInitializeEx.Call(0, 0)
	ret, _, _ := procSHQueryUserNotificationState.Call(uintptr(unsafe.Pointer(&qState)))
	if coRes == 0 || coRes == 1 {
		_, _, _ = procCoUninitialize.Call()
	}
	goruntime.UnlockOSThread()

	if ret == 0 && (qState == QUNS_PRESENTATION_MODE || qState == QUNS_BUSY) {
		return true
	}

	foundShareToolbar := false
	cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		vis, _, _ := procIsWindowVisible.Call(hwnd)
		if vis == 0 {
			return 1
		}
		var clsBuf [128]uint16
		n, _, _ := procGetClassName.Call(hwnd, uintptr(unsafe.Pointer(&clsBuf[0])), uintptr(len(clsBuf)))
		if n == 0 {
			return 1
		}
		cls := syscall.UTF16ToString(clsBuf[:n])
		if cls == "ScreenSharingToolbar" || cls == "ZPFloatToolbarClass" || cls == "ZPToolBarParentWndClass" {
			foundShareToolbar = true
			return 0
		}
		return 1
	})
	_, _, _ = procEnumWindows.Call(cb, 0)
	return foundShareToolbar
}
