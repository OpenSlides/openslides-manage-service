package setpassword

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	// SetPasswordHelp contains the short help text for the command.
	SetPasswordHelp = "Sets the password of an user in OpenSlides"

	// SetPasswordHelpExtra contains the long help text for the command without
	// the headline.
	SetPasswordHelpExtra = "This command sets the password of an user by a given user ID."
)

// Cmd returns the subcommand.
func Cmd(cmd *cobra.Command, cfg connection.Params) *cobra.Command {
	cmd.Use = "set-password"
	cmd.Short = SetPasswordHelp
	cmd.Long = SetPasswordHelp + "\n\n" + SetPasswordHelpExtra
	cmd.Args = cobra.NoArgs

	userID := cmd.Flags().Int64P("user_id", "u", 0, "ID of the user account")
	cmd.MarkFlagRequired("user_id")
	password := cmd.Flags().StringP("password", "p", "", "new password of the user")
	cmd.MarkFlagRequired("password")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout())
		defer cancel()

		cl, close, err := connection.Dial(ctx, cfg.Addr(), cfg.PasswordFile(), !cfg.NoSSL())
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		if err := Run(ctx, cl, *userID, *password); err != nil {
			return fmt.Errorf("setting password: %w", err)
		}
		return nil
	}
	return cmd
}

// Client

type gRPCClient interface {
	SetPassword(ctx context.Context, in *proto.SetPasswordRequest, opts ...grpc.CallOption) (*proto.SetPasswordResponse, error)
}

// Run calls respective procedure to set password of the given user.
func Run(ctx context.Context, gc gRPCClient, userID int64, password string) error {
	in := &proto.SetPasswordRequest{
		UserID:   userID,
		Password: password,
	}
	if _, err := gc.SetPassword(ctx, in); err != nil {
		s, _ := status.FromError(err) // The ok value does not matter here.
		return fmt.Errorf("calling manage service (setting password of user %d): %s", userID, s.Message())
	}
	return nil
}

// Server

type action interface {
	Single(ctx context.Context, name string, data json.RawMessage) (json.RawMessage, error)
}

// SetPassword gets the hash and sets the password for the given user.
// This function is the server side entrypoint for this package.
func SetPassword(ctx context.Context, in *proto.SetPasswordRequest, a action) (*proto.SetPasswordResponse, error) {
	if err := Execute(ctx, in.UserID, in.Password, a); err != nil {
		return nil, fmt.Errorf("setting password for user %d: %w", in.UserID, err)
	}
	return &proto.SetPasswordResponse{}, nil
}

// Execute gets the hash and sets the password for the given user.
func Execute(ctx context.Context, userID int64, password string, a action) error {
	name := "user.set_password"
	payload := []struct {
		ID       int64  `json:"id"`
		Password string `json:"password"`
	}{
		{
			ID:       userID,
			Password: password,
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshalling action data: %w", err)
	}
	if _, err := a.Single(ctx, name, data); err != nil {
		// There is no action result in success case.
		return fmt.Errorf("requesting backend action %q: %w", name, err)
	}
	return nil
}
