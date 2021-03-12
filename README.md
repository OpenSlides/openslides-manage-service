# OpenSlides Manage Service

Manage service for OpenSlides which provides the following management commands.
The service listens on the port given by environment variable
`MANAGE_SERVICE_PORT` (default 8001). Just send a POST request with some JSON.


## createUser

Path:

    /create-user

Data:

    username: string
    password: string
    organisation_management_level: string  // Optional





https://grpc.io/docs/languages/go/

https://github.com/grpc/grpc-go/tree/master/examples/helloworld

https://github.com/OpenSlides/OpenSlides/pull/5916