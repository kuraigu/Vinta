// Package backend provides the core application logic for Vinta.
package backend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
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
	procSetWindowLongPtr         = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProc           = user32.NewProc("CallWindowProcW")
	procFindWindow               = user32.NewProc("FindWindowW")
	procEnumChildWindows         = user32.NewProc("EnumChildWindows")
	procGetDesktopWindow         = user32.NewProc("GetDesktopWindow")
	procGetWindow                = user32.NewProc("GetWindow")
	procGetWindowThreadProcessID = user32.NewProc("GetWindowThreadProcessId")
	procGetClassName             = user32.NewProc("GetClassNameW")
	procSetWinEventHook          = user32.NewProc("SetWinEventHook")
	procEnumWindows              = user32.NewProc("EnumWindows")

	gdi32                  = syscall.NewLazyDLL("gdi32.dll")
	procCreateRoundRectRgn = gdi32.NewProc("CreateRoundRectRgn")
	procCombineRgn         = gdi32.NewProc("CombineRgn")
	procSetWindowRgn       = user32.NewProc("SetWindowRgn")
	procDeleteObject       = gdi32.NewProc("DeleteObject")
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
	OutsidePadding   = 15
	rgnOr            = 2

	//nolint:revive // Windows API constants
	GW_CHILD = 5
	//nolint:revive // Windows API constants
	GW_HWNDNEXT = 2
	//nolint:revive // Windows API constants
	EVENT_SYSTEM_FOREGROUND = 0x0003
	//nolint:revive // Windows API constants
	WINEVENT_OUTOFCONTEXT = 0x0000
)

func getMyWindowHwnd() uintptr {
	titlePtr, _ := syscall.UTF16PtrFromString("Vinta")
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd != 0 {
		return hwnd
	}

	myPid := uint32(os.Getpid())
	var targetHwnd uintptr
	cb := syscall.NewCallback(func(h uintptr, _ uintptr) uintptr {
		var pid uint32
		_, _, _ = procGetWindowThreadProcessID.Call(h, uintptr(unsafe.Pointer(&pid)))
		if pid == myPid {
			var classBuf [256]uint16
			_, _, _ = procGetClassName.Call(h, uintptr(unsafe.Pointer(&classBuf[0])), 256)
			className := syscall.UTF16ToString(classBuf[:])
			if className != "IME" && className != "MSCTFIME UI" && className != "Chrome_WidgetWin_0" && className != "Intermediate D3D Window" {
				targetHwnd = h
				return 0
			}
		}
		return 1
	})
	_, _, _ = procEnumWindows.Call(cb, 0)
	if targetHwnd != 0 {
		return targetHwnd
	}

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
				return child
			}
		}
		child, _, _ = procGetWindow.Call(child, uintptr(GW_HWNDNEXT))
	}
	return 0
}

var (
	wndProcCallback uintptr
	subclassedMap   = make(map[uintptr]uintptr)
)

