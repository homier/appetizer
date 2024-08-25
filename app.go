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
	Name     string
	Services []Service
	Debug    bool

	log log.Logger
}

func (a *App) Init() (errs error) {
	a.log = a.appLogger()

	for _, service := range a.Services {
		log := a.serviceLogger(service.Name)

		if err := service.Lifecycle.Init(log, service.Deps); err != nil {
			errs = stdErrors.Join(errs, err)
		}
	}

	return
}

func (a *App) Run(ctx context.Context) error {
	a.log.Info().Msg("Application started")
	err := <-a.RunCh(ctx)
	a.log.Info().Msg("Application stopped")

	return err
}

func (a *App) RunCh(ctx context.Context) <-chan error {
	errCh := make(chan error, 1)

	ctx, cancel := context.WithCancel(ctx)
	pool := pool.New().WithContext(ctx).WithCancelOnError().WithFirstError()

	for _, service := range a.Services {
		service := service

		pool.Go(func(ctx context.Context) error {
			var err error

			enableRestart := service.RestartEnabled
			if enableRestart && service.RestartOpts.Opts == nil {
				a.log.Warn().Str("service", service.Name).Msg(
					"Service is set up as restartable," +
						" but no restart options were provided." +
						" Restart is skipped.",
				)
				enableRestart = false
			}

			if enableRestart {
				err = retry.With(ctx, service.RestartOpts)
			} else {
				err = service.Lifecycle.Run(ctx)
			}

			if err != nil {
				err = errors.Wrapf(err, "service '%s' crashed", service.Name)
			}

			return err
		})
	}

	go func() {
		defer cancel()
		defer close(errCh)

		if err := pool.Wait(); err != nil {
			errCh <- err
		}
	}()

	return errCh
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
