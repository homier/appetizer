// Appetizer is simple yet effective way to create applications with background services.
// See README.md for more information.
package appetizer

import (
	"context"
	stdErrors "errors"
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/conc/pool"

	"github.com/homier/appetizer/log"
	"github.com/homier/appetizer/retry"
)

var (
	ErrStarted = errors.New("application is already started")
)

type App struct {
	// Application name. This name will be used as the "app" field value in logger
	Name string

	// A list of services to run. If empty, app will exit immediately.
	Services []Service

	// Configure app to run in debug mode. Will set logger level to `zerolog.DebugLevel`.
	Debug bool

	log     log.Logger
	logOnce sync.Once

	startedWaiter Waiter
}

// Run application, blocking until error or nil is returned.
func (a *App) Run(ctx context.Context) error {
	return <-a.RunCh(ctx)
}

// Run application in background, returning an error channel.
// Application is considered stopped when that channel is closed
// or has an error within.
func (a *App) RunCh(ctx context.Context) <-chan error {
	a.ensureLog()

	a.log.Debug().Msg("app: run: starting...")
	errCh := make(chan error, 1)
	if a.startedWaiter.Is(true) {
		errCh <- ErrStarted
		close(errCh)

		a.log.Debug().Msg("app: run: already started")
		return errCh
	}

	if len(a.Services) == 0 {
		a.log.Debug().Msg("app: run: no services, exiting")
		close(errCh)

		return errCh
	}

	if err := a.init(); err != nil {
		errCh <- err
		close(errCh)

		return errCh
	}

	pool := pool.New().WithContext(ctx).
		WithCancelOnError().
		WithFirstError().
		WithMaxGoroutines(len(a.Services))

	a.log.Debug().Msg("app: run: pool: starting services...")

	readyWg := &sync.WaitGroup{}
	for _, service := range a.Services {
		service := service
		readyWg.Add(1)

		a.log.Debug().Msgf("app: run: pool: service: '%s': starting...", service.Name)
		pool.Go(func(ctx context.Context) error {
			readyWg.Done()

			return a.runService(ctx, &service)
		})
	}

	a.log.Debug().Msg("app: run: pool: waiting for all services to be started")
	readyWg.Wait()

	a.log.Info().Msg("app: run: started")
	a.startedWaiter.Set(true)

	go func() {
		defer close(errCh)
		defer func() { a.startedWaiter.Set(false) }()

		if err := pool.Wait(); err != nil {
			errCh <- err
		}

		a.log.Debug().Msg("app: run: pool: stopped")
	}()

	return errCh
}

func (a *App) Log() *log.Logger {
	log := a.log.With().Logger()

	return &log
}

func (a *App) WaitCh() <-chan struct{} {
	return a.startedWaiter.WaitCh()
}

func (a *App) Wait(ctx context.Context) error {
	return a.startedWaiter.Wait(ctx)
}

func (a *App) init() (errs error) {
	a.log.Debug().Msg("app: init: starting...")

	if len(a.Services) == 0 {
		a.log.Debug().Msg("app: init: no services, exiting")
		return
	}

	for _, service := range a.Services {
		log := a.serviceLogger(service.Name)

		a.log.Debug().Msgf("app: init: service: '%s': initializing", service.Name)
		if err := service.Servicer.Init(log); err != nil {
			log.Debug().Err(err).Msgf("app: init: service: '%s': failed to initialize", service.Name)
			errs = stdErrors.Join(errs, err)
			continue
		}

		a.log.Debug().Msgf("app: init: service: '%s': initialized", service.Name)
	}

	a.log.Debug().Msg("app: init: done")
	return
}

func (a *App) runService(ctx context.Context, service *Service) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	enableRestart := service.RestartEnabled
	if enableRestart && service.RestartOpts.Opts == nil {
		a.log.Warn().Str("service", service.Name).Msgf(
			"app: run: service: '%s': service is set up as restartable,"+
				" but no options were provided."+
				" Restart is skipped.", service.Name,
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

func (a *App) ensureLog() {
	a.logOnce.Do(func() {
		a.log = log.Setup(a.Debug, log.ContextualField{Name: "app", Value: a.Name})
	})
}

func (a *App) serviceLogger(name string) log.Logger {
	return log.EnrichLogger(a.log, a.Debug, log.ContextualField{
		Name:  "service",
		Value: name,
	})
}
