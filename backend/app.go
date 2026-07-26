// Package backend provides the core application logic for Vinta.
package backend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	user32            = syscall.NewLazyDLL("user32.dll")
	procGetWindowLong = user32.NewProc("GetWindowLongW")
	procSetWindowLong = user32.NewProc("SetWindowLongW")
	procFindWindow    = user32.NewProc("FindWindowW")

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
	WS_EX_LAYERED = 0x00080000
	//nolint:revive // Windows API constants
	WS_EX_APPWINDOW = 0x00040000
	//nolint:revive // Windows API constants
	WS_EX_TOOLWINDOW = 0x00000080
	//nolint:revive // Windows API constants
	WS_EX_NOACTIVATE = 0x08000000
	OutsidePadding   = 15
	rgnOr            = 2
)

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

	// Expand the window to fill the entire screen so CSS has full real estate.
	// SetWindowRgn is used to define the interactive region — only the pill
	// area (plus GlowPadding) receives mouse events; everything else is
	// click-through. The extra padding lets glows and borders render freely.
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
		// DO NOT REMOVE -1 HERE. IT PREVENTS STREAMER MODE FROM OTHER APPS TO BE TRIGGERED BY VINTA
		runtime.WindowSetSize(ctx, screen.Size.Width, screen.Size.Height-1)
		runtime.WindowSetPosition(ctx, 0, 0)
	}

	// Hide the window from the Windows taskbar by setting the ToolWindow extended style.
	// Also initialise the click region to empty until Svelte calls SetWidgetBounds.
	titlePtr, _ := syscall.UTF16PtrFromString("Vinta")
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd != 0 {
		gwl := int32(GWL_EXSTYLE)
		style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwl))
		style = style | WS_EX_TOOLWINDOW | WS_EX_NOACTIVATE
		style = style &^ WS_EX_APPWINDOW
		_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)

		// Start native services interceptor (Notifications + Media controls)
		go services.StartServicesListener(ctx)

		// Start fully click-through; Svelte will call SetWidgetBounds on mount.
		a.updateWindowRegion()
	}

	// Launch background loop to re-assert AlwaysOnTop.
	// This ensures Vinta stays above games running in Borderless Windowed mode
	// that might launch after Vinta and steal the Topmost z-band.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				runtime.WindowSetAlwaysOnTop(ctx, true)
				time.Sleep(2 * time.Second)
			}
		}
	}()

	// Launch the system tray
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

	// If any fullscreen game (League of Legends, EVE Online, CS2, Valorant) is active/visible, DO NOT grab focus to prevent tab-outs!
	if a.isAnyFullscreenGameRunning() {
		return
	}

	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)

	titlePtr, _ := syscall.UTF16PtrFromString("Vinta")
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd != 0 {
		procGetForegroundWindow := user32.NewProc("GetForegroundWindow")
		procGetWindowThreadProcessID := user32.NewProc("GetWindowThreadProcessId")
		procAttachThreadInput := user32.NewProc("AttachThreadInput")
		procSetForegroundWindow := user32.NewProc("SetForegroundWindow")
		procSetActiveWindow := user32.NewProc("SetActiveWindow")
		procSetFocus := user32.NewProc("SetFocus")

		foreHwnd, _, _ := procGetForegroundWindow.Call()
		if foreHwnd != 0 && foreHwnd != hwnd {
			foreThread, _, _ := procGetWindowThreadProcessID.Call(foreHwnd, 0)
			appThread, _, _ := procGetWindowThreadProcessID.Call(hwnd, 0)

			if foreThread != appThread {
				_, _, _ = procAttachThreadInput.Call(foreThread, appThread, 1)
				_, _, _ = procSetForegroundWindow.Call(hwnd)
				_, _, _ = procSetActiveWindow.Call(hwnd)
				_, _, _ = procSetFocus.Call(hwnd)
				_, _, _ = procAttachThreadInput.Call(foreThread, appThread, 0)
				return
			}
		}
		_, _, _ = procSetForegroundWindow.Call(hwnd)
		_, _, _ = procSetActiveWindow.Call(hwnd)
		_, _, _ = procSetFocus.Call(hwnd)
	}
}

func (a *App) isAnyFullscreenGameRunning() bool {
	procEnumWindows := user32.NewProc("EnumWindows")
	procIsWindowVisible := user32.NewProc("IsWindowVisible")
	procGetWindowRect := user32.NewProc("GetWindowRect")
	procGetClassName := user32.NewProc("GetClassNameW")

	titlePtr, _ := syscall.UTF16PtrFromString("Vinta")
	vintaHwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))

	var foundFullscreenGame bool

	type RECT struct {
		Left, Top, Right, Bottom int32
	}

	cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		if hwnd == vintaHwnd {
			return 1 // continue
		}

		vis, _, _ := procIsWindowVisible.Call(hwnd)
		if vis == 0 {
			return 1
		}

		var classBuf [256]uint16
		_, _, _ = procGetClassName.Call(hwnd, uintptr(unsafe.Pointer(&classBuf[0])), 256)
		className := syscall.UTF16ToString(classBuf[:])

		// Ignore Windows system desktop/shell windows
		if className == "Progman" || className == "WorkerW" || className == "Shell_TrayWnd" || className == "Windows.UI.Core.CoreWindow" {
			return 1
		}

		var rect RECT
		_, _, _ = procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))

		width := rect.Right - rect.Left
		height := rect.Bottom - rect.Top

		scaleInt := func(v int) int {
			if a.scale <= 0 {
				return v
			}
			return int(float64(v) * a.scale)
		}

		sw := int32(scaleInt(screenWidth))
		sh := int32(scaleInt(screenHeight))

		// Check if window covers screen bounds (exclusive or borderless fullscreen game)
		if rect.Left <= 0 && rect.Top <= 0 && width >= sw && height >= (sh-10) {
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
	titlePtr, _ := syscall.UTF16PtrFromString("Vinta")
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
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
