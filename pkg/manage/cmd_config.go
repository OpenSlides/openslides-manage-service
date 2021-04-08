package manage

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/pkg/datastore"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
)

const configHelp = `Get or set config values

This command gets or sets the config values for an organisation.

Example:

$ manage config get electronic_voting
disabled
$ manage config set electronic_voting enabled
$ manage config electronic_voting
enabled
`

// CmdConfig initializes the config command.
func CmdConfig(cfg *ClientConfig) *cobra.Command {
	values := []string{
		"electronic_voting",
	}

	cmd := &cobra.Command{
		Use:       "config",
		Short:     "Get or set config values.",
		Long:      configHelp,
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: values,
	}

	cmd.AddCommand(&cobra.Command{
		Use:       "get",
		Short:     "Get a value",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: values,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
			defer cancel()

			service, close, err := Dial(ctx, cfg.Address)
			if err != nil {
				return fmt.Errorf("connecting to gRPC server: %w", err)
			}
			defer close()

			req := &proto.ConfigRequest{
				Field: proto.ConfigRequest_Field(proto.ConfigRequest_Field_value[args[0]]),
			}

			resp, err := service.Config(ctx, req)
			if err != nil {
				return fmt.Errorf("get config value %s: %w", args[0], err)
			}

			fmt.Println(resp.Value)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set",
		Short: "Set a value",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return err
			}
			for _, a := range values {
				if args[0] == a {
					return nil
				}
			}
			return fmt.Errorf("invalid argument %s", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
			defer cancel()

			service, close, err := Dial(ctx, cfg.Address)
			if err != nil {
				return fmt.Errorf("connecting to gRPC server: %w", err)
			}
			defer close()

			req := &proto.ConfigRequest{
				Field:    proto.ConfigRequest_Field(proto.ConfigRequest_Field_value[args[0]]),
				NewValue: args[1],
			}

			if _, err := service.Config(ctx, req); err != nil {
				return fmt.Errorf("set config value %s to %s: %w", args[0], args[1], err)
			}

			return nil
		},
	})

	return cmd
}

// Config gets or sets an organisation config value.
func (s *Server) Config(ctx context.Context, in *proto.ConfigRequest) (*proto.ConfigResponse, error) {
	var key string
	switch in.Field {
	case proto.ConfigRequest_ELECTRONIC_VOTING:
		key = "organisation/1/enable_electronic_voting"
	default:
		return nil, fmt.Errorf("Invalid request")
	}

	if in.NewValue == "" {
		// Fetch value
		waitForService(ctx, s.config.DatastoreReaderHost, s.config.DatastoreReaderPort)

		addr := fmt.Sprintf("%s://%s:%s", s.config.DatastoreReaderProtocol, s.config.DatastoreReaderHost, s.config.DatastoreReaderPort)
		var enabled bool
		if err := datastore.Get(ctx, addr, key, &enabled); err != nil {
			return nil, fmt.Errorf("getting key %s from %s: %w", key, addr, err)
		}

		value := "disabled"
		if enabled {
			value = "enabled"
		}

		return &proto.ConfigResponse{Value: value}, nil
	}

	// Write value
	waitForService(ctx, s.config.DatastoreWriterHost, s.config.DatastoreWriterPort)

	var value []byte
	switch in.NewValue {
	case "enabled", "true", "1", "on":
		value = []byte("true")
	case "disabled", "false", "0", "off":
		value = []byte("false")
	default:
		return nil, fmt.Errorf("invalid new value `%s`, expected `enabled` or `disabled` ", in.NewValue)
	}

	addr := fmt.Sprintf("%s://%s:%s", s.config.DatastoreWriterProtocol, s.config.DatastoreWriterHost, s.config.DatastoreWriterPort)
	if err := datastore.Set(ctx, addr, key, value); err != nil {
		return nil, fmt.Errorf("writing key %s to %s: %w", key, addr, err)
	}

	return &proto.ConfigResponse{}, nil
}
