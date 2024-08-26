package appetizer

import (
	"os"
	"testing"

	"github.com/homier/appetizer/log"
)

func TestMain(m *testing.M) {
	log.Disable()
	code := m.Run()
	log.Enable()

	os.Exit(code)
}
