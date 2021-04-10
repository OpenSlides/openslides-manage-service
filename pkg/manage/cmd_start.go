package manage

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

const startHelp = `Builds images and starts OpenSlides with Docker Compose

This command executes the following steps to start OpenSlides:
- Create a local docker-compose.yml.
- Create local secrets for the auth service.
- Run the docker compose file with the "up" command in daemonized mode.
- TODO ...
`

// CmdStart creates docker-compose.yml, secrets, runs docker-compose up in daemonized mode and ... TODO
func CmdStart(cfg *ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Builds images and starts OpenSlides with Docker Compose.",
		Long:  startHelp,
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()

		if err := createDockerComposeYML(ctx); err != nil {
			return fmt.Errorf("creating Docker Compose YML: %w", err)
		}

		return nil
	}

	return cmd
}
