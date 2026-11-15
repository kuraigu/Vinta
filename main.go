// Package main is the entrypoint for the Vinta application.
// It initializes single-instance locking, configures the Wails runtime
// with frameless transparent window flags, sets up system tray integration, and
// starts background hardware monitoring services.
package main

import (
	"Vinta/backend"
	"Vinta/backend/devtest"
	"Vinta/backend/services"
	"embed"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/windows/icon.ico
var icon []byte

//go:embed build/appicon.png
var appIconPNG []byte

//go:embed docs/CONTRIBUTORS.md
var contributorsMD string

func main() {
	// Ensure panic recovery is always first in the defer stack
	defer backend.SetupPanicRecovery()

	// Windows Explorer launches Run entries with C:\Windows\System32 as the working directory.
	if wd, wdErr := os.Getwd(); wdErr == nil {
		if sysRoot := os.Getenv("SystemRoot"); sysRoot != "" && strings.EqualFold(filepath.Clean(wd), filepath.Join(filepath.Clean(sysRoot), "System32")) {
			if exePath, exeErr := os.Executable(); exeErr == nil {
				_ = os.Chdir(filepath.Dir(exePath))
			}
		}
	}

	backend.SetEmbeddedContributors(contributorsMD)

	// Initialize AppData icon asset safely
	appData := os.Getenv("APPDATA")
	if appData != "" && len(appIconPNG) > 0 {
		vintaDir := filepath.Join(appData, "FreeMatrix Studios", "Vinta")
		if err := os.MkdirAll(vintaDir, 0755); err == nil {
			iconDest := filepath.Join(vintaDir, "appicon.png")
			if _, err := os.Stat(iconDest); os.IsNotExist(err) {
				_ = os.WriteFile(iconDest, appIconPNG, 0644)
			}
		}
	}

	app := backend.NewApp(icon)
	devTools := devtest.NewDevTest()
	cache := backend.NewCache()
	themeService := backend.NewThemeService()

	hwAccel := backend.GetHardwareAccelerationConfig()

	browserArgs := []string{
		"--disable-extensions",
		"--disable-component-extensions-with-background-pages",
		"--disable-component-update",
		"--disable-background-networking",
		"--disable-domain-reliability",
		"--disable-client-side-phishing-detection",
		"--disable-sync",
		"--disable-translate",
		"--disable-default-apps",
		"--disable-spell-checking",
		"--disable-speech-api",
		"--disable-notifications",
		"--disable-print-preview",
		"--disable-pdf-extension",
		"--disable-webgl",
		"--disable-3d-apis",
		"--metrics-recording-only",
		"--no-first-run",
		"--no-pings",
		"--renderer-process-limit=1",
		// Keep Chromium from throttling rAF, timers, and compositing when another app has OS focus.
		"--disable-renderer-backgrounding",
	}

	disabledFeatures := []string{
		"msWebOOUI",
		"msPdfOOUI",
		"msSmartScreenProtection",
		"Translate",
		"TranslateUI",
		"AutofillServerCommunication",
		"PasswordImport",
		"OptimizationHints",
		"OptimizationGuideModelDownloading",
		"InterestFeedContentSuggestions",
		"MediaRouter",
		"DialMediaRouteProvider",
		"GlobalMediaControls",
		"HardwareMediaKeyHandling",
		"MediaSessionService",
		"AudioServiceOutOfProcess",
		"WinUseBrowserSpellChecker",
		"BackForwardCache",
		"Prerender2",
		"SpeculativeServiceWorkerWarmUp",
		// Required for transparent layered overlays: otherwise Windows marks Vinta occluded behind
		// maximized windows and freezes CSS/canvas animations.
		"CalculateNativeWinOcclusion",
	}

	if hwAccel {
		browserArgs = append(browserArgs,
			"--ignore-gpu-blocklist",
			"--enable-gpu-rasterization",
			"--enable-zero-copy",
			"--enable-features=CanvasOopRasterization",
		)
	} else {
		browserArgs = append(browserArgs, "--disable-gpu")
	}

	wailsApp := application.New(application.Options{
		Name:        "Vinta",
		Description: "Dynamic Island for Windows PC",
		Icon:        icon,
		Services: []application.Service{
			application.NewService(app),
			application.NewService(devTools),
			application.NewService(cache),
			application.NewService(themeService),
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "vinta-dynamic-island",
			OnSecondInstanceLaunch: func(_ application.SecondInstanceData) {
				if a := application.Get(); a != nil {
					a.Event.Emit("show-pill")
				}
			},
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
			Middleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasPrefix(r.URL.Path, "/icon-cache/") {
						cache.ServeHTTP(w, r)
						return
					}
					next.ServeHTTP(w, r)
				})
			},
		},
		Windows: application.WindowsOptions{
			WebviewUserDataPath:           filepath.Join(backend.GetAppDataDir(), "webview2"),
			DisableQuitOnLastWindowClosed: false,
			DisabledFeatures:              disabledFeatures,
			AdditionalBrowserArgs:         browserArgs,
		},
	})

	// Calculate multi-monitor spanning geometry
	vLeft, vTop, vWidth, vHeight, _, monitors := backend.GetDisplayMetrics()
	overlayHeight := backend.ComputeOverlayHeight(vTop, vHeight, monitors)
	if vWidth <= 0 {
		vWidth = 1920
	}

	win := wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:                      "Vinta",
		Width:                      vWidth,
		Height:                     overlayHeight,
		InitialPosition:            application.WindowXY,
		X:                          vLeft,
		Y:                          vTop,
		Frameless:                  true,
		AlwaysOnTop:                true,
		DisableResize:              true,
		Hidden:                     false,
		BackgroundType:             application.BackgroundTypeTransparent,
		BackgroundColour:           application.NewRGBA(0, 0, 0, 0),
		IgnoreMouseEvents:          false,
		EnableFileDrop:             true,
		DefaultContextMenuDisabled: true,
		ZoomControlEnabled:         false,
		DevToolsEnabled:            app.IsTestMode(),
		Windows: application.WindowsWindow{
			DisableFramelessWindowDecorations: true,
			HiddenOnTaskbar:                   true,
			BackdropType:                      application.None,
			DisableMenu:                       true,
			GeneralAutofillEnabled:            false,
			PasswordAutosaveEnabled:           false,
			EnableSwipeGestures:               false,
		},
	})

	app.Setup(wailsApp, win)

	if app.IsTestMode() {
		win.Show()
		devURL := "/?window=devtest"
		if os.Getenv("VINTA_SHOWCASE") == "1" {
			devURL = "/?window=devtest&showcase=true"
		}
		devWin := wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
			Title:            "Vinta Dev Tools",
			Width:            500,
			Height:           650,
			InitialPosition:  application.WindowCenter,
			Frameless:        false,
			AlwaysOnTop:      true,
			DisableResize:    false,
			URL:              devURL,
			BackgroundType:   application.BackgroundTypeSolid,
			BackgroundColour: application.NewRGBA(24, 24, 27, 255),
		})
		app.SetDevWindow(devWin)
		devWin.Show()
	}

	// Clean signal shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		if a := application.Get(); a != nil {
			a.Quit()
		}
	}()

	// Single source of truth for cleanup on application exit
	defer services.StopServicesListener()

	err := wailsApp.Run()
	if err != nil {
		println("Error:", err.Error())
	}
}
