package appetizer

import (
	"context"
	stdErrors "errors"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/homier/appetizer/retry"
)

func TestApp_init(t *testing.T) {
	tests := []struct {
		name          string
		setupServices func(t *testing.T) []Service
		wantErr       bool
		err           error
	}{
		{
			name: "no services provided",
			setupServices: func(_ *testing.T) []Service {
				return []Service{}
			},
		},
		{
			name: "no errors on services init",
			setupServices: func(t *testing.T) []Service {
				srv1 := NewMockServicer(t)
				srv1.EXPECT().Init(mock.AnythingOfType("zerolog.Logger")).Return(nil).Once()

				srv2 := NewMockServicer(t)
				srv2.EXPECT().Init(mock.AnythingOfType("zerolog.Logger")).Return(nil).Once()

				return []Service{
					{
						Name:     t.Name() + "_srv1",
						Servicer: srv1,
					},
					{
						Name:     t.Name() + "_srv2",
						Servicer: srv2,
					},
				}
			},
		},
		{
			name: "1 out of 2 service failed",
			setupServices: func(t *testing.T) []Service {
				srv1 := NewMockServicer(t)
				srv1.EXPECT().Init(mock.AnythingOfType("zerolog.Logger")).
					Return(errors.New("unexpected error")).Once()

				srv2 := NewMockServicer(t)
				srv2.EXPECT().Init(mock.AnythingOfType("zerolog.Logger")).Return(nil).Once()

				return []Service{
					{
						Name:     t.Name() + "_srv1",
						Servicer: srv1,
					},
					{
						Name:     t.Name() + "_srv2",
						Servicer: srv2,
					},
				}
			},
			wantErr: true,
			err:     errors.New("unexpected error"),
		},
		{
			name: "both services failed",
			setupServices: func(t *testing.T) []Service {
				srv1 := NewMockServicer(t)
				srv1.EXPECT().Init(mock.AnythingOfType("zerolog.Logger")).
					Return(errors.New("unexpected error1")).Once()

				srv2 := NewMockServicer(t)
				srv2.EXPECT().Init(mock.AnythingOfType("zerolog.Logger")).
					Return(errors.New("unexpected error2")).Once()

				return []Service{
					{
						Name:     t.Name() + "_srv1",
						Servicer: srv1,
					},
					{
						Name:     t.Name() + "_srv2",
						Servicer: srv2,
					},
				}
			},
			wantErr: true,
			err:     stdErrors.Join(errors.New("unexpected error1"), errors.New("unexpected error2")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				Name:     t.Name(),
				Services: tt.setupServices(t),
				Debug:    true,
			}

			err := app.init()
			if tt.wantErr {
				if assert.Error(t, err) {
					assert.ErrorContains(t, err, tt.err.Error())
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestApp_runService(t *testing.T) {
	app := &App{
		Name: t.Name(),
	}

	ErrCritical := errors.New("some very critical error")

	tests := []struct {
		name         string
		setupCtx     func() (context.Context, context.CancelFunc)
		setupService func(t *testing.T) Service
		wantErr      bool
		err          error
	}{
		{
			name: "service with no error returned",
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			setupService: func(t *testing.T) Service {
				srv := NewMockServicer(t)
				srv.EXPECT().Run(mock.AnythingOfType("*context.cancelCtx")).
					Return(nil).Once()

				return Service{
					Name:     t.Name(),
					Servicer: srv,
				}
			},
		},
		{
			name: "service with error and no restart",
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			setupService: func(t *testing.T) Service {
				srv := NewMockServicer(t)
				srv.EXPECT().Run(mock.AnythingOfType("*context.cancelCtx")).
					Return(errors.New("unexpected error")).Once()

				return Service{
					Name:     "failed service",
					Servicer: srv,
				}
			},
			wantErr: true,
			err:     errors.Wrap(errors.New("unexpected error"), "service 'failed service' crashed"),
		},
		{
			name: "service with error and not configured enabled restart",
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			setupService: func(t *testing.T) Service {
				srv := NewMockServicer(t)
				srv.EXPECT().Run(mock.AnythingOfType("*context.cancelCtx")).
					Return(errors.New("unexpected error")).Once()

				return Service{
					Name:           "failed service",
					Servicer:       srv,
					RestartEnabled: true,
				}
			},
			wantErr: true,
			err:     errors.Wrap(errors.New("unexpected error"), "service 'failed service' crashed"),
		},
		{
			name: "service with critical error and restart",
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			setupService: func(t *testing.T) Service {
				srv := NewMockServicer(t)
				srv.EXPECT().Run(mock.AnythingOfType("*context.cancelCtx")).
					Return(ErrCritical).Once()

				return Service{
					Name:           "failed service",
					Servicer:       srv,
					RestartEnabled: true,
					RestartOpts: retry.Opts{
						Opts: &backoff.ExponentialBackOff{
							InitialInterval: time.Millisecond,
							MaxInterval:     time.Millisecond * 5,
							MaxElapsedTime:  time.Millisecond * 10,
							Clock:           backoff.SystemClock,
						},
						CriticalError: ErrCritical,
					},
				}
			},
			wantErr: true,
			err:     errors.Wrap(ErrCritical, "service 'failed service' crashed"),
		},
		{
			name: "service with configured critical error and restart",
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			setupService: func(t *testing.T) Service {
				srv := NewMockServicer(t)
				srv.EXPECT().Run(mock.AnythingOfType("*context.cancelCtx")).
					Return(errors.New("something")).Twice()

				return Service{
					Name:           "failed service",
					Servicer:       srv,
					RestartEnabled: true,
					RestartOpts: retry.Opts{
						Opts: &backoff.ExponentialBackOff{
							InitialInterval: time.Millisecond,
							MaxInterval:     time.Millisecond * 5,
							MaxElapsedTime:  time.Millisecond * 10,
							Clock:           backoff.SystemClock,
						},
						CriticalError: ErrCritical,
						MaxRetry:      1,
					},
				}
			},
			wantErr: true,
			err:     errors.Wrap(errors.New("something"), "service 'failed service' crashed"),
		},
		{
			name: "service with error and restart",
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			setupService: func(t *testing.T) Service {
				srv := NewMockServicer(t)
				srv.EXPECT().Run(mock.AnythingOfType("*context.cancelCtx")).
					Return(errors.New("unexpected error")).Twice()

				return Service{
					Name:           "failed service",
					Servicer:       srv,
					RestartEnabled: true,
					RestartOpts: retry.Opts{
						Opts: &backoff.ExponentialBackOff{
							InitialInterval: time.Millisecond,
							MaxInterval:     time.Millisecond * 5,
							MaxElapsedTime:  time.Millisecond * 10,
							Clock:           backoff.SystemClock,
						},
						MaxRetry: 1,
					},
				}
			},
			wantErr: true,
			err:     errors.Wrap(errors.New("unexpected error"), "service 'failed service' crashed"),
		},
		{
			name: "service without error and restart",
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			setupService: func(t *testing.T) Service {
				srv := NewMockServicer(t)
				srv.EXPECT().Run(mock.AnythingOfType("*context.cancelCtx")).Return(nil).Once()

				return Service{
					Name:           "ok service",
					Servicer:       srv,
					RestartEnabled: true,
					RestartOpts: retry.Opts{
						Opts: &backoff.ExponentialBackOff{
							InitialInterval: time.Millisecond,
							MaxInterval:     time.Millisecond * 5,
							MaxElapsedTime:  time.Millisecond * 10,
							Clock:           backoff.SystemClock,
						},
						MaxRetry: 1,
					},
				}
			},
		},
		{
			name: "service with error and cancelled context",
			setupCtx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx, cancel
			},
			setupService: func(t *testing.T) Service {
				srv := NewMockServicer(t)
				srv.EXPECT().Run(mock.AnythingOfType("*context.cancelCtx")).
					Return(errors.New("unexpected error")).Once()

				return Service{
					Name:           "failed service",
					Servicer:       srv,
					RestartEnabled: true,
					RestartOpts: retry.Opts{
						Opts: &backoff.ExponentialBackOff{
							InitialInterval: time.Millisecond,
							MaxInterval:     time.Millisecond * 5,
							MaxElapsedTime:  time.Millisecond * 10,
							Clock:           backoff.SystemClock,
						},
						MaxRetry: 1,
					},
				}
			},
			wantErr: true,
			err:     errors.Wrap(context.Canceled, "service 'failed service' crashed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.setupCtx()
			defer cancel()

			service := tt.setupService(t)
			err := app.runService(ctx, &service)

			if !tt.wantErr {
				assert.NoError(t, err)
				return
			}

			if assert.Error(t, err) {
				assert.ErrorContains(t, err, tt.err.Error())
			}
		})
	}
}

func TestApp_Run(t *testing.T) {
	defaultContextFunc := func() (context.Context, context.CancelFunc) {
		return context.WithCancel(context.Background())
	}
	cancelledContextFunc := func() (context.Context, context.CancelFunc) {
		ctx, cancel := defaultContextFunc()
		cancel()

		return ctx, cancel
	}

	tests := []struct {
		name          string
		setupCtx      func() (context.Context, context.CancelFunc)
		setupServices func(t *testing.T) []Service
		wantErr       bool
		err           error
	}{
		{
			name:     "no services",
			setupCtx: defaultContextFunc,
			setupServices: func(_ *testing.T) []Service {
				return nil
			},
			wantErr: false,
		},
		{
			name:     "context cancelled",
			setupCtx: cancelledContextFunc,
			setupServices: func(t *testing.T) []Service {
				srv := NewMockServicer(t)
				srv.EXPECT().Init(mock.AnythingOfType("zerolog.Logger")).Return(nil).Once()
				srv.EXPECT().Run(mock.AnythingOfType("*context.cancelCtx")).
					RunAndReturn(func(ctx context.Context) error {
						<-ctx.Done()
						return ctx.Err()
					})

				return []Service{{
					Name:           "srv1",
					Servicer:       srv,
					RestartEnabled: false,
				}}
			},
			wantErr: true,
			err:     context.Canceled,
		},
		{
			name:     "init failed",
			setupCtx: defaultContextFunc,
			setupServices: func(t *testing.T) []Service {
				srv1 := NewMockServicer(t)
				srv1.EXPECT().Init(mock.AnythingOfType("zerolog.Logger")).
					Return(errors.New("init failed")).Once()

				return []Service{
					{
						Name:           "srv1",
						Servicer:       srv1,
						RestartEnabled: false,
					},
				}
			},
			wantErr: true,
			err:     errors.New("init failed"),
		},
		{
			name:     "one service failed",
			setupCtx: defaultContextFunc,
			setupServices: func(t *testing.T) []Service {
				srv1 := NewMockServicer(t)
				srv1.EXPECT().Init(mock.AnythingOfType("zerolog.Logger")).Return(nil).Once()
				srv1.EXPECT().Run(mock.AnythingOfType("*context.cancelCtx")).
					RunAndReturn(func(ctx context.Context) error {
						<-ctx.Done()
						return nil
					})

				srv2 := NewMockServicer(t)
				srv2.EXPECT().Init(mock.AnythingOfType("zerolog.Logger")).Return(nil).Once()
				srv2.EXPECT().Run(mock.AnythingOfType("*context.cancelCtx")).
					Return(errors.New("unexpected error from srv2")).Once()

				return []Service{
					{
						Name:           "srv1",
						Servicer:       srv1,
						RestartEnabled: false,
					},
					{
						Name:           "srv2",
						Servicer:       srv2,
						RestartEnabled: false,
					},
				}
			},
			wantErr: true,
			err:     errors.Wrap(errors.New("unexpected error from srv2"), "service 'srv2' crashed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.setupCtx()
			defer cancel()

			app := &App{
				Name:     t.Name(),
				Services: tt.setupServices(t),
			}

			err := app.Run(ctx)
			if !tt.wantErr {
				assert.NoError(t, err)
				return
			}

			if assert.Error(t, err) {
				assert.ErrorContains(t, err, tt.err.Error())
			}
		})
	}
}
