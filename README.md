# OpenSlides Manage Service

Manage service and tool for OpenSlides which provides some management commands
to setup and control OpenSlides instances.

The tool can be used as follows:

    $ ./openslides

You will get a help text with all management commands.


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
