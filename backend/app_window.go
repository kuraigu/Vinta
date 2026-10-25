// Package backend provides core application services and state management for Vinta.
package backend

import (
	"fmt"
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

	// Windows Power Throttling / EcoQoS constants
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
	if time.Since(lastDisplayFetch) < 1*time.Second && cachedDisplayLayout.VirtualWidth > 0 &&
		cachedDisplayLayout.VirtualWidth == currVw && cachedDisplayLayout.VirtualHeight == currVh {
		layout := cachedDisplayLayout
		cachedDisplayMu.RUnlock()
		return layout.VirtualLeft, layout.VirtualTop, layout.VirtualWidth, layout.VirtualHeight, layout.PrimaryCenter, layout.Monitors
	}
	cachedDisplayMu.RUnlock()

	cachedDisplayMu.Lock()
	defer cachedDisplayMu.Unlock()

	if time.Since(lastDisplayFetch) < 1*time.Second && cachedDisplayLayout.VirtualWidth > 0 &&
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
)

// getMyWindowHwnd locates the native window handle for the active Vinta window.
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

// getHwnd retrieves the window handle, preferring the native Wails v3 window handle.
func (a *App) getHwnd() uintptr {
	if a.window != nil {
		if h := uintptr(a.window.NativeWindow()); h != 0 {
			return h
		}
	}
	return getMyWindowHwnd()
}

// isDesktopOrShell determines whether the window belongs to standard desktop backgrounds, shells, or taskbars.
func isDesktopOrShell(hwnd uintptr) bool {
	if hwnd == 0 {
		return false
	}
	var classBuf [256]uint16
	_, _, _ = procGetClassName.Call(hwnd, uintptr(unsafe.Pointer(&classBuf[0])), 256)
	className := syscall.UTF16ToString(classBuf[:])
	return className == "WorkerW" || className == "Progman" ||
		className == "Shell_TrayWnd" || className == "Shell_SecondaryTrayWnd" ||
		className == "Windows.UI.Core.CoreWindow"
}

// isMonitorCoveredByFullscreen checks whether the specified monitor is occupied by a fullscreen application.
func isMonitorCoveredByFullscreen(m MonitorInfo, vintaHwnd uintptr) bool {
	centerX := m.Left + m.Width/2
	centerY := m.Top + m.Height/2
	ptPacked := (uintptr(uint32(centerX)) & 0xFFFFFFFF) | ((uintptr(uint32(centerY)) & 0xFFFFFFFF) << 32)
	topHwnd, _, _ := procWindowFromPoint.Call(ptPacked)
	if topHwnd == 0 {
		return false
	}
	rootHwnd, _, _ := procGetAncestor.Call(topHwnd, uintptr(GA_ROOT))
	if rootHwnd == 0 || rootHwnd == vintaHwnd {
		return false
	}

	if isDesktopOrShell(rootHwnd) || isSnippingTool(rootHwnd) || isScreenClippingHost(rootHwnd) {
		return false
	}

	vis, _, _ := procIsWindowVisible.Call(rootHwnd)
	iconic, _, _ := procIsIconic.Call(rootHwnd)
	if vis == 0 || iconic != 0 {
		return false
	}

	var rc rect32
	res, _, _ := procGetWindowRect.Call(rootHwnd, uintptr(unsafe.Pointer(&rc)))
	if res == 0 {
		return false
	}

	covers := int(rc.Left) <= m.Left+4 &&
		int(rc.Top) <= m.Top+4 &&
		int(rc.Right) >= m.Left+m.Width-4 &&
		int(rc.Bottom) >= m.Top+m.Height-4

	if !covers {
		return false
	}

	var style uint32
	if procGetWindowLong.Find() == nil {
		gwl := int32(GWL_STYLE)
		resStyle, _, _ := procGetWindowLong.Call(rootHwnd, uintptr(gwl))
		style = uint32(resStyle)
	}

	// Maximized windows with a caption/titlebar are not fullscreen applications
	if (style&WS_CAPTION == WS_CAPTION) && ((style & WS_MAXIMIZE) != 0) {
		return false
	}

	return (style&WS_CAPTION) == 0 || (style&WS_POPUP) != 0
}

