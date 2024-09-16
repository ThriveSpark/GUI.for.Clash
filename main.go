package main

import (
	"embed"
	"guiforcores/bridge"
	"log"
	"log/slog"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed frontend/dist
var assets embed.FS

//go:embed frontend/dist/favicon.ico
var icon []byte

//go:embed frontend/dist/icons/tray_normal_dark.png
var trayIcon []byte

// main function serves as the application's entry point. It initializes the application, creates a window,
// and starts a goroutine that emits a time-based event every second. It subsequently runs the application and
// logs any error that might occur.
func main() {
	appService := &bridge.App{}

	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	app := application.New(application.Options{
		Name:        "GUI.for.Cores",
		Description: "A GUI program developed by vue3 + wails3.",
		Icon:        icon,
		LogLevel:    slog.LevelWarn,
		Services: []application.Service{
			application.NewService(appService),
		},
		PanicHandler: func(a any) {
			log.Println(a)
		},
		Assets: application.AssetOptions{
			Handler:        application.AssetFileServerFS(assets),
			Middleware:     application.ChainMiddleware(appService.BridgeHTTPApi, appService.BridgeRollingReleaseApi),
			DisableLogging: true,
		},
	})

	appService.Ctx = app
	bridge.InitApp()
	bridge.InitTray(app, trayIcon, assets)
	bridge.InitNotification(assets)
	bridge.InitScheduledTasks()

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	window := app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Name:                   "Main",
		URL:                    "/",
		MinWidth:               600,
		MinHeight:              400,
		Centered:               true,
		DisableResize:          false,
		OpenInspectorOnStartup: true,
		Title:                  bridge.Env.AppName,
		Width:                  bridge.Config.Width,
		Height:                 bridge.Config.Height,
		Frameless:              bridge.Env.OS == "windows",
		Hidden:                 bridge.Config.Hidden,
		EnableDragAndDrop:      true,
		BackgroundType:         application.BackgroundType(bridge.Config.BackgroundType),
		BackgroundColour:       application.NewRGBA(255, 255, 255, 1),
		StartState:             application.WindowState(bridge.Config.WindowStartState),
		DevToolsEnabled:        true,
		Windows: application.WindowsWindow{
			BackdropType: application.Acrylic,
		},
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		Linux: application.LinuxWindow{
			Icon:                icon,
			WindowIsTranslucent: true,
			WebviewGpuPolicy:    application.WebviewGpuPolicyNever,
		},
		ShouldClose: func(window *application.WebviewWindow) bool {
			appService.Ctx.EmitEvent("onBeforeExitApp")
			return true
		},
		// SingleInstanceLock: &options.SingleInstanceLock{
		// 	UniqueId: func() string {
		// 		if bridge.Config.MultipleInstance {
		// 			return uuid.New().String()
		// 		}
		// 		return bridge.Env.AppName
		// 	}(),
		// 	OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
		// 		runtime.Show(app.Ctx)
		// 		runtime.EventsEmit(app.Ctx, "launchArgs", data.Args)
		// 	},
		// },
	})

	window.OnWindowEvent(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {
		files := event.Context().DroppedFiles()
		app.EmitEvent("onFilesDropped", files)
	})

	// Run the application. This blocks until the application has been exited.
	err := app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
