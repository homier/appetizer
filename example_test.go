package appetizer_test

import (
	"context"
	"fmt"
	"sort"

	"github.com/homier/appetizer"
	"github.com/homier/appetizer/log"
	"github.com/pkg/errors"
)

type (
	QueueFiller struct {
		logger log.Logger
		deps   QueueFillerDeps
	}

	QueueFillerDeps chan int
)

var _ appetizer.Servicer = (*QueueFiller)(nil)

// Init implements appetizer.Servicer.
func (q *QueueFiller) Init(log log.Logger, deps appetizer.Dependencies) error {
	q.logger = log

	var ok bool
	q.deps, ok = deps.(QueueFillerDeps)
	if !ok {
		return errors.New("invalid dependencies provided")
	}

	return nil
}

// Run implements appetizer.Servicer.
func (q *QueueFiller) Run(ctx context.Context) error {
	q.logger.Info().Msg("Started")

	for i := range 10 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			q.deps <- i
		}
	}

	q.logger.Info().Msg("Finished")
	close(q.deps)
	return nil
}

func ExampleApp() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queue := make(QueueFillerDeps, 10)
	app := &appetizer.App{
		Name: "simple",
		Services: []appetizer.Service{
			{
				Name:     "Queue filler",
				Servicer: &QueueFiller{},
				Deps:     queue,
			},
		},
	}

	doneCh := make(chan struct{}, 1)
	go func() {
		defer close(doneCh)

		numbers := make([]int, 0, 10)
		for i := range queue {
			numbers = append(numbers, i)
		}

		sort.Slice(numbers, func(i, j int) bool {
			return numbers[i] < numbers[j]
		})

		fmt.Println(numbers)
	}()

	select {
	case err := <-app.RunCh(ctx):
		if err != nil {
			panic(err)
		}
	case <-ctx.Done():
	}

	<-doneCh
	//Output: [0 1 2 3 4 5 6 7 8 9]
}
