package manage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/pkg/datastore"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
)

const initialDataHelp = `Creates initial data if there is an empty datastore

This command also sets admin password to "admin".

It does nothing if the datastore is not empty.
`

// CmdInitialData creates given initial data and sets admin password if there is an
// empty datastore. This does nothing if the datastore is not empty.
func CmdInitialData(cfg *ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initial-data",
		Short: "Creates initial data if there is an empty datastore.",
		Long:  initialDataHelp,
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()

		service, close, err := Dial(ctx, cfg.Address)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		// TODO: Use initial data from file given in an env variable.
		d := `{"foo":"bar"}`
		req := &proto.InitialDataRequest{
			Data: d,
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

// Sets initial data in datastore.
func (s *Server) InitialData(ctx context.Context, in *proto.InitialDataRequest) (*proto.InitialDataResponse, error) {
	initialData, err := parseData(in.Data)
	if err != nil {
		return nil, fmt.Errorf("parsing initial data: %w", err)
	}

	const magicKey = "organisation/1/id"
	var existingData bool
	if err := datastore.Get(ctx, s.config.DatastoreReaderURL(), magicKey, &existingData); err != nil {
		return nil, fmt.Errorf("reading key `%s` from datastore: %w", magicKey, err)
	}

	if existingData {
		return &proto.InitialDataResponse{Initialized: false}, nil
	}

	for k, v := range initialData {
		if err := datastore.Set(ctx, s.config.DatastoreWriterURL(), k, v); err != nil {
			return nil, fmt.Errorf("setting datastore key `%s`: %w", k, err)
		}
	}
	// TODO: Set admin password.

	return &proto.InitialDataResponse{Initialized: true}, nil
}

func parseData(d string) (map[string]json.RawMessage, error) {
	// TODO: Take JSON encoded string and validate and check it an make an nice map.
	return nil, nil
}
