package fehler_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/fehler"
)

func TestExitCode(t *testing.T) {
	myErr := errors.New("some error")
	errExitCode := fehler.ExitCode(42, myErr)
	err := fmt.Errorf("got error: %w", errExitCode)

	t.Run("Find exit code error", func(t *testing.T) {
		var errExit interface {
			ExitCode() int
		}
		if !errors.As(err, &errExit) {
			t.Errorf("unwrapping error did not return exit code error, got: %v", err)
		}
	})
}
