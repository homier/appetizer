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
* Integrated pprof (coming soon)

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

