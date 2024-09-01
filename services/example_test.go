package services_test

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/homier/appetizer"
	"github.com/homier/appetizer/services"
)

func ExampleHTTPServer() {
	app := appetizer.App{
		Name: "ExampleHTTPServer",
		Services: []appetizer.Service{{
			Name: "http_server",
			Servicer: &services.HTTPServer{
				Handlers: []services.Handler{{
					Path: "GET /hello",
					Handler: func(w http.ResponseWriter, _ *http.Request) {
						if _, err := io.WriteString(w, "world\n"); err != nil {
							panic(err)
						}
					},
				}},
				PprofEnabled: true,
			},
		}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := app.Run(ctx); err != nil {
			app.Log().Fatal().Err(err).Msg("Fatal error while running an application")
		}
	}()

	<-app.WaitCh()

	resp, err := http.Get("http://" + services.DefaultAddress + "/hello")
	if err != nil {
		app.Log().Fatal().Err(err).Msg("Could not query test HTTP server")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Log().Fatal().Err(err).
			Msg("Could not read response body from test HTTP server endpoint")
	}

	fmt.Println(string(body))
	//Output: world
}
