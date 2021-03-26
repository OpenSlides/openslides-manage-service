package manage

import (
	"context"
	"fmt"

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

			resp, err := service.CheckServer(ctx, req)
			if err != nil {
				return fmt.Errorf("check server: %w", err)
			}

			fmt.Printf("%s (code %d)\n", resp.StatusMessage, resp.StatusCode)
			return nil
		},
	}

	cmd.Flags().BoolVar(&skipClient, "skip-client", false, "Skip checking the client")

	return cmd
}

// CheckServer sets hashes and sets the password
func (s *Server) CheckServer(ctx context.Context, in *proto.CheckServerRequest) (*proto.CheckServerResponse, error) {

	resp := new(proto.CheckServerResponse)
	resp.StatusCode = 1
	resp.StatusMessage = fmt.Sprintf("Something went wrong: %v", in.SkipClient)

	return resp, nil
}
