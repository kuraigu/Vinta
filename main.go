// Package main provides the entry point and window configuration for Vinta.
package main

import (
	"Vinta/backend"
	"Vinta/backend/devtest"
	"embed"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/windows/icon.ico
var icon []byte

//go:embed docs/CONTRIBUTORS.md
var contributorsMD string

func main() {
	defer backend.SetupPanicRecovery()

	backend.SetEmbeddedContributors(contributorsMD)

	app := backend.NewApp(icon)
	devTools := devtest.NewDevTest()
	cache := backend.NewCache()

	wailsApp := application.New(application.Options{
		Name:        "Vinta",
		Description: "Dynamic Island for Windows PC",
		Icon:        icon,
		Services: []application.Service{
			application.NewService(app),
			application.NewService(devTools),
			application.NewService(cache),
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
					if strings.HasPrefix(r.URL.Path, "/icon-cache/") || strings.HasPrefix(r.URL.Path, "/local/") {
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
		},
	})

	vLeft, vTop, vWidth, vHeight, _, monitors := backend.GetDisplayMetrics()
	maxTopOffset := 0
	for _, mon := range monitors {
		if offset := mon.Top - vTop; offset > maxTopOffset {
			maxTopOffset = offset
		}
	}
	overlayHeight := maxTopOffset + 680
	if vHeight > 0 && overlayHeight > vHeight {
		overlayHeight = vHeight
	}
	if vWidth <= 0 {
		vWidth = 1920
	}

	win := wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:             "Vinta",
		Width:             vWidth,
		Height:            overlayHeight,
		InitialPosition:   application.WindowXY,
		X:                 vLeft,
		Y:                 vTop,
		Frameless:         true,
		AlwaysOnTop:       true,
		DisableResize:     true,
		BackgroundType:    application.BackgroundTypeTransparent,
		BackgroundColour:  application.NewRGBA(0, 0, 0, 0),
		IgnoreMouseEvents: false,
		Windows: application.WindowsWindow{
			DisableFramelessWindowDecorations: true,
			HiddenOnTaskbar:                   true,
		},
	})

	app.Setup(wailsApp, win)

	if app.IsTestMode() {
		devWin := wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
			Title:            "Vinta Dev Tools",
			Width:            500,
			Height:           600,
			InitialPosition:  application.WindowCenter,
			Frameless:        false,
			AlwaysOnTop:      true,
			DisableResize:    false,
			URL:              "/?window=devtest",
			BackgroundType:   application.BackgroundTypeSolid,
			BackgroundColour: application.NewRGBA(24, 24, 27, 255),
		})
		app.SetDevWindow(devWin)
		devWin.Show()
	}

	err := wailsApp.Run()
	if err != nil {
		println("Error:", err.Error())
	}
}
