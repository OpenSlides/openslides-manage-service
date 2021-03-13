package client

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/management"
	"github.com/spf13/cobra"
)

const resetPasswordHelp = `Resets the Password of an user

This command resets the password of a user by a given user id.
`

func cmdResetPassword(cfg *config) *cobra.Command {
	var userID int64
	var password string

	cmd := &cobra.Command{
		Use:   "reset_password",
		Short: "Resets a user password.",
		Long:  resetPasswordHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true

			ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
			defer cancel()

			service := connect(ctx, cfg.address)

			req := &management.ResetPasswordRequest{
				UserID:   userID,
				Password: password,
			}

			if _, err := service.ResetPassword(ctx, req); err != nil {
				return fmt.Errorf("reset password: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().Int64VarP(&userID, "user_id", "u", 1, "ID of the user account")
	cmd.Flags().StringVarP(&password, "password", "p", "admin", "New password for the user")

	return cmd
}
