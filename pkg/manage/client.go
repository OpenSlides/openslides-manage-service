package manage

import (
	"context"
	"fmt"
	"time"

	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const rootHelp = `manage is an admin tool to perform manager actions on an OpenSlides instance.`

func cmdRoot(cfg *ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "manage",
		Short:        "manage swiss army knife for OpenSlides admins.",
		Long:         rootHelp,
		SilenceUsage: true,
	}

	cmd.PersistentFlags().StringVarP(&cfg.Address, "address", "a", "localhost:9008", "Address of the OpenSlides manage service.")
	cmd.PersistentFlags().DurationVarP(&cfg.Timeout, "timeout", "t", 5*time.Second, "Time to wait for the command's response.")

	return cmd
}

// RunClient starts the root command.
func RunClient() error {
	cfg := new(ClientConfig)
	cmd := cmdRoot(cfg)
	cmd.AddCommand(
		CmdSetup(cfg),
		CmdCompose(cfg),
		CmdCreateUser(cfg),
		CmdSetPassword(cfg),
		CmdCheckServer(cfg),
		CmdConfig(cfg),
	)
	return cmd.Execute()
}

// ClientConfig holds the top level arguments.
type ClientConfig struct {
	Address string
	Timeout time.Duration
}

// Dial creates a grpc connection to the server.
func Dial(ctx context.Context, address string) (proto.ManageClient, func() error, error) {
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, fmt.Errorf("creating gRPC client connection with grpc.DialContect(): %w", err)
	}
	return proto.NewManageClient(conn), conn.Close, nil
}
