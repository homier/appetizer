package appetizer

import (
	"context"

	"github.com/homier/appetizer/log"
	"github.com/homier/appetizer/retry"
)

type Lifecycle interface {
	Init(log log.Logger, deps Dependencies) error
	Run(ctx context.Context) error
	Stop() error
}

type Service struct {
	Name      string
	Lifecycle Lifecycle
	Deps      Dependencies

	RestartEnabled bool
	RestartOpts    retry.Opts
}

type Dependencies interface {
}
