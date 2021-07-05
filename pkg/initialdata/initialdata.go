package initialdata

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	// InitialDataHelp contains the short help text for the command.
	InitialDataHelp = "Creates initial data if there is an empty datastore"

	// InitialDataHelpExtra contains the long help text for the command without the headline.
	InitialDataHelpExtra = `This command also sets password of user 1 to the value in the docker
secret "admin". It does nothing if the datastore is not empty.`
)

// Cmd returns the initial-data subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initial-data",
		Short: InitialDataHelp,
		Long:  InitialDataHelp + "\n\n" + InitialDataHelpExtra,
		Args:  cobra.NoArgs,
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}
	return cmd
}

type gRPCClient interface {
	InitialData(ctx context.Context, in *proto.InitialDataRequest, opts ...grpc.CallOption) (*proto.InitialDataResponse, error)
}

// Initialdata calls respective procedure to set initial data to an empty database via given gRPC client.
func Initialdata(ctx context.Context, c gRPCClient) error {
	req := &proto.InitialDataRequest{
		Data: []byte("harr"),
	}
	_, err := c.InitialData(ctx, req)
	if err != nil {
		return fmt.Errorf("setting initial data: %w", err)
	}

	// ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	// defer cancel()

	// service, close, err := Dial(ctx, cfg.Address)
	// if err != nil {
	// 	return fmt.Errorf("connecting to gRPC server: %w", err)
	// }
	// defer close()

	// iniD := defaultInitialData
	// if *path != "" {
	// 	c, err := os.ReadFile(*path)
	// 	if err != nil {
	// 		return fmt.Errorf("reading initial data file `%s`: %w", *path, err)
	// 	}
	// 	iniD = c
	// }
	// req := &proto.InitialDataRequest{
	// 	Data: iniD,
	// }

	// resp, err := service.InitialData(ctx, req)
	// if err != nil {
	// 	return fmt.Errorf("setting initial data: %w", err)
	// }

	// msg := "Datastore contains data. Initial data were NOT set."
	// if resp.Initialized {
	// 	msg = "Initial data were set successfully."
	// }
	// fmt.Println(msg)

	return nil
}
