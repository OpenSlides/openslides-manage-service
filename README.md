# OpenSlides Manage Service

Manage service for OpenSlides which provides some management commands. The
service listens on the port given by environment variable `MANAGE_SERVICE_PORT`
(default 9008) and uses [gRPC](https://grpc.io/).

The tool (client) can be used as follows:

    $ ./manage

You can find all management commands in the help text.

## Development

For development you need [Go](https://golang.org/) and the [Protocol Buffer
Compiler](https://grpc.io/docs/protoc-installation/).

Build the server with:

    $ go build ./cmd/server

The client can be build with

    $ go build ./cmd/manage

To compile changed `.proto` files, run `protoc`:

    $ make protoc


## Docker

You can build the following Docker images:

    $ docker build .  # Builds the manage service server.
    $ docker build --target manage-tool-productive .  # Builds the client e. g. for use as one shot container with customized command.


## TODO

This is temporary text during development:

    $ ./manage setup --local
    $ docker-compose build
    $ docker-compose up -d
    $ ./manage initial-data
