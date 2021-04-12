# This configuration was created from a template file. The accompanying .env file
# might be the correct place for customizations.

version: '3.4'

services:
  proxy:
    build:
      context: https://github.com/OpenSlides/OpenSlides.git#{{ .Ref }}:proxy
    depends_on:
      - client
      - backend
      - autoupdate
      - auth
      - media
    env_file: services.env
    networks:
      - uplink
      - frontend
    ports:
      - "127.0.0.1:{{ .ExternalHTTPPort }}:8000"

  client:
    build:
      context: https://github.com/OpenSlides/openslides-client.git#{{ .CommitID.client }}
    depends_on:
      - backend
      - autoupdate
    networks:
      - frontend

  backend:
    build:
      context: https://github.com/OpenSlides/openslides-backend.git#{{ .CommitID.backend }}
    depends_on:
      - datastore-reader
      - datastore-writer
      - auth
    env_file: services.env
    networks:
      - frontend
      - backend
    secrets:
      - auth_token_key
      - auth_cookie_key

  datastore-reader:
    build:
      context: https://github.com/OpenSlides/openslides-datastore-service.git#{{ .CommitID.datastore }}
      args:
        MODULE: reader
        PORT: 9010
    depends_on:
      - postgres
    env_file: services.env
    environment:
      - NUM_WORKERS=8
    networks:
      - backend
      - datastore-reader
      - postgres

  datastore-writer:
    build:
      context: https://github.com/OpenSlides/openslides-datastore-service.git#{{ .CommitID.datastore }}
      args:
        MODULE: writer
        PORT: 9011
    depends_on:
      - postgres
      - message-bus
    env_file: services.env
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
    build:
      context: https://github.com/OpenSlides/openslides-autoupdate-service.git#{{ .CommitID.autoupdate }}
    depends_on:
      - datastore-reader
      - message-bus
    env_file: services.env
    networks:
      - frontend
      - datastore-reader
      - message-bus
    secrets:
      - auth_token_key
      - auth_cookie_key

  auth:
    build:
      context: https://github.com/OpenSlides/openslides-auth-service.git#{{ .CommitID.auth }}
    depends_on:
      - datastore-reader
      - message-bus
      - cache
    env_file: services.env
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
    build:
      context: https://github.com/OpenSlides/openslides-media-service.git#{{ .CommitID.media }}
    depends_on:
      - backend
      - postgres
    env_file: services.env
    networks:
      - frontend
      - backend
      - postgres

  manage:
    build:
      context: https://github.com/OpenSlides/openslides-manage-service.git#{{ .CommitID.manage }}
    depends_on:
    - datastore-reader
    - datastore-writer
    - auth
    env_file: services.env
    ports:
    - "127.0.0.1:{{ .ExternalManagePort }}:9008"
    networks:
    - uplink
    - frontend
    - backend

# TODO: Remove this service so the networks won't matter any more.
  permission:
    build:
      context: https://github.com/OpenSlides/openslides-permission-service.git#{{ .CommitID.permission }}
    depends_on:
    - datastore-reader
    env_file: services.env
    networks:
    - frontend
    - backend

# Setup: host <-uplink-> proxy <-frontend-> services that are reachable from the client <-backend-> services that are internal-only
# There are special networks for some services only, e.g. postgres only for the postgresql, datastore reader and datastore writer
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
