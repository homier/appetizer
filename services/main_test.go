package services

import (
	"os"
	"testing"

	"github.com/homier/appetizer/log"
)

func TestMain(m *testing.M) {
	log.Disable()
	defer log.Enable()

	os.Exit(m.Run())
}
