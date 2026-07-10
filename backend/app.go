// Package backend provides the core application logic for Vinta.
package backend

import (
	"context"
	"fmt"
	"syscall"
	"time"
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

	// GlowPadding is the extra space around the pill given to the OS region so
	// that CSS box-shadow and border-glow can render without being clipped.
	GlowPadding = 10
)

// screenWidth caches the screen width so SetWidgetBounds can center the region correctly.
var screenWidth int

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
		screenWidth = screen.Width
		runtime.WindowSetSize(ctx, screen.Width, screen.Height)
		runtime.WindowSetPosition(ctx, 0, 0)
	}

	// Hide the window from the Windows taskbar by setting the ToolWindow extended style.
	// Also initialise the click region to empty until Svelte calls SetWidgetBounds.
	titlePtr, _ := syscall.UTF16PtrFromString("Vinta")
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd != 0 {
		gwl := int32(GWL_EXSTYLE)
		style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwl))
		style = style | WS_EX_TOOLWINDOW
		style = style &^ WS_EX_APPWINDOW
		_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)

		// Start fully click-through; Svelte will call SetWidgetBounds on mount.
		setClickRegion(hwnd, 0, 0, 0)
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

// SetWidgetBounds is called by the frontend whenever the pill size changes.
// It updates the Win32 window region so only the pill (plus GlowPadding) receives
// mouse events; everything else passes through to the desktop.
// GlowPadding ensures CSS box-shadow and border-glow are not clipped by the OS mask.
func (a *App) SetWidgetBounds(width int, height int, _ int) {
	titlePtr, _ := syscall.UTF16PtrFromString("Vinta")
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd != 0 {
		setClickRegion(hwnd, width, height, GlowPadding)
	}
}

// setClickRegion updates the Win32 window region to a rounded rect around the pill.
// padding expands the region beyond the visual pill so CSS effects render freely.
// bRedraw is always 0 — Webview2 manages its own compositing.
func setClickRegion(hwnd uintptr, width int, height int, padding int) {
	var hrgn uintptr

	if width == 0 || height == 0 {
		// Empty region = fully click-through
		hrgn = 0
	} else {
		// Center the pill horizontally; place it near the top (CSS margin-top: 5px)
		visualTop := 5
		centerX := screenWidth / 2

		left := centerX - (width / 2) - padding
		right := centerX + (width / 2) + padding
		top := visualTop - padding
		bottom := visualTop + height + padding

		// CreateRoundRectRgn takes the ellipse diameter, not radius.
		// We approximate the rounded corners using the padding as extra radius.
		ellipseSize := padding * 2

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