// setupForegroundEventHook registers a WinEvent hook on a dedicated OS-locked thread
// with a Win32 message pump to maintain topmost Z-order on window changes without
// stealing focus or causing fullscreen games to minimize.
func setupForegroundEventHook(vintaHwnd uintptr) {
	if vintaHwnd == 0 {
		return
	}

	procSetWindowPos := user32.NewProc("SetWindowPos")
	procGetMessage := user32.NewProc("GetMessageW")
	procDispatchMessage := user32.NewProc("DispatchMessageW")
	procTranslateMessage := user32.NewProc("TranslateMessage")
	procUnhookWinEvent := user32.NewProc("UnhookWinEvent")

	type MSG struct {
		Hwnd    uintptr
		Message uint32
		WParam  uintptr
		LParam  uintptr
		Time    uint32
		Pt      [2]int32
	}

	reassertTopmost := func() {
		// HWND_TOPMOST = ^uintptr(0), SWP_NOSIZE (0x0001) | SWP_NOMOVE (0x0002) | SWP_NOACTIVATE (0x0010) | SWP_NOOWNERZORDER (0x0200) | SWP_NOSENDCHANGING (0x0400)
		_, _, _ = procSetWindowPos.Call(
			vintaHwnd,
			uintptr(^uintptr(0)),
			0, 0, 0, 0,
			0x0001|0x0002|0x0010|0x0200|0x0400,
		)
	}

	go func() {
		goruntime.LockOSThread()

		cb := syscall.NewCallback(func(_ uintptr, event uint32, fgHwnd uintptr, _ int32, _ int32, _ uint32, _ uint32) uintptr {
			if event == EVENT_SYSTEM_FOREGROUND && fgHwnd != 0 && fgHwnd != vintaHwnd {
				// Only re-assert if foreground window is not a fullscreen exclusive game to avoid minimizing games
				if !isAnyFullscreenForegroundWindow(fgHwnd) {
					reassertTopmost()
				}
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

func isAnyFullscreenForegroundWindow(fgHwnd uintptr) bool {
	if fgHwnd == 0 {
		return false
	}
	procGetClassName := user32.NewProc("GetClassNameW")
	procGetWindowRect := user32.NewProc("GetWindowRect")
	procGetSystemMetrics := user32.NewProc("GetSystemMetrics")

	var classBuf [256]uint16
	_, _, _ = procGetClassName.Call(fgHwnd, uintptr(unsafe.Pointer(&classBuf[0])), 256)
	className := syscall.UTF16ToString(classBuf[:])
	if className == "Progman" || className == "WorkerW" || className == "Shell_TrayWnd" || className == "Windows.UI.Core.CoreWindow" {
		return false
	}

	smwVal, _, _ := procGetSystemMetrics.Call(0) // SM_CXSCREEN
	smhVal, _, _ := procGetSystemMetrics.Call(1) // SM_CYSCREEN
	smw := int32(smwVal)
	smh := int32(smhVal)
	if smw == 0 || smh == 0 {
		smw = int32(screenWidth)
		smh = int32(screenHeight)
	}

	type RECT struct {
		Left, Top, Right, Bottom int32
	}
	var rect RECT
	_, _, _ = procGetWindowRect.Call(fgHwnd, uintptr(unsafe.Pointer(&rect)))
	width := rect.Right - rect.Left
	height := rect.Bottom - rect.Top
	return rect.Left <= 0 && rect.Top <= 0 && width >= (smw-15) && height >= (smh-15)
}

func setWindowLongPtr(hwnd uintptr, index int32, value uintptr) uintptr {
	if procSetWindowLongPtr.Find() == nil {
		res, _, _ := procSetWindowLongPtr.Call(hwnd, uintptr(index), value)
		return res
	}
	res, _, _ := procSetWindowLong.Call(hwnd, uintptr(index), value)
	return res
}

func initNoActivateHook(hwnd uintptr) {
	if wndProcCallback == 0 {
		wndProcCallback = syscall.NewCallback(func(h uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
			switch msg {
			case WM_MOUSEACTIVATE:
				return MA_NOACTIVATE
			case WM_PARENTNOTIFY:
				if (wParam & 0xFFFF) == 1 { // WM_CREATE
					if lParam != 0 {
						subclassWindow(lParam)
					}
				}
			}

			oldProc := subclassedMap[h]
			if oldProc != 0 {
				res, _, _ := procCallWindowProc.Call(oldProc, h, uintptr(msg), wParam, lParam)
				return res
			}
			res, _, _ := user32.NewProc("DefWindowProcW").Call(h, uintptr(msg), wParam, lParam)
			return res
		})
	}

	subclassWindow(hwnd)

	enumCb := syscall.NewCallback(func(childHwnd uintptr, _ uintptr) uintptr {
		subclassWindow(childHwnd)
		return 1
	})
	_, _, _ = procEnumChildWindows.Call(hwnd, enumCb, 0)
}

func subclassWindow(h uintptr) {
	if h == 0 {
		return
	}
	if _, exists := subclassedMap[h]; !exists {
		old := setWindowLongPtr(h, GWLP_WNDPROC, wndProcCallback)
		if old != 0 && old != wndProcCallback {
			subclassedMap[h] = old
		}
	}
}

// screenWidth and screenHeight cache the screen dimensions so SetWidgetBounds can center regions correctly.
var screenWidth, screenHeight int

// App struct
type App struct {
	ctx      context.Context
	icon     []byte
	pillW    int
	pillH    int
	pillR    int
	offsetX  int
	devW     int
	devH     int
	devExtra int
	scale    float64
}

// NewApp creates a new App application struct
func NewApp(icon []byte) *App {
	return &App{
		icon: icon,
	}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	screens, _ := runtime.ScreenGetAll(ctx)
	if len(screens) > 0 {
		screen := screens[0]
		for _, s := range screens {
			if s.IsCurrent {
				screen = s
				break
			}
		}
		screenWidth = screen.Size.Width
		screenHeight = screen.Size.Height
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
			style = style | WS_EX_TOOLWINDOW | WS_EX_NOACTIVATE | WS_EX_TOPMOST
			style = style &^ WS_EX_APPWINDOW
			_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)

			initNoActivateHook(hwnd)

			setupForegroundEventHook(hwnd)

			time.AfterFunc(1500*time.Millisecond, func() {
				initNoActivateHook(hwnd)
			})
			time.AfterFunc(3000*time.Millisecond, func() {
				initNoActivateHook(hwnd)
			})

			procSetWindowPos := user32.NewProc("SetWindowPos")
			// HWND_TOPMOST = ^uintptr(0), SWP_NOACTIVATE (0x0010) | SWP_FRAMECHANGED (0x0020) | SWP_SHOWWINDOW (0x0040)
			overlayHeight := 500
			if screenHeight > 0 && overlayHeight > screenHeight {
				overlayHeight = screenHeight
			}
			if screenWidth > 0 {
				_, _, _ = procSetWindowPos.Call(
					hwnd,
					uintptr(^uintptr(0)),
					0, 0,
					uintptr(screenWidth), uintptr(overlayHeight),
					0x0010|0x0020|0x0040,
				)
			}

			go services.StartServicesListener(ctx)

			a.updateWindowRegion()
		}
	}()

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
	services.StopServicesListener()
}

// Shutdown is called when Wails application shuts down.
func (a *App) Shutdown(_ context.Context) {
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

// Quit terminates the application
func (a *App) Quit() {
	services.StopServicesListener()
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
}

// FocusWindow requests OS input focus so mouse wheel scroll events work immediately on hover
func (a *App) FocusWindow() {
	if a.ctx == nil {
		return
	}

	// If any fullscreen game or application is active/visible, DO NOT grab focus to prevent tab-outs!
	if a.isAnyFullscreenGameRunning() {
		return
	}

	// Do NOT grab focus if the current active foreground window is an external application
	procGetForegroundWindow := user32.NewProc("GetForegroundWindow")
	procGetClassName := user32.NewProc("GetClassNameW")
	foreHwnd, _, _ := procGetForegroundWindow.Call()
	if foreHwnd != 0 {
		var classBuf [256]uint16
		_, _, _ = procGetClassName.Call(foreHwnd, uintptr(unsafe.Pointer(&classBuf[0])), 256)
		className := syscall.UTF16ToString(classBuf[:])
		titlePtr, _ := syscall.UTF16PtrFromString("Vinta")
		hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))

		if foreHwnd != hwnd && className != "Progman" && className != "WorkerW" && className != "Shell_TrayWnd" {
			return
		}
	}

	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
}

func (a *App) isAnyFullscreenGameRunning() bool {
	procEnumWindows := user32.NewProc("EnumWindows")
	procIsWindowVisible := user32.NewProc("IsWindowVisible")
	procGetWindowRect := user32.NewProc("GetWindowRect")
	procGetClassName := user32.NewProc("GetClassNameW")
	procGetSystemMetrics := user32.NewProc("GetSystemMetrics")
	procGetForegroundWindow := user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessID := user32.NewProc("GetWindowThreadProcessId")

	myPid := uint32(os.Getpid())

	smwVal, _, _ := procGetSystemMetrics.Call(0) // SM_CXSCREEN
	smhVal, _, _ := procGetSystemMetrics.Call(1) // SM_CYSCREEN
	smw := int32(smwVal)
	smh := int32(smhVal)
	if smw == 0 || smh == 0 {
		smw = int32(screenWidth)
		smh = int32(screenHeight)
	}

	type RECT struct {
		Left, Top, Right, Bottom int32
	}

	// 1. Check if active foreground window is an external fullscreen application
	foreHwnd, _, _ := procGetForegroundWindow.Call()
	if foreHwnd != 0 {
		var forePid uint32
		_, _, _ = procGetWindowThreadProcessID.Call(foreHwnd, uintptr(unsafe.Pointer(&forePid)))
		if forePid != myPid {
			var classBuf [256]uint16
			_, _, _ = procGetClassName.Call(foreHwnd, uintptr(unsafe.Pointer(&classBuf[0])), 256)
			className := syscall.UTF16ToString(classBuf[:])
			if className != "Progman" && className != "WorkerW" && className != "Shell_TrayWnd" && className != "Windows.UI.Core.CoreWindow" {
				var rect RECT
				_, _, _ = procGetWindowRect.Call(foreHwnd, uintptr(unsafe.Pointer(&rect)))
				width := rect.Right - rect.Left
				height := rect.Bottom - rect.Top
				if rect.Left <= 0 && rect.Top <= 0 && width >= (smw-15) && height >= (smh-15) {
					return true
				}
			}
		}
	}

	// 2. Check if any top-level visible window (excluding Vinta's own windows) covers full screen bounds
	var foundFullscreenGame bool
	cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		var pid uint32
		_, _, _ = procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
		if pid == myPid {
			return 1 // ignore Vinta's own process windows
		}

		vis, _, _ := procIsWindowVisible.Call(hwnd)
		if vis == 0 {
			return 1
		}

		var classBuf [256]uint16
		_, _, _ = procGetClassName.Call(hwnd, uintptr(unsafe.Pointer(&classBuf[0])), 256)
		className := syscall.UTF16ToString(classBuf[:])

		if className == "Progman" || className == "WorkerW" || className == "Shell_TrayWnd" || className == "Windows.UI.Core.CoreWindow" {
			return 1
		}

		var rect RECT
		_, _, _ = procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))

		width := rect.Right - rect.Left
		height := rect.Bottom - rect.Top

		if rect.Left <= 0 && rect.Top <= 0 && width >= (smw-15) && height >= (smh-15) {
			foundFullscreenGame = true
			return 0 // stop enumeration
		}

		return 1
	})

	_, _, _ = procEnumWindows.Call(cb, 0)
	return foundFullscreenGame
}

