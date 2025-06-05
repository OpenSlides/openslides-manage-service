ARG CONTEXT=prod

FROM golang:1.19-alpine as base

## Setup
ARG CONTEXT
WORKDIR /root
ENV ${CONTEXT}=1

## Installs
RUN apk add git --no-cache

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY pkg pkg
COPY proto proto
COPY Makefile Makefile


## External Information
LABEL org.opencontainers.image.title="OpenSlides Manage Service"
LABEL org.opencontainers.image.description="Manage service and tool for OpenSlides which \
    provides some management commands to setup and control OpenSlides instances."
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/OpenSlides/openslides-manage-service"
LABEL org.opencontainers.image.documentation="https://github.com/OpenSlides/openslides-manage-service/blob/main/README.md"    

EXPOSE 9008

## Command
COPY ./dev/command.sh ./
RUN chmod +x command.sh
CMD ["./command.sh"]
HEALTHCHECK CMD ["/healthcheck"]


# Development Image

FROM base as dev

RUN ["go", "install", "github.com/githubnemo/CompileDaemon@latest"]



# Testing Image

FROM base as tests

RUN apk add build-base --no-cache



# Production Image

FROM base as builder

RUN CGO_ENABLED=0 go build ./cmd/openslides && \
    CGO_ENABLED=0 go build ./cmd/server && \ 
    CGO_ENABLED=0 go build ./cmd/healthcheck


FROM scratch as client

WORKDIR /

COPY --from=builder /root/openslides .

ENTRYPOINT ["/openslides"]


FROM scratch as prod

WORKDIR /

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
