# OpenSlides Manage Service

Manage service and tool for OpenSlides which provides some management commands
to setup and control OpenSlides instances.

The tool can be used as follows:

    $ ./openslides

You will get a help text with all management commands.


## How to start the full system

You can start OpenSlides with Docker Compose. It is also possible to use Docker
Swarm or other container orchestration systems instead of Docker Compose for
bigger setups, but this is not described here.

So for now let's use Docker Compose. Be sure you have Docker and Docker Compose
installed and the Docker daemon is running. Check if you have to run docker as
local user or as root.

    $ docker info

Go to a nice place in your filesystem and build or
[download](https://github.com/OpenSlides/openslides-manage-service/releases) the
manage tool `openslides`. To setup the instance in this directory run:

    $ ./openslides setup .

This will give you a default setup with local self-signed SSL certificates. You
will get [a browser warning and have to care yourself about checking the
fingerprint of the
certificate](https://en.wikipedia.org/wiki/Self-signed_certificate).

[Below](#SSL-encryption) you find for more information and how to use
[caddys](https://github.com/OpenSlides/OpenSlides/blob/master/proxy) integrated
certificate retrieval or how to disable the proxy expecting SSL. Disabling SSL
is only feasible if you use an extra proxy in front of OpenSlides. Keep in mind
that the browser client requires a HTTPS connection to the server. It is NOT
possible to use OpenSlides without any SSL encryption at all.

Now have a look at the `docker-compose.yml` and customize it if you want. Then
run:

    $ docker-compose pull
    $ docker-compose up

Wait until all services are available. Then run in a second terminal in the same
directory:

    $ ./openslides initial-data

Now open https://localhost:8000, login with superuser credentials (default
username and password: `superadmin`) and have fun.


## Stop the server and remove the containers

To stop the server run:

    $ docker-compose stop

To remove all containers run:

    $ docker-compose rm

To remove the database you have to remove the content of the `db-data`
directory.


## Configuration of the generated Docker Compose YAML file

The `setup` command generates a Docker Compose YAML file (default filename:
`docker-compose.yml`) with the container configuration for all services. This
step can be configured with a (second) YAML formated configuration file.

E. g. create a file `my-config.yml` with the following content:

    ---
    port: 9000

See the [default config](pkg/config/default-config.yml) for syntax and defaults
of this file.

After you have such a file remove your Docker Compose YAML file and rerun the
`setup` command:

    $ ./openslides setup --config my-config.yml .

This way you get a Docker Compose YAML file which let OpenSlides listen on the
configured custom port.

You may also use  the `--force` flag in some cases which also resets secrets and
much more. To rebuild the Docker Compose YAML file without resetting the whole
directory (including secrets) use the `config` command instead of the `setup`
command. E. g. run:

    $ ./openslides config --config my-config.yml .

This command will just rebuild your Docker Compose YAML file.


## SSL encryption

The manage tool provides settable options for using SSL encryption, which can be
set in a [custom YAML configuration
file](#Configuration-of-the-generated-Docker-Compose-YAML-file).

If you do not use any customization the `setup` command generates a self-signed
certificate by default.

If you want to use any other certificate you posses, just replace `cert_crt` and
`cert_key` files in the `secrets` directory before starting Docker.

If you want to disable SSL encryption, because you use OpenSlides behind your
own proxy that provides SSL encryption, just  add the following line to your
YAML configuration file.

    enableLocalHTTPS: false

If you run OpenSlides behind a publicly accessible domain, you can use caddys
integrated certificate retrieval. Add the following lines to your YAML
configuration file before running the setup command:

    enableAutoHTTPS: true
    defaultEnvironment:
      EXTERNAL_ADDRESS: openslides.example.com
      # Use letsencrypt staging environment for testing
      # ACME_ENDPOINT: https://acme-staging-v02.api.letsencrypt.org/directory

See [the proxy service](https://github.com/OpenSlides/OpenSlides/blob/main/proxy) for
details on provided methods for HTTPS activation.


## Under the hood

The manage service uses [gRPC](https://grpc.io/) and can be reached directly via
the OpenSlides proxy service.


## Development

For development you need [Go](https://golang.org/) and the [Protocol Buffer
Compiler](https://grpc.io/docs/protoc-installation/).

The tool can be build with

    $ go build ./cmd/openslides

The server can be build with:

    $ go build ./cmd/server

To compile changed `.proto` files, install the [Protocol Buffer
Compiler](https://grpc.io/docs/protoc-installation/) and its [Go
plugins](https://grpc.io/docs/languages/go/quickstart/). Then run `protoc`:

    $ make protoc


## Using Docker images

You can build the following Docker images.

To build the manage service server use:

    $ docker build .

To build the tool e. g. for use as one shot container with customized command
use:

    $ docker build --target client .

Finally you can use Docker to build the tool even without having Go installed.
Just run:

    $ make openslides
