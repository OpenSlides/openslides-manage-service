package setup_test

import (
	"errors"
	"os"
	"path"
	"reflect"
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
		if err := setup.Setup(testDir, false, nil); err != nil {
			t.Fatalf("running Setup() failed with error: %v", err)
		}
		secDir := path.Join(testDir, setup.SecretsDirName)
		testContentFile(t, testDir, "docker-compose.yml", defaultDockerComposeYml)
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
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

		if err := setup.Setup(testDir, false, nil); err != nil {
			t.Fatalf("running Setup() failed with error: %v", err)
		}
		secDir := path.Join(testDir, setup.SecretsDirName)
		testContentFile(t, testDir, "docker-compose.yml", testContent)
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
		testContentFile(t, secDir, setup.SuperadminFileName, setup.DefaultSuperadminPassword)
		testDirectory(t, testDir, "db-data")
	})

	t.Run("running setup.Setup() with force flag with changing existant files", func(t *testing.T) {
		if err := setup.Setup(testDir, true, nil); err != nil {
			t.Fatalf("running Setup() failed with error: %v", err)
		}
		secDir := path.Join(testDir, setup.SecretsDirName)
		testContentFile(t, testDir, "docker-compose.yml", defaultDockerComposeYml)
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
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
		if err := setup.Setup(dir, false, nil); err != nil {
			t.Fatalf("running Setup() failed with error: %v", err)
		}
		secDir := path.Join(dir, setup.SecretsDirName)
		testContentFile(t, dir, "docker-compose.yml", defaultDockerComposeYml)
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
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
		if err := setup.Setup(testDir, false, []byte(tplText)); err != nil {
			t.Fatalf("running Setup() failed with error: %v", err)
		}
		secDir := path.Join(testDir, setup.SecretsDirName)
		testContentFile(t, testDir, "docker-compose.yml", tplText)
		testKeyFile(t, secDir, "auth_token_key")
		testKeyFile(t, secDir, "auth_cookie_key")
		testContentFile(t, secDir, setup.SuperadminFileName, setup.DefaultSuperadminPassword)
		testDirectory(t, testDir, "db-data")
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
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("wrong content of file %q, got %q, expected %q", p, got, expected)
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

  AUTOUPDATE_HOST: autoupdate
  AUTOUPDATE_PORT: 9012

  AUTH_HOST: auth
  AUTH_PORT: 9004

  CACHE_HOST: cache
  CACHE_PORT: 6379

  MESSAGE_BUS_HOST: message-bus
  MESSAGE_BUS_PORT: 6379

  MEDIA_HOST: media
  MEDIA_PORT: 9006
  MEDIA_DATABASE_HOST: postgres
  MEDIA_DATABASE_NAME: openslides

  MANAGE_HOST: manage
  MANAGE_PORT: 9008

services:
  proxy:
    image: ghcr.io/openslides/openslides/openslides-proxy:4.0.0-dev
    depends_on:
      - client
      - backend
      - autoupdate
      - auth
      - media
    environment:
      << : *default-environment
    networks:
      - uplink
      - frontend
    ports:
      - 127.0.0.1:8000:8000

  client:
    image: ghcr.io/openslides/openslides/openslides-client:4.0.0-dev
    depends_on:
      - backend
      - autoupdate
    environment:
      << : *default-environment
    networks:
      - frontend

  backend:
    image: ghcr.io/openslides/openslides/openslides-backend:4.0.0-dev
    depends_on:
      - datastore-reader
      - datastore-writer
      - auth
    environment:
      << : *default-environment
    networks:
      - frontend
      - backend
    secrets:
      - auth_token_key
      - auth_cookie_key

  datastore-reader:
    image: ghcr.io/openslides/openslides/openslides-datastore-reader:4.0.0-dev
    depends_on:
      - postgres
    environment:
      << : *default-environment
      NUM_WORKERS: 8
    networks:
      - backend
      - datastore-reader
      - postgres

  datastore-writer:
    image: ghcr.io/openslides/openslides/openslides-datastore-writer:4.0.0-dev
    depends_on:
      - postgres
      - message-bus
    environment:
      << : *default-environment
    networks:
      - backend
      - postgres
      - message-bus

  postgres:
    image: postgres:11
    environment:
      << : *default-environment
      POSTGRES_USER: openslides
      POSTGRES_PASSWORD: openslides
      POSTGRES_DB: openslides
      PGDATA: /var/lib/postgresql/data/pgdata
    networks:
      - postgres
    volumes:
      - ./db-data:/var/lib/postgresql/data

  autoupdate:
    image: ghcr.io/openslides/openslides/openslides-autoupdate:4.0.0-dev
    depends_on:
      - datastore-reader
      - message-bus
    environment:
      << : *default-environment
    networks:
      - frontend
      - datastore-reader
      - message-bus
    secrets:
      - auth_token_key
      - auth_cookie_key

  auth:
    image: ghcr.io/openslides/openslides/openslides-auth:4.0.0-dev
    depends_on:
      - datastore-reader
      - message-bus
      - cache
    environment:
      << : *default-environment
    networks:
      - frontend
      - datastore-reader
      - message-bus
      - cache
    secrets:
      - auth_token_key
      - auth_cookie_key

  cache:
    image: redis:latest
    environment:
      << : *default-environment
    networks:
      - cache

  message-bus:
    image: redis:latest
    environment:
      << : *default-environment
    networks:
      - message-bus

  media:
    image: ghcr.io/openslides/openslides/openslides-media:4.0.0-dev
    depends_on:
      - backend
      - postgres
    environment:
      << : *default-environment
    networks:
      - frontend
      - backend
      - postgres

  manage:
    image: ghcr.io/openslides/openslides/openslides-manage:4.0.0-dev
    depends_on:
      - datastore-reader
      - datastore-writer
      - auth
    environment:
      << : *default-environment
    networks:
      - uplink
      - frontend
      - backend
    secrets:
      - admin
    ports:
      - 127.0.0.1:9008:9008

networks:
  uplink:
  frontend:
    internal: true
  backend:
    internal: true
  postgres:
    internal: true
  datastore-reader:
    internal: true
  message-bus:
    internal: true
  cache:
    internal: true

secrets:
  auth_token_key:
    file: ./secrets/auth_token_key
  auth_cookie_key:
    file: ./secrets/auth_cookie_key
  admin:
    file: ./secrets/admin
`

func TestSetupNoDirectory(t *testing.T) {
	hasErrMsg := "not a directory"
	err := setup.Setup("setup_test.go", false, nil)
	if !strings.Contains(err.Error(), hasErrMsg) {
		t.Fatalf("running Setup() with invalid directory, got error message %q, expected %q", err.Error(), hasErrMsg)
	}
}