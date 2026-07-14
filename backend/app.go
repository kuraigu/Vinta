// Package backend provides the core application logic for Vinta.
package backend

import (
	"context"
	"fmt"
	"os"
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
	procCombineRgn         = gdi32.NewProc("CombineRgn")
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

	// OutsidePadding is the extra space outside the window
	// This is useful for any objects that will render outside the pill without getting clipped
	OutsidePadding = 15

	// rgnOr is used to combine two regions.
	rgnOr = 2
)

// screenWidth caches the screen dimensions so SetWidgetBounds can center regions correctly.
var screenWidth int

// App struct
type App struct {
	ctx   context.Context
	icon  []byte
	pillW int
	pillH int
	pillR int
	devW  int
	devH  int
	scale float64
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
		runtime.WindowSetSize(ctx, screen.Size.Width, screen.Size.Height)
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
	// Clean up here if necessary
}

// SetWidgetBounds is called by the frontend whenever the widget dimensions change.
// It updates the native Win32 click-through region so only the visible widget is clickable.
func (a *App) SetWidgetBounds(width int, height int, radius int, scale float64) {
	a.pillW = width
	a.pillH = height
	a.pillR = radius
	if scale > 0 {
		a.scale = scale
	}
	a.updateWindowRegion()
}

// Quit terminates the application
func (a *App) Quit() {
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
}

// SetDevPillBounds is called by the frontend to show a separate developer menu.
func (a *App) SetDevPillBounds(width int, height int, scale float64) {
	a.devW = width
	a.devH = height
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
		// Use extremely tight bounds when hidden so we don't block background clicks
		if a.pillW <= 80 && a.pillH <= 15 {
			padding = 1
		}
		setClickRegion(hwnd, a.pillW, a.pillH, a.pillR, a.devW, a.devH, padding, a.scale)
	}
}

// setClickRegion updates the Win32 window region to combine the main pill and the dev pill.
func setClickRegion(hwnd uintptr, w int, h int, r int, dw int, dh int, padding int, scale float64) {
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
		physicalCenterX := scaleInt(screenWidth) / 2
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

		devPadding := 20
		physicalCenterX := scaleInt(screenWidth) / 2
		pdw := scaleInt(dw)
		pdh := scaleInt(dh)

		devLeft := physicalCenterX - (pdw / 2) - devPadding
		devRight := physicalCenterX + (pdw / 2) + devPadding
		devTop := devTopOffset - devPadding
		devBottom := devTop + pdh + devPadding

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
	for _, arg := range os.Args {
		if arg == "-test" {
			return true
		}
	}
	return false
}