// SetDevPillBounds is called by the frontend to show a separate developer menu.
func (a *App) SetDevPillBounds(width int, height int, extraH int, scale float64) {
	a.devW = width
	a.devH = height
	a.devExtra = extraH
	if scale > 0 {
		a.scale = scale
	}
	a.updateWindowRegion()
}

func (a *App) updateWindowRegion() {
	hwnd := getMyWindowHwnd()
	if hwnd != 0 {
		padding := OutsidePadding
		// Use tight padding (8px) when hidden to allow ambient glow and smooth drag without clipping
		if a.pillW <= 110 && a.pillH <= 30 {
			padding = 8
		}
		setClickRegion(hwnd, a.pillW, a.pillH, a.pillR, a.offsetX, a.devW, a.devH, a.devExtra, padding, a.scale)

		scale := a.scale
		if scale <= 0 {
			scale = 1.0
		}
		scaleInt := func(v int) int {
			return int(float64(v) * scale)
		}

		// Dynamically calculate exact required OS window height based on active widget & dropdown dimensions
		neededH := scaleInt(5) + scaleInt(a.pillH) + padding*2
		if a.devH > 0 {
			neededH += scaleInt(15) + scaleInt(a.devH) + scaleInt(a.devExtra)
		}

		smwVal, _, _ := user32.NewProc("GetSystemMetrics").Call(0) // SM_CXSCREEN
		winW := int(smwVal)
		if winW <= 0 {
			winW = screenWidth
		}
		if winW > 0 {
			winW = scaleInt(winW)
		}

		procSetWindowPos := user32.NewProc("SetWindowPos")
		// SWP_NOMOVE (0x0002) | SWP_NOACTIVATE (0x0010) | SWP_NOOWNERZORDER (0x0200) | SWP_NOSENDCHANGING (0x0400)
		_, _, _ = procSetWindowPos.Call(
			hwnd,
			uintptr(^uintptr(0)),
			0, 0,
			uintptr(winW), uintptr(neededH),
			0x0002|0x0010|0x0200|0x0400,
		)
	}
}

