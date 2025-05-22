ARG CONTEXT=prod
ARG GO_IMAGE_VERSION=1.19

FROM golang:${GO_IMAGE_VERSION}-alpine as base

ARG CONTEXT
ARG GO_IMAGE_VERSION

WORKDIR /root

## Installs
RUN apk add git

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY pkg pkg
COPY proto proto
COPY Makefile Makefile


LABEL org.opencontainers.image.title="OpenSlides Manage Service"
LABEL org.opencontainers.image.description="Manage service and tool for OpenSlides which \
    provides some management commands to setup and control OpenSlides instances."
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/OpenSlides/openslides-manage-service"
LABEL org.opencontainers.image.documentation="https://github.com/OpenSlides/openslides-manage-service/blob/main/README.md"    

EXPOSE 9008
HEALTHCHECK CMD ["/healthcheck"]


# Development Image

FROM base as dev

RUN ["go", "install", "github.com/githubnemo/CompileDaemon@latest"]

CMD CompileDaemon -log-prefix=false -build="go build ./cmd/server" -command="./server"



# Testing Image

FROM base as tests

RUN apk add build-base

CMD make test



# Production Image

FROM base as builder

RUN CGO_ENABLED=0 go build ./cmd/openslides
RUN CGO_ENABLED=0 go build ./cmd/server
RUN CGO_ENABLED=0 go build ./cmd/healthcheck


FROM scratch as client

COPY --from=builder /root/openslides .

ENTRYPOINT ["/openslides"]


FROM scratch as prod

LABEL org.opencontainers.image.title="OpenSlides Manage Service"
LABEL org.opencontainers.image.description="Manage service and tool for OpenSlides which \
    provides some management commands to setup and control OpenSlides instances."
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/OpenSlides/openslides-manage-service"
LABEL org.opencontainers.image.documentation="https://github.com/OpenSlides/openslides-manage-service/blob/main/README.md"    

COPY --from=builder /root/healthcheck .
COPY --from=builder /root/server .

EXPOSE 9008

ENTRYPOINT ["/server"]
