package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RunClient is the entrypoint for the client tool of this service. It starts the root command.
func RunClient() error {
	if err := rootCmd().Execute(); err != nil {
		return fmt.Errorf("executing root command: %w", err)
	}
	return nil
}

const rootHelp = `openslides is an admin tool to setup an OpenSlides instance and perform manager actions on it.`

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "openslides",
		Short:        "Swiss army knife for OpenSlides admins",
		Long:         rootHelp,
		SilenceUsage: true,
	}

	// Add subcommands here.
	// cmd.AddCommand(
	//)

	return cmd
}