// setClickRegion updates the Win32 window region to combine the main pill and the dev pill.
func setClickRegion(hwnd uintptr, w int, h int, r int, offsetX int, dw int, dh int, extraH int, padding int, scale float64) {
	var hrgnMain uintptr
	var hrgnDev uintptr
	var hrgnFinal uintptr

	if scale <= 0 {
		scale = 1.0
	}

	scaleInt := func(v int) int {
		return int(float64(v) * scale)
	}

	// 1. Create main pill region
	if w > 0 && h > 0 {
		visualTop := scaleInt(5)
		physicalCenterX := (scaleInt(screenWidth) / 2) + scaleInt(offsetX)
		pw := scaleInt(w)
		ph := scaleInt(h)

		left := physicalCenterX - (pw / 2) - padding
		right := physicalCenterX + (pw / 2) + padding
		top := visualTop - padding
		bottom := visualTop + ph + padding
		ellipseSize := (scaleInt(r) + padding) * 2

		hrgnMain, _, _ = procCreateRoundRectRgn.Call(
			uintptr(left), uintptr(top),
			uintptr(right), uintptr(bottom),
			uintptr(ellipseSize), uintptr(ellipseSize),
		)
	}

	// 2. Create dev pill region (stacked under main pill)
	if dw > 0 && dh > 0 {
		gap := scaleInt(15) // Must match the CSS flex gap
		visualTop := scaleInt(5)

		// If main pill exists, stack it below it. Otherwise, default to some position.
		devTopOffset := visualTop
		if w > 0 && h > 0 {
			ph := scaleInt(h)
			devTopOffset = visualTop + ph + gap
		}

		devPadding := 6
		extraMargin := scaleInt(extraH) // ONLY expand region when a dropdown is actively open!
		physicalCenterX := scaleInt(screenWidth) / 2
		pdw := scaleInt(dw)
		pdh := scaleInt(dh)

		devLeft := physicalCenterX - (pdw / 2) - devPadding
		devRight := physicalCenterX + (pdw / 2) + devPadding
		devTop := devTopOffset - devPadding - extraMargin
		devBottom := devTopOffset + pdh + devPadding + extraMargin

		hrgnDev, _, _ = procCreateRoundRectRgn.Call(
			uintptr(devLeft), uintptr(devTop),
			uintptr(devRight), uintptr(devBottom),
			uintptr(24), uintptr(24),
		)
	}

	// 3. Combine regions
	if hrgnMain != 0 && hrgnDev != 0 {
		// Both exist, combine into hrgnMain
		_, _, _ = procCombineRgn.Call(hrgnMain, hrgnMain, hrgnDev, uintptr(rgnOr))
		hrgnFinal = hrgnMain
		_, _, _ = procDeleteObject.Call(hrgnDev) // cleanup
	} else if hrgnMain != 0 {
		hrgnFinal = hrgnMain
	} else if hrgnDev != 0 {
		hrgnFinal = hrgnDev
	} else {
		hrgnFinal = 0 // fully transparent
	}

	oldRgn, _, _ := procSetWindowRgn.Call(hwnd, hrgnFinal, 0)
	if oldRgn != 0 {
		_, _, _ = procDeleteObject.Call(oldRgn)
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// IsTestMode returns true if the application was launched with the -test flag.
func (a *App) IsTestMode() bool {
	return isTestModeSync()
}

func isTestModeSync() bool {
	for _, arg := range os.Args {
		if arg == "-test" {
			return true
		}
	}
	return false
}

// GetHardwareAccelerationConfig reads stored GPU acceleration setting from user config dir.
func GetHardwareAccelerationConfig() bool {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return true // Default enabled
	}
	filePath := filepath.Join(configDir, "Vinta", "hw_accel.txt")
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
	dirPath := filepath.Join(configDir, "Vinta")
	_ = os.MkdirAll(dirPath, 0755)
	filePath := filepath.Join(dirPath, "hw_accel.txt")
	val := "true"
	if !enabled {
		val = "false"
	}
	return os.WriteFile(filePath, []byte(val), 0644)
}
