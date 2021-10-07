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

Then run:

    $ ./openslides setup .
    $ docker-compose up --detach
    $ ./openslides initial-data

Now open https://localhost:8000, login and have fun (TODO: HTTPS-support is still missing). Afterwars run:

    $ docker-compose stop

To remove all containers run:

    $ docker-compose rm

It is also possible to use Docker Swarm instead of Docker Compose for bigger
setups. Let us know if this is interesting for you.


## Configuration of the generated YAML file

The setup command generates also a YAML file (default filename: `docker-compose.yml`)
with the container configuration for all services. This step can be configured with
a YAML formated config file. E. g. to get a customized YAML file run:

    $ ./openslides setup --config my-config.yml .

See the [default config](pkg/setup/default-config.yml) for syntax and defaults.


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

    $ docker build --target client .
