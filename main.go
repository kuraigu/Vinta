// Package main provides the entry point and window configuration for Vinta.
package main

import (
	"Vinta/backend"
	"Vinta/backend/devtest"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/windows/icon.ico
var icon []byte

func main() {
	// Create an instance of the app structure
	app := backend.NewApp(icon)
	devTools := devtest.NewDevTest()
	cache := backend.NewCache()

	// Create application with options
	err := wails.Run(&options.App{
		Title:         "Vinta",
		Frameless:     true,
		AlwaysOnTop:   true,
		DisableResize: true,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "vinta-dynamic-island",
		},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		Bind: []interface{}{
			app,
			devTools,
			cache,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               true,
			BackdropType:                      windows.None,
			DisableWindowIcon:                 true,
			DisableFramelessWindowDecorations: true,
			WebviewGpuIsDisabled:              false,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
