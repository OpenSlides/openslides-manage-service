package setup_test

import (
	"errors"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
)

func TestCmd(t *testing.T) {
	t.Run("executing setup.Cmd() with new directory", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "openslides-manage-service-")
		if err != nil {
			t.Fatalf("generating temporary directory failed: %v", err)
		}
		defer os.RemoveAll(testDir)

		cmd := setup.Cmd()
		cmd.SetArgs([]string{testDir})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("executing setup subcommand: %v", err)
		}

		secDir := path.Join(testDir, setup.SecretsDirName)
		testContentFile(t, testDir, "docker-compose.yml", defaultDockerComposeYml)
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
		testKeyFile(t, secDir, "manage_auth_password")
		testPasswordFile(t, secDir, "datastore_postgres_password")
		testPasswordFile(t, secDir, "media_postgres_password")
		testContentFile(t, secDir, setup.SuperadminFileName, setup.DefaultSuperadminPassword)
		testDirectory(t, testDir, "db-data")
	})

	t.Run("executing setup.Cmd() with new directory with --force flag", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "openslides-manage-service-")
		if err != nil {
			t.Fatalf("generating temporary directory failed: %v", err)
		}
		defer os.RemoveAll(testDir)

		cmd := setup.Cmd()
		cmd.SetArgs([]string{testDir, "--force"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("executing setup subcommand: %v", err)
		}

		secDir := path.Join(testDir, setup.SecretsDirName)
		testContentFile(t, testDir, "docker-compose.yml", defaultDockerComposeYml)
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
		testKeyFile(t, secDir, "manage_auth_password")
		testPasswordFile(t, secDir, "datastore_postgres_password")
		testPasswordFile(t, secDir, "media_postgres_password")
		testContentFile(t, secDir, setup.SuperadminFileName, setup.DefaultSuperadminPassword)
		testDirectory(t, testDir, "db-data")
	})

	t.Run("executing setup.Cmd() with new directory with --template flag", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "openslides-manage-service-")
		if err != nil {
			t.Fatalf("generating temporary directory failed: %v", err)
		}
		defer os.RemoveAll(testDir)

		customTplFileName := path.Join(testDir, "custom-template.yaml.tpl")
		f, err := os.OpenFile(customTplFileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			t.Fatalf("creating custom template file: %v", err)
		}
		defer f.Close()
		customTpl := "custom template Oht2oph9qu"
		if _, err := f.WriteString(customTpl); err != nil {
			t.Fatalf("writing custom template to file %q: %v", customTplFileName, err)
		}

		cmd := setup.Cmd()
		cmd.SetArgs([]string{testDir, "--template", customTplFileName})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("executing setup subcommand: %v", err)
		}

		secDir := path.Join(testDir, setup.SecretsDirName)
		testContentFile(t, testDir, "docker-compose.yml", customTpl)
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
		testKeyFile(t, secDir, "manage_auth_password")
		testPasswordFile(t, secDir, "datastore_postgres_password")
		testPasswordFile(t, secDir, "media_postgres_password")
		testContentFile(t, secDir, setup.SuperadminFileName, setup.DefaultSuperadminPassword)
		testDirectory(t, testDir, "db-data")
	})

	t.Run("executing setup.Cmd() with new directory with --config flag", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "openslides-manage-service-")
		if err != nil {
			t.Fatalf("generating temporary directory failed: %v", err)
		}
		defer os.RemoveAll(testDir)

		customConfigFileName := path.Join(testDir, "custom-config.yml")
		f, err := os.OpenFile(customConfigFileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			t.Fatalf("creating custom config file: %v", err)
		}
		defer f.Close()
		customConfigContent := `---
defaults:
  containerRegistry: example.com/test_fahNae5i
services:
  backend:
    tag: 2.0.1
`
		if _, err := f.WriteString(customConfigContent); err != nil {
			t.Fatalf("writing custom config to file %q: %v", customConfigFileName, err)
		}

		cmd := setup.Cmd()
		cmd.SetArgs([]string{testDir, "--config", customConfigFileName})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("executing setup subcommand: %v", err)
		}

		secDir := path.Join(testDir, setup.SecretsDirName)
		testFileContains(t, testDir, "docker-compose.yml", "image: example.com/test_fahNae5i/openslides-proxy:latest")
		testFileContains(t, testDir, "docker-compose.yml", "image: example.com/test_fahNae5i/openslides-backend:2.0.1")
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
		testKeyFile(t, secDir, "manage_auth_password")
		testPasswordFile(t, secDir, "datastore_postgres_password")
		testPasswordFile(t, secDir, "media_postgres_password")
		testContentFile(t, secDir, setup.SuperadminFileName, setup.DefaultSuperadminPassword)
		testDirectory(t, testDir, "db-data")
	})

	t.Run("executing setup.Cmd() with new directory with --config flag twice", func(t *testing.T) {
		testDir, err := os.MkdirTemp("", "openslides-manage-service-")
		if err != nil {
			t.Fatalf("generating temporary directory failed: %v", err)
		}
		defer os.RemoveAll(testDir)

		customConfigFileName := path.Join(testDir, "custom-config.yml")
		f, err := os.OpenFile(customConfigFileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			t.Fatalf("creating custom config file: %v", err)
		}
		defer f.Close()
		customConfigContent := `---
defaults:
  containerRegistry: example.com/test_Ohm7uafo
  tag: wrong_tag_to_be_overridden
`
		if _, err := f.WriteString(customConfigContent); err != nil {
			t.Fatalf("writing custom config to file %q: %v", customConfigFileName, err)
		}

		customConfigFileName2 := path.Join(testDir, "custom-config-2.yml")
		f2, err := os.OpenFile(customConfigFileName2, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			t.Fatalf("creating custom config file: %v", err)
		}
		defer f2.Close()
		customConfigContent2 := `---
defaults:
  tag: test_Ra9va3ie
`
		if _, err := f2.WriteString(customConfigContent2); err != nil {
			t.Fatalf("writing custom config to file %q: %v", customConfigFileName2, err)
		}

		cmd := setup.Cmd()
		cmd.SetArgs([]string{testDir, "--config", customConfigFileName, "--config", customConfigFileName2})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("executing setup subcommand: %v", err)
		}

		secDir := path.Join(testDir, setup.SecretsDirName)
		testFileContains(t, testDir, "docker-compose.yml", "image: example.com/test_Ohm7uafo/openslides-proxy:test_Ra9va3ie")
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
		testKeyFile(t, secDir, "manage_auth_password")
		testPasswordFile(t, secDir, "datastore_postgres_password")
		testPasswordFile(t, secDir, "media_postgres_password")
		testContentFile(t, secDir, setup.SuperadminFileName, setup.DefaultSuperadminPassword)
		testDirectory(t, testDir, "db-data")
	})

}

