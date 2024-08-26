package appetizer

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// A list of default os signals for NotifyContext.
var SignalsDefault = []os.Signal{
	syscall.SIGINT,
	syscall.SIGTERM,
}

// Returns a `context.Context` with its cancel function,
// that will be cancelled on catching specified signals.
// If no signals are provided, the `SignalsDefault` will be used.
func NotifyContext(signals ...os.Signal) (context.Context, context.CancelFunc) {
	if len(signals) == 0 {
		signals = []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	}

	return signal.NotifyContext(context.Background(), signals...)
}
