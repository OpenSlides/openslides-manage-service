package setpassword

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	// SetPasswordHelp contains the short help text for the command.
	SetPasswordHelp = "Sets the password of an user"

	// SetPasswordHelpExtra contains the long help text for the command without the headline.
	SetPasswordHelpExtra = "This command sets the password of an user by a given user ID."
)

// Cmd returns the set-password subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-password",
		Short: SetPasswordHelp,
		Long:  SetPasswordHelp + "\n\n" + SetPasswordHelpExtra,
		Args:  cobra.NoArgs,
	}

	userID := cmd.Flags().Int64P("user_id", "u", 0, "ID of the user account")
	cmd.MarkFlagRequired("user_id")
	password := cmd.Flags().StringP("password", "p", "", "New password of the user")
	cmd.MarkFlagRequired("password")
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
		return fmt.Errorf("setting password of user %d: %w", userID, err)
	}
	return nil
}

// Server

type datastore interface {
	Set(ctx context.Context, fqfield string, value json.RawMessage) error
}

type auth interface {
	Hash(ctx context.Context, password string) (string, error)
}

// SetPassword gets the hash and sets the password for the given user.
func SetPassword(ctx context.Context, in *proto.SetPasswordRequest, ds datastore, auth auth) (*proto.SetPasswordResponse, error) {
	hash, err := auth.Hash(ctx, in.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing passwort: %w", err)
	}

	key := fmt.Sprintf("user/%d/password", in.UserID)
	value := []byte(`"` + hash + `"`)
	if err := ds.Set(ctx, key, value); err != nil {
		return nil, fmt.Errorf("writing key %q to datastore: %w", key, err)
	}

	return &proto.SetPasswordResponse{}, nil
}
