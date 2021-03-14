package create_user

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/pkg/client/clientutil"
	pb "github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
)

const createUsersHelp = `Creates a user account

This command creates a user account on the server.
`

func Command(cfg *clientutil.Config) *cobra.Command {
	var username string
	var password string
	var orgaLvl string

	cmd := &cobra.Command{
		Use:   "create-user",
		Short: "Creates a user account.",
		Long:  createUsersHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
			defer cancel()

			service := clientutil.Connect(ctx, cfg.Address)

			req := &pb.CreateUserRequest{
				Username:                    username,
				Password:                    password,
				OrganisationManagementLevel: orgaLvl,
			}

			if _, err := service.CreateUser(ctx, req); err != nil {
				return fmt.Errorf("create user: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "admin", "Name of the user account")
	cmd.Flags().StringVarP(&password, "password", "p", "admin", "Password for the user")
	cmd.Flags().StringVarP(&orgaLvl, "organisation_management_level", "m", "superadmin", "Set organisation management level")

	return cmd
}
