package manage

import (
	"context"
	_ "embed" // Blank import required to use go directive.
	"encoding/json"
	"fmt"
	"os"

	"github.com/OpenSlides/openslides-manage-service/pkg/datastore"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
)

const initialDataHelp = `Creates initial data if there is an empty datastore

This command also sets password of user 1 to the value in the docker secret "admin".

It does nothing if the datastore is not empty.
`

//go:embed default-initial-data.json
var defaultInitialData []byte

const adminSecretPath = "/run/secrets/admin"

// CmdInitialData creates given initial data if there is an empty datastore. It
// also sets password of user 1 to the value in the docker secret "admin".
// This does nothing if the datastore is not empty.
func CmdInitialData(cfg *ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initial-data",
		Short: "Creates initial data if there is an empty datastore",
		Long:  initialDataHelp,
	}

	path := cmd.Flags().StringP("file", "f", "", "JSON-formated file with initial data")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()

		service, close, err := Dial(ctx, cfg.Address)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		iniD := defaultInitialData
		if *path != "" {
			c, err := os.ReadFile(*path)
			if err != nil {
				return fmt.Errorf("reading initial data file `%s`: %w", *path, err)
			}
			iniD = c
		}
		req := &proto.InitialDataRequest{
			Data: iniD,
		}

		resp, err := service.InitialData(ctx, req)
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

	return cmd
}

// InitialData sets initial data in datastore.
func (s *Server) InitialData(ctx context.Context, in *proto.InitialDataRequest) (*proto.InitialDataResponse, error) {
	exists, err := datastore.Exists(ctx, s.config.DatastoreReaderURL(), "organisation", 1)
	if err != nil {
		return nil, fmt.Errorf("checking existance in datastore: %w", err)
	}
	if exists {
		return &proto.InitialDataResponse{Initialized: false}, nil
	}

	initialData, err := parseData(in.Data)
	if err != nil {
		return nil, fmt.Errorf("parsing initial data: %w", err)
	}

	for collection, elements := range initialData {
		for id, fields := range elements {
			fqid := fmt.Sprintf("%s/%s", collection, id)
			if err := datastore.Create(ctx, s.config.DatastoreWriterURL(), fqid, fields); err != nil {
				return nil, fmt.Errorf("creating datastore object `%s`: %w", fqid, err)
			}
		}
	}

	if err := s.setAdminPassword(ctx); err != nil {
		return nil, fmt.Errorf("setting admin password: %w", err)
	}

	return &proto.InitialDataResponse{Initialized: true}, nil
}

// parseData takes a JSON encoded string and transforms it into a map of FQField and value.
func parseData(d []byte) (map[string]map[string]map[string]json.RawMessage, error) {
	var data map[string]map[string]map[string]json.RawMessage

	if err := json.Unmarshal(d, &data); err != nil {
		return nil, fmt.Errorf("unmarshaling JSON: %w", err)
	}

	return data, nil
}

// setAdminPassword reads the docker secret "admin" and sets the password
// for user 1 to this value.
func (s *Server) setAdminPassword(ctx context.Context) error {
	sec, err := os.ReadFile(adminSecretPath)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", adminSecretPath, err)
	}

	in := &proto.SetPasswordRequest{
		UserID:   1,
		Password: string(sec),
	}
	if _, err := s.SetPassword(ctx, in); err != nil {
		return fmt.Errorf("setting password: %w", err)
	}

	return nil
}
