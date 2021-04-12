package manage

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
)

const helpCheckServer = `Checks if all OpenSlides services are running

This command checks if all OpenSlides services are currently running and
listening on their respective ports.
`

// CmdCheckServer checks if all OpenSlides services are running.
func CmdCheckServer(cfg *ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-server",
		Short: "Checks if all services are running.",
		Long:  helpCheckServer,
	}

	skipClient := cmd.Flags().Bool("skip-client", false, "Skip checking the client")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()

		service, close, err := Dial(ctx, cfg.Address)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		req := &proto.CheckServerRequest{
			SkipClient: *skipClient,
		}

		if _, err := service.CheckServer(ctx, req); err != nil {
			return fmt.Errorf("check server: %w", err)
		}

		return nil
	}

	return cmd
}

// CheckServer sets hashes and sets the password
func (s *Server) CheckServer(ctx context.Context, in *proto.CheckServerRequest) (*proto.CheckServerResponse, error) {
	// TODO: Use grpc streaming to inform the client.
	// TODO: Let the client define the services that should be checked.

	waitForService(ctx, s.config.DatastoreWriterURL().Host, s.config.AuthURL().Host)

	return &proto.CheckServerResponse{}, ctx.Err()
}
