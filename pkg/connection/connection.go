package connection

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	// DefaultAddr holds the default host and port to be used for the gRPC connection that is established by some commands.
	DefaultAddr = "localhost:8000"

	// DefaultTimeout holds the default timeout for the gRPC connection that is established by some commands.
	DefaultTimeout = 5 * time.Second
)

// Params provides the parameters for the connection to the manage server.
type Params interface {
	Addr() string
	PasswordFile() string
	Timeout() time.Duration
	NoSSL() bool
}

// Dial creates a gRPC connection to the server.
func Dial(ctx context.Context, address, passwordFile string, ssl bool) (proto.ManageClient, func() error, error) {
	pw, err := shared.AuthSecret(passwordFile, os.Getenv("OPENSLIDES_DEVELOPMENT"))
	if err != nil {
		return nil, nil, fmt.Errorf("getting server auth secret: %w", err)
	}
	creds := shared.BasicAuth{
		Password: pw,
	}

	transportOption := grpc.WithInsecure() // Option for unencrypted HTTP connection
	if ssl {
		// TODO: Have a look at https://itnext.io/practical-guide-to-securing-grpc-connections-with-go-and-tls-part-1-f63058e9d6d1 and
		// try not to use InsecureSkipVerify by default.
		transCreds := credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})
		transportOption = grpc.WithTransportCredentials(transCreds)
	}

	conn, err := grpc.DialContext(ctx, address,
		transportOption,
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(creds),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("creating gRPC client connection with grpc.DialContext(): %w", err)
	}
	return proto.NewManageClient(conn), conn.Close, nil
}
