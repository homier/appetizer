# appetizer

## TL;DR
Appetizer is simple yet effective way to create applications with background services.

## Why?
My requirements have been always simple: I need a tool to run multiple instances of something in background,
while being able to restart them if needed or to fail fast otherwise.

## Features
* Declarative approach to define your application
* Consistent logging for each service, by injecting [zerolog.Logger](https://github.com/rs/zerolog) instance on service initialization
* Any service could be configured as restartable thanks to awesome [cenkalti/backoff](https://github.com/cenkalti/backoff) library.
* Integrated HTTP servicer with pprof

## Examples
### Simple time printer
```go
package main

import (
    "context"
    "errors"
    "time"

    "github.com/homier/appetizer"
    "github.com/homier/appetizer/log"
)

type TimePrinter struct {
    TickDuration time.Duration

    log log.Logger
}

func (tp *TimePrinter) Init(log log.Logger) error {
    tp.log = log

    if tp.TickDuration == time.Duration(0) {
        return errors.New("TickDuration must be defined")
    }

    return nil
}

func (tp *TimePrinter) Run(ctx context.Context) error {
    tp.log.Msg("Started")
    defer tp.log.Msg("Stopped")

    ticker := time.NewTicker(tp.TickDuration)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil
        case tick := <-ticker.C:
            tp.log.Info().Msgf("Ticked at: %s", tick.Format(time.RFC822Z))
        }
    }
}

func main() {
    app := &appetizer.App{
        Name: "SimpleApplication"
        Services: []appetizer.Service{{
            Name: "TimePrinter",
            Servicer: &TimePrinter{TickDuration: time.Second},
        }},
    }

    ctx, cancel := appetizer.NotifyContext()
    defer cancel()

    if err := app.Run(ctx); err != nil {
        panic(err)
    }
}
```

### HTTP servicer
```go
package main

import (
	"io"
	"net/http"

	"github.com/homier/appetizer"
	"github.com/homier/appetizer/services"
)

func main() {
	app := appetizer.App{
		Name:  "WebServer",
		Debug: true,
		Services: []appetizer.Service{{
			Name: "httpServer",
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

	ctx, cancel := appetizer.NotifyContext()
	defer cancel()

	runCh := app.RunCh(ctx)

	<-app.WaitCh()

	select {
	case <-ctx.Done():
		return
	case err := <-runCh:
		if err != nil {
			app.Log().Fatal().Err(err).Msg("App crashed")
		}
	}
}
```

