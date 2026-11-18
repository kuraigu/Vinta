// Package backend provides core application services and state management for Vinta.
package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"Vinta/backend/services"
)

// screenWidth and screenHeight cache the display dimensions for window bounds centering.
var screenWidth, screenHeight int

// App represents the primary application orchestrator and Wails v3 service bridge.
type App struct {
	ctx                      context.Context
	wailsApp                 *application.App
	window                   *application.WebviewWindow
	devWindow                *application.WebviewWindow
	icon                     []byte
	pillW                    int
	pillH                    int
	pillR                    int
	restingW                 int
	restingH                 int
	restingR                 int
	offsetX                  int
	devW                     int
	devH                     int
	devExtra                 int
	satellitesMu             sync.RWMutex
	satellites               map[string]SatelliteBounds
	tooltipsMu               sync.RWMutex
	tooltips                 map[string]TooltipBounds
	scale                    float64
	isHidden                 bool
	autoHideBorderless       bool
	relocateOnFullscreen     bool
	returnOnFullscreenExit   bool
	autoHideMu               sync.RWMutex
	updaterMu                sync.Mutex
	downloadCancel           context.CancelFunc
	downloadPaused           bool
	activeDownloadURL        string
	activeChecksumURL        string
	pendingInstallPath       string
	customUpdateRepo         string
	fullscreenRelocatedMu    sync.Mutex
	isFullscreenRelocated    bool
	originalFullscreenOffset int
	fullscreenTriggerHwnd    uintptr
	fullscreenTriggerMonID   int
	pillHitMu                sync.RWMutex
	pillHitRects             []pillHitBounds
	hitRectsVersion          uint64
	uiScale                  float64
	overlayHeight            int
	overlayNeeded            int
	overlayFitTimer          *time.Timer
	isForceInteractive       bool
	// Tracks whether WS_EX_TRANSPARENT is currently cleared on the overlay HWND so mouse
	// events go to WebView2 instead of clicking through to desktop windows behind Vinta.
	// Protected by pillHitMu.
	isCurrentlyInteractive bool
	isOutboundDragging     bool
	initialShowOnce        sync.Once
	themeService           *ThemeService
}

// NewApp initializes and returns a new App instance.
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
		icon:                   icon,
		pillW:                  340,
		pillH:                  34,
		pillR:                  24,
		scale:                  1.0,
		uiScale:                1.0,
		autoHideBorderless:     false,
		relocateOnFullscreen:   true,
		returnOnFullscreenExit: false,
		tooltips:               make(map[string]TooltipBounds),
		themeService:           NewThemeService(),
	}
}

