package log

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	outStream io.Writer = os.Stderr
	writer              = zerolog.ConsoleWriter{
		Out:        outStream,
		TimeFormat: time.RFC3339Nano,
		NoColor:    false,
	}

	log = zerolog.New(writer).
		With().
		Timestamp().
		Logger().
		Level(zerolog.InfoLevel)

	mu sync.Mutex
)

type ContextualField struct {
	Name  string
	Value string
}

type Logger = zerolog.Logger

// Returns a copy of default `zerolog.Logger` with specified contextual fields.
// If `debug` is true, a returning logger will be configured to `zerolog.DebugLevel` level.
func Setup(debug bool, fields ...ContextualField) Logger {
	mu.Lock()
	defer mu.Unlock()

	return EnrichLogger(log, debug, fields...)
}

// Returns a copy of provided `log` with additional contextual fields.
// If `debug` is true, a returning logger will be configured to `zerolog.DebugLevel` level.
func EnrichLogger(log Logger, debug bool, fields ...ContextualField) Logger {
	ctx := log.With()

	for _, field := range fields {
		ctx = ctx.Str(field.Name, field.Value)
	}

	logger := ctx.Logger()

	if debug {
		logger = logger.Level(zerolog.DebugLevel)
	}

	return logger
}

// Sets global logging output to `os.Stderr`.
func Enable() {
	mu.Lock()
	defer mu.Unlock()

	outStream = os.Stderr
	log = log.Output(outStream)
}

// Sets global logging output to `io.Discard`, meaning no log messages will be produced.
func Disable() {
	mu.Lock()
	defer mu.Unlock()

	outStream = io.Discard
	log = log.Output(io.Discard)
}
