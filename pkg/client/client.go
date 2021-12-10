package client

import (
	"fmt"
	"path"
	"time"

	"github.com/OpenSlides/openslides-manage-service/pkg/config"
	"github.com/OpenSlides/openslides-manage-service/pkg/connection"
	"github.com/OpenSlides/openslides-manage-service/pkg/createuser"
	"github.com/OpenSlides/openslides-manage-service/pkg/initialdata"
	"github.com/OpenSlides/openslides-manage-service/pkg/setpassword"
	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
	"github.com/OpenSlides/openslides-manage-service/pkg/tunnel"
	"github.com/spf13/cobra"
)

// RunClient is the entrypoint for the client tool of this service. It starts the root command.
func RunClient() error {
	if err := RootCmd().Execute(); err != nil {
		return fmt.Errorf("executing root command: %w", err)
	}
	return nil
}

// RootHelp is the main help text for the client tool.
const RootHelp = `openslides is an admin tool to setup an OpenSlides instance and perform manager actions on it.`

// RootCmd returns the root cobra command.
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "openslides",
		Short:        "Swiss army knife for OpenSlides admins",
		Long:         RootHelp,
		SilenceUsage: true,
	}

	cmd.AddCommand(
		setup.Cmd(),
		config.Cmd(),
		withConnectionFlags(initialdata.Cmd),
		withConnectionFlags(setpassword.Cmd),
		withConnectionFlags(createuser.Cmd),
		tunnel.Cmd(),
	)

	return cmd
}

type cmdFunc func(cmd *cobra.Command, addr, passwordFile *string, timeout *time.Duration, noSSL *bool) *cobra.Command

func withConnectionFlags(fn cmdFunc) *cobra.Command {
	cmd := &cobra.Command{}
	addr := cmd.Flags().StringP("address", "a", connection.DefaultAddr, "address of the OpenSlides manage service")
	defaultPasswordFile := path.Join(".", setup.SecretsDirName, setup.ManageAuthPasswordFileName)
	passwordFile := cmd.Flags().String("password-file", defaultPasswordFile, "file with password for authorization to manage service, not usable in development mode")
	timeout := cmd.Flags().DurationP("timeout", "t", connection.DefaultTimeout, "time to wait for the command's response")
	noSSL := cmd.Flags().Bool("no-ssl", false, "use an unencrypted connection to manage service")
	return fn(cmd, addr, passwordFile, timeout, noSSL)
}
