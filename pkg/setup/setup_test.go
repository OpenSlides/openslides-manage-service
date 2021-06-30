package setup_test

import (
	"errors"
	"os"
	"path"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
)

func TestSetup(t *testing.T) {
	testDir, err := os.MkdirTemp("", "openslides-manage-service-")
	if err != nil {
		t.Error("generating temporary directory failed")
	}
	defer os.RemoveAll(testDir)

	t.Run("running setup.Setup() and create all stuff in tmp directory", func(t *testing.T) {
		if err := setup.Setup(testDir); err != nil {
			t.Errorf("Setup returned error %w, expected nil", err)
		}
		testDockerComposeYML(t, testDir)
	})

}

func testDockerComposeYML(t testing.TB, dir string) {
	t.Helper()

	dcYml := path.Join(dir, "docker-compose.yml")
	if _, err := os.Stat(dcYml); errors.Is(err, os.ErrNotExist) {
		t.Errorf("file %s does not exist, expected existance", dcYml)
	}
	dcYmlContent, err := os.ReadFile(dcYml)
	if err != nil {
		t.Errorf("reading file %s: %w", dcYml, err)
	}

	got := string(dcYmlContent[:])
	expected := defaultDockerComposeYml
	if got != expected {
		t.Errorf("checking content of YML file, expected %q, got %q", expected, got)
	}
}

const defaultDockerComposeYml = `---
foo: bar
`
