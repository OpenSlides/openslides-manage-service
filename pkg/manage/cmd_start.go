package manage

import (
	"context"
	"fmt"
	"io/fs"
	"os"

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

		if err := createSecrets(); err != nil {
			return fmt.Errorf("creating secrets: %w", err)
		}

		return nil
	}

	return cmd
}

// create Secrets creates random values uses as secrets in Docker Compose file.
func createSecrets() error {
	path := "secrets/"
	if err := os.MkdirAll(path, fs.ModePerm); err != nil {
		return fmt.Errorf("creating directory `%s`: %w", path, err)
	}

	secrets := []string{
		"auth_token_key",
		"auth_cookie_key",
	}
	for _, s := range secrets {
		f, err := os.Create(path + s)
		if err != nil {
			return fmt.Errorf("creating file `%s`: %w", path+s, err)
		}
		defer f.Close()

		r := "my random value" // TODO: Use a random value here.
		if _, err := f.WriteString(r); err != nil {
			return fmt.Errorf("writing secret to file `%s`: %w", f.Name(), err)
		}
	}

	return nil
}
