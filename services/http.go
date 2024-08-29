package services

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/homier/appetizer/log"
)

var (
	DefaultAdress              = "127.0.0.1:9000"
	DefaultGracefulStopEnabled = true
	DefaultGracefulStopTimeout = time.Second * 5
)

type Muxer = http.Handler
type Handler struct {
	Path    string
	Handler http.HandlerFunc
}

type ServerFactory func(config HTTPServerConfig, handlers []Handler, muxers ...Muxer) *http.Server

func DefaultServerFactory(config HTTPServerConfig, handlers []Handler, muxers ...Muxer) *http.Server {
	srv := &http.Server{
		Addr:              config.Address,
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.WriteTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,
		Handler:           NewMuxer(config.BaseURL, handlers, muxers...),
	}

	return srv
}

func NewMuxer(uri string, handlers []Handler, muxers ...Muxer) *http.ServeMux {
	if uri == "" {
		uri = "/"
	}

	root := http.NewServeMux()
	for _, handler := range handlers {
		root.HandleFunc(handler.Path, handler.Handler)
	}

	for _, muxer := range muxers {
		if uri == "/" {
			root.Handle(uri, muxer)
		} else {
			root.Handle(uri, http.StripPrefix(uri, muxer))
		}
	}

	return root
}

type HTTPServerConfig struct {
	Address string `json:"address" default:"127.0.0.1:9000"`
	BaseURL string `json:"base_url" default:"/"`

	ReadTimeout       time.Duration `json:"read_timeout" default:"0s"`
	ReadHeaderTimeout time.Duration `json:"read_header_timeout" default:"1s"`
	WriteTimeout      time.Duration `json:"write_timeout" default:"0s"`
	IdleTimeout       time.Duration `json:"idle_timeout" default:"0s"`

	MaxHeaderBytes int `json:"max_header_bytes" default:"0"`

	// If enabled, you must configure TLS for server by yourself using
	// a ServerFactory function
	TLSEnabled bool `json:"tls_enabled" default:"false"`
}

type HTTPServer struct {
	Config HTTPServerConfig

	Handlers      []Handler
	Muxers        []Muxer
	ServerFactory ServerFactory

	GracefulStopEnabled bool
	GracefulStopTimeout time.Duration

	PprofEnabled   bool
	PprofURIPrefix string

	server *http.Server

	log log.Logger
	mu  sync.Mutex
}

func (hs *HTTPServer) Init(log log.Logger) error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	hs.log = log

	factory := DefaultServerFactory
	if hs.ServerFactory != nil {
		factory = hs.ServerFactory
	}

	muxers := hs.Muxers
	if hs.PprofEnabled {
		muxers = append([]Muxer{PprofMuxer(hs.PprofURIPrefix)}, muxers...)
	}

	if hs.Config.Address == "" {
		hs.Config.Address = DefaultAdress
	}

	hs.server = factory(hs.Config, hs.Handlers, muxers...)
	return nil
}

func (hs *HTTPServer) Run(ctx context.Context) error {
	runCh := hs.runServer()

	select {
	case err := <-runCh:
		return err
	case <-ctx.Done():
		if !hs.GracefulStopEnabled {
			return hs.gracefulStop()
		}

		return hs.forceStop()
	}
}

func (hs *HTTPServer) runServer() <-chan error {
	ch := make(chan error, 1)

	hs.mu.Lock()
	server := hs.server
	hs.mu.Unlock()

	if server == nil {
		ch <- errors.Wrap(http.ErrServerClosed, "HTTP server is not initialized")
		close(ch)

		return ch
	}

	go func(server *http.Server) {
		defer close(ch)

		ch <- server.ListenAndServe()
	}(server)

	hs.log.Info().Msgf("Listening on %s", hs.server.Addr)
	return ch
}

func (hs *HTTPServer) gracefulStop() error {
	timeout := hs.GracefulStopTimeout
	if timeout <= time.Duration(0) {
		timeout = DefaultGracefulStopTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	hs.mu.Lock()
	if hs.server == nil {
		hs.mu.Unlock()
		return nil
	}

	server := hs.server
	hs.mu.Unlock()

	err := server.Shutdown(ctx)
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return errors.Wrap(err, "failed gracefully stop HTTP server")
}

func (hs *HTTPServer) forceStop() error {
	hs.mu.Lock()
	if hs.server == nil {
		hs.mu.Unlock()
		return nil
	}

	server := hs.server
	hs.mu.Unlock()

	err := server.Close()
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return errors.Wrap(err, "failed stop HTTP server")
}
