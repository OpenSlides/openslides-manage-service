package initialdata

import (
	"context"
	_ "embed" // Blank import required to use go directive.
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
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

//go:embed default-initial-data.json
// DefaultInitialData contains default initial data as JSON
var DefaultInitialData []byte

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
// If dataFile is an empty string, the default initial data are used.
func Run(ctx context.Context, gc gRPCClient, dataFile string) error {
	iniD := DefaultInitialData
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

type datastore interface {
	Exists(ctx context.Context, collection string, id int) (bool, error)
	Create(ctx context.Context, creatables map[string]map[string]json.RawMessage, migrationIndex int) error
	Set(ctx context.Context, fqfield string, value json.RawMessage) error
}

type auth interface {
	Hash(ctx context.Context, password string) (string, error)
}

// InitialData sets initial data in the datastore.
func InitialData(ctx context.Context, in *proto.InitialDataRequest, runPath string, ds datastore, auth auth) (*proto.InitialDataResponse, error) {
	exists, err := CheckDatastore(ctx, ds)
	if err != nil {
		return nil, fmt.Errorf("checking existance in datastore: %w", err)
	}
	if exists {
		return &proto.InitialDataResponse{Initialized: false}, nil
	}

	if err := InsertIntoDatastore(ctx, ds, in.Data); err != nil {
		return nil, fmt.Errorf("inserting initial data into datastore: %w", err)
	}

	p := path.Join(runPath, setup.SecretsDirName, setup.SuperadminFileName)
	if err := SetSuperadminPassword(ctx, p, ds, auth); err != nil {
		return nil, fmt.Errorf("setting superadmin password: %w", err)
	}

	return &proto.InitialDataResponse{Initialized: true}, nil
}

// CheckDatastore checks if the object organization/1 exists in the datastore.
func CheckDatastore(ctx context.Context, ds datastore) (bool, error) {
	exists, err := ds.Exists(ctx, "organization", 1)
	if err != nil {
		return false, fmt.Errorf("checking existance in datastore: %w", err)
	}
	return exists, nil
}

// InsertIntoDatastore inserts the given JSON data into datastore with write requests.
//
// The given data should be formatted like this: '{"collection_a": {"1": {"field_a": "foo", "field_b": "bar"}}}'
// but the datastore.Create expected a map with FQId like this: '{"collection_a/1": {"field_a": "foo", "field_b": "bar"}}'
// so this function also transforms the data into correct format.
func InsertIntoDatastore(ctx context.Context, ds datastore, data []byte) error {
	var decodedData map[string]json.RawMessage
	if err := json.Unmarshal(data, &decodedData); err != nil {
		return fmt.Errorf("unmarshaling JSON: %w", err)
	}

	creatables := make(map[string]map[string]json.RawMessage)
	migrationIndex := -1
	for collection, value := range decodedData {
		if collection == "_migration_index" {
			if err := json.Unmarshal(value, &migrationIndex); err != nil {
				return fmt.Errorf("unmarshaling JSON: %w", err)
			}
			continue
		}
		var elements map[string]map[string]json.RawMessage
		if err := json.Unmarshal(value, &elements); err != nil {
			return fmt.Errorf("unmarshaling JSON: %w", err)
		}
		for id, fields := range elements {
			fqid := fmt.Sprintf("%s/%s", collection, id)
			creatables[fqid] = fields
		}
	}

	if err := ds.Create(ctx, creatables, migrationIndex); err != nil {
		return fmt.Errorf("creating datastore objects: %w", err)
	}

	return nil
}

// SetSuperadminPassword sets the first password for the superadmin according to respective secret.
func SetSuperadminPassword(ctx context.Context, superadminSecretFile string, ds datastore, auth auth) error {
	sapw, err := os.ReadFile(superadminSecretFile)
	if err != nil {
		return fmt.Errorf("reading file %q: %w", superadminSecretFile, err)
	}
	_ = sapw
	// if err := setpassword.Execute(ctx, 1, string(sapw), ds, auth); err != nil {
	// 	return fmt.Errorf("setting superadmin password: %w", err)
	// }
	return nil
}
