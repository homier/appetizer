package log

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var outStream io.Writer = os.Stderr
var writer = zerolog.ConsoleWriter{
	Out:        outStream,
	TimeFormat: time.RFC3339Nano,
	NoColor:    false,
}

var log = zerolog.New(writer).
	With().
	Timestamp().
	Logger().
	Level(zerolog.InfoLevel)

type ContextualField struct {
	Name  string
	Value string
}

type Logger = zerolog.Logger

func Setup(debug bool, fields ...ContextualField) Logger {
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
