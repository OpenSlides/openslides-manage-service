package version

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	// VersionHelp contains the short help text for the command.
	VersionHelp = "Retrieves OpenSlides client version tag"

	// VersionHelpExtra contains the long help text for the command without
	// the headline.
	VersionHelpExtra = ``
)

// Cmd returns the subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: VersionHelp,
		Long:  VersionHelp + "\n\n" + VersionHelpExtra,
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
			return fmt.Errorf("run version call: %w", err)
		}
		return nil
	}
	return cmd
}

// Client

type gRPCClient interface {
	Version(ctx context.Context, in *proto.VersionRequest, opts ...grpc.CallOption) (*proto.VersionResponse, error)
}

// Run calls respective procedure via given gRPC client.
func Run(ctx context.Context, gc gRPCClient) error {
	in := &proto.VersionRequest{}

	resp, err := gc.Version(ctx, in)
	if err != nil {
		s, _ := status.FromError(err) // The ok value does not matter here.
		return fmt.Errorf("calling manage service (retrieving version): %s", s.Message())
	}
	fmt.Printf(resp.Version)
	return nil
}
