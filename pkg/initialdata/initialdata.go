package initialdata

import (
	"context"
	_ "embed" // Blank import required to use go directive.
	"fmt"
	"os"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
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

//go:embed default-initial-data.json
// DefaultInitialData contains default initial data as JSON
var DefaultInitialData []byte

// Cmd returns the initial-data subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initial-data",
		Short: InitialDataHelp,
		Long:  InitialDataHelp + "\n\n" + InitialDataHelpExtra,
		Args:  cobra.NoArgs,
	}
	dataFile := cmd.Flags().StringP("file", "f", "", "custom JSON file with initial data")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		c, close, err := connection.Dial(cmd.Context())
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if err := Initialdata(cmd.Context(), c, *dataFile); err != nil {
			return fmt.Errorf("setting initial data: %w", err)
		}

		return nil
	}
	return cmd
}

type gRPCClient interface {
	InitialData(ctx context.Context, in *proto.InitialDataRequest, opts ...grpc.CallOption) (*proto.InitialDataResponse, error)
}

// Initialdata calls respective procedure to set initial data to an empty database via given gRPC client.
// If dataFile is an empty string, the default initial data are used.
func Initialdata(ctx context.Context, c gRPCClient, dataFile string) error {
	iniD := DefaultInitialData
	if dataFile != "" {
		c, err := os.ReadFile(dataFile)
		if err != nil {
			return fmt.Errorf("reading initial data file %q: %w", dataFile, err)
		}
		iniD = c
	}
	req := &proto.InitialDataRequest{
		Data: iniD,
	}

	resp, err := c.InitialData(ctx, req)
	if err != nil {
		return fmt.Errorf("setting initial data: %w", err)
	}

	msg := "Datastore contains data. Initial data were NOT set."
	if resp.Initialized {
		msg = "Initial data were set successfully."
	}
	fmt.Println(msg)

	return nil
}
