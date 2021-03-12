FROM golang:1.16.2-alpine3.12 as base
LABEL maintainer="OpenSlides Team <info@openslides.com>"
WORKDIR /root/

RUN apk add git

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY pkg pkg
COPY proto proto


# Build service in seperate stage.
FROM base as builder
RUN go build ./cmd/server


# Test build.
FROM base as testing

RUN apk add build-base

CMD go vet ./... && go test ./...


# Development build.
FROM base as development

RUN ["go", "install", "github.com/githubnemo/CompileDaemon@latest"]
EXPOSE 8001

CMD CompileDaemon -log-prefix=false -build="go build ./cmd/server" -command="./server"


# Productive build.
FROM alpine:3.13.2
WORKDIR /root/

COPY --from=builder /root/server .
EXPOSE 8001

CMD ./server