func TestSetupCommon(t *testing.T) {
	testDir, err := os.MkdirTemp("", "openslides-manage-service-")
	if err != nil {
		t.Fatalf("generating temporary directory failed: %v", err)
	}
	defer os.RemoveAll(testDir)

	t.Run("running setup.Setup() and create all stuff in tmp directory", func(t *testing.T) {
		if err := setup.Setup(testDir, false, nil, nil); err != nil {
			t.Fatalf("running Setup() failed with error: %v", err)
		}
		secDir := path.Join(testDir, setup.SecretsDirName)
		testContentFile(t, testDir, "docker-compose.yml", defaultDockerComposeYml)
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
		testKeyFile(t, secDir, "manage_auth_password")
		testPasswordFile(t, secDir, "datastore_postgres_password")
		testPasswordFile(t, secDir, "media_postgres_password")
		testContentFile(t, secDir, setup.SuperadminFileName, setup.DefaultSuperadminPassword)
		testDirectory(t, testDir, "db-data")
	})

	t.Run("running setup.Setup() twice without changing existant files", func(t *testing.T) {
		p := path.Join(testDir, "docker-compose.yml")
		testContent := "test-content-of-a-file"
		f, err := os.OpenFile(p, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
		if err != nil {
			t.Fatalf("opening file %q: %v", p, err)
		}
		if _, err := f.WriteString(testContent); err != nil {
			t.Fatalf("writing to file %q: %v", p, err)
		}

		if err := setup.Setup(testDir, false, nil, nil); err != nil {
			t.Fatalf("running Setup() failed with error: %v", err)
		}
		secDir := path.Join(testDir, setup.SecretsDirName)
		testContentFile(t, testDir, "docker-compose.yml", testContent)
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
		testKeyFile(t, secDir, "manage_auth_password")
		testPasswordFile(t, secDir, "datastore_postgres_password")
		testPasswordFile(t, secDir, "media_postgres_password")
		testContentFile(t, secDir, setup.SuperadminFileName, setup.DefaultSuperadminPassword)
		testDirectory(t, testDir, "db-data")
	})

	t.Run("running setup.Setup() with force flag with changing existant files", func(t *testing.T) {
		if err := setup.Setup(testDir, true, nil, nil); err != nil {
			t.Fatalf("running Setup() failed with error: %v", err)
		}
		secDir := path.Join(testDir, setup.SecretsDirName)
		testContentFile(t, testDir, "docker-compose.yml", defaultDockerComposeYml)
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
		testKeyFile(t, secDir, "manage_auth_password")
		testPasswordFile(t, secDir, "datastore_postgres_password")
		testPasswordFile(t, secDir, "media_postgres_password")
		testContentFile(t, secDir, setup.SuperadminFileName, setup.DefaultSuperadminPassword)
		testDirectory(t, testDir, "db-data")
	})
}

func TestSetupNonExistingSubdirectory(t *testing.T) {
	testDir, err := os.MkdirTemp("", "openslides-manage-service-")
	if err != nil {
		t.Fatalf("generating temporary directory failed: %v", err)
	}
	defer os.RemoveAll(testDir)

	t.Run("running setup.Setup() and give a previously not existing subdirectory", func(t *testing.T) {
		dir := path.Join(testDir, "new_directory")
		if err := setup.Setup(dir, false, nil, nil); err != nil {
			t.Fatalf("running Setup() failed with error: %v", err)
		}
		secDir := path.Join(dir, setup.SecretsDirName)
		testContentFile(t, dir, "docker-compose.yml", defaultDockerComposeYml)
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
		testKeyFile(t, secDir, "manage_auth_password")
		testPasswordFile(t, secDir, "datastore_postgres_password")
		testPasswordFile(t, secDir, "media_postgres_password")
		testContentFile(t, secDir, setup.SuperadminFileName, setup.DefaultSuperadminPassword)
		testDirectory(t, dir, "db-data")
	})
}

func TestSetupExternalTemplate(t *testing.T) {
	testDir, err := os.MkdirTemp("", "openslides-manage-service-")
	if err != nil {
		t.Fatalf("generating temporary directory failed: %v", err)
	}
	defer os.RemoveAll(testDir)

	t.Run("running setup.Setup() and give an external template", func(t *testing.T) {
		tplText := "test-from-external-template"
		if err := setup.Setup(testDir, false, []byte(tplText), nil); err != nil {
			t.Fatalf("running Setup() failed with error: %v", err)
		}
		secDir := path.Join(testDir, setup.SecretsDirName)
		testContentFile(t, testDir, "docker-compose.yml", tplText)
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
		testKeyFile(t, secDir, "manage_auth_password")
		testPasswordFile(t, secDir, "datastore_postgres_password")
		testPasswordFile(t, secDir, "media_postgres_password")
		testContentFile(t, secDir, setup.SuperadminFileName, setup.DefaultSuperadminPassword)
		testDirectory(t, testDir, "db-data")
	})
}

func TestSetupCommonWithConfig(t *testing.T) {
	testDir, err := os.MkdirTemp("", "openslides-manage-service-")
	if err != nil {
		t.Fatalf("generating temporary directory failed: %v", err)
	}
	defer os.RemoveAll(testDir)

	t.Run("running setup.Setup() and create all stuff in tmp directory using a custom config", func(t *testing.T) {
		customConfig := `---
filename: my-filename-ooph1OhShi.yml
port: 8001
defaults:
  containerRegistry: example.com/test_Waetai0ohf
services:
  proxy:
    tag: 2.0.0
`
		myFileName := "my-filename-ooph1OhShi.yml"
		c := make([][]byte, 1)
		c[0] = []byte(customConfig)
		if err := setup.Setup(testDir, false, nil, c); err != nil {
			t.Fatalf("running Setup() failed with error: %v", err)
		}
		secDir := path.Join(testDir, setup.SecretsDirName)
		testFileContains(t, testDir, myFileName, "image: example.com/test_Waetai0ohf/openslides-proxy:2.0.0")
		testFileContains(t, testDir, myFileName, "image: example.com/test_Waetai0ohf/openslides-client:latest")
		testFileContains(t, testDir, myFileName, "ports:\n      - 127.0.0.1:8001:8000")
		testFileContains(t, testDir, myFileName, "image: postgres:11")
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
		testKeyFile(t, secDir, "manage_auth_password")
		testPasswordFile(t, secDir, "datastore_postgres_password")
		testPasswordFile(t, secDir, "media_postgres_password")
		testContentFile(t, secDir, setup.SuperadminFileName, setup.DefaultSuperadminPassword)
		testDirectory(t, testDir, "db-data")
	})

	t.Run("running setup.Setup() and create all stuff in tmp directory using another custom config", func(t *testing.T) {
		customConfig := `---
filename: my-filename-eab7iv8Oom.yml
disablePostgres: true
`
		myFileName := "my-filename-eab7iv8Oom.yml"
		c := make([][]byte, 1)
		c[0] = []byte(customConfig)
		if err := setup.Setup(testDir, false, nil, c); err != nil {
			t.Fatalf("running Setup() failed with error: %v", err)
		}
		testFileNotContains(t, testDir, myFileName, "image: postgres:11")
	})

	t.Run("running setup.Setup() and create all stuff in tmp directory using yet another custom config", func(t *testing.T) {
		customConfig := `---
filename: my-filename-Koo0eidifg.yml
disableDependsOn: true
`
		myFileName := "my-filename-Koo0eidifg.yml"
		c := make([][]byte, 1)
		c[0] = []byte(customConfig)
		if err := setup.Setup(testDir, false, nil, c); err != nil {
			t.Fatalf("running Setup() failed with error: %v", err)
		}
		testFileNotContains(t, testDir, myFileName, "depends_on")
	})
}

func testContentFile(t testing.TB, dir, name, expected string) {
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
	if got != expected {
		t.Fatalf("wrong content of file %q, got %q, expected %q", p, got, expected)
	}
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

func testKeyFile(t testing.TB, dir, name string) {
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
	expected := 40 // 32 bytes base64 encoded give 40 characters
	if len(got) != expected {
		t.Fatalf("wrong length of key file %q, got %d, expected %d", p, len(got), expected)
	}
}

func testPasswordFile(t testing.TB, dir, name string) {
	t.Helper()

	p := path.Join(dir, name)
	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file %q does not exist, expected existance", p)
	}
}

func testDirectory(t testing.TB, dir, name string) {
	t.Helper()

	subdir := path.Join(dir, name)
	if _, err := os.Stat(subdir); err != nil {
		t.Fatalf("missing (sub-)directory %q", subdir)
	}
}

const defaultDockerComposeYml = `---
x-default-environment: &default-environment
  ACTION_HOST: backend
  ACTION_PORT: 9002
  PRESENTER_HOST: backend
  PRESENTER_PORT: 9003

  DATASTORE_READER_HOST: datastore-reader
  DATASTORE_READER_PORT: 9010
  DATASTORE_WRITER_HOST: datastore-writer
  DATASTORE_WRITER_PORT: 9011
  DATASTORE_DATABASE_HOST: postgres
  DATASTORE_DATABASE_PORT: 5432
  DATASTORE_DATABASE_NAME: openslides
  DATASTORE_DATABASE_USER: openslides
  DATASTORE_DATABASE_PASSWORD: openslides
  DATASTORE_DATABASE_PASSWORD_FILE: /run/secrets/datastore_postgres_password

  AUTOUPDATE_HOST: autoupdate
  AUTOUPDATE_PORT: 9012

  AUTH_HOST: auth
  AUTH_PORT: 9004

  CACHE_HOST: redis
  CACHE_PORT: 6379

  MESSAGE_BUS_HOST: redis
  MESSAGE_BUS_PORT: 6379

  MEDIA_HOST: media
  MEDIA_PORT: 9006
  MEDIA_DATABASE_HOST: postgres
  MEDIA_DATABASE_PORT: 5432
  MEDIA_DATABASE_NAME: openslides
  MEDIA_DATABASE_USER: openslides
  MEDIA_DATABASE_PASSWORD: openslides
  MEDIA_DATABASE_PASSWORD_FILE: /run/secrets/media_postgres_password
  MEDIA_BLOCK_SIZE: 4096
  MEDIA_PRESENTER_HOST: backend
  MEDIA_PRESENTER_PORT: 9003

  ICC_HOST: icc
  ICC_PORT: 9013
  ICC_REDIS_HOST: redis
  ICC_REDIS_PORT: 6379

  MANAGE_HOST: manage
  MANAGE_PORT: 9008

services:
  proxy:
    image: ghcr.io/openslides/openslides/openslides-proxy:latest
    depends_on:
      - client
      - backend
      - autoupdate
      - auth
      - media
      - icc
    environment:
      << : *default-environment
    networks:
      - uplink
      - frontend
    ports:
      - 127.0.0.1:8000:8000

  client:
    image: ghcr.io/openslides/openslides/openslides-client:latest
    depends_on:
      - backend
      - autoupdate
      - auth
      - media
      - icc
    environment:
      << : *default-environment
    networks:
      - frontend

  backend:
    image: ghcr.io/openslides/openslides/openslides-backend:latest
    depends_on:
      - datastore-reader
      - datastore-writer
      - auth
      - postgres
    environment:
      << : *default-environment
    networks:
      - frontend
      - datastore-reader
      - datastore-writer
      - postgres
    secrets:
      - auth_token_key
      - auth_cookie_key
      - datastore_postgres_password

  datastore-reader:
    image: ghcr.io/openslides/openslides/openslides-datastore-reader:latest
    depends_on:
      - postgres
    environment:
      << : *default-environment
      NUM_WORKERS: 8
    networks:
      - datastore-reader
      - postgres
    secrets:
      - datastore_postgres_password

  datastore-writer:
    image: ghcr.io/openslides/openslides/openslides-datastore-writer:latest
    depends_on:
      - postgres
      - redis
    environment:
      << : *default-environment
    networks:
      - datastore-reader
      - datastore-writer
      - postgres
      - redis
    secrets:
      - datastore_postgres_password

  postgres:
    image: postgres:11
    environment:
      << : *default-environment
      POSTGRES_USER: openslides
      POSTGRES_PASSWORD_FILE: /run/secrets/datastore_postgres_password
      POSTGRES_DB: openslides
      PGDATA: /var/lib/postgresql/data/pgdata
    networks:
      - postgres
    secrets:
      - datastore_postgres_password
    volumes:
      - ./db-data:/var/lib/postgresql/data

  autoupdate:
    image: ghcr.io/openslides/openslides/openslides-autoupdate:latest
    depends_on:
      - datastore-reader
      - redis
    environment:
      << : *default-environment
      MESSAGING: redis
      AUTH: ticket
    networks:
      - frontend
      - datastore-reader
      - redis
    secrets:
      - auth_token_key
      - auth_cookie_key

  auth:
    image: ghcr.io/openslides/openslides/openslides-auth:latest
    depends_on:
      - datastore-reader
      - redis
    environment:
      << : *default-environment
    networks:
      - frontend
      - datastore-reader
      - redis
    secrets:
      - auth_token_key
      - auth_cookie_key

  redis:
    image: redis:latest
    environment:
      << : *default-environment
    networks:
      - redis

  media:
    image: ghcr.io/openslides/openslides/openslides-media:latest
    depends_on:
      - backend
      - postgres
    environment:
      << : *default-environment
    networks:
      - frontend
      - datastore-reader
      - datastore-writer
      - postgres
    secrets:
      - media_postgres_password

  icc:
    image: ghcr.io/openslides/openslides/openslides-icc:latest
    depends_on:
      - datastore-reader
      - postgres
      - redis
    environment:
      << : *default-environment
      MESSAGING: redis
      AUTH: ticket
    networks:
      - frontend
      - datastore-reader
      - postgres
      - redis
    secrets:
      - auth_token_key
      - auth_cookie_key

  manage:
    image: ghcr.io/openslides/openslides/openslides-manage:latest
    depends_on:
      - datastore-reader
      - datastore-writer
      - auth
    environment:
      << : *default-environment
    networks:
      - frontend
      - datastore-reader
      - datastore-writer
      - postgres
      - redis
    secrets:
      - superadmin
      - manage_auth_password

networks:
  uplink:
  frontend:
    internal: true
  datastore-reader:
    internal: true
  datastore-writer:
    internal: true
  postgres:
    internal: true
  redis:
    internal: true

secrets:
  auth_token_key:
    file: ./secrets/auth_token_key
  auth_cookie_key:
    file: ./secrets/auth_cookie_key
  superadmin:
    file: ./secrets/superadmin
  manage_auth_password:
    file: ./secrets/manage_auth_password
  datastore_postgres_password:
    file: ./secrets/datastore_postgres_password
  media_postgres_password:
    file: ./secrets/media_postgres_password
`

func TestSetupNoDirectory(t *testing.T) {
	hasErrMsg := "not a directory"
	err := setup.Setup("setup_test.go", false, nil, nil)
	if !strings.Contains(err.Error(), hasErrMsg) {
		t.Fatalf("running Setup() with invalid directory, got error message %q, expected %q", err.Error(), hasErrMsg)
	}
}
