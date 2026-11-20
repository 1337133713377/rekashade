package main

import (
	"embed"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/1337133713377/rekashade/internal/updater"
)

//go:embed all:frontend/dist
var assets embed.FS

// version is set at build time: -ldflags "-X main.version=1.2.3"
var version = "1.0.1"

type launchArgs struct {
	minimized bool
	updated   bool
	waitPID   int
}

func parseArgs(argv []string) launchArgs {
	var a launchArgs
	for i := 0; i < len(argv); i++ {
		switch argv[i] {
		case "--minimized", "/minimized":
			a.minimized = true
		case "--updated":
			a.updated = true
		case "--wait-pid":
			if i+1 < len(argv) {
				a.waitPID, _ = strconv.Atoi(argv[i+1])
				i++
			}
		}
	}
	return a
}

func main() {
	args := parseArgs(os.Args[1:])
	if args.waitPID > 0 {
		// Started by the updater: let the previous version finish exiting
		// before we take the single-instance lock.
		updater.WaitForExit(args.waitPID, 20*time.Second)
	}

	app := NewApp(args)

	webviewData := ""
	if dir, err := os.UserCacheDir(); err == nil {
		webviewData = filepath.Join(dir, "RekaShade", "WebView2")
	}

	err := wails.Run(&options.App{
		Title:            "Reka Shade",
		Width:            1160,
		Height:           800,
		MinWidth:         minWindowWidth,
		MinHeight:        minWindowHeight,
		Frameless:        true,
		StartHidden:      args.minimized,
		BackgroundColour: &options.RGBA{R: 11, G: 9, B: 20, A: 255},
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		Bind:             []interface{}{app},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "8f2b6c1e-reka-shade-3d1a7e5f",
			OnSecondInstanceLaunch: func(options.SecondInstanceData) { app.ShowWindow() },
		},
		EnableDefaultContextMenu: false,
		Windows: &windows.Options{
			Theme:                windows.Dark,
			WindowClassName:      mainWindowClass,
			WebviewUserDataPath:  webviewData,
			IsZoomControlEnabled: false,
			ZoomFactor:           1.0,
			DisablePinchZoom:     true,
			OnResume:             app.onResume,
		},
	})
	if err != nil {
		println("Error:", err.Error())
		os.Exit(1)
	}
}
