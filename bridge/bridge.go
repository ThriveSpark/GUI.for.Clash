package bridge

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/klauspost/cpuid/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
	"gopkg.in/yaml.v3"
)

type App struct {
	Ctx *application.App
}

var Env = &EnvResult{
	BasePath:    "",
	AppName:     "",
	OS:          runtime.GOOS,
	ARCH:        runtime.GOARCH,
	X64Level:    cpuid.CPU.X64Level(),
	FromTaskSch: false,
}

var Config = &AppConfig{}

var isStartup = true

func InitApp() {
	// step1: Set Env
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	for _, v := range os.Args {
		if v == "tasksch" {
			Env.FromTaskSch = true
			break
		}
	}

	Env.BasePath = filepath.Dir(exePath)
	Env.AppName = filepath.Base(exePath)

	// step2: Read Config
	b, err := os.ReadFile(Env.BasePath + UserProfile)
	if err == nil {
		yaml.Unmarshal(b, &Config)
	}

	if Config.Width == 0 {
		Config.Width = 800
	}

	if Config.Height == 0 {
		if Env.OS == "linux" {
			Config.Height = 510
		} else {
			Config.Height = 540
		}
	}

	if Env.OS == "windows" {
		Config.BackgroundType = int(application.BackgroundTypeTranslucent)
	} else {
		Config.BackgroundType = int(application.BackgroundTypeSolid)
	}

	Config.Hidden = Env.FromTaskSch && (Config.WindowStartState == int(application.WindowStateMinimised))

	if !Env.FromTaskSch {
		Config.WindowStartState = int(application.WindowStateNormal)
	}
}

func (a *App) ExitApp() {
	a.Ctx.Quit()
}

func (a *App) RestartApp() FlagResult {
	exePath := Env.BasePath + "/" + Env.AppName

	cmd := exec.Command(exePath)
	HideExecWindow(cmd)

	err := cmd.Start()
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	a.Ctx.Quit()

	return FlagResult{true, "Success"}
}

func (a *App) GetEnv() EnvResult {
	return EnvResult{
		AppName:  Env.AppName,
		BasePath: Env.BasePath,
		OS:       Env.OS,
		ARCH:     Env.ARCH,
		X64Level: Env.X64Level,
	}
}

func (a *App) IsStartup() bool {
	if isStartup {
		isStartup = false
		return true
	}
	return false
}

func (a *App) GetInterfaces() FlagResult {
	log.Printf("GetInterfaces")

	interfaces, err := net.Interfaces()
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	var interfaceNames []string

	for _, inter := range interfaces {
		interfaceNames = append(interfaceNames, inter.Name)
	}

	return FlagResult{true, strings.Join(interfaceNames, "|")}
}

func (a *App) BridgeHTTPApi(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/bridge/fs/write" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.Write(GetBridgeHTTPApiError(err))
				return
			}

			params := BridgeHTTPApiIOParams{}

			err = json.Unmarshal(body, &params)
			if err != nil {
				w.Write(GetBridgeHTTPApiError(err))
				return
			}

			result := a.Writefile(params.Path, params.Content, params.Options)

			b, err := json.Marshal(result)
			if err != nil {
				w.Write(GetBridgeHTTPApiError(err))
				return
			}

			w.Write(b)
			return
		}

		if r.Method == "POST" && r.URL.Path == "/bridge/fs/read" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.Write(GetBridgeHTTPApiError(err))
				return
			}

			params := BridgeHTTPApiIOParams{}

			err = json.Unmarshal(body, &params)
			if err != nil {
				w.Write(GetBridgeHTTPApiError(err))
				return
			}

			result := a.Readfile(params.Path, params.Options)

			b, err := json.Marshal(result)
			if err != nil {
				w.Write(GetBridgeHTTPApiError(err))
				return
			}

			w.Write(b)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) BridgeRollingReleaseApi(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !Config.RollingRelease {
			next.ServeHTTP(w, r)
			return
		}

		url := r.URL.Path
		if url == "/" {
			url = "/index.html"
		}

		log.Printf("[Rolling Release] %v %v\n", r.Method, url)

		file := GetPath("data/rolling-release" + url)

		bytes, err := os.ReadFile(file)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ext := path.Ext(url)
		mime := "application/octet-stream"

		switch ext {
		case ".html":
			mime = "text/html"
		case ".ico":
			mime = "image/x-icon"
		case ".png":
			mime = "image/png"
		case ".css":
			mime = "text/css"
		case ".js":
			mime = "text/javascript"
		}

		w.Header().Set("Content-Type", mime)
		w.Write(bytes)
	})
}