// Setup initializes the Wails v3 application instance, window hooks, and native system tray.
func (a *App) Setup(wailsApp *application.App, win *application.WebviewWindow) {
	defer SetupPanicRecovery()
	a.wailsApp = wailsApp
	a.window = win
	if wailsApp != nil {
		a.ctx = wailsApp.Context()
		wailsApp.Event.On("cancel-fullscreen-relocate", func(_ *application.CustomEvent) {
			a.CancelFullscreenRelocate()
		})
		wailsApp.Event.On("set-autohide-borderless", func(e *application.CustomEvent) {
			if enabled, ok := e.Data.(bool); ok {
				a.SetAutoHideBorderless(enabled)
			}
		})
		wailsApp.Event.On("send-native-notification", func(e *application.CustomEvent) {
			if data, ok := e.Data.(map[string]interface{}); ok {
				title, _ := data["title"].(string)
				msg, _ := data["message"].(string)
				if title != "" || msg != "" {
					a.SendNativeNotification(title, msg)
				}
			}
		})
		wailsApp.Event.On("devtest:open-devtools", func(_ *application.CustomEvent) {
			if a.IsTestMode() && a.window != nil {
				a.window.OpenDevTools()
			}
		})
	}

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

	var initOnce sync.Once
	applyInit := func(hwnd uintptr) {
		if hwnd == 0 {
			return
		}
		initOnce.Do(func() {
			cachedHwndMu.Lock()
			cachedWindowHwnd = hwnd
			cachedHwndMu.Unlock()

			gwlStyle := int32(GWL_STYLE)
			baseStyle, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwlStyle))
			baseStyle = (baseStyle | WS_POPUP) &^ (WS_CAPTION | WS_THICKFRAME | WS_SYSMENU)
			_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwlStyle), baseStyle)

			gwl := int32(GWL_EXSTYLE)
			style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwl))
			style = (style | WS_EX_TOOLWINDOW | WS_EX_TOPMOST | WS_EX_NOACTIVATE | WS_EX_LAYERED | WS_EX_TRANSPARENT) &^ WS_EX_APPWINDOW
			_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)

			if procDwmSetWindowAttribute.Find() == nil {
				ncPolicy := uint32(DWMNCRP_DISABLED)
				transitionsDisabled := uint32(1)
				_, _, _ = procDwmSetWindowAttribute.Call(hwnd, DWMWA_NCRENDERING_POLICY, uintptr(unsafe.Pointer(&ncPolicy)), 4)
				_, _, _ = procDwmSetWindowAttribute.Call(hwnd, DWMWA_TRANSITIONS_FORCEDISABLED, uintptr(unsafe.Pointer(&transitionsDisabled)), 4)
			}

			initAlpha := uintptr(0)
			if a.IsTestMode() {
				initAlpha = 255
			}
			_, _, _ = procSetLayeredWindowAttrs.Call(hwnd, 0, initAlpha, 2)

			_, _, _ = procSetWindowRgn.Call(hwnd, 0, 0)

			a.setupForegroundEventHook(hwnd)

			vLeft, vTop, vWidth, vHeight, _, monitors := GetDisplayMetrics()
			a.pillHitMu.RLock()
			uiScale := a.uiScale
			a.pillHitMu.RUnlock()
			overlayHeight := ComputeOverlayHeightForScale(vTop, vHeight, monitors, uiScale, a.effectiveScale())
			a.pillHitMu.Lock()
			a.overlayHeight = overlayHeight
			a.pillHitMu.Unlock()
			if vWidth > 0 {
				_, _, _ = procSetWindowPos.Call(
					hwnd,
					uintptr(^uintptr(0)),
					uintptr(int32(vLeft)), uintptr(int32(vTop)),
					uintptr(vWidth), uintptr(overlayHeight),
					SWP_NOACTIVATE|SWP_NOOWNERZORDER|SWP_FRAMECHANGED,
				)
			}

			primaryCenter := (vWidth / 2)
			for _, mon := range monitors {
				if mon.IsPrimary {
					primaryCenter = (mon.Left - vLeft) + (mon.Width / 2)
					break
				}
			}
			initW := int32(360)
			initH := int32(50)
			initLeft := int32(primaryCenter) - (initW / 2)
			initTop := int32(0)
			a.pillHitMu.Lock()
			if len(a.pillHitRects) == 0 {
				a.pillHitRects = []pillHitBounds{{
					left:   initLeft,
					top:    initTop,
					right:  initLeft + initW,
					bottom: initTop + initH,
				}}
			}
			a.pillHitMu.Unlock()

			go services.StartServicesListener()

			a.startHitTestingLoop(hwnd)
		})
	}

	if win != nil {
		win.OnWindowEvent(events.Common.WindowShow, func(_ *application.WindowEvent) {
			if h := uintptr(win.NativeWindow()); h != 0 {
				applyInit(h)
			}
		})
		win.OnWindowEvent(events.Windows.WindowShow, func(_ *application.WindowEvent) {
			if h := uintptr(win.NativeWindow()); h != 0 {
				applyInit(h)
			}
		})
		win.OnWindowEvent(events.Common.WindowRuntimeReady, func(_ *application.WindowEvent) {
			if h := uintptr(win.NativeWindow()); h != 0 {
				applyInit(h)
			}
			services.SetFrontendReady(true)
		})
		win.OnWindowEvent(events.Common.WindowClosing, func(_ *application.WindowEvent) {
			services.SetFrontendReady(false)
			services.StopServicesListener()
		})
		win.OnWindowEvent(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {
			if event == nil || event.Context() == nil || a.IsOutboundDragging() {
				return
			}
			dropped := event.Context().DroppedFiles()
			if len(dropped) > 0 && wailsApp != nil {
				wailsApp.Event.Emit("vinta-files-dropped", dropped)
			}
		})

		go func() {
			if a.IsTestMode() {
				time.Sleep(300 * time.Millisecond)
			} else {
				time.Sleep(3000 * time.Millisecond)
			}
			a.initialShowOnce.Do(func() {
				if h := a.getHwnd(); h != 0 {
					_, _, _ = procSetLayeredWindowAttrs.Call(h, 0, 255, 2)
				}
				if a.window != nil {
					a.window.Show()
				}
			})
		}()
	}

	go func() {
		for i := 0; i < 400; i++ {
			if win != nil {
				if h := uintptr(win.NativeWindow()); h != 0 {
					applyInit(h)
					return
				}
			}
			if h := getMyWindowHwnd(); h != 0 {
				applyInit(h)
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()

	a.StartBackgroundUpdateChecker()
	go a.SyncAutostartOnBoot()

	go func() {
		time.Sleep(600 * time.Millisecond)
		toggleHk := "Alt+Shift+V"
		quickNotesHk := "Alt+Shift+N"
		modeSwitchHk := "Alt+Shift+D"

		if settingsJSON, err := a.GetSettings(); err == nil && settingsJSON != "" && settingsJSON != "{}" {
			var m map[string]any
			if json.Unmarshal([]byte(settingsJSON), &m) == nil {
				if hk, ok := m["vinta_global_hotkey"].(string); ok && strings.TrimSpace(hk) != "" {
					toggleHk = strings.TrimSpace(hk)
				}
				if hk, ok := m["vinta_hotkey_quick_notes"].(string); ok && strings.TrimSpace(hk) != "" {
					quickNotesHk = strings.TrimSpace(hk)
				}
				if hk, ok := m["vinta_hotkey_mode_switch"].(string); ok && strings.TrimSpace(hk) != "" {
					modeSwitchHk = strings.TrimSpace(hk)
				}
			}
		}
		_, _ = a.RegisterNamedGlobalShortcut("toggle", toggleHk)
		_, _ = a.RegisterNamedGlobalShortcut("quick_notes", quickNotesHk)
		_, _ = a.RegisterNamedGlobalShortcut("mode_switch", modeSwitchHk)
	}()

	if wailsApp != nil {
		systray := wailsApp.SystemTray.New()
		systray.SetIcon(a.icon)
		systray.SetDarkModeIcon(a.icon)
		systray.SetTooltip("Vinta Dynamic Island")
		systray.OnClick(func() {
			if app := application.Get(); app != nil {
				app.Event.Emit("show-pill")
			}
		})

		menu := wailsApp.NewMenu()
		menu.Add("Show Vinta").OnClick(func(_ *application.Context) {
			if app := application.Get(); app != nil {
				app.Event.Emit("show-pill")
			}
		})
		menu.Add("Settings").OnClick(func(_ *application.Context) {
			if app := application.Get(); app != nil {
				app.Event.Emit("open-settings")
			}
		})
		menu.Add("Reset Island Position").OnClick(func(_ *application.Context) {
			if app := application.Get(); app != nil {
				app.Event.Emit("reset-position")
			}
		})
		menu.AddSeparator()
		menu.Add("Toggle Theme").OnClick(func(_ *application.Context) {
			if app := application.Get(); app != nil {
				app.Event.Emit("toggle-theme")
			}
		})
		menu.Add("Run Onboarding").OnClick(func(_ *application.Context) {
			if app := application.Get(); app != nil {
				app.Event.Emit("open-onboarding")
			}
		})
		menu.AddSeparator()
		menu.Add("Quit Vinta").OnClick(func(_ *application.Context) {
			a.Quit()
		})

		systray.SetMenu(menu)

		wailsApp.OnShutdown(func() {
			services.SetFrontendReady(false)
			services.StopServicesListener()
		})
	}
}

// Startup satisfies legacy startup calls if invoked.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Shutdown stops native background listeners.
func (a *App) Shutdown(_ context.Context) {
	services.StopServicesListener()
}

// SetDevWindow registers the developer tools window for lifecycle management.
func (a *App) SetDevWindow(win *application.WebviewWindow) {
	a.devWindow = win
}

// Quit terminates the application and all child windows.
func (a *App) Quit() {
	services.StopServicesListener()

	go func() {
		time.Sleep(50 * time.Millisecond)

		if a.devWindow != nil {
			a.devWindow.Close()
		}

		if a.wailsApp != nil {
			a.wailsApp.Quit()
		} else if app := application.Get(); app != nil {
			app.Quit()
		}

		if a.window != nil {
			a.window.Close()
		}

		time.Sleep(100 * time.Millisecond)
		os.Exit(0)
	}()
}

// Greet returns a greeting string for the provided name.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// IsTestMode returns true if the application was compiled with the devtest tag.
func (a *App) IsTestMode() bool {
	return isTestModeSync()
}

// SetClipboardMonitoring satisfies existing frontend bindings as clipboard monitoring is retired.
func (a *App) SetClipboardMonitoring(copyEnabled bool, pasteEnabled bool) {
	_, _ = copyEnabled, pasteEnabled
}

// SendNativeNotification dispatches an OS toast notification via native Windows WinRT services.
func (a *App) SendNativeNotification(title string, message string) {
	services.SendNativeToast(title, message)
}

// updateWindowRegion is a no-op under the unclipped transparent canvas hit-testing architecture.
func (a *App) updateWindowRegion() {
}

// ListThemes returns all available built-in and user-installed themes.
func (a *App) ListThemes() ([]ThemeInfo, error) {
	if a.themeService == nil {
		a.themeService = NewThemeService()
	}
	return a.themeService.ListThemes()
}

// GetThemeCSS returns the CSS content for the specified theme ID.
func (a *App) GetThemeCSS(id string) (string, error) {
	if a.themeService == nil {
		a.themeService = NewThemeService()
	}
	return a.themeService.GetThemeCSS(id)
}

// ImportThemeDialog opens a native file dialog to import a .vintatheme package.
func (a *App) ImportThemeDialog() (*ThemeInfo, error) {
	if a.themeService == nil {
		a.themeService = NewThemeService()
	}
	return a.themeService.ImportThemeDialog()
}

// OpenThemesFolder launches Windows Explorer targeting %APPDATA%/FreeMatrix Studios/Vinta/themes.
func (a *App) OpenThemesFolder() error {
	if a.themeService == nil {
		a.themeService = NewThemeService()
	}
	return a.themeService.OpenThemesFolder()
}
