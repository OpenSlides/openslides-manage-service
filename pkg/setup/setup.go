package setup

import (
	_ "embed" // Blank import required to use go directive.
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

const ymlFileName = "docker-compose.yml"

//go:embed default-docker-compose.yml
var defaultDockerComposeYml []byte

// Setup creates YAML file for Docker Compose or Docker Swarm and secrets directory.
func Setup(d string) error {
	if err := createYMLFile(d); err != nil {
		return fmt.Errorf("creating YML file at %s: %w", d, err)
	}
	return nil
}

func createYMLFile(d string) error {
	p := path.Join(d, ymlFileName)

	pExists, err := fileExists(p)
	if err != nil {
		return fmt.Errorf("checking file existance: %w", err)
	}
	if pExists {
		// File already exists, so skip this step.
		return nil
	}

	w, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("creating file `%s`: %w", p, err)
	}
	defer w.Close()

	if err := writeContent(w); err != nil {
		return fmt.Errorf("writing content to file %s: %w", p, err)
	}
	return nil
}

func writeContent(w io.Writer) error {
	if _, err := w.Write(defaultDockerComposeYml); err != nil {
		return fmt.Errorf("writing content: %w", err)
	}
	return nil
}

func fileExists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("checking existance of file %s: %w", p, err)
}
