package log

import (
	"io"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestEnable(t *testing.T) {
	old := outStream
	defer func() {
		outStream = old
	}()

	Enable()

	assert.Equal(t, outStream, os.Stderr)
}

func TestDisable(t *testing.T) {
	old := outStream
	defer func() {
		outStream = old
	}()

	Disable()
	assert.Equal(t, outStream, io.Discard)
}

func TestSetup(t *testing.T) {
	log := Setup(true, ContextualField{Name: "app", Value: t.Name()})
	assert.Equal(t, zerolog.DebugLevel, log.GetLevel())
}
