package initialdata_test

import (
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/initialdata"
)

func TestCmd(t *testing.T) {
	t.Run("executing initialdata.Cmd() without flags so using default initial data", func(t *testing.T) {
		cmd := initialdata.Cmd()
		if err := cmd.Execute(); err != nil {
			t.Fatalf("executing initial-data subcommand: %v", err)
		}
	})
}
