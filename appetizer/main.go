package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/homier/appetizer"
	"github.com/homier/appetizer/log"
)

var Version string

type srv struct {
	l    log.Logger
	deps appetizer.Dependencies
}

func (s *srv) Init(l log.Logger, deps appetizer.Dependencies) error {
	s.l = l
	s.deps = deps

	s.l.Info().Msg("Initialized")

	return nil
}

func (s *srv) Run(ctx context.Context) error {
	s.l.Info().Msg("Started")

	<-ctx.Done()

	s.l.Info().Msg("Stopped")

	return nil
}

func main() {
	cli.VersionPrinter = func(ctx *cli.Context) {
		fmt.Printf("%s\n", ctx.App.Version)
	}

	ctx, cancel := appetizer.NotifyContext()
	defer cancel()

	app := &cli.App{
		Name:    "appetizer",
		Version: Version,
		Commands: []*cli.Command{
			{
				Name:  "test",
				Flags: []cli.Flag{},
				Action: func(_ *cli.Context) error {
					app := &appetizer.App{Name: "test", Services: []appetizer.Service{{
						Name:           "Test node",
						Servicer:       &srv{},
						RestartEnabled: true,
					}}}

					if err := app.Init(); err != nil {
						return err
					}

					return app.Run(ctx)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
