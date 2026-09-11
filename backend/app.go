// Package backend provides the core application logic for Vinta.
package backend

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows/registry"

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
	procSetWinEventHook          = user32.NewProc("SetWinEventHook")
	procEnumDisplayMonitors      = user32.NewProc("EnumDisplayMonitors")
	procGetMonitorInfoW          = user32.NewProc("GetMonitorInfoW")
	procGetWindowRect            = user32.NewProc("GetWindowRect")
	procMonitorFromWindow        = user32.NewProc("MonitorFromWindow")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procEnumChildWindows         = user32.NewProc("EnumChildWindows")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procIsIconic                 = user32.NewProc("IsIconic")

	shell32                          = syscall.NewLazyDLL("shell32.dll")
	procSHQueryUserNotificationState = shell32.NewProc("SHQueryUserNotificationState")

	ole32              = syscall.NewLazyDLL("ole32.dll")
	procCoInitializeEx = ole32.NewProc("CoInitializeEx")
	procCoUninitialize = ole32.NewProc("CoUninitialize")

	gdi32                  = syscall.NewLazyDLL("gdi32.dll")
	procCreateRoundRectRgn = gdi32.NewProc("CreateRoundRectRgn")
	procCombineRgn         = gdi32.NewProc("CombineRgn")
	procSetWindowRgn       = user32.NewProc("SetWindowRgn")
	procDeleteObject       = gdi32.NewProc("DeleteObject")
	procSetCapture         = user32.NewProc("SetCapture")
	procReleaseCapture     = user32.NewProc("ReleaseCapture")

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

	dwmapi       = syscall.NewLazyDLL("dwmapi.dll")
	procDwmFlush = dwmapi.NewProc("DwmFlush")
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
	OutsidePadding    = 20
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

type monitorInfoW struct {
	CbSize    uint32
	RcMonitor rect32
	RcWork    rect32
	DwFlags   uint32
}

