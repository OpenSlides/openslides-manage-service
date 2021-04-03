package manage

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
)

const checkServerHelp = `Checks if all OpenSlides services are running

This command checks if all OpenSlides services are currently running and
listening on their respective ports.
`

// CmdCheckServer checks if all OpenSlides services are running.
func CmdCheckServer(cfg *ClientConfig) *cobra.Command {
	var skipClient bool

	cmd := &cobra.Command{
		Use:   "check-server",
		Short: "Checks if all services are running.",
		Long:  checkServerHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
			defer cancel()

			service, close, err := Dial(ctx, cfg.Address)
			if err != nil {
				return fmt.Errorf("connecting to gRPC server: %w", err)
			}
			defer close()

			req := &proto.CheckServerRequest{
				SkipClient: skipClient,
			}

			if _, err := service.CheckServer(ctx, req); err != nil {
				return fmt.Errorf("check server: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&skipClient, "skip-client", false, "Skip checking the client")

	return cmd
}

// CheckServer sets hashes and sets the password
func (s *Server) CheckServer(ctx context.Context, in *proto.CheckServerRequest) (*proto.CheckServerResponse, error) {
	// TODO: Check in parallel and use grpc streaming to inform the client.
	// TODO: Let the client define the services that should be checked.

	waitForService(ctx, s.config.DatastoreWriterHost, s.config.DatastoreWriterPort)
	waitForService(ctx, s.config.AuthHost, s.config.AuthPort)

	return new(proto.CheckServerResponse), ctx.Err()
}

// waitForService checks if the service at host:port is available.
//
// Blocks until the connection is established or the context is canceled or
// expired.
func waitForService(ctx context.Context, host, port string) {
	addr := net.JoinHostPort(host, port)
	d := net.Dialer{}
	_, err := d.DialContext(ctx, "tcp", addr)
	for err != nil && ctx.Err() == nil {
		time.Sleep(100 * time.Millisecond)
		_, err = d.DialContext(ctx, "tcp", addr)
	}
}
