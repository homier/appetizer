package appetizer

import (
	"context"
	"sync"
	"sync/atomic"
)

type Waiter struct {
	ready atomic.Bool

	mu   sync.Mutex
	once sync.Once
	cond *sync.Cond
}

func (w *Waiter) Wait(ctx context.Context) error {
	select {
	case <-w.WaitCh():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

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

func (w *Waiter) Set(ready bool) {
	if swapped := w.ready.CompareAndSwap(!ready, ready); ready && swapped {
		w.ensureCond()
		w.cond.Broadcast()
	}
}

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
