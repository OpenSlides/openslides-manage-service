package client

import (
	"time"

	"github.com/OpenSlides/openslides-manage-service/pkg/client/clientutil"
	"github.com/OpenSlides/openslides-manage-service/pkg/set_password"
	"github.com/spf13/cobra"
)

const rootHelp = `manage is an admin tool to perform manager actions on an OpenSlides instance.`

// func (c config) timeout() (time.Duration, error) {
// 	return time.ParseDuration(c.rawTimeout)
// }

func cmdRoot(cfg *clientutil.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manage",
		Short: "manage swiss army knife for OpenSlides admins.",
		Long:  rootHelp,
	}

	cmd.PersistentFlags().StringVarP(&cfg.Address, "address", "a", "localhost:8001", "Address of the OpenSlides manage service.")
	cmd.PersistentFlags().DurationVarP(&cfg.Timeout, "timeout", "t", time.Second, "Time to wait for the command's response.")

	return cmd
}

// Execute starts the root command.
func Execute() error {
	cfg := new(clientutil.Config)
	cmd := cmdRoot(cfg)
	cmd.AddCommand(
		cmdCreateUser(cfg),
		set_password.Command(cfg),
	)
	return cmd.Execute()
}
