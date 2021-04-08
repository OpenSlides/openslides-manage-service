FROM golang:1.16.3-alpine3.13 as base
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
RUN CGO_ENABLED=0 go build ./cmd/server
RUN CGO_ENABLED=0 go build ./cmd/manage

# Test build.
FROM base as testing

RUN apk add build-base

CMD go vet ./... && go test ./...


# Development build.
FROM base as development

RUN ["go", "install", "github.com/githubnemo/CompileDaemon@latest"]
EXPOSE 9008

CMD CompileDaemon -log-prefix=false -build="go build ./cmd/server" -command="./server"


# Productive build.
FROM alpine:3.13.2
WORKDIR /root/
RUN apk add bash

COPY --from=builder /root/server .
COPY --from=builder /root/manage .
COPY entrypoint-setup .

EXPOSE 9008
CMD ["/root/server"]
