package main

import (
	"embed"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func goos() string   { return runtime.GOOS }
func goarch() string { return runtime.GOARCH }

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "DJI Caster",
		Width:     420,
		Height:    860,
		MinWidth:  380,
		MinHeight: 640,
		MaxWidth:  560,
		MaxHeight: 1200,
		// The UI is designed for a cast-sized window: a single narrow column.
		BackgroundColour: &options.RGBA{R: 251, G: 249, B: 243, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