// MonitorInfo describes a connected physical display's boundaries and properties.
type MonitorInfo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Left      int    `json:"left"`
	Top       int    `json:"top"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	IsPrimary bool   `json:"isPrimary"`
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

func getDisplayMetrics() (vLeft, vTop, vWidth, vHeight, pCenter int, monitors []MonitorInfo) {
	cachedDisplayMu.RLock()
	if time.Since(lastDisplayFetch) < 2*time.Second && cachedDisplayLayout.VirtualWidth > 0 {
		layout := cachedDisplayLayout
		cachedDisplayMu.RUnlock()
		return layout.VirtualLeft, layout.VirtualTop, layout.VirtualWidth, layout.VirtualHeight, layout.PrimaryCenter, layout.Monitors
	}
	cachedDisplayMu.RUnlock()

	cachedDisplayMu.Lock()
	defer cachedDisplayMu.Unlock()

	if time.Since(lastDisplayFetch) < 2*time.Second && cachedDisplayLayout.VirtualWidth > 0 {
		layout := cachedDisplayLayout
		return layout.VirtualLeft, layout.VirtualTop, layout.VirtualWidth, layout.VirtualHeight, layout.PrimaryCenter, layout.Monitors
	}

	procGetSystemMetrics := user32.NewProc("GetSystemMetrics")
	vx, _, _ := procGetSystemMetrics.Call(SM_XVIRTUALSCREEN)
	vy, _, _ := procGetSystemMetrics.Call(SM_YVIRTUALSCREEN)
	vw, _, _ := procGetSystemMetrics.Call(SM_CXVIRTUALSCREEN)
	vh, _, _ := procGetSystemMetrics.Call(SM_CYVIRTUALSCREEN)

	vLeft = int(int32(vx))
	vTop = int(int32(vy))
	vWidth = int(vw)
	vHeight = int(vh)

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
		id := 0
		cb := syscall.NewCallback(func(hMonitor, _, _, _ uintptr) uintptr {
			var mi monitorInfoW
			mi.CbSize = uint32(unsafe.Sizeof(mi))
			ret, _, _ := procGetMonitorInfoW.Call(hMonitor, uintptr(unsafe.Pointer(&mi)))
			if ret != 0 {
				id++
				w := int(mi.RcMonitor.Right - mi.RcMonitor.Left)
				h := int(mi.RcMonitor.Bottom - mi.RcMonitor.Top)
				isPrimary := (mi.DwFlags & 1) != 0
				monitors = append(monitors, MonitorInfo{
					ID:        id,
					Name:      fmt.Sprintf("Display %d", id),
					Left:      int(mi.RcMonitor.Left),
					Top:       int(mi.RcMonitor.Top),
					Width:     w,
					Height:    h,
					IsPrimary: isPrimary,
				})
			}
			return 1
		})
		_, _, _ = procEnumDisplayMonitors.Call(0, 0, cb, 0)
	}

	if len(monitors) == 0 {
		monitors = append(monitors, MonitorInfo{
			ID:        1,
			Name:      "Primary Display",
			Left:      0,
			Top:       0,
			Width:     primaryW,
			Height:    vHeight,
			IsPrimary: true,
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
		ID:        1,
		Name:      "Primary Display",
		Left:      0,
		Top:       0,
		Width:     1920,
		Height:    1080,
		IsPrimary: true,
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
			var classBuf [256]uint16
			_, _, _ = procGetClassName.Call(child, uintptr(unsafe.Pointer(&classBuf[0])), 256)
			className := syscall.UTF16ToString(classBuf[:])
			if className != "IME" && className != "MSCTFIME UI" {
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

// isBorderlessFullscreen determines whether the foreground window is a borderless
// window that covers the monitor where Vinta is currently hosted.
func isBorderlessFullscreen(fgHwnd, vintaHwnd uintptr) bool {
	if fgHwnd == 0 || fgHwnd == vintaHwnd {
		return false
	}

	var classBuf [256]uint16
	_, _, _ = procGetClassName.Call(fgHwnd, uintptr(unsafe.Pointer(&classBuf[0])), 256)
	className := syscall.UTF16ToString(classBuf[:])
	if className == "WorkerW" || className == "Progman" ||
		className == "Shell_TrayWnd" || className == "Shell_SecondaryTrayWnd" ||
		className == "Windows.UI.Core.CoreWindow" {
		return false
	}

	hMonFg, _, _ := procMonitorFromWindow.Call(fgHwnd, uintptr(MONITOR_DEFAULTTONEAREST))
	hMonVinta, _, _ := procMonitorFromWindow.Call(vintaHwnd, uintptr(MONITOR_DEFAULTTONEAREST))
	if hMonFg == 0 || hMonFg != hMonVinta {
		return false
	}

	var mi monitorInfoW
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	res, _, _ := procGetMonitorInfoW.Call(hMonFg, uintptr(unsafe.Pointer(&mi)))
	if res == 0 {
		return false
	}

	var rc rect32
	res, _, _ = procGetWindowRect.Call(fgHwnd, uintptr(unsafe.Pointer(&rc)))
	if res == 0 {
		return false
	}

	coversMonitor := rc.Left <= mi.RcMonitor.Left &&
		rc.Top <= mi.RcMonitor.Top &&
		rc.Right >= mi.RcMonitor.Right &&
		rc.Bottom >= mi.RcMonitor.Bottom
	if !coversMonitor {
		return false
	}

	gwlStyle := int32(GWL_STYLE)
	styleVal, _, _ := procGetWindowLong.Call(fgHwnd, uintptr(gwlStyle))
	style := uint32(styleVal)
	hasCaption := (style & WS_CAPTION) == WS_CAPTION

	return !hasCaption
}

// isSnippingTool checks whether the given window belongs to Windows Snipping Tool or ScreenClippingHost.
func isSnippingTool(hwnd uintptr) bool {
	if hwnd == 0 {
		return false
	}

	var pid uint32
	_, _, _ = procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return false
	}

	hProc, _, _ := procOpenProcess.Call(0x1000, 0, uintptr(pid)) // PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	if hProc == 0 {
		return false
	}
	defer func() {
		_, _, _ = procCloseHandle.Call(hProc)
	}()

	var buf [1024]uint16
	size := uint32(len(buf))
	res, _, _ := procQueryFullProcessImageName.Call(hProc, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if res == 0 {
		return false
	}

	exePath := syscall.UTF16ToString(buf[:size])
	baseName := strings.ToLower(filepath.Base(exePath))
	return baseName == "screenclippinghost.exe" || baseName == "snippingtool.exe"
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
	procGetAncestor := user32.NewProc("GetAncestor")
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

	var activeSnipHwnd uintptr
	cbEnumSnip := syscall.NewCallback(func(h, _ uintptr) uintptr {
		vis, _, _ := procIsWindowVisible.Call(h)
		if vis == 0 {
			return 1
		}

		// A minimized window must never be treated as an active overlay.
		iconic, _, _ := procIsIconic.Call(h)
		if iconic != 0 {
			return 1
		}

		// Snipping overlay toolbars are explicitly WS_EX_TOPMOST (0x00000008).
		// Normal Snipping Tool application windows or dialogs are not topmost.
		exStyle, _, _ := procGetWindowLong.Call(h, uintptr(0xFFFFFFEC)) // GWL_EXSTYLE = -20
		if exStyle&0x00000008 == 0 {
			return 1
		}

		var rc struct {
			Left, Top, Right, Bottom int32
		}
		res, _, _ := procGetWindowRect.Call(h, uintptr(unsafe.Pointer(&rc)))
		if res == 0 || rc.Left <= -10000 || rc.Right <= rc.Left || rc.Bottom <= rc.Top {
			return 1
		}

		if isSnippingTool(h) {
			activeSnipHwnd = h
			return 0
		}
		return 1
	})

	findActiveSnippingWindow := func() uintptr {
		activeSnipHwnd = 0
		_, _, _ = procEnumWindows.Call(cbEnumSnip, 0)
		return activeSnipHwnd
	}

	reassertTopmost := func() {
		snipHwnd := findActiveSnippingWindow()
		if snipHwnd != 0 {
			// Place Vinta immediately behind Snipping Tool so its top toolbar controls stay on top.
			_, _, _ = procSetWindowPos.Call(
				vintaHwnd,
				snipHwnd,
				0, 0, 0, 0,
				0x0001|0x0002|0x0010|0x0400,
			)
			return
		}

		// HWND_TOPMOST = ^uintptr(0), SWP_NOSIZE (0x0001) | SWP_NOMOVE (0x0002) | SWP_NOACTIVATE (0x0010) | SWP_NOOWNERZORDER (0x0200) | SWP_NOSENDCHANGING (0x0400)
		_, _, _ = procSetWindowPos.Call(
			vintaHwnd,
			uintptr(^uintptr(0)),
			0, 0, 0, 0,
			0x0001|0x0002|0x0010|0x0200|0x0400,
		)
	}

	var (
		autoHiddenByFullscreen bool
		autoHiddenMu           sync.Mutex
	)

	evaluateForegroundState := func() {
		curFg, _, _ := procGetForegroundWindow.Call()
		if curFg == 0 || isVintaForeground(curFg) {
			reassertTopmost()
			return
		}

		a.autoHideMu.RLock()
		hideBorderless := a.autoHideBorderless
		a.autoHideMu.RUnlock()

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

		if ret == 0 {
			switch qState {
			case QUNS_RUNNING_D3D_FULL_SCREEN:
				shouldHide = true
				reason = "exclusive_fullscreen"
			case QUNS_PRESENTATION_MODE:
				shouldHide = true
				reason = "presentation_mode"
			}
		}

		if !shouldHide && hideBorderless {
			if isBorderlessFullscreen(curFg, vintaHwnd) {
				shouldHide = true
				reason = "borderless_fullscreen"
			}
		}

		autoHiddenMu.Lock()
		defer autoHiddenMu.Unlock()

		if shouldHide {
			if !autoHiddenByFullscreen {
				autoHiddenByFullscreen = true
				services.SetAppHidden(true)
				if a.ctx != nil {
					runtime.WindowHide(a.ctx)
					runtime.EventsEmit(a.ctx, "game-mode-enter", reason)
				}
			}
			return
		}

		if autoHiddenByFullscreen {
			autoHiddenByFullscreen = false
			services.SetAppHidden(false)
			if a.ctx != nil {
				runtime.WindowShow(a.ctx)
				runtime.EventsEmit(a.ctx, "game-mode-exit")
			}
		}

		reassertTopmost()
	}

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
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
	ctx                context.Context
	icon               []byte
	pillW              int
	pillH              int
	pillR              int
	offsetX            int
	devW               int
	devH               int
	devExtra           int
	scale              float64
	autoHideBorderless bool
	autoHideMu         sync.RWMutex
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
	}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	defer SetupPanicRecovery()
	a.ctx = ctx

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

	go func() {
		var hwnd uintptr
		for i := 0; i < 30; i++ {
			hwnd = getMyWindowHwnd()
			if hwnd != 0 {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		if hwnd != 0 {
			gwl := int32(GWL_EXSTYLE)
			style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwl))
			style = (style | WS_EX_TOOLWINDOW | WS_EX_TOPMOST) &^ WS_EX_NOACTIVATE &^ WS_EX_APPWINDOW &^ WS_EX_LAYERED
			_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)

			a.setupForegroundEventHook(hwnd)

			go services.StartServicesListener(ctx)
			go services.StartClipboardListener(ctx)

			a.updateWindowRegion()
		}
	}()

	a.StartBackgroundUpdateChecker()

	go systray.Run(a.onReady, a.onExit)
}

func (a *App) onReady() {
	systray.SetIcon(a.icon)
	systray.SetTitle("Vinta")
	systray.SetTooltip("Vinta Dynamic Island")

	mShow := systray.AddMenuItem("Show Vinta", "Bring the widget back into view")
	systray.AddSeparator()
	mTheme := systray.AddMenuItem("Toggle Theme", "Switch between Dark and Light mode")
	mShape := systray.AddMenuItem("Toggle Shape", "Switch between Pill and Box shape")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Vinta", "Close the application")

	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				if a.ctx != nil {
					runtime.EventsEmit(a.ctx, "show-pill")
				}
			case <-mTheme.ClickedCh:
				if a.ctx != nil {
					runtime.EventsEmit(a.ctx, "toggle-theme")
				}
			case <-mShape.ClickedCh:
				if a.ctx != nil {
					runtime.EventsEmit(a.ctx, "toggle-shape")
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				if a.ctx != nil {
					runtime.Quit(a.ctx)
				}
				return
			}
		}
	}()
}

func (a *App) onExit() {
	services.StopClipboardListener()
	services.StopServicesListener()
}

// Shutdown is called when Wails application shuts down.
func (a *App) Shutdown(_ context.Context) {
	services.StopClipboardListener()
	services.StopServicesListener()
	systray.Quit()
}

// SetWidgetBounds is called by the frontend whenever the widget dimensions change.
// It updates the native Win32 click-through region so only the visible widget is clickable.
func (a *App) SetWidgetBounds(width int, height int, radius int, scale float64, offsetX int) {
	a.pillW = width
	a.pillH = height
	a.pillR = radius
	a.offsetX = offsetX
	if scale > 0 {
		a.scale = scale
	}
	a.updateWindowRegion()
}

// SetClipboardMonitoring updates the native clipboard copy and paste monitoring state.
func (a *App) SetClipboardMonitoring(copyEnabled bool, pasteEnabled bool) {
	services.SetClipboardMonitoring(copyEnabled, pasteEnabled)
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
	vLeft, vTop, vWidth, vHeight, pCenter, monitors := getDisplayMetrics()
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
	vLeft, _, _, _, pCenter, monitors := getDisplayMetrics()
	scale := a.scale
	if scale <= 0 {
		scale = 1.0
	}
	screenCenterX := vLeft + pCenter + int(float64(a.offsetX)*scale)
	return resolveCurrentMonitor(screenCenterX, monitors)
}

// Quit terminates the application
func (a *App) Quit() {
	services.StopServicesListener()
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
}

// FocusWindow requests OS input focus so typing and keyboard events work immediately on click/focus
func (a *App) FocusWindow() {
	hwnd := getMyWindowHwnd()
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
// When enabled, mouse messages across the monitor are delivered to the window,
// preventing dropped gestures when bounds tighten or cursor moves rapidly.
func (a *App) SetDragCapture(enable bool) {
	hwnd := getMyWindowHwnd()
	if hwnd != 0 {
		if enable {
			_, _, _ = procSetCapture.Call(hwnd)
		} else {
			_, _, _ = procReleaseCapture.Call()
		}
	}
}

type glideController struct {
	mu         sync.Mutex
	stopCh     chan struct{}
	generation uint64
}

var activeGlideController glideController

// StartInertiaGlide runs the physical deceleration natively synchronized with DWM VSync.
func (a *App) StartInertiaGlide(startOffset float64, velocity float64, minOffset float64, maxOffset float64) {
	activeGlideController.mu.Lock()
	if activeGlideController.stopCh != nil {
		close(activeGlideController.stopCh)
		activeGlideController.stopCh = nil
	}
	activeGlideController.generation++
	myGen := activeGlideController.generation
	stopCh := make(chan struct{})
	activeGlideController.stopCh = stopCh
	activeGlideController.mu.Unlock()

	go func() {
		curOffset := startOffset
		curVel := velocity
		friction := 0.0035
		lastTime := time.Now()
		lastEmit := time.Now()
		hasDwmFlush := procDwmFlush.Find() == nil

		for {
			select {
			case <-stopCh:
				return
			default:
			}

			if hasDwmFlush {
				ret, _, _ := procDwmFlush.Call()
				if ret != 0 {
					// Fall back to sleep if DWM composition drops or returns an error
					time.Sleep(8 * time.Millisecond)
				}
			} else {
				time.Sleep(8 * time.Millisecond)
			}

			select {
			case <-stopCh:
				return
			default:
			}

			// Validate generation before updating window state or emitting events
			activeGlideController.mu.Lock()
			if activeGlideController.generation != myGen {
				activeGlideController.mu.Unlock()
				return
			}
			activeGlideController.mu.Unlock()

			now := time.Now()
			dt := float64(now.Sub(lastTime).Nanoseconds()) / 1e6
			lastTime = now
			if dt < 1.0 {
				time.Sleep(2 * time.Millisecond)
				continue
			}

			// Sub-step physics if dt is large due to a system lag spike
			if dt > 32 {
				steps := int(math.Ceil(dt / 16.0))
				if steps > 4 {
					steps = 4
				}
				stepDt := dt / float64(steps)
				for s := 0; s < steps; s++ {
					curVel *= math.Exp(-friction * stepDt)
					curOffset += curVel * stepDt
				}
			} else {
				curVel *= math.Exp(-friction * dt)
				curOffset += curVel * dt
			}

			// Clamp to monitor boundaries
			if curOffset >= maxOffset {
				curOffset = maxOffset
				a.offsetX = int(curOffset)
				a.updateWindowRegion()
				activeGlideController.mu.Lock()
				if activeGlideController.generation == myGen && a.ctx != nil {
					runtime.EventsEmit(a.ctx, "inertia-stopped", curOffset)
				}
				activeGlideController.mu.Unlock()
				return
			}
			if curOffset <= minOffset {
				curOffset = minOffset
				a.offsetX = int(curOffset)
				a.updateWindowRegion()
				activeGlideController.mu.Lock()
				if activeGlideController.generation == myGen && a.ctx != nil {
					runtime.EventsEmit(a.ctx, "inertia-stopped", curOffset)
				}
				activeGlideController.mu.Unlock()
				return
			}

			if math.Abs(curVel) < 0.04 {
				a.offsetX = int(math.Round(curOffset))
				a.updateWindowRegion()
				activeGlideController.mu.Lock()
				if activeGlideController.generation == myGen && a.ctx != nil {
					runtime.EventsEmit(a.ctx, "inertia-stopped", float64(a.offsetX))
				}
				activeGlideController.mu.Unlock()
				return
			}

			a.offsetX = int(curOffset)
			a.updateWindowRegion()

			if now.Sub(lastEmit) >= 15*time.Millisecond {
				lastEmit = now
				activeGlideController.mu.Lock()
				if activeGlideController.generation == myGen && a.ctx != nil {
					runtime.EventsEmit(a.ctx, "inertia-tick", curOffset)
				}
				activeGlideController.mu.Unlock()
			}
		}
	}()
}

// StopInertiaGlide immediately halts any in-progress native inertia glide and returns current offset.
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
	hwnd := getMyWindowHwnd()
	if hwnd == 0 {
		return
	}

	setExStyle := func(h uintptr) {
		gwl := int32(GWL_EXSTYLE)
		style, _, _ := procGetWindowLong.Call(h, uintptr(gwl))
		if ignore {
			style = (style | WS_EX_TRANSPARENT) &^ WS_EX_LAYERED
		} else {
			style = style &^ WS_EX_TRANSPARENT &^ WS_EX_LAYERED
		}
		_, _, _ = procSetWindowLong.Call(h, uintptr(gwl), style)
		_, _, _ = procSetWindowPos.Call(
			h, 0, 0, 0, 0, 0,
			uintptr(SWP_NOSIZE|SWP_NOMOVE|SWP_NOZORDER|SWP_NOACTIVATE|SWP_FRAMECHANGED),
		)
	}

	setExStyle(hwnd)

	// Propagate to all child windows (e.g. WebView2 Chrome_WidgetWin_0 and intermediate windows)
	cb := syscall.NewCallback(func(childHwnd uintptr, _ uintptr) uintptr {
		setExStyle(childHwnd)
		return 1
	})
	_, _, _ = procEnumChildWindows.Call(hwnd, cb, 0)

	if ignore {
		// Lift clipping region completely so island renders seamlessly without edge clipping during flight
		_, _, _ = procSetWindowRgn.Call(hwnd, 0, 0)
	} else {
		// Invalidate region cache and restore tight hit-testing region around the pill
		regionMu.Lock()
		lastRegion = regionCache{}
		regionMu.Unlock()
		a.updateWindowRegion()
	}
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
	services.SetAppHidden(hidden)
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

type regionCache struct {
	w, h, r, offsetX int
	dw, dh, extraH   int
	padding          int
	scale            float64
	winW, winH       int
	windowLeft       int
	windowTop        int
	currentMonID     int
}

var (
	regionMu   sync.Mutex
	lastRegion regionCache
)

// updateWindowRegion positions, sizes, and hit-tests the Win32 window tightly around the dynamic island.
func (a *App) updateWindowRegion() {
	hwnd := getMyWindowHwnd()
	if hwnd == 0 {
		return
	}

	padding := OutsidePadding
	// Use compact padding when in minimal collapsed state
	if a.pillW > 0 && a.pillW <= 110 && a.pillH <= 30 {
		padding = 16
	}

	scale := a.scale
	if scale <= 0 {
		scale = 1.0
	}

	scaleInt := func(v int) int {
		return int(float64(v) * scale)
	}

	pad := scaleInt(padding)
	visualTop := scaleInt(5)
	gap := scaleInt(15)

	pw := scaleInt(a.pillW)
	if pw <= 0 {
		pw = scaleInt(340)
	}
	ph := scaleInt(a.pillH)
	if ph <= 0 {
		ph = scaleInt(34)
	}

	totalW := pw
	pdw := 0
	pdh := 0
	pextra := 0
	if a.devW > 0 && a.devH > 0 {
		pdw = scaleInt(a.devW)
		pdh = scaleInt(a.devH)
		pextra = scaleInt(a.devExtra)
		if pdw > totalW {
			totalW = pdw
		}
	}

	baseW := scaleInt(650)
	baseH := scaleInt(320)
	winW := baseW
	winH := baseH

	neededW := totalW + (pad * 2)
	neededH := visualTop + ph + (pad * 2)
	if pdh > 0 {
		neededH = visualTop + ph + gap + pdh + pextra + (pad * 2)
	}

	if neededW > winW {
		winW = neededW
	}
	if neededH > winH {
		winH = neededH
	}

	vLeft, _, _, _, pCenter, monitors := getDisplayMetrics()
	// Screen center of the primary display plus the scaled user offset
	screenCenterX := vLeft + pCenter + scaleInt(a.offsetX)
	currentMon := resolveCurrentMonitor(screenCenterX, monitors)

	screenWidth = currentMon.Width
	screenHeight = currentMon.Height

	windowLeft := screenCenterX - (winW / 2)
	windowTop := currentMon.Top

	regionMu.Lock()
	defer regionMu.Unlock()

	if lastRegion.currentMonID != currentMon.ID {
		lastRegion.currentMonID = currentMon.ID
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "monitor-changed", currentMon)
		}
	}

	isFirstPlacement := lastRegion.winW == 0
	needMove := windowLeft != lastRegion.windowLeft || windowTop != lastRegion.windowTop
	needResize := winW != lastRegion.winW || winH != lastRegion.winH

	// Early return if window size, position, and pill parameters are identical
	if !isFirstPlacement && !needMove && !needResize &&
		lastRegion.w == a.pillW && lastRegion.h == a.pillH && lastRegion.r == a.pillR &&
		lastRegion.dw == a.devW && lastRegion.dh == a.devH && lastRegion.extraH == a.devExtra &&
		lastRegion.padding == padding && lastRegion.scale == scale {
		return
	}

	flags := uintptr(SWP_NOACTIVATE | SWP_NOOWNERZORDER | SWP_NOSENDCHANGING | SWP_NOCOPYBITS)
	if isFirstPlacement {
		flags |= SWP_SHOWWINDOW | SWP_FRAMECHANGED
	} else {
		flags |= SWP_NOZORDER
		if !needResize {
			flags |= SWP_NOSIZE
		}
		if !needMove {
			flags |= SWP_NOMOVE
		}
	}

	_, _, _ = procSetWindowPos.Call(
		hwnd,
		uintptr(^uintptr(0)),
		uintptr(int32(windowLeft)),
		uintptr(int32(windowTop)),
		uintptr(winW),
		uintptr(winH),
		flags,
	)

	// If dimensions and shape did not change, the local HRGN is already valid
	if !isFirstPlacement && !needResize &&
		lastRegion.w == a.pillW && lastRegion.h == a.pillH && lastRegion.r == a.pillR &&
		lastRegion.dw == a.devW && lastRegion.dh == a.devH && lastRegion.extraH == a.devExtra &&
		lastRegion.padding == padding && lastRegion.scale == scale {
		lastRegion.windowLeft = windowLeft
		lastRegion.windowTop = windowTop
		lastRegion.offsetX = a.offsetX
		return
	}

	lastRegion = regionCache{
		w:            a.pillW,
		h:            a.pillH,
		r:            a.pillR,
		offsetX:      a.offsetX,
		dw:           a.devW,
		dh:           a.devH,
		extraH:       a.devExtra,
		padding:      padding,
		scale:        scale,
		winW:         winW,
		winH:         winH,
		windowLeft:   windowLeft,
		windowTop:    windowTop,
		currentMonID: currentMon.ID,
	}

	winCenterX := winW / 2
	hrgnMain := createLocalMainPillRegion(winCenterX, visualTop, pw, ph, a.pillR, pad, scale)
	var hrgnDev uintptr
	if pdw > 0 && pdh > 0 {
		hrgnDev = createLocalDevPillRegion(winCenterX, visualTop+ph+gap, pdw, pdh, pextra, pad, scale)
	}

	hrgnFinal := combineWindowRegions(hrgnMain, hrgnDev)
	if hrgnFinal != 0 {
		_, _, _ = procSetWindowRgn.Call(hwnd, hrgnFinal, 0)
	}
}

// createLocalMainPillRegion generates the Win32 HRGN for the primary dynamic island in window-local coordinates.
func createLocalMainPillRegion(winCenterX, visualTop, pw, ph, r, pad int, scale float64) uintptr {
	scaleInt := func(v int) int {
		return int(float64(v) * scale)
	}

	left := winCenterX - (pw / 2) - pad
	if left < 0 {
		left = 0
	}
	right := winCenterX + (pw / 2) + pad
	top := visualTop - pad
	if top < 0 {
		top = 0
	}
	bottom := visualTop + ph + pad

	ellipseSize := scaleInt(r) * 2
	if hLimit := bottom - top; ellipseSize > hLimit {
		ellipseSize = hLimit
	}
	if wLimit := right - left; ellipseSize > wLimit {
		ellipseSize = wLimit
	}

	hrgn, _, _ := procCreateRoundRectRgn.Call(
		uintptr(left), uintptr(top),
		uintptr(right), uintptr(bottom),
		uintptr(ellipseSize), uintptr(ellipseSize),
	)
	return hrgn
}

// createLocalDevPillRegion generates the Win32 HRGN for the secondary dev panel in window-local coordinates.
func createLocalDevPillRegion(winCenterX, devTop, pdw, pdh, pextra, pad int, scale float64) uintptr {
	scaleInt := func(v int) int {
		return int(float64(v) * scale)
	}

	left := winCenterX - (pdw / 2) - pad
	if left < 0 {
		left = 0
	}
	right := winCenterX + (pdw / 2) + pad
	top := devTop - pad
	if top < 0 {
		top = 0
	}
	bottom := devTop + pdh + pextra + pad

	ellipseSize := scaleInt(14) * 2
	if hLimit := bottom - top; ellipseSize > hLimit {
		ellipseSize = hLimit
	}
	if wLimit := right - left; ellipseSize > wLimit {
		ellipseSize = wLimit
	}

	hrgn, _, _ := procCreateRoundRectRgn.Call(
		uintptr(left), uintptr(top),
		uintptr(right), uintptr(bottom),
		uintptr(ellipseSize), uintptr(ellipseSize),
	)
	return hrgn
}

// combineWindowRegions merges a slice of Win32 HRGNs into a single composite region.
func combineWindowRegions(regions ...uintptr) uintptr {
	var finalRgn uintptr
	for _, rgn := range regions {
		if rgn == 0 {
			continue
		}
		if finalRgn == 0 {
			finalRgn = rgn
		} else {
			_, _, _ = procCombineRgn.Call(finalRgn, finalRgn, rgn, uintptr(rgnOr))
			_, _, _ = procDeleteObject.Call(rgn)
		}
	}
	return finalRgn
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

// NoteFileInfo describes a saved markdown note file.
type NoteFileInfo struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	FileName string `json:"fileName"`
	ModTime  int64  `json:"modTime"`
	Size     int64  `json:"size"`
}

func getNotesDir() (string, error) {
	notesDir := filepath.Join(GetAppDataDir(), "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return "", err
	}
	return notesDir, nil
}

// ListNotes scans the notes directory and returns all markdown note files metadata.
func (a *App) ListNotes() ([]NoteFileInfo, error) {
	dir, err := getNotesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var notes []NoteFileInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fileName := entry.Name()
		id := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		notes = append(notes, NoteFileInfo{
			ID:       id,
			Title:    id,
			FileName: fileName,
			ModTime:  info.ModTime().UnixMilli(),
			Size:     info.Size(),
		})
	}
	return notes, nil
}

// ReadNote reads and returns the text content of a note file.
func (a *App) ReadNote(fileName string) (string, error) {
	dir, err := getNotesDir()
	if err != nil {
		return "", err
	}
	cleanName := filepath.Base(fileName)
	if !strings.HasSuffix(strings.ToLower(cleanName), ".md") {
		cleanName += ".md"
	}
	filePath := filepath.Join(dir, cleanName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveNote writes or updates a markdown note file.
func (a *App) SaveNote(fileName string, content string) error {
	dir, err := getNotesDir()
	if err != nil {
		return err
	}
	cleanName := filepath.Base(fileName)
	if !strings.HasSuffix(strings.ToLower(cleanName), ".md") {
		cleanName += ".md"
	}
	filePath := filepath.Join(dir, cleanName)
	return os.WriteFile(filePath, []byte(content), 0644)
}

// RenameNote renames an existing note file using the new title as the filename.
// Returns the new filename on success.
func (a *App) RenameNote(oldFileName string, newTitle string) (string, error) {
	dir, err := getNotesDir()
	if err != nil {
		return "", err
	}
	// Sanitize the new title into a safe filename
	sanitized := strings.NewReplacer(
		"/", "-", "\\", "-", ":", "-", "*", "-",
		"?", "-", "\"", "-", "<", "-", ">", "-", "|", "-",
	).Replace(strings.TrimSpace(newTitle))
	if sanitized == "" {
		sanitized = "untitled"
	}
	newFileName := sanitized + ".md"

	oldClean := filepath.Base(oldFileName)
	if !strings.HasSuffix(strings.ToLower(oldClean), ".md") {
		oldClean += ".md"
	}

	oldPath := filepath.Join(dir, oldClean)
	newPath := filepath.Join(dir, newFileName)

	// Avoid clobbering a different existing file
	if !strings.EqualFold(oldPath, newPath) {
		if _, statErr := os.Stat(newPath); statErr == nil {
			// Target exists — append a counter to make it unique
			base := strings.TrimSuffix(newFileName, ".md")
			for n := 2; ; n++ {
				candidate := fmt.Sprintf("%s (%d).md", base, n)
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

// DeleteNote deletes a markdown note file from disk.
func (a *App) DeleteNote(fileName string) error {
	dir, err := getNotesDir()
	if err != nil {
		return err
	}
	cleanName := filepath.Base(fileName)
	if !strings.HasSuffix(strings.ToLower(cleanName), ".md") {
		cleanName += ".md"
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
