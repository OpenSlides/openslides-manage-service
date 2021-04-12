package manage

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/pkg/datastore"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
)

const helpConfig = `Gets or sets config values

This command gets or sets the config values for an organisation.

Example:

$ manage config get electronic_voting
disabled

$ manage config set electronic_voting enabled
$ manage config get electronic_voting
enabled
`

// CmdConfig initializes the config command.
func CmdConfig(cfg *ClientConfig) *cobra.Command {
	values := []string{
		"electronic_voting",
	}

	cmd := &cobra.Command{
		Use:       "config",
		Short:     "Gets or sets config values.",
		Long:      helpConfig,
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
		v, err := getConfig(ctx, s.config, key)
		if err != nil {
			return nil, fmt.Errorf("getting config value: %w", err)
		}
		return &proto.ConfigResponse{Value: v}, nil
	}

	if err := setConfig(ctx, s.config, key, in.NewValue); err != nil {
		return nil, fmt.Errorf("setting config value: %w", err)
	}
	return &proto.ConfigResponse{}, nil
}

// getConfig fetches the organisation config value from datastore.
func getConfig(ctx context.Context, cfg *ServerConfig, key string) (string, error) {
	waitForService(ctx, cfg.DatastoreReaderURL().Host)

	var enabled bool
	if err := datastore.Get(ctx, cfg.DatastoreReaderURL(), key, &enabled); err != nil {
		return "", fmt.Errorf("getting key %s from datastore: %w", key, err)
	}

	if enabled {
		return "enabled", nil
	}
	return "disabled", nil

}

// setConfig sets the given organisation config value in datastore.
func setConfig(ctx context.Context, cfg *ServerConfig, key string, newValue string) error {
	waitForService(ctx, cfg.DatastoreReaderURL().Host)

	var value []byte
	switch newValue {
	case "enabled", "true", "1", "on":
		value = []byte("true")
	case "disabled", "false", "0", "off":
		value = []byte("false")
	default:
		return fmt.Errorf("invalid new value `%s`, expected `enabled` or `disabled` ", newValue)
	}

	if err := datastore.Set(ctx, cfg.DatastoreWriterURL(), key, value); err != nil {
		return fmt.Errorf("writing key %s to datastore: %w", key, err)
	}
	return nil
}
