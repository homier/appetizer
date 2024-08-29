package appetizer

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaiter_ensureCond(t *testing.T) {
	t.Run("condition is set up", func(t *testing.T) {
		w := &Waiter{}
		cond := sync.NewCond(&w.mu)
		w.cond = cond

		w.ensureCond()

		assert.Equal(t, cond, w.cond)
	})

	t.Run("exactly once", func(t *testing.T) {
		w := &Waiter{}
		w.ensureCond()

		if assert.NotNil(t, w.cond) {
			w.cond = nil
			w.ensureCond()

			assert.Nil(t, w.cond)
		}
	})
}

func TestWaiter_Is(t *testing.T) {
	w := &Waiter{}

	if assert.True(t, w.Is(false)) {
		w.ready.Store(true)
		assert.True(t, w.Is(true))
	}
}

func TestWaiter_Set(t *testing.T) {
	t.Run("ready is false", func(t *testing.T) {
		w := &Waiter{}
		w.Set(false)

		assert.Nil(t, w.cond)
	})

	t.Run("ready is true", func(t *testing.T) {
		w := &Waiter{}

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
		defer cancel()

		go func() {
			w.Set(true)
		}()

		select {
		case <-ctx.Done():
			t.Fatal("wait should've completed")
		case <-w.WaitCh():
		}
	})
}

func TestWaiter_WaitCh(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		w := &Waiter{}
		w.Set(true)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
		defer cancel()

		select {
		case <-ctx.Done():
			t.Fatal("wait should've completed")
		case <-w.WaitCh():
		}
	})

	t.Run("condition is created", func(t *testing.T) {
		w := &Waiter{}

		if assert.Nil(t, w.cond) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
			defer cancel()

			select {
			case <-w.WaitCh():
				t.Fatal("wait should've failed")
			case <-ctx.Done():
				assert.NotNil(t, w.cond)
			}
		}
	})

	t.Run("wait", func(t *testing.T) {
		w := &Waiter{}

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
		defer cancel()

		ch := w.WaitCh()

		go func() {
			<-time.After(time.Millisecond)
			w.Set(true)
		}()

		select {
		case <-ctx.Done():
			t.Fatal("wait should've completed")
		case <-ch:
			assert.True(t, w.Is(true))
		}
	})
}

func TestWaiter_Wait(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		w := &Waiter{}
		w.Set(true)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
		defer cancel()

		assert.NoError(t, w.Wait(ctx))
	})

	t.Run("timeout", func(t *testing.T) {
		w := &Waiter{}

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
		defer cancel()

		if err := w.Wait(ctx); assert.Error(t, err) {
			assert.ErrorIs(t, err, context.DeadlineExceeded)
		}
	})
}
