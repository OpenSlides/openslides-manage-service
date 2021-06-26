package setup_test

import (
	"errors"
	"os"
	"path"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
)

func TestSetup(t *testing.T) {
	d, err := os.MkdirTemp("", "openslides-manage-service-")
	if err != nil {
		t.Error("generating temporary directory failed")
	}
	defer os.RemoveAll(d)

	if err := setup.Setup(d); err != nil {
		t.Errorf("Setup returned error %w, expected nil", err)
	}

	testDockerComposeYML(t, d)

}

func testDockerComposeYML(t *testing.T, d string) {
	dcYml := path.Join(d, "docker-compose.yml")
	if _, err := os.Stat(dcYml); errors.Is(err, os.ErrNotExist) {
		t.Errorf("file %s does not exist, expected existance", dcYml)
	}
	dcYmlContent, err := os.ReadFile(dcYml)
	if err != nil {
		t.Errorf("reading file %s: %w", dcYml, err)
	}
	got := string(dcYmlContent[:])
	if got != defaultDockerComposeYml {
		t.Errorf("checking content of YML file, expected `%s`, got `%s`", defaultDockerComposeYml, got)
	}
}

const defaultDockerComposeYml = `---
foo: bar
`
