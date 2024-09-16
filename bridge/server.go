package bridge

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v3/pkg/application"
)

var serverMap = make(map[string]*http.Server)

type ResponseData struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Options IOOptions         `json:"options"`
}

type RequestData struct {
	Id      string              `json:"id"`
	Method  string              `json:"method"`
	Url     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

func (a *App) StartServer(address string, serverID string) FlagResult {
	log.Printf("StartServer: %s", address)

	server := &http.Server{
		Addr: address,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(500)
				log.Printf("reading body error: %v", err)
				return
			}

			requestID := uuid.New().String()
			respChan := make(chan ResponseData)
			respBody := []byte{}

			defer close(respChan)

			a.Ctx.OnEvent(requestID, func(event *application.CustomEvent) {
				a.Ctx.OffEvent(requestID)
				resp := ResponseData{200, map[string]string{}, "A sample http server", IOOptions{Mode: Text}}

				if data, ok := event.Data.(map[string]interface{}); ok {
					if status, ok := data["status"].(float64); ok {
						resp.Status = int(status)
					}
					if headers, ok := data["headers"].(map[string]interface{}); ok {
						for key, value := range headers {
							if strValue, ok := value.(string); ok {
								resp.Headers[key] = strValue
							}
						}
					}
					if body, ok := data["body"].(string); ok {
						resp.Body = body
					}
					if options, ok := data["options"].(map[string]interface{}); ok {
						if mode, ok := options["Mode"].(string); ok {
							resp.Options.Mode = mode
						}
					}
					if resp.Options.Mode == Text {
						respBody = []byte(resp.Body)
					} else {
						respBody, err = base64.StdEncoding.DecodeString(resp.Body)
						if err != nil {
							resp.Status = 500
							respBody = []byte(err.Error())
						}
					}
				}
				respChan <- resp
			})

			a.Ctx.EmitEvent(serverID, &RequestData{
				Id:      requestID,
				Method:  r.Method,
				Url:     r.URL.RequestURI(),
				Headers: r.Header,
				Body:    string(body),
			})

			res := <-respChan
			for key, value := range res.Headers {
				w.Header().Set(key, value)
			}
			w.WriteHeader(res.Status)
			w.Write(respBody)
		}),
	}

	var result error

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			result = err
		}
	}()

	time.Sleep(1 * time.Second)

	if result != nil {
		return FlagResult{false, result.Error()}
	}

	serverMap[serverID] = server

	return FlagResult{true, "Success"}
}

func (a *App) StopServer(id string) FlagResult {
	log.Printf("StopServer: %s", id)

	server, ok := serverMap[id]
	if !ok {
		return FlagResult{false, "server not found"}
	}

	err := server.Close()
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	delete(serverMap, id)

	return FlagResult{true, "Success"}
}

func (a *App) ListServer() FlagResult {
	log.Printf("ListServer: %v", serverMap)

	var servers []string

	for serverID := range serverMap {
		servers = append(servers, serverID)
	}

	return FlagResult{true, strings.Join(servers, "|")}
}
