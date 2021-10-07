package config_test

import (
	"errors"
	"os"
	"path"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/config"
)

func TestCmd(t *testing.T) {
	t.Run("executing setup.Cmd() with existing directory", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "openslides-manage-service-")
		if err != nil {
			t.Fatalf("generating temporary directory failed: %v", err)
		}
		defer os.RemoveAll(testDir)

		cmd := config.Cmd()
		cmd.SetArgs([]string{testDir})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("executing config subcommand: %v", err)
		}

		p := path.Join(testDir, "docker-compose.yml")
		if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
			t.Fatalf("file %q does not exist, expected existance", p)
		}
	})

}
