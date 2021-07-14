package client

import (
	"fmt"

	"github.com/OpenSlides/openslides-manage-service/pkg/initialdata"
	"github.com/OpenSlides/openslides-manage-service/pkg/setup"
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
		initialdata.Cmd(),
	)

	return cmd
}
