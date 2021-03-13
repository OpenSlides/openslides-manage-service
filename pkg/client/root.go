package client

import (
	"time"

	"github.com/spf13/cobra"
)

const rootHelp = `manage is an admin tool to perform manager actions on an OpenSlides instance.`

type config struct {
	address string
	timeout time.Duration
}

// func (c config) timeout() (time.Duration, error) {
// 	return time.ParseDuration(c.rawTimeout)
// }

func cmdRoot(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manage",
		Short: "manage swiss army knife for OpenSlides admins.",
		Long:  rootHelp,
	}

	cmd.PersistentFlags().StringVarP(&cfg.address, "address", "a", "localhost:8001", "Address of the OpenSlides manage service.")
	cmd.PersistentFlags().DurationVarP(&cfg.timeout, "timeout", "t", time.Second, "Time to wait for the command's response.")

	return cmd
}

// Execute starts the root command.
func Execute() error {
	cfg := new(config)
	cmd := cmdRoot(cfg)
	cmd.AddCommand(
		cmdCreateUser(cfg),
		cmdSetPassword(cfg),
	)
	return cmd.Execute()
}
