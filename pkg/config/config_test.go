package config_test

import (
	"errors"
	"os"
	"path"
	"strings"
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

	t.Run("executing setup.CmdCreateDefault() with existing directory", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "openslides-manage-service-")
		if err != nil {
			t.Fatalf("generating temporary directory failed: %v", err)
		}
		defer os.RemoveAll(testDir)

		cmd := config.CmdCreateDefault()
		cmd.SetArgs([]string{testDir})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("executing config-create-default subcommand: %v", err)
		}

		p := path.Join(testDir, "config.yml")
		if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
			t.Fatalf("file %q does not exist, expected existance", p)
		}
	})
}

func TestConfigWithCustomConfig(t *testing.T) {
	testDir, err := os.MkdirTemp("", "openslides-manage-service-")
	if err != nil {
		t.Fatalf("generating temporary directory failed: %v", err)
	}
	defer os.RemoveAll(testDir)

	t.Run("running config.Config() using a custom config twice 1", func(t *testing.T) {
		customConfig1 := `---
defaults:
  containerRegistry: example.com/test_Waetai0ohf
`
		customConfig2 := `---
defaults:
  containerRegistry: example.com/test_Aeghies3me
`

		c := make([][]byte, 2)
		c[0] = []byte(customConfig1)
		c[1] = []byte(customConfig2)
		if err := config.Config(testDir, nil, c); err != nil {
			t.Fatalf("running config.Config() failed with error: %v", err)
		}
		testFileContains(t, testDir, "docker-compose.yml", "image: example.com/test_Aeghies3me/openslides-proxy:latest")
	})

	t.Run("running config.Config() using a custom config twice 2", func(t *testing.T) {
		customConfig1 := `---
disablePostgres: false
`
		customConfig2 := `---
disablePostgres: true
`
		c := make([][]byte, 2)
		c[0] = []byte(customConfig1)
		c[1] = []byte(customConfig2)
		if err := config.Config(testDir, nil, c); err != nil {
			t.Fatalf("running config.Config() failed with error: %v", err)
		}
		testFileNotContains(t, testDir, "docker-compose.yml", "image: postgres:11")
	})

	t.Run("running config.Config() using a custom config twice 3", func(t *testing.T) {
		customConfig1 := `---
disablePostgres: true
`
		customConfig2 := `---
disablePostgres: false
`
		c := make([][]byte, 2)
		c[0] = []byte(customConfig1)
		c[1] = []byte(customConfig2)
		if err := config.Config(testDir, nil, c); err != nil {
			t.Fatalf("running config.Config() failed with error: %v", err)
		}
		testFileContains(t, testDir, "docker-compose.yml", "image: postgres:11")
	})
}

func testFileContains(t testing.TB, dir, name, exp string) {
	t.Helper()
	p := path.Join(dir, name)
	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file %q does not exist, expected existance", p)
	}
	content, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("error reading file %q: %v", p, err)
	}
	got := string(content)
	if !strings.Contains(got, exp) {
		t.Fatalf("wrong content of file %q, which should contain %q", p, exp)
	}
}

func testFileNotContains(t testing.TB, dir, name, exp string) {
	t.Helper()
	p := path.Join(dir, name)
	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file %q does not exist, expected existance", p)
	}
	content, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("error reading file %q: %v", p, err)
	}
	got := string(content)
	if strings.Contains(got, exp) {
		t.Fatalf("wrong content of file %q, which should not contain %q", p, exp)
	}
}
