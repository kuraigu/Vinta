// Package backend provides the core application logic for Vinta.
package backend

import (
	"context"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procGetWindowLong = user32.NewProc("GetWindowLongW")
	procSetWindowLong = user32.NewProc("SetWindowLongW")
	procFindWindow    = user32.NewProc("FindWindowW")

	gdi32                  = syscall.NewLazyDLL("gdi32.dll")
	procCreateRoundRectRgn = gdi32.NewProc("CreateRoundRectRgn")
	procSetWindowRgn       = user32.NewProc("SetWindowRgn")
	procDeleteObject       = gdi32.NewProc("DeleteObject")
)

const (
	//nolint:revive // Windows API constants
	GWL_EXSTYLE = -20
	//nolint:revive // Windows API constants
	WS_EX_APPWINDOW = 0x00040000
	//nolint:revive // Windows API constants
	WS_EX_TOOLWINDOW = 0x00000080
	// MaxWindowWidth defines maximum native OS window bounds to support future widget expansion
	MaxWindowWidth  = 600
	MaxWindowHeight = 400
)

// App struct
type App struct {
	ctx  context.Context
	icon []byte
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

	// Center the window horizontally at the top tightly wrapped
	screens, _ := runtime.ScreenGetAll(ctx)
	if len(screens) > 0 {
		screen := screens[0]
		for _, s := range screens {
			if s.IsCurrent {
				screen = s
				break
			}
		}

		windowX := (screen.Size.Width / 2) - (MaxWindowWidth / 2)
		runtime.WindowSetSize(ctx, MaxWindowWidth, MaxWindowHeight)
		runtime.WindowSetPosition(ctx, windowX, 15)
	}

	// Hide the window from the Windows taskbar by finding its handle and setting the ToolWindow extended style
	titlePtr, _ := syscall.UTF16PtrFromString("Vinta")
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd != 0 {
		gwl := int32(GWL_EXSTYLE)
		style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwl))
		style = style | WS_EX_TOOLWINDOW
		style = style &^ WS_EX_APPWINDOW
		_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)

		// Clip the interactive region to nothing until Svelte loads and pushes the true sizes.
		// The OS window itself stays at MaxWindowWidth x MaxWindowHeight forever — we never resize it.
		// This is the key trick: SetWindowRgn controls click-through, CSS controls visuals.
		setClickRegion(hwnd, 0, 0, 0)
	}

	// Launch the system tray
	go systray.Run(a.onReady, a.onExit)
}

func (a *App) onReady() {
	systray.SetIcon(a.icon)
	systray.SetTitle("Vinta")
	systray.SetTooltip("Vinta Dynamic Island")
	mQuit := systray.AddMenuItem("Quit Vinta", "Close the application")

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
		runtime.Quit(a.ctx)
	}()
}

func (a *App) onExit() {
	// Clean up here if necessary
}

// SetWidgetBounds is called by the frontend to update the interactive click region of the widget.
// The OS window is NEVER resized — it stays permanently at MaxWindowWidth x MaxWindowHeight.
// This eliminates all Webview2 flicker caused by native window resize operations.
// CSS is solely responsible for the visual pill shape and animation.
// SetWindowRgn is used only to control which pixels can receive mouse clicks,
// so that clicks on the transparent area outside the pill pass through to windows below.
func (a *App) SetWidgetBounds(width int, height int, borderRadius int) {
	titlePtr, _ := syscall.UTF16PtrFromString("Vinta")
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd != 0 {
		setClickRegion(hwnd, width, height, borderRadius)
	}
}

// setClickRegion updates the Win32 window region to match the pill shape.
// bRedraw is always 0 — we never ask the OS to force-repaint, since Webview2
// manages its own compositing. The visual rendering is driven entirely by CSS.
func setClickRegion(hwnd uintptr, width int, height int, borderRadius int) {
	// CSS caps border-radius at half the height
	if borderRadius > height/2 {
		borderRadius = height / 2
	}

	// Center the clickable region within the fixed-size OS window
	left := (MaxWindowWidth - width) / 2
	right := left + width
	top := 0
	bottom := height

	// CreateRoundRectRgn takes the ellipse diameter, not radius
	ellipseSize := borderRadius * 2

	var hrgn uintptr
	if width == 0 || height == 0 {
		// Empty region = fully click-through (invisible)
		hrgn = 0
	} else {
		hrgn, _, _ = procCreateRoundRectRgn.Call(
			uintptr(left), uintptr(top),
			uintptr(right), uintptr(bottom),
			uintptr(ellipseSize), uintptr(ellipseSize),
		)
	}

	// bRedraw = 0: do NOT force a repaint. Webview2 composites its own frames.
	oldRgn, _, _ := procSetWindowRgn.Call(hwnd, hrgn, 0)
	if oldRgn != 0 {
		_, _, _ = procDeleteObject.Call(oldRgn)
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
