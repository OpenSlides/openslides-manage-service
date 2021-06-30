package setup

import (
	_ "embed" // Blank import required to use go directive.
	"errors"
	"fmt"
	"os"
	"path"
)

const (
	ymlFileName = "docker-compose.yml"
	envFileName = ".env"
)

//go:embed default-docker-compose.yml
var defaultDockerComposeYml []byte

//go:embed default-environment.env
var defaultEnvFile []byte

// Setup creates YAML file for Docker Compose or Docker Swarm with .env file and secrets directory.
func Setup(dir string) error {
	if err := createYMLFile(dir); err != nil {
		return fmt.Errorf("creating YAML file at %q: %w", dir, err)
	}
	if err := createEnvFile(dir); err != nil {
		return fmt.Errorf("creating .env file at %q: %w", dir, err)
	}
	return nil
}

func createYMLFile(dir string) error {
	if err := createFile(dir, ymlFileName, defaultDockerComposeYml); err != nil {
		return fmt.Errorf("creating YAML file at %q: %w", dir, err)
	}
	return nil
}

func createEnvFile(dir string) error {
	if err := createFile(dir, envFileName, defaultEnvFile); err != nil {
		return fmt.Errorf("creating YAML file at %q: %w", dir, err)
	}
	return nil
}

func createFile(dir string, name string, content []byte) error {
	p := path.Join(dir, name)

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
		return fmt.Errorf("creating file %q: %w", p, err)
	}
	defer w.Close()

	if _, err := w.Write(content); err != nil {
		return fmt.Errorf("writing content to file %q: %w", p, err)
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
