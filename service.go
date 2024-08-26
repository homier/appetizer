package appetizer

import (
	"context"

	"github.com/homier/appetizer/log"
	"github.com/homier/appetizer/retry"
)

//go:generate mockery --name Servicer
type Servicer interface {
	Init(log log.Logger, deps Dependencies) error
	Run(ctx context.Context) error
}

type Service struct {
	Name     string
	Servicer Servicer
	Deps     Dependencies

	RestartEnabled bool
	RestartOpts    retry.Opts
}

//go:generate mockery --name Dependencies
type Dependencies interface {
}
