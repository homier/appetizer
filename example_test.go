package appetizer_test

import (
	"context"
	"fmt"
	"sort"

	"github.com/pkg/errors"

	"github.com/homier/appetizer"
	"github.com/homier/appetizer/log"
)

type QueueFiller struct {
	Queue  chan<- int
	logger log.Logger
}

var _ appetizer.Servicer = (*QueueFiller)(nil)

// Init implements appetizer.Servicer.
func (q *QueueFiller) Init(log log.Logger) error {
	q.logger = log

	if q.Queue == nil {
		return errors.New("queue has not been provided")
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
			q.Queue <- i
		}
	}

	q.logger.Info().Msg("Finished")
	close(q.Queue)
	return nil
}

func ExampleApp() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queue := make(chan int, 10)
	app := &appetizer.App{
		Name: "simple",
		Services: []appetizer.Service{
			{
				Name: "Queue filler",
				Servicer: &QueueFiller{
					Queue: queue,
				},
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
