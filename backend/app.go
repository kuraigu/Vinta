// Package backend provides the core application logic for Vinta.
package backend

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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
	procFindWindow               = user32.NewProc("FindWindowW")
	procGetDesktopWindow         = user32.NewProc("GetDesktopWindow")
	procGetWindow                = user32.NewProc("GetWindow")
	procGetWindowThreadProcessID = user32.NewProc("GetWindowThreadProcessId")
	procGetClassName             = user32.NewProc("GetClassNameW")
	procSetWinEventHook          = user32.NewProc("SetWinEventHook")
	procEnumWindows              = user32.NewProc("EnumWindows")

	gdi32                  = syscall.NewLazyDLL("gdi32.dll")
	procCreateRectRgn      = gdi32.NewProc("CreateRectRgn")
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
		icon:  icon,
		pillW: 340,
		pillH: 34,
		pillR: 24,
		scale: 1.0,
	}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	defer SetupPanicRecovery()
	a.ctx = ctx

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
			style = (style | WS_EX_TOOLWINDOW | WS_EX_TOPMOST) &^ WS_EX_NOACTIVATE &^ WS_EX_APPWINDOW
			_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)

			setupForegroundEventHook(hwnd)

			procSetWindowPos := user32.NewProc("SetWindowPos")
			// HWND_TOPMOST = ^uintptr(0), SWP_NOACTIVATE (0x0010) | SWP_FRAMECHANGED (0x0020) | SWP_SHOWWINDOW (0x0040)
			overlayHeight := 650
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

// FocusWindow requests OS input focus so typing and keyboard events work immediately on click/focus
func (a *App) FocusWindow() {
	if a.ctx == nil {
		return
	}

	// If any fullscreen game is active/visible, DO NOT grab focus to prevent tab-outs
	if a.isAnyFullscreenGameRunning() {
		return
	}

	hwnd := getMyWindowHwnd()
	if hwnd != 0 {
		procSetForegroundWindow := user32.NewProc("SetForegroundWindow")
		_, _, _ = procSetForegroundWindow.Call(hwnd)
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

	sw := screenWidth
	if sw <= 0 {
		smwVal, _, _ := user32.NewProc("GetSystemMetrics").Call(0)
		if smwVal > 0 {
			sw = int(smwVal)
			screenWidth = sw
		} else {
			sw = 1920
		}
	}

	// 1. Create main pill region
	if w == -1 && h > 0 {
		visualTop := scaleInt(5)
		ph := scaleInt(h)
		top := visualTop - padding
		if top < 0 {
			top = 0
		}
		bottom := visualTop + ph + padding

		// In-flight full-width rectangular glide track: allows smooth 120 FPS glide across screen with zero edge clipping
		hrgnMain, _, _ = procCreateRectRgn.Call(
			0, uintptr(top),
			uintptr(sw), uintptr(bottom),
		)
	} else if w > 0 && h > 0 {
		visualTop := scaleInt(5)
		physicalCenterX := (sw / 2) + scaleInt(offsetX)
		pw := scaleInt(w)
		ph := scaleInt(h)

		left := physicalCenterX - (pw / 2) - padding
		if left < 0 {
			left = 0
		}
		right := physicalCenterX + (pw / 2) + padding
		top := visualTop - padding
		if top < 0 {
			top = 0
		}
		bottom := visualTop + ph + padding
		ellipseSize := (scaleInt(r) + padding) * 2
		if maxH := bottom - top; ellipseSize > maxH {
			ellipseSize = maxH
		}

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
		physicalCenterX := sw / 2
		pdw := scaleInt(dw)
		pdh := scaleInt(dh)

		devLeft := physicalCenterX - (pdw / 2) - devPadding
		if devLeft < 0 {
			devLeft = 0
		}
		devRight := physicalCenterX + (pdw / 2) + devPadding
		devTop := devTopOffset - devPadding - extraMargin
		if devTop < 0 {
			devTop = 0
		}
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

	_, _, _ = procSetWindowRgn.Call(hwnd, hrgnFinal, 1)
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
