ARG CONTEXT=prod

FROM golang:1.24.4-alpine AS base

## Setup
ARG CONTEXT
WORKDIR /app
ENV APP_CONTEXT=${CONTEXT}

## Installs
RUN apk add git --no-cache

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY pkg pkg
COPY proto proto
COPY Makefile Makefile

## External Information
EXPOSE 9008

## Healthcheck
HEALTHCHECK CMD ["/app/healthcheck"]

# Development Image

FROM base AS dev

RUN ["go", "install", "github.com/githubnemo/CompileDaemon@latest"]

## Command
CMD CompileDaemon -log-prefix=false -build="go build ./cmd/server" -command="./server"

# Testing Image

FROM base AS tests

COPY dev/container-tests.sh ./dev/container-tests.sh

RUN apk add --no-cache \
    build-base \
    docker && \
    go get -u github.com/ory/dockertest/v3 && \
    go install golang.org/x/lint/golint@latest && \
    chmod +x dev/container-tests.sh

## Command
STOPSIGNAL SIGKILL
CMD ["sleep", "inf"]

# Production Image

FROM base AS builder

RUN CGO_ENABLED=0 go build ./cmd/openslides && \
    CGO_ENABLED=0 go build ./cmd/server && \
    CGO_ENABLED=0 go build ./cmd/healthcheck

FROM scratch AS client

WORKDIR /
ENV APP_CONTEXT=prod

COPY --from=builder /app/openslides .

ENTRYPOINT ["/openslides"]

FROM scratch AS prod

## Setup
ARG CONTEXT
ENV APP_CONTEXT=prod

LABEL org.opencontainers.image.title="OpenSlides Manage Service"
LABEL org.opencontainers.image.description="Manage service and tool for OpenSlides which \
    provides some management commands to setup and control OpenSlides instances."
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/OpenSlides/openslides-manage-service"
LABEL org.opencontainers.image.documentation="https://github.com/OpenSlides/openslides-manage-service/blob/main/README.md"

COPY --from=builder /app/healthcheck /
COPY --from=builder /app/server /

EXPOSE 9008

ENTRYPOINT ["/server"]

HEALTHCHECK CMD ["/healthcheck"]
