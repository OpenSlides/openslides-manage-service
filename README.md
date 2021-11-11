# OpenSlides Manage Service

Manage service and tool for OpenSlides which provides some management commands
to setup and control OpenSlides instances.

The service uses [gRPC](https://grpc.io/) and can be reached directly via the OpenSlides proxy service.

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
    $ docker-compose pull

Before starting the instance you must setup a means of SSL encryption as OpenSlides 4 only runs over HTTPS. See [HTTPS](#HTTPS)

When HTTPS is set up continue with:

    $ docker-compose up

Wait until all services are available. Then run in a second terminal in the same directory:

    $ ./openslides initial-data

Now open https://localhost:8000, login and have fun. Afterwars run:

## Stop the server

To stop the server run:

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

To rebuild the YAML file without resetting the whole directory (including secrets) run:

    $ ./openslides config --config my-config.yml .

See the [default config](pkg/setup/default-config.yml) for syntax and defaults.

## HTTPS
The manage tool provides two settable options for using HTTPS, both of which can be set in a `config.yml` file.
When running OpenSlides on localhost a locally generated self-signed certificate can be used.
To do so add the following line to your `config.yml`.

    enableLocalHTTPS: true

The `cert_crt` and `cert_key` files are now expected within the `secrets` directory. Use your favorite tool to generate them, e.g.

    openssl req -x509 -newkey rsa:4096 -nodes -days 3650 \
        -subj "/C=DE/O=Selfsigned Test/CN=localhost" \
        -keyout secrets/cert_key -out secrets/cert_crt

If OpenSlides is ran behind a publicly available domain, caddys integrated certificate retrieval can be utilized.
The following lines in `config.yml` are necessary to do so.

    enableAutoHTTPS: true
    defaultEnvironment:
      EXTERNAL_ADDRESS: openslides.example.com
      # use letsencrypt staging environment for testing
      # ACME_ENDPOINT: https://acme-staging-v02.api.letsencrypt.org/directory

See [proxy](https://github.com/OpenSlides/OpenSlides/blob/master/proxy) for details on provided methods for HTTPS activation.


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
