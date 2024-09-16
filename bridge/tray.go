package bridge

import (
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var (
	systemTray *application.SystemTray
)

func InitTray(app *application.App, icon []byte, fs embed.FS) {
	src := "frontend/dist/icons/"
	dst := "data/.cache/icons/"

	icons := []string{
		"tray_normal_light.png",
		"tray_normal_dark.png",
		"tray_proxy_light.png",
		"tray_proxy_dark.png",
		"tray_tun_light.png",
		"tray_tun_dark.png",
	}

	os.MkdirAll(GetPath(dst), os.ModePerm)

	for _, icon := range icons {
		path := GetPath(dst + icon)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Printf("InitTray [Icon]: %s", dst+icon)
			b, _ := fs.ReadFile(src + icon)
			os.WriteFile(path, b, os.ModePerm)
		}
	}

	systemTray = app.NewSystemTray()

	systemTray.SetDarkModeIcon(icon)
	systemTray.SetIcon(icon)

	systemTray.OnClick(func() {
		win := app.GetWindowByName("Main")
		win.UnMinimise()
		win.Show()
	})
}

func (a *App) UpdateTray(tray TrayContent) {
	if tray.Icon != "" {
		icon, err := os.ReadFile(GetPath(tray.Icon))
		if err == nil {
			systemTray.SetIcon(icon)
			systemTray.SetDarkModeIcon(icon)
		}
	}
	if tray.Title != "" {
		systemTray.SetLabel(tray.Title)
		a.Ctx.GetWindowByName("Main").SetTitle(tray.Title)
	}
}

func (a *App) UpdateTrayMenus(menus []MenuItem) {
	log.Printf("UpdateTrayMenus")

	appMenu := a.Ctx.NewMenu()

	for _, menu := range menus {
		createMenuItem(menu, a, appMenu)
	}

	systemTray.SetMenu(appMenu)
}

func createMenuItem(menu MenuItem, a *App, parent *application.Menu) {
	if menu.Hidden {
		return
	}

	if len(menu.Children) != 0 {
		subMenu := parent.AddSubmenu(menu.Text)
		for _, child := range menu.Children {
			createMenuItem(child, a, subMenu)
		}
		return
	}

	onClick := func(ctx *application.Context) {
		log.Printf("%v", menu)
		a.Ctx.EmitEvent("onTrayMenuClick", menu.Id)
	}

	switch menu.Type {
	case "item":
		parent.Add(menu.Text).SetTooltip(menu.Tooltip).SetChecked(menu.Checked).OnClick(onClick)
	case "radio":
		parent.AddRadio(menu.Text, menu.Checked).SetTooltip(menu.Tooltip).OnClick(onClick)
	case "separator":
		parent.AddSeparator()
	}
}
