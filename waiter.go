package appetizer

import (
	"context"
	"sync"
	"sync/atomic"
)

// Waiter is a special struct that provides a group of methods to wait the
// internal condition to become true and to manage that condition.
type Waiter struct {
	ready atomic.Bool

	mu   sync.Mutex
	once sync.Once
	cond *sync.Cond
}

// Blocks until the internal condition is set to true,
// or the provided context is either cancelled or timed out.
// If context is cancelled/timed out, a context error is returning,
// otherwise no error is returning.
func (w *Waiter) Wait(ctx context.Context) error {
	select {
	case <-w.WaitCh():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Returns a channel that will be closed when the internal condition
// is set to true.
func (w *Waiter) WaitCh() <-chan struct{} {
	ch := make(chan struct{}, 1)
	if w.ready.Load() {
		close(ch)

		return ch
	}

	w.ensureCond()

	go func() {
		defer close(ch)

		w.cond.L.Lock()
		for !w.ready.Load() {
			w.cond.Wait()
		}
		w.cond.L.Unlock()

		ch <- struct{}{}
	}()

	return ch
}

// Sets the internal condition to the provided ready value.
// If the provided value is true, all of the wait channels
// will be closed, signalling the readiness of the waiter.
func (w *Waiter) Set(ready bool) {
	if swapped := w.ready.CompareAndSwap(!ready, ready); ready && swapped {
		w.ensureCond()
		w.cond.Broadcast()
	}
}

// Checks that the internal condition is equal to the provided boolean value.
func (w *Waiter) Is(ready bool) bool {
	return w.ready.Load() == ready
}

func (w *Waiter) ensureCond() {
	w.once.Do(func() {
		if w.cond != nil {
			return
		}

		w.mu.Lock()
		if w.cond == nil {
			w.cond = sync.NewCond(&w.mu)
		}
		w.mu.Unlock()
	})
}
