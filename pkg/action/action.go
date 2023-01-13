package action

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	// ActionHelp contains the short help text for the command.
	ActionHelp = "Calls an arbitrary OpenSlides action"

	// ActionHelpExtra contains the long help text for the command without
	// the headline.
	ActionHelpExtra = `This command calls an OpenSlides backend action with the given YAML or JSON
formatted payload. Provide the payload directly or use the --file flag with a
file or use this flag with - to read from stdin.`
)

// Cmd returns the subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "action name [payload]",
		Short: ActionHelp,
		Long:  ActionHelp + "\n\n" + ActionHelpExtra,
		Args:  cobra.RangeArgs(1, 2),
	}
	cp := connection.Unary(cmd)

	payloadFileHelpText := "YAML or JSON file with the payload; you can use - to provide the payload via stdin"
	payloadFile := cmd.Flags().StringP("file", "f", "", payloadFileHelpText)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		args = append(args, "") // This is to ensure that the slice always has enough values.
		action := args[0]
		payload, err := shared.InputOrFileOrStdin(args[1], *payloadFile)
		if err != nil {
			return fmt.Errorf("reading payload from positional argument or file or stdin: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), *cp.Timeout)
		defer cancel()

		cl, close, err := connection.Dial(ctx, *cp.Addr, *cp.PasswordFile, !*cp.NoSSL)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if err := Run(ctx, cl, action, payload); err != nil {
			return fmt.Errorf("run backend action: %w", err)
		}
		return nil
	}
	return cmd
}

// Client

type gRPCClient interface {
	Action(ctx context.Context, in *proto.ActionRequest, opts ...grpc.CallOption) (*proto.ActionResponse, error)
}

// Run calls respective procedure via given gRPC client.
func Run(ctx context.Context, gc gRPCClient, action string, payload []byte) error {
	in := &proto.ActionRequest{
		Action:  action,
		Payload: payload,
	}

	resp, err := gc.Action(ctx, in)
	if err != nil {
		s, _ := status.FromError(err) // The ok value does not matter here.
		return fmt.Errorf("calling manage service (calling backend action): %s", s.Message())
	}
	fmt.Printf("Request was successful with following response: %s\n", string(resp.Payload))
	return nil
}

// Server

type backendAction interface {
	Single(ctx context.Context, name string, data json.RawMessage) (json.RawMessage, error)
}

// Action calls the given backend action with the given payload.
// This function is the server side entrypoint for this package.
func Action(ctx context.Context, in *proto.ActionRequest, ba backendAction) (*proto.ActionResponse, error) {
	name := in.Action

	c, err := yaml.YAMLToJSON(in.Payload)
	if err != nil {
		return nil, fmt.Errorf("converting YAML to JSON: %w", err)
	}

	result, err := ba.Single(ctx, name, c)
	if err != nil {
		return nil, fmt.Errorf("requesting backend action %q: %w", name, err)
	}
	return &proto.ActionResponse{Payload: result}, nil
}
