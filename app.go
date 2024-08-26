// Appetizer is simple yet effective way to create applications with background services.
// See README.md for more information.
package appetizer

import (
	"context"
	stdErrors "errors"

	"github.com/pkg/errors"
	"github.com/sourcegraph/conc/pool"

	"github.com/homier/appetizer/log"
	"github.com/homier/appetizer/retry"
)

type App struct {
	// Application name. This name will be used as the "app" field value in logger
	Name string

	// A list of services to run. If empty, app will exit immediately.
	Services []Service

	// Configure app to run in debug mode. Will set logger level to `zerolog.DebugLevel`.
	Debug bool

	log log.Logger
}

// Run application, blocking until error or nil is returned.
func (a *App) Run(ctx context.Context) error {
	return <-a.RunCh(ctx)
}

// Run application in background, returning an error channel.
// Application is considered stopped when that channel is closed
// or has an error within.
func (a *App) RunCh(ctx context.Context) <-chan error {
	errCh := make(chan error, 1)
	if len(a.Services) == 0 {
		close(errCh)
		return errCh
	}

	if err := a.init(); err != nil {
		errCh <- err
		close(errCh)

		return errCh
	}

	pool := pool.New().WithContext(ctx).WithCancelOnError().WithFirstError()

	for _, service := range a.Services {
		service := service

		pool.Go(func(ctx context.Context) error {
			return a.runService(ctx, &service)
		})
	}

	a.log.Info().Msg("Application started")

	go func() {
		defer close(errCh)

		if err := pool.Wait(); err != nil {
			errCh <- err
		}
	}()

	return errCh
}

func (a *App) init() (errs error) {
	a.log = a.appLogger()

	if len(a.Services) == 0 {
		return
	}

	for _, service := range a.Services {
		log := a.serviceLogger(service.Name)

		// TODO: Concrete types for dependencies
		if err := service.Servicer.Init(log, service.Deps); err != nil {
			errs = stdErrors.Join(errs, err)
		}
	}

	return
}

func (a *App) runService(ctx context.Context, service *Service) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var err error

	enableRestart := service.RestartEnabled
	if enableRestart && service.RestartOpts.Opts == nil {
		a.log.Warn().Str("service", service.Name).Msg(
			"Service is set up as restartable," +
				" but no options were provided." +
				" Restart is skipped.",
		)
		enableRestart = false
	}

	if enableRestart {
		err = retry.With(ctx, service.Servicer.Run, service.RestartOpts)
	} else {
		err = service.Servicer.Run(ctx)
	}

	if err != nil {
		err = errors.Wrapf(err, "service '%s' crashed", service.Name)
	}

	return err
}

func (a *App) appLogger() log.Logger {
	return log.Setup(a.Debug, log.ContextualField{Name: "app", Value: a.Name})
}

func (a *App) serviceLogger(name string) log.Logger {
	return log.EnrichLogger(a.log, a.Debug, log.ContextualField{
		Name:  "service",
		Value: name,
	})
}
