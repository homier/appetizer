package appetizer

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestNotifyContext(t *testing.T) {
	tests := []struct {
		name    string
		signals []os.Signal
	}{
		{
			name:    "default signals",
			signals: []os.Signal{},
		},
		{
			name: "custom signals",
			signals: []os.Signal{
				syscall.SIGQUIT,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := NotifyContext(tt.signals...)
			defer cancel()

			p, err := os.FindProcess(os.Getpid())
			if err != nil {
				t.Fatal(err)
			}

			if len(tt.signals) > 0 {
				if err := p.Signal(tt.signals[0]); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := p.Signal(SignalsDefault[0]); err != nil {
					t.Fatal(err)
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				t.Fatal("context has not been cancelled by specified signal")
			}
		})
	}
}
