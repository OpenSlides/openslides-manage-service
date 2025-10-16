package connection

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	// defaultAddr holds the default host and port to be used for the gRPC connection that is established by some commands.
	defaultAddr = "localhost:8000"

	// defaultTimeout holds the default timeout for the gRPC connection that is established by some commands.
	defaultTimeout = 30 * time.Second
)

// Params provides the parameters for the connection to the manage server.
type Params struct {
	Addr         *string
	PasswordFile *string
	Timeout      *time.Duration
	NoSSL        *bool
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

// Unary provides parameters for an unary connection like address, passwordfile,
// timeout and the noSSL flag to the given cobra command.
func Unary(cmd *cobra.Command) Params {
	addr := cmd.Flags().StringP("address", "a", defaultAddr, "address of the OpenSlides manage service")
	defaultPasswordFile := path.Join(".", setup.SecretsDirName, setup.ManageAuthPasswordFileName)
	passwordFile := cmd.Flags().String("password-file", defaultPasswordFile, "file with password for authorization to manage service, not usable in development mode")
	noSSL := cmd.Flags().Bool("no-ssl", false, "use an unencrypted connection to manage service")
	timeout := cmd.Flags().DurationP("timeout", "t", defaultTimeout, "time to wait for the command's response")
	return Params{
		Addr:         addr,
		PasswordFile: passwordFile,
		NoSSL:        noSSL,
		Timeout:      timeout,
	}
}
