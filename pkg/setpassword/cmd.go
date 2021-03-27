package setpassword

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/pkg/util"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
)

const setPasswordHelp = `Sets the password of an user

This command sets the password of a user by a given user id.
`

// Command initializes the set-password command.
func Command(cfg *util.ClientConfig) *cobra.Command {
	var userID int64
	var password string

	cmd := &cobra.Command{
		Use:   "set-password",
		Short: "Sets an user password.",
		Long:  setPasswordHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
			defer cancel()

			service := util.Connect(ctx, cfg.Address)

			req := &proto.SetPasswordRequest{
				UserID:   userID,
				Password: password,
			}

			if _, err := service.SetPassword(ctx, req); err != nil {
				return fmt.Errorf("reset password: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().Int64VarP(&userID, "user_id", "u", 1, "ID of the user account")
	cmd.Flags().StringVarP(&password, "password", "p", "admin", "New password for the user")

	return cmd
}
