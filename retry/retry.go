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

	Target func(ctx context.Context) error
}

func With(ctx context.Context, opts Opts) error {
	var strategy backoff.BackOff

	strategy = backoff.WithContext(opts.Opts, ctx)
	strategy.Reset()

	if opts.MaxRetry > 0 {
		strategy = backoff.WithMaxRetries(strategy, opts.MaxRetry)
	}

	return backoff.Retry(func() error {
		err := opts.Target(ctx)
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
