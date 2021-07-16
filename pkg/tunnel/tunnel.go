package tunnel

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	// TunnelHelp contains the short help text for the command.
	TunnelHelp = "Creates tcp tunnels to the OpenSlides network"

	// TunnelHelpExtra contains the long help text for the command without the headline.
	TunnelHelpExtra = `Opens local ports to all services and creates
tunnels into the OpenSlides network to the services they belong to.

Without any argument the command creates tunnels to all services at their
default ports. To specify tunnels, the flag "-L" can be used or the name
of a known service as argument.

The syntax of the -L argument is the same as "ssh -L". The argument can be
used more then one time.

Example:

Open tunnels to all known services:
$ manage tunnel

Open tunnels to the datastore and auth

$ manage tunnel datastore-reader datastore-writer auth

Open a tunnel to auth on localhost:8080
$ manage tunnel -L localhost:8080:auth:9004
`
)

// Cmd returns the set-password subcommand.
func Cmd() *cobra.Command {
	services := map[string]string{
		"message-bus":       ":6379:message-bus:6379",
		"backend-action":    ":9002:backend:9002",
		"backend-presenter": ":9003:backend:9003",
		"auth":              ":9004:auth:9004",
		"media":             ":9006:media:9006",
		"datastore-reader":  ":9010:datastore-reader:9010",
		"datastore-writer":  ":9011:datastore-writer:9011",
		"autoupdate":        ":9012:autoupdate:9012",
		"postgres":          ":5432:postgres:5432",
		"cache":             ":6379:cache:6379", // TODO: Use another port.
		// TODO: Add voting and icc.
	}

	var serviceNames []string
	for service := range services {
		serviceNames = append(serviceNames, service)
	}

	cmd := &cobra.Command{
		Use:       "tunnel",
		Short:     TunnelHelp,
		Long:      TunnelHelp + "\n\n" + TunnelHelpExtra,
		Args:      cobra.OnlyValidArgs,
		ValidArgs: serviceNames,
	}

	addrsF := cmd.Flags().StringArrayP("addr", "L", nil, "[bind_address:]port:host:hostport")
	addr := cmd.Flags().StringP("address", "a", connection.DefaultAddr, "address of the OpenSlides manage service")
	timeout := cmd.Flags().DurationP("timeout", "t", connection.DefaultTimeout, "time to wait for the command's response")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		defer cancel()

		cl, close, err := connection.Dial(ctx, *addr)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if err := Run(ctx, cl, args, *addrsF); err != nil {
			return fmt.Errorf("opening tunnel: %w", err)
		}
		return nil
	}
	return cmd
}

// Client

type gRPCClient interface {
	Tunnel(ctx context.Context, opts ...grpc.CallOption) (proto.Manage_TunnelClient, error)
}

// Run calls respective procedure to open a tunnel to one or more services.
func Run(ctx context.Context, gc gRPCClient, serviceNames []string, addrF []string) error {
	fmt.Printf("Service names: %v\n", serviceNames)
	fmt.Printf("Bindings: %v\n", addrF)
	// TODO: https://github.com/OpenSlides/openslides-manage-service/blob/main/pkg/manage/cmd_tunnel.go
	return nil
}

// Server

// TODO
