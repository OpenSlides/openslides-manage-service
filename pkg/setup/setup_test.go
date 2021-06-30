package setup_test

import (
	"errors"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
	"github.com/go-test/deep"
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

	got := strings.Split(string(dcYmlContent[:]), "\n")
	expected := strings.Split(defaultDockerComposeYml, "\n")
	diff := deep.Equal(got, expected)
	if diff != nil {
		t.Errorf("wrong content of YML file: %s", diff)
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
