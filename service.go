package appetizer

import (
	"context"

	"github.com/homier/appetizer/log"
	"github.com/homier/appetizer/retry"
)

// Service logic.
// No explicit `Stop` method is required, so if you want to gracefully stop your service,
// consider handling context cancellation.
//
//go:generate mockery --name Servicer
type Servicer interface {
	// An initial stage for every service lifecycle.
	// It could be possible called more than once, so
	// you need to decide by yourself, whether you'll support this or not.
	Init(log log.Logger, deps Dependencies) error

	// Run your logic here.
	// If this method returns `nil`, a service is considered stopped,
	// it won't be restarted event if the `Service.RestartEnabled` is true.
	// If this method returns some kind of error, a service is considered failed,
	// and it'll be restarted depending on the `Service.RestartEnabled` and
	// `Service.RestartOpts` policy.
	Run(ctx context.Context) error
}

// Service descriptor. Here you describe your service and specify the target `Servicer`.
type Service struct {
	// Service name. This name will be used as the "service" field value in logger.
	Name string

	// Servicer value. Actual logic for the service.
	Servicer Servicer

	// Optional dependencies for the servicer. It will be passed as the parameter
	// during `Init` call of the servicer.
	Deps Dependencies

	// Whether to restart failed service or not.
	RestartEnabled bool

	// If `RestartEnabled` is `true`, this must be defined to describe
	// a restart policy you need.
	// NOTE: `RestartOpts.Opts` must be defined, otherwise service won't be restarted.
	RestartOpts retry.Opts
}

// An interface each `Dependencies` type must implement.
//
//go:generate mockery --name Dependencies
type Dependencies interface {
}
