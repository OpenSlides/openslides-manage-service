package initialdata

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/pkg/fehler"
	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	// InitialDataHelp contains the short help text for the command.
	InitialDataHelp = "Creates initial data if there is an empty datastore"

	// InitialDataHelpExtra contains the long help text for the command without
	// the headline.
	InitialDataHelpExtra = `This command also sets password of user 1 to the value of the docker secret
"superadmin" which is "superadmin" by default. It returns an error if the
datastore is not empty.`
)

// Cmd returns the subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initial-data",
		Short: InitialDataHelp,
		Long:  InitialDataHelp + "\n\n" + InitialDataHelpExtra,
		Args:  cobra.NoArgs,
	}
	cp := connection.Unary(cmd)

	dataFileHelpText := "custom JSON file with initial data; you can use - to provide the data via stdin"
	dataFile := cmd.Flags().StringP("file", "f", "", dataFileHelpText)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var data []byte
		if *dataFile != "" {
			d, err := shared.ReadFromFileOrStdin(*dataFile)
			if err != nil {
				return fmt.Errorf("reading initial-data file: %w", err)
			}
			data = d
		}

		ctx, cancel := context.WithTimeout(context.Background(), *cp.Timeout)
		defer cancel()

		cl, close, err := connection.Dial(ctx, *cp.Addr, *cp.PasswordFile, !*cp.NoSSL)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if err := Run(ctx, cl, data); err != nil {
			return fmt.Errorf("setting initial data: %w", err)
		}
		return nil
	}
	return cmd
}

// Client

type gRPCClient interface {
	InitialData(ctx context.Context, in *proto.InitialDataRequest, opts ...grpc.CallOption) (*proto.InitialDataResponse, error)
}

// Run calls respective procedure to set initial data to an empty database via given gRPC client.
func Run(ctx context.Context, gc gRPCClient, data []byte) error {
	req := &proto.InitialDataRequest{
		Data: data,
	}

	resp, err := gc.InitialData(ctx, req)
	if err != nil {
		s, _ := status.FromError(err) // The ok value does not matter here.
		return fmt.Errorf("calling manage service (setting initial data): %s", s.Message())
	}
	if !resp.Initialized {
		return fehler.ExitCode(2, fmt.Errorf("datastore contains data, initial data were NOT set"))
	}
	fmt.Println("Initial data were set successfully.")
	return nil
}

// Server

const datastoreNotEmptyMsg = "Datastore is not empty"

type backendAction interface {
	Single(ctx context.Context, name string, data json.RawMessage) (json.RawMessage, error)
}

// InitialData sets initial data in the datastore.
func InitialData(ctx context.Context, in *proto.InitialDataRequest, superadminSecretFile string, ba backendAction) (*proto.InitialDataResponse, error) {
	initialData := in.Data
	if initialData == nil {
		// The backend expects at least an empty object.
		initialData = []byte("{}")
	} else {
		// TODO: validate data
	}

	name := "organization.initial_import"
	payload := []struct {
		Data json.RawMessage `json:"data"`
	}{
		{
			Data: json.RawMessage(initialData),
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling action data: %w", err)
	}

	if _, err := ba.Single(ctx, name, data); err != nil {
		// There is no action result in success case.
		if strings.Contains(err.Error(), datastoreNotEmptyMsg) {
			return &proto.InitialDataResponse{Initialized: false}, nil
		}
		return nil, fmt.Errorf("requesting backend action %q: %w", name, err)
	}

	if err := SetSuperadminPassword(ctx, superadminSecretFile, ba); err != nil {
		return nil, fmt.Errorf("setting superadmin password: %w", err)
	}

	return &proto.InitialDataResponse{Initialized: true}, nil
}

// SetSuperadminPassword sets the first password for the superadmin according to respective secret.
func SetSuperadminPassword(ctx context.Context, superadminSecretFile string, ba backendAction) error {
	sapw, err := os.ReadFile(superadminSecretFile)
	if err != nil {
		return fmt.Errorf("reading file %q: %w", superadminSecretFile, err)
	}
	if err := setpassword.Execute(ctx, 1, string(sapw), ba); err != nil {
		return fmt.Errorf("setting superadmin password: %w", err)
	}
	return nil
}
