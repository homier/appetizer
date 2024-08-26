package retry

import (
	"context"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
)

// Retry options definition.
type Opts struct {
	// See `https://github.com/cenkalti/backoff` for more information.
	Opts backoff.BackOff

	// If set, no retries will be attempted if a target callable returns
	// this exact error.
	CriticalError error

	// If set to something greater than 0, a target callable won't be run
	// more than `MaxRetry + 1` times.
	MaxRetry uint64
}

// Run provided `target` callable with retry policy.
// Based on `https://github.com/cenkalti/backoff` library.
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

		if opts.CriticalError == nil {
			return err
		}

		if errors.Is(err, opts.CriticalError) {
			return backoff.Permanent(err)
		}

		return err
	}, strategy)
}
