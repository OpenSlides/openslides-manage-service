package initialdata

import (
	"context"
	_ "embed" // Blank import required to use go directive.
	"encoding/json"
	"fmt"
	"os"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
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

		if err := Run(cmd.Context(), c, *dataFile); err != nil {
			return fmt.Errorf("setting initial data: %w", err)
		}

		return nil
	}
	return cmd
}

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

type datastore interface {
	Exists(collection string, id int) (bool, error)
	Create(fqid string, fields map[string]json.RawMessage) error
	Set(fqfield string, value json.RawMessage) error
}

type auth interface {
	Hash(password string) (string, error)
}

// InitialData sets initial data in the datastore.
func InitialData(ctx context.Context, in *proto.InitialDataRequest, ds datastore) (*proto.InitialDataResponse, error) {
	exists, err := CheckDatastore(ds)
	if err != nil {
		return nil, fmt.Errorf("checking existance in datastore: %w", err)
	}
	if exists {
		return &proto.InitialDataResponse{Initialized: false}, nil
	}

	if err := InsertIntoDatastore(ds, in.Data); err != nil {
		return nil, fmt.Errorf("inserting initial data into datastore: %w", err)
	}

	// if err := SetAdminPassword(ds); err != nil {
	// 	return nil, fmt.Errorf("setting admin password: %w", err)
	// }

	return &proto.InitialDataResponse{Initialized: true}, nil
}

// CheckDatastore checks if the object organization/1 exists in the datastore.
func CheckDatastore(ds datastore) (bool, error) {
	exists, err := ds.Exists("organization", 1)
	if err != nil {
		return false, fmt.Errorf("checking existance in datastore: %w", err)
	}
	return exists, nil
}

// InsertIntoDatastore inserts the given JSON data into datastore with write requests.
func InsertIntoDatastore(ds datastore, data []byte) error {
	var parsedData map[string]map[string]map[string]json.RawMessage
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return fmt.Errorf("unmarshaling JSON: %w", err)
	}

	for collection, elements := range parsedData {
		for id, fields := range elements {
			fqid := fmt.Sprintf("%s/%s", collection, id)
			if err := ds.Create(fqid, fields); err != nil {
				return fmt.Errorf("creating datastore object %q: %w", fqid, err)
			}
		}
	}

	return nil
}

// SetSuperadminPassword sets the first password for the superadmin according to respective secret.
func SetSuperadminPassword(superadminSecretFile string, ds datastore, auth auth) error {
	sapw, err := os.ReadFile(superadminSecretFile)
	if err != nil {
		return fmt.Errorf("reading file %q: %w", superadminSecretFile, err)
	}

	in := &proto.SetPasswordRequest{
		UserID:   1,
		Password: string(sapw),
	}
	if _, err := setpassword.SetPassword(in, ds, auth); err != nil {
		return fmt.Errorf("setting superadmin password: %w", err)
	}
	return nil
}
