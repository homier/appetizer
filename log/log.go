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

func Setup(debug bool, fields ...ContextualField) Logger {
	mu.Lock()
	defer mu.Unlock()

	return EnrichLogger(log, debug, fields...)
}

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

func Enable() {
	mu.Lock()
	defer mu.Unlock()

	outStream = os.Stderr
	log = log.Output(outStream)
}

func Disable() {
	mu.Lock()
	defer mu.Unlock()

	outStream = io.Discard
	log = log.Output(io.Discard)
}
