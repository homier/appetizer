package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

var (
	version string
)

func main() {
	cli.VersionPrinter = func(ctx *cli.Context) {
		fmt.Printf("%s\n", ctx.App.Version)
	}

	app := &cli.App{
		Name:    "appetizer",
		Version: version,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
