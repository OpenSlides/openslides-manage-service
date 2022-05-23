package checkserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	// CheckServerHelp contains the short help text for the command.
	CheckServerHelp = "Checks if the server and its services are ready"

	// CheckServerHelpExtra contains the long help text for the command without
	// the headline.
	CheckServerHelpExtra = `At the moment this only checks the health route of the backendManage service.`
)

// Cmd returns the subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-server",
		Short: CheckServerHelp,
		Long:  CheckServerHelp + "\n\n" + CheckServerHelpExtra,
		Args:  cobra.NoArgs,
	}
	cp := connection.Unary(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), *cp.Timeout)
		defer cancel()

		cl, close, err := connection.Dial(ctx, *cp.Addr, *cp.PasswordFile, !*cp.NoSSL)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if err := Run(ctx, cl); err != nil {
			return fmt.Errorf("checking server: %w", err)
		}
		return nil
	}
	return cmd
}

// Client

type gRPCClient interface {
	CheckServer(ctx context.Context, in *proto.CheckServerRequest, opts ...grpc.CallOption) (*proto.CheckServerResponse, error)
}

// Run calls respective procedure to check the server via given gRPC client.
func Run(ctx context.Context, gc gRPCClient) error {
	req := &proto.CheckServerRequest{}

	for {
		resp, err := gc.CheckServer(ctx, req)
		if err != nil {
			s, _ := status.FromError(err) // The ok value does not matter here.
			return fmt.Errorf("calling manage service (checking server): %s", s.Message())
		}
		if resp.Ready {
			break
		}
	}

	// We reach this line only if the check server request was successfuls and
	// the context was not canceled (e. g. deadline exceeded).
	fmt.Println("Server is ready.")
	return nil
}

// Server

type action interface {
	Health(context.Context) (json.RawMessage, error)
}

// CheckServer sends a health request to backend manage service.
func CheckServer(ctx context.Context, in *proto.CheckServerRequest, a action) *proto.CheckServerResponse {
	_, err := a.Health(ctx)
	if err != nil {
		// Special error handling here: We do not return the (wrapped) error but
		// a response with falsy value.
		return &proto.CheckServerResponse{Ready: false}

	}
	return &proto.CheckServerResponse{Ready: true}
}
