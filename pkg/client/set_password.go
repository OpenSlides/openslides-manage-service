package client

import (
	"context"
	"fmt"

	pb "github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
)

const setPasswordHelp = `Sets the password of an user

This command sets the password of a user by a given user id.
`

func cmdSetPassword(cfg *config) *cobra.Command {
	var userID int64
	var password string

	cmd := &cobra.Command{
		Use:   "set-password",
		Short: "Sets an user password.",
		Long:  setPasswordHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true

			ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
			defer cancel()

			service := connect(ctx, cfg.address)

			req := &pb.SetPasswordRequest{
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
