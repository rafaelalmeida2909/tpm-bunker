package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "TPM-BUNKER",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		// Opcional: Você pode adicionar mais configurações aqui, como:
		MinWidth:         800,            // Largura mínima da janela
		MinHeight:        600,            // Altura mínima da janela
		DisableResize:    false,          // Se deseja permitir redimensionamento
		Fullscreen:       false,          // Se deve iniciar em tela cheia
		WindowStartState: options.Normal, // Estado inicial da janela
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
