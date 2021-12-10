package initialdata

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	// InitialDataHelp contains the short help text for the command.
	InitialDataHelp = "Creates initial data if there is an empty datastore"

	// InitialDataHelpExtra contains the long help text for the command without
	// the headline.
	InitialDataHelpExtra = `This command also sets password of user 1 to the value of the docker
secret "superadmin". It does nothing if the datastore is not empty.`
)

// Cmd returns the initial-data subcommand.
func Cmd(cmd *cobra.Command, cfg connection.Params) *cobra.Command {
	cmd.Use = "initial-data"
	cmd.Short = InitialDataHelp
	cmd.Long = InitialDataHelp + "\n\n" + InitialDataHelpExtra
	cmd.Args = cobra.NoArgs

	dataFile := cmd.Flags().StringP("file", "f", "", "custom JSON file with initial data")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout())
		defer cancel()

		cl, close, err := connection.Dial(ctx, cfg.Addr(), cfg.PasswordFile(), !cfg.NoSSL())
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if err := Run(ctx, cl, *dataFile); err != nil {
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
func Run(ctx context.Context, gc gRPCClient, dataFile string) error {
	var iniD []byte
	if dataFile != "" {
		content, err := os.ReadFile(dataFile)
		if err != nil {
			return fmt.Errorf("reading initial data file %q: %w", dataFile, err)
		}
		iniD = content
	}
	req := &proto.InitialDataRequest{
		Data: iniD,
	}

	resp, err := gc.InitialData(ctx, req)
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

// Server

type action interface {
	Single(ctx context.Context, name string, data json.RawMessage) (json.RawMessage, error)
}

// InitialData sets initial data in the datastore.
func InitialData(ctx context.Context, in *proto.InitialDataRequest, runPath string, a action) (*proto.InitialDataResponse, error) {
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

	if _, err := a.Single(ctx, name, data); err != nil {
		// There is no action result in success case.
		return nil, fmt.Errorf("requesting backend action %q: %w", name, err)
	}

	p := path.Join(runPath, setup.SecretsDirName, setup.SuperadminFileName)
	if err := SetSuperadminPassword(ctx, p, a); err != nil {
		return nil, fmt.Errorf("setting superadmin password: %w", err)
	}

	return &proto.InitialDataResponse{Initialized: true}, nil
}

// SetSuperadminPassword sets the first password for the superadmin according to respective secret.
func SetSuperadminPassword(ctx context.Context, superadminSecretFile string, a action) error {
	sapw, err := os.ReadFile(superadminSecretFile)
	if err != nil {
		return fmt.Errorf("reading file %q: %w", superadminSecretFile, err)
	}
	if err := setpassword.Execute(ctx, 1, string(sapw), a); err != nil {
		return fmt.Errorf("setting superadmin password: %w", err)
	}
	return nil
}
