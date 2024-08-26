package retry

import (
	"context"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
)

type Opts struct {
	Opts backoff.BackOff

	StopError error
	MaxRetry  uint64
}

func With(ctx context.Context, target func(context.Context) error, opts Opts) error {
	var strategy backoff.BackOff

	strategy = backoff.WithContext(opts.Opts, ctx)
	strategy.Reset()

	if opts.MaxRetry > 0 {
		strategy = backoff.WithMaxRetries(strategy, opts.MaxRetry)
	}

	return backoff.Retry(func() error {
		err := target(ctx)
		if err == nil {
			return nil
		}

		if opts.StopError == nil {
			return err
		}

		if errors.Is(err, opts.StopError) {
			return backoff.Permanent(err)
		}

		return err
	}, strategy)
}
