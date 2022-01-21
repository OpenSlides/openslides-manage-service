package set

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	// SetHelp contains the short help text for the command.
	SetHelp = "Calls an OpenSlides action to set some settings"

	// SetHelpExtra contains the long help text for the command without
	// the headline.
	SetHelpExtra = `This command calls an OpenSlides backend action with the given payload. Only
some actions to update organization or meeting settings are supportet.`
)

// Cmd returns the subcommand.
func Cmd(cmd *cobra.Command, cfg connection.Params) *cobra.Command {
	cmd.Use = "set"
	cmd.Short = SetHelp
	cmd.Long = SetHelp + "\n\n" + SetHelpExtra
	cmd.Args = cobra.ExactArgs(2)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		action := args[0]
		fileName := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout())
		defer cancel()

		cl, close, err := connection.Dial(ctx, cfg.Addr(), cfg.PasswordFile(), !cfg.NoSSL())
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if err := Run(ctx, cl, action, fileName); err != nil {
			return fmt.Errorf("run backend action: %w", err)
		}
		return nil
	}
	return cmd
}

// Client

type gRPCClient interface {
	Set(ctx context.Context, in *proto.SetRequest, opts ...grpc.CallOption) (*proto.SetResponse, error)
}

// Run calls respective procedure via given gRPC client.
func Run(ctx context.Context, gc gRPCClient, action, payloadFile string) error {
	actionMap := map[string]string{
		"organization": "organization.update",
	}

	actionName, ok := actionMap[action]
	if !ok {
		return fmt.Errorf("unknown action %q", action)
	}

	payload, err := os.ReadFile(payloadFile)
	if err != nil {
		return fmt.Errorf("reading payload file %q: %w", payloadFile, err)
	}

	in := &proto.SetRequest{
		Action:  actionName,
		Payload: payload,
	}
	resp, err := gc.Set(ctx, in)
	if err != nil {
		s, _ := status.FromError(err) // The ok value does not matter here.
		return fmt.Errorf("calling manage service (calling backend action): %s", s.Message())
	}

	fmt.Printf("Request was successful. Response: %s", string(resp.Payload))
	return nil
}

// Server

type action interface {
	Single(ctx context.Context, name string, data json.RawMessage) (json.RawMessage, error)
}

// Set calls the given backend action with the given payload.
// This function is the server side entrypoint for this package.
func Set(ctx context.Context, in *proto.SetRequest, a action) (*proto.SetResponse, error) {
	name := in.Action
	result, err := a.Single(ctx, name, in.Payload)
	if err != nil {
		return nil, fmt.Errorf("requesting backend action %q: %w", name, err)
	}
	return &proto.SetResponse{Payload: result}, nil
}
