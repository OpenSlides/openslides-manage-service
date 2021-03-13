package client

import (
	"time"

	"github.com/spf13/cobra"
)

const rootHelp = `manage is an admin tool to perform manager actions at a OpenSlides instance.`

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

	cmd.PersistentFlags().StringVarP(&cfg.address, "address", "a", "localhost:8001", "Host of the OpenSlides service.")
	cmd.PersistentFlags().DurationVarP(&cfg.timeout, "timeout", "t", time.Second, "Time to wait for the command.")

	return cmd
}

// Execute starts the root command.
func Execute() error {
	cfg := new(config)
	cmd := cmdRoot(cfg)
	cmd.AddCommand(
		cmdCreateUser(cfg),
		cmdResetPassword(cfg),
	)

	return cmd.Execute()
}
