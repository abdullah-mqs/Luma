package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "luma",
		Description: "Fast, beautiful note-taking app",
		Services: []application.Service{
			application.NewService(NewApp()),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Luma",
		Width:            1024,
		Height:           768,
		StartState:       application.WindowStateMaximised,
		BackgroundColour: application.NewRGBA(27, 38, 54, 1),
		URL:              "/",
	})

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