// windowCoversMonitor determines whether the foreground window covers the monitor where Vinta is hosted.
func windowCoversMonitor(fgHwnd uintptr, vintaMon MonitorInfo) (bool, monitorInfoW) {
	var mi monitorInfoW
	if fgHwnd == 0 {
		return false, mi
	}

	if isSnippingTool(fgHwnd) || isScreenClippingHost(fgHwnd) || isDesktopOrShell(fgHwnd) {
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

// hasExclusiveDisplayOwnership queries whether any graphics context currently holds exclusive hardware ownership.
func hasExclusiveDisplayOwnership() bool {
	if procD3DKMTCheckExclusiveOwnership.Find() != nil {
		return false
	}
	res, _, _ := procD3DKMTCheckExclusiveOwnership.Call()
	return res != 0
}

// isSnippingTool checks whether the window belongs to Windows Snipping Tool, ScreenClippingHost, or related overlays.
func isSnippingTool(hwnd uintptr) bool {
	if hwnd == 0 {
		return false
	}

	checkHwnd := func(h uintptr) bool {
		var pid uint32
		_, _, _ = procGetWindowThreadProcessID.Call(h, uintptr(unsafe.Pointer(&pid)))
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
	root, _, _ := procGetAncestor.Call(hwnd, 2)
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

			if isScreenClippingHost(h) || isSnippingTool(h) {
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

// setupForegroundEventHook registers WinEvent hooks on a dedicated OS thread to maintain topmost Z-order.
func (a *App) setupForegroundEventHook(vintaHwnd uintptr) {
	if vintaHwnd == 0 {
		return
	}

	procGetMessage := user32.NewProc("GetMessageW")
	procDispatchMessage := user32.NewProc("DispatchMessageW")
	procTranslateMessage := user32.NewProc("TranslateMessage")
	procUnhookWinEvent := user32.NewProc("UnhookWinEvent")

	myPid := uint32(os.Getpid())

	isVintaForeground := func(fg uintptr) bool {
		if fg == 0 {
			return false
		}
		if fg == vintaHwnd {
			return true
		}
		root, _, _ := procGetAncestor.Call(fg, 2)
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

	var (
		autoHiddenByFullscreen bool
		autoHiddenMu           sync.Mutex
	)

	reassertTopmost := func(curFg uintptr) {
		snipHwnd := findActiveSnippingWindow(curFg)
		if snipHwnd != 0 {
			// Subordinate Vinta immediately behind Snipping Tool window so its top toolbar controls stay visible and interactive
			_, _, _ = procSetWindowPos.Call(
				vintaHwnd,
				snipHwnd,
				0, 0, 0, 0,
				0x0001|0x0002|0x0010|0x0400, // SWP_NOSIZE | SWP_NOMOVE | SWP_NOACTIVATE | SWP_NOSENDCHANGING
			)
			return
		}

		autoHiddenMu.Lock()
		isHiddenByFS := autoHiddenByFullscreen
		autoHiddenMu.Unlock()
		if isHiddenByFS {
			return
		}

		// Reassert Vinta as topmost when no Snipping Tool overlay is active
		_, _, _ = procSetWindowPos.Call(
			vintaHwnd,
			uintptr(^uintptr(0)), // HWND_TOPMOST
			0, 0, 0, 0,
			0x0001|0x0002|0x0010, // SWP_NOSIZE | SWP_NOMOVE | SWP_NOACTIVATE
		)
	}

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

		// 1. If currently relocated, check whether the triggering fullscreen application is still active/open
		a.fullscreenRelocatedMu.Lock()
		if a.isFullscreenRelocated {
			triggerHwnd := a.fullscreenTriggerHwnd
			isAlive := false
			if triggerHwnd != 0 {
				ret, _, _ := procIsWindow.Call(triggerHwnd)
				if ret != 0 {
					vis, _, _ := procIsWindowVisible.Call(triggerHwnd)
					iconic, _, _ := procIsIconic.Call(triggerHwnd)
					if vis != 0 && iconic == 0 {
						isAlive = true
					}
				}
			}

			if isAlive {
				// The fullscreen application is still running. Keep Vinta on the secondary monitor.
				a.fullscreenRelocatedMu.Unlock()
				reassertTopmost(curFg)
				return
			}

			// The trigger window has closed or minimized. Restore Vinta to its original monitor.
			a.isFullscreenRelocated = false
			a.fullscreenTriggerHwnd = 0
			a.fullscreenTriggerMonID = 0
			restoreOffset := a.originalFullscreenOffset
			origMonID := a.originalFullscreenMonID
			a.offsetX = restoreOffset
			a.lastRelocateExitTime = time.Now()
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
					app.Event.Emit("game-mode-exit-relocate", map[string]interface{}{
						"restoredMonitor": origMonName,
						"restoreOffset":   restoreOffset,
					})
				}
			}
			reassertTopmost(curFg)
			return
		}
		timeSinceRelocateExit := time.Since(a.lastRelocateExitTime)
		a.fullscreenRelocatedMu.Unlock()

		// Settle cooldown immediately after returning from relocation to prevent bounce loops
		if timeSinceRelocateExit < 3*time.Second {
			reassertTopmost(curFg)
			return
		}

		vintaMon := a.GetCurrentMonitor()
		covers, _ := windowCoversMonitor(curFg, vintaMon)

		// Dynamic Fullscreen & Exclusive Ownership Evaluation
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
		isD3DFullScreen := ret == 0 && qState == QUNS_RUNNING_D3D_FULL_SCREEN

		if covers {
			var style uint32
			if procGetWindowLong.Find() == nil {
				gwl := int32(GWL_STYLE)
				resStyle, _, _ := procGetWindowLong.Call(curFg, uintptr(gwl))
				style = uint32(resStyle)
			}

			// Standard desktop windows with a titlebar are never fullscreen games
			isMaximizedStandard := (style&WS_CAPTION == WS_CAPTION) && ((style & WS_MAXIMIZE) != 0)
			isFrameless := (style & (WS_CAPTION | WS_THICKFRAME)) == 0
			isPopupFullscreen := (style & WS_POPUP) != 0

			if !isMaximizedStandard {
				a.autoHideMu.RLock()
				hideBorderless := a.autoHideBorderless
				a.autoHideMu.RUnlock()

				appNameLower := strings.ToLower(appName)
				isKnownUnsupportedGame := strings.Contains(appNameLower, "stardew")

				if hasEx || isD3DFullScreen {
					shouldHide = true
					reason = "exclusive_fullscreen"
				} else if isKnownUnsupportedGame {
					shouldHide = true
					reason = "unsupported_app"
				} else if (isFrameless || isPopupFullscreen) && hideBorderless {
					shouldHide = true
					reason = "borderless_fullscreen"
				}
			}
		}

		if shouldHide {
			if len(monitors) > 1 {
				var targetMon MonitorInfo
				foundTarget := false
				for _, m := range monitors {
					if m.ID != vintaMon.ID {
						// Ensure this candidate monitor is not also occupied by a fullscreen application
						if !isMonitorCoveredByFullscreen(m, vintaHwnd) {
							targetMon = m
							foundTarget = true
							break
						}
					}
				}

				if foundTarget {
					a.fullscreenRelocatedMu.Lock()
					if !a.isFullscreenRelocated {
						a.isFullscreenRelocated = true
						a.fullscreenTriggerHwnd = curFg
						a.fullscreenTriggerMonID = vintaMon.ID
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
				if vintaHwnd != 0 {
					// Drop topmost and hide native window completely to avoid interfering with exclusive DirectX mode
					_, _, _ = procSetWindowPos.Call(
						vintaHwnd,
						uintptr(^uintptr(1)), // HWND_NOTOPMOST
						0, 0, 0, 0,
						SWP_NOSIZE|SWP_NOMOVE|SWP_NOACTIVATE|SWP_HIDEWINDOW,
					)
				}
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

		autoHiddenMu.Lock()
		if autoHiddenByFullscreen {
			autoHiddenByFullscreen = false
			services.SetAppHidden(false)
			if vintaHwnd != 0 {
				_, _, _ = procSetWindowPos.Call(
					vintaHwnd,
					uintptr(^uintptr(0)), // HWND_TOPMOST
					0, 0, 0, 0,
					SWP_NOSIZE|SWP_NOMOVE|SWP_NOACTIVATE|SWP_SHOWWINDOW,
				)
			}
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

		// Active Display Follow Mode
		a.autoHideMu.RLock()
		followDisplay := a.followActiveDisplay
		a.autoHideMu.RUnlock()

		if followDisplay && len(monitors) > 1 && !a.isFullscreenRelocated {
			hMonFg, _, _ := procMonitorFromWindow.Call(curFg, uintptr(MONITOR_DEFAULTTONEAREST))
			if hMonFg != 0 {
				var fgMi monitorInfoW
				fgMi.CbSize = uint32(unsafe.Sizeof(fgMi))
				if res, _, _ := procGetMonitorInfoW.Call(hMonFg, uintptr(unsafe.Pointer(&fgMi))); res != 0 {
					var activeMon MonitorInfo
					foundMon := false
					for _, m := range monitors {
						if m.Left == int(fgMi.RcMonitor.Left) && m.Top == int(fgMi.RcMonitor.Top) {
							activeMon = m
							foundMon = true
							break
						}
					}
					if foundMon && activeMon.ID != vintaMon.ID {
						targetCenterScreenX := activeMon.Left + (activeMon.Width / 2)
						scaledOffset := targetCenterScreenX - (vLeft + pCenter)
						targetOffsetX := int(float64(scaledOffset) / scale)
						if targetOffsetX != a.offsetX {
							a.offsetX = targetOffsetX
							a.updateWindowRegion()
							if services.IsFrontendReady() {
								if app := application.Get(); app != nil {
									app.Event.Emit("active-monitor-follow", map[string]interface{}{
										"targetMonitor": activeMon,
										"targetOffset":  targetOffsetX,
									})
								}
							}
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

		cb := syscall.NewCallback(func(_ uintptr, event uint32, fgHwnd uintptr, _ int32, _ int32, _ uint32, _ uint32) uintptr {
			if event == EVENT_SYSTEM_FOREGROUND && fgHwnd != 0 && !isVintaForeground(fgHwnd) {
				time.AfterFunc(40*time.Millisecond, func() {
					evaluateForegroundState()
				})
				time.AfterFunc(200*time.Millisecond, func() {
					evaluateForegroundState()
				})
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

// GetDisplayLayout returns multi-monitor geometry and composite virtual desktop layout.
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

// CancelFullscreenRelocate clears any temporary relocation state.
func (a *App) CancelFullscreenRelocate() {
	a.fullscreenRelocatedMu.Lock()
	a.isFullscreenRelocated = false
	a.fullscreenRelocatedMu.Unlock()
}

// FocusWindow requests operating system input focus for the Vinta window.
func (a *App) FocusWindow() {
	hwnd := a.getHwnd()
	if hwnd != 0 {
		curFg, _, _ := procGetForegroundWindow.Call()
		if curFg != hwnd {
			root, _, _ := procGetAncestor.Call(curFg, 2)
			if root != hwnd {
				_, _, _ = procSetForegroundWindow.Call(hwnd)
			}
		}
	}
}

// SetDragCapture controls Win32 OS-level mouse capture during drag gestures.
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

var (
	lastTrimTime time.Time
	trimMu       sync.Mutex
)

// TrimMemory requests Windows to reclaim unused physical RAM pages from Vinta and child WebView2 processes.
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
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err == nil {
		defer func() {
			_ = windows.CloseHandle(snapshot)
		}()

		var pe windows.ProcessEntry32
		pe.Size = uint32(unsafe.Sizeof(pe))

		type procItem struct {
			pid       uint32
			parentPID uint32
			exe       string
		}
		var allProcs []procItem

		if err = windows.Process32First(snapshot, &pe); err == nil {
			for {
				allProcs = append(allProcs, procItem{
					pid:       pe.ProcessID,
					parentPID: pe.ParentProcessID,
					exe:       strings.ToLower(windows.UTF16ToString(pe.ExeFile[:])),
				})
				if err = windows.Process32Next(snapshot, &pe); err != nil {
					break
				}
			}
		}

		childPIDs := make(map[uint32]bool)
		for _, p := range allProcs {
			if p.parentPID == currentPID && strings.Contains(p.exe, "msedgewebview2") {
				childPIDs[p.pid] = true
			}
		}
		for _, p := range allProcs {
			if childPIDs[p.parentPID] && strings.Contains(p.exe, "msedgewebview2") {
				childPIDs[p.pid] = true
			}
		}

		for pid := range childPIDs {
			hChild, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
			if err == nil {
				_, _, _ = procEmptyWorkingSet.Call(uintptr(hChild))
				_ = windows.CloseHandle(hChild)
			}
		}
	}

	hProcess, _, _ := procGetCurrentProcess.Call()
	if hProcess != 0 {
		_, _, _ = procEmptyWorkingSet.Call(hProcess)
	}
}

// disableProcessPowerThrottling opts the application out of Windows 11 Efficiency Mode and timer throttling.
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

// SetIsHidden informs native background services whether Vinta is hidden to throttle high-frequency IPC.
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

// SetFollowActiveDisplay configures whether Vinta dynamically tracks the active monitor with the mouse or foreground window.
func (a *App) SetFollowActiveDisplay(enabled bool) {
	a.autoHideMu.Lock()
	a.followActiveDisplay = enabled
	a.autoHideMu.Unlock()
}

// SetWidgetBounds is called by the frontend whenever the widget dimensions change.
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
