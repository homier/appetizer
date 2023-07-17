package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/homier/appetizer"
	"github.com/homier/appetizer/log"
)

var (
	version string
)

type srv struct {
	l log.Logger
}

func (s *srv) Init(l log.Logger, deps appetizer.Dependency) error {
	s.l = l

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
		Version: version,
		Commands: []*cli.Command{
			{
				Name:  "test",
				Flags: []cli.Flag{},
				Action: func(cCtx *cli.Context) error {
					app := &appetizer.App{Name: "test", Nodes: []appetizer.Node{{
						Name: "Test node",
						Leaf: &srv{},
					}}}
					return app.Run(ctx)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
