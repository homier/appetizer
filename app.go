package appetizer

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/sourcegraph/conc/pool"

	"github.com/homier/appetizer/log"
)

type Leaf interface {
	Init(log log.Logger, deps Dependency) error
	Run(ctx context.Context) error
}

type Node struct {
	Name string
	Leaf Leaf
	Deps Dependency
}

type Dependency interface {
}

type App struct {
	Name  string
	Nodes []Node
	Debug bool

	log log.Logger
}

func (a *App) Init() error {
	for _, node := range a.Nodes {
		log := log.EnrichLogger(a.log, a.Debug, log.ContextualField{Name: "node", Value: node.Name})

		if err := node.Leaf.Init(log, node.Deps); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) Run(ctx context.Context) error {
	a.initLogger()

	if err := a.Init(); err != nil {
		return err
	}

	a.log.Info().Msg("Started")
	err := <-a.RunCh(ctx)
	a.log.Info().Msg("Stopped")

	return err
}

func (a *App) RunCh(ctx context.Context) <-chan error {
	errCh := make(chan error, 1)
	p := pool.New().WithContext(ctx).WithCancelOnError().WithFirstError()

	for _, node := range a.Nodes {
		node := node

		p.Go(func(ctx context.Context) error {
			nodeErrCh := make(chan error, 1)

			go func() {
				defer close(nodeErrCh)
				if err := node.Leaf.Run(ctx); err != nil {
					nodeErrCh <- errors.Wrapf(err, "node '%s' crashed", node.Name)
				}
			}()

			select {
			case <-ctx.Done():
				return nil
			case err := <-nodeErrCh:
				return err
			}
		})
	}

	go func() {
		defer close(errCh)
		if err := p.Wait(); err != nil {
			errCh <- err
		}
	}()

	return errCh
}

func (a *App) initLogger() {
	a.log = log.Setup(a.Debug, log.ContextualField{Name: "app", Value: a.Name})
}

func NotifyContext(signals ...os.Signal) (context.Context, context.CancelFunc) {
	if len(signals) == 0 {
		signals = []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	}

	return signal.NotifyContext(context.Background(), signals...)
}
