package setup_test

import (
	"errors"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
)

func TestSetup(t *testing.T) {
	testDir, err := os.MkdirTemp("", "openslides-manage-service-")
	if err != nil {
		t.Errorf("generating temporary directory failed: %v", err)
	}
	defer os.RemoveAll(testDir)

	t.Run("running setup.Setup() and create all stuff in tmp directory", func(t *testing.T) {
		if err := setup.Setup(testDir); err != nil {
			t.Errorf("running Setup() failed with error: %v", err)
		}
		testFile(t, testDir, "docker-compose.yml", defaultDockerComposeYml)
		testFile(t, testDir, ".env", defaultEnvFile)
	})

}

func testFile(t testing.TB, dir, name, expected string) {
	t.Helper()

	p := path.Join(dir, name)
	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		t.Errorf("file %q does not exist, expected existance", p)
	}
	content, err := os.ReadFile(p)
	if err != nil {
		t.Errorf("error reading file %q: %v", p, err)
	}

	got := string(content[:])
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("wrong content of YML file, expected %q, got %q", expected, got)
	}

}

const defaultDockerComposeYml = `---
services:
  proxy:
    image: ghcr.io/openslides/openslides/openslides-proxy:4.0.0-dev
    depends_on:
      - client
      - backend
      - autoupdate
      - auth
      - media
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
    networks:
      - frontend

  backend:
    image: ghcr.io/openslides/openslides/openslides-backend:4.0.0-dev
    depends_on:
      - datastore-reader
      - datastore-writer
      - auth
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
      - NUM_WORKERS=8
    networks:
      - backend
      - datastore-reader
      - postgres

  datastore-writer:
    image: ghcr.io/openslides/openslides/openslides-datastore-writer:4.0.0-dev
    depends_on:
      - postgres
      - message-bus
    networks:
      - backend
      - postgres
      - message-bus

  postgres:
    image: postgres:11
    environment:
      - POSTGRES_USER=openslides
      - POSTGRES_PASSWORD=openslides
      - POSTGRES_DB=openslides
    networks:
      - postgres

  autoupdate:
    image: ghcr.io/openslides/openslides/openslides-autoupdate:4.0.0-dev
    depends_on:
      - datastore-reader
      - message-bus
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
    networks:
      - cache

  message-bus:
    image: redis:latest
    networks:
      - message-bus

  media:
    image: ghcr.io/openslides/openslides/openslides-media:4.0.0-dev
    depends_on:
      - backend
      - postgres
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
    ports:
      - 127.0.0.1:9008:9008
    networks:
      - uplink
      - frontend
      - backend
    secrets:
      - admin

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

const defaultEnvFile = `MESSAGE_BUS_HOST=message-bus
MESSAGE_BUS_PORT=6379

DATASTORE_READER_HOST=datastore-reader
DATASTORE_READER_PORT=9010
DATASTORE_WRITER_HOST=datastore-writer
DATASTORE_WRITER_PORT=9011
DATASTORE_DATABASE_HOST=postgres

ACTION_HOST=backend
ACTION_PORT=9002
PRESENTER_HOST=backend
PRESENTER_PORT=9003

AUTOUPDATE_HOST=autoupdate
AUTOUPDATE_PORT=9012

PERMISSION_HOST=permission
PERMISSION_PORT=9005

AUTH_HOST=auth
AUTH_PORT=9004
CACHE_HOST=cache
CACHE_PORT=6379

MEDIA_HOST=media
MEDIA_PORT=9006
MEDIA_DATABASE_HOST=postgres
MEDIA_DATABASE_NAME=openslides

MANAGE_HOST=manage
MANAGE_PORT=9008
`
