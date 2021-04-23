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

You can build the following Docker images.

To build the manage service server use:

    $ docker build .

To build the client e. g. for use as one shot container with customized command
use:

    $ docker build --target manage-tool-productive .


## How to start the full system

You can start OpenSlides with Docker Compose as follows:

Be sure you have Docker and Docker Compose installed and Docker daemon is
running. Check if you have to run docker as local user or as root.

    $ docker info

First go to a nice place in your filesystem. Then run:

    $ ./manage setup --local
    $ docker-compose up --build --detach
    $ ./manage initial-data

Now open http://localhost:8000, login and have fun. Afterwars run:

    $ docker-compose stop

To remove all containers including the complete database run:

    $ docker-compose rm
