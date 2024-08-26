package appetizer

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

var SignalsDefault = []os.Signal{
	syscall.SIGINT,
	syscall.SIGTERM,
}

func NotifyContext(signals ...os.Signal) (context.Context, context.CancelFunc) {
	if len(signals) == 0 {
		signals = []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	}

	return signal.NotifyContext(context.Background(), signals...)
}
