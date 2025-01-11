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
	t.Run("executing config.Cmd() with existing directory", func(t *testing.T) {
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

	t.Run("executing config.Cmd() with existing directory with builtin-template flag and template flag", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "openslides-manage-service-")
		if err != nil {
			t.Fatalf("generating temporary directory failed: %v", err)
		}
		defer os.RemoveAll(testDir)

		templateFilePath := path.Join(testDir, "some-template.yml")
		if err := os.WriteFile(templateFilePath, []byte(""), os.ModePerm); err != nil {
			t.Fatalf("writing custom template failed: %v", err)
		}
		cmd := config.Cmd()
		cmd.SetArgs([]string{testDir, "--builtin-template", "kubernetes", "--template", templateFilePath})

		err = cmd.Execute()
		if err == nil {
			t.Fatalf("executing config subcommand: expected error but err is nil")
		}
		errMsg := "flag --builtin-template must not be used together with flag --template"
		if err.Error() != errMsg {
			t.Fatalf("wrong error message, expected %q, got %q", errMsg, err.Error())
		}
	})

	t.Run("executing config.CmdCreateDefault() with existing directory", func(t *testing.T) {
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
	testLibDir, err := os.MkdirTemp("", "openslides-manage-service-lib-")
	if err != nil {
		t.Fatalf("generating temporary directory failed: %v", err)
	}
	defer os.RemoveAll(testLibDir)

	t.Run("running config.Config() using a custom config twice 1", func(t *testing.T) {
		customConfig1 := `---
defaults:
  containerRegistry: example.com/test_Waetai0ohf
`
		customConfig2 := `---
defaults:
  containerRegistry: example.com/test_Aeghies3me
`
		c := make([]string, 2)
		cfgPath1 := path.Join(testLibDir, "ext-conf1")
		if err := os.WriteFile(cfgPath1, []byte(customConfig1), os.ModePerm); err != nil {
			t.Fatalf("writing custom config failed: %v", err)
		}
		c[0] = cfgPath1
		cfgPath2 := path.Join(testLibDir, "ext-conf2")
		if err := os.WriteFile(cfgPath2, []byte(customConfig2), os.ModePerm); err != nil {
			t.Fatalf("writing custom config failed: %v", err)
		}
		c[1] = cfgPath2
		tech := "docker-compose"

		if err := config.Config(testDir, tech, "", c); err != nil {
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
		c := make([]string, 2)
		cfgPath1 := path.Join(testLibDir, "ext-conf1")
		if err := os.WriteFile(cfgPath1, []byte(customConfig1), os.ModePerm); err != nil {
			t.Fatalf("writing custom config failed: %v", err)
		}
		c[0] = cfgPath1
		cfgPath2 := path.Join(testLibDir, "ext-conf2")
		if err := os.WriteFile(cfgPath2, []byte(customConfig2), os.ModePerm); err != nil {
			t.Fatalf("writing custom config failed: %v", err)
		}
		c[1] = cfgPath2
		tech := "docker-compose"

		if err := config.Config(testDir, tech, "", c); err != nil {
			t.Fatalf("running config.Config() failed with error: %v", err)
		}
		testFileNotContains(t, testDir, "docker-compose.yml", "image: postgres:15")
	})

	t.Run("running config.Config() using a custom config twice 3", func(t *testing.T) {
		customConfig1 := `---
disablePostgres: true
`
		customConfig2 := `---
disablePostgres: false
`
		c := make([]string, 2)
		cfgPath1 := path.Join(testLibDir, "ext-conf1")
		if err := os.WriteFile(cfgPath1, []byte(customConfig1), os.ModePerm); err != nil {
			t.Fatalf("writing custom config failed: %v", err)
		}
		c[0] = cfgPath1
		cfgPath2 := path.Join(testLibDir, "ext-conf2")
		if err := os.WriteFile(cfgPath2, []byte(customConfig2), os.ModePerm); err != nil {
			t.Fatalf("writing custom config failed: %v", err)
		}
		c[1] = cfgPath2
		tech := "docker-compose"

		if err := config.Config(testDir, tech, "", c); err != nil {
			t.Fatalf("running config.Config() failed with error: %v", err)
		}
		testFileContains(t, testDir, "docker-compose.yml", "image: postgres:15")
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
