# OpenSlides Manage Service

Manage service and tool for OpenSlides which provides some management commands
to setup and control OpenSlides instances.

The service listens on the port given by environment variable
`MANAGE_SERVICE_PORT` (default 9008) and uses [gRPC](https://grpc.io/).

The (client) tool can be used as follows:

    $ ./openslides

You can find all management commands in the help text.


## How to start the full system

You can start OpenSlides with Docker Compose as follows:

Be sure you have Docker and Docker Compose installed and Docker daemon is
running. Check if you have to run docker as local user or as root.

    $ docker info

First go to a nice place in your filesystem. Then run:

    $ ./manage setup --cwd
    $ docker-compose up --detach
    $ ./manage initial-data

Now open http://localhost:8000, login and have fun. Afterwars run:

    $ docker-compose stop

To remove all containers including the complete database run:

    $ docker-compose rm


## Development

For development you need [Go](https://golang.org/) and the [Protocol Buffer
Compiler](https://grpc.io/docs/protoc-installation/).

The (client) tool can be build with

    $ go build ./cmd/openslides

The server part can be build with:

    $ go build ./cmd/server

To compile changed `.proto` files, run `protoc`:

    $ make protoc


## Docker

You can build the following Docker images.

To build the manage service server use:

    $ docker build .

To build the client e. g. for use as one shot container with customized command
use:

    $ docker build --target manage-tool-productive .
