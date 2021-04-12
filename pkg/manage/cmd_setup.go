package manage

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
)

const setupHelp = `Builds required files and docker images

This command executes the following steps to start OpenSlides:
- Create a local docker-compose.yml.
- Create local secrets for the auth service.
- Creates the services.env.
- Runs docker-compose build to build images. TODO
`

const appName = "openslides"

// CmdSetup creates docker-compose.yml, secrets and services.env. Also runs
// docker-compose build to build all images.
func CmdSetup(cfg *ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Builds the required files and docker images.",
		Long:  setupHelp,
	}

	local := cmd.Flags().BoolP("local", "l", false, "Create required files in currend working directory.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()

		dataPath := path.Join(xdg.DataHome, appName)
		if *local {
			p, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting current directory: %w", err)
			}
			dataPath = p
		}

		if err := os.MkdirAll(dataPath, fs.ModePerm); err != nil {
			return fmt.Errorf("creating directory `%s`: %w", dataPath, err)
		}

		if err := createDockerComposeYML(ctx, dataPath); err != nil {
			return fmt.Errorf("creating Docker Compose YML: %w", err)
		}

		if err := createEnvFile(dataPath); err != nil {
			return fmt.Errorf("creating .env file: %w", err)
		}

		if err := createSecrets(dataPath); err != nil {
			return fmt.Errorf("creating secrets: %w", err)
		}

		return nil
	}

	return cmd
}

// createSecrets creates random values uses as secrets in Docker Compose file.
func createSecrets(dataPath string) error {
	dataPath = path.Join(dataPath, "secrets")
	if err := os.MkdirAll(dataPath, fs.ModePerm); err != nil {
		return fmt.Errorf("creating directory `%s`: %w", dataPath, err)
	}

	secrets := []string{
		"auth_token_key",
		"auth_cookie_key",
	}
	for _, s := range secrets {
		e := func() error {
			f, err := os.Create(path.Join(dataPath, s))
			if err != nil {
				return fmt.Errorf("creating file `%s`: %w", path.Join(dataPath, s), err)
			}
			defer f.Close()

			// This creates cryptographically secure random bytes. 32 Bytes means
			// 256bit. The output can contain zero bytes.
			if _, err := io.Copy(f, io.LimitReader(rand.Reader, 32)); err != nil {
				return fmt.Errorf("creating and writing cryptographically secure random bytes: %w", err)
			}

			return nil
		}()
		if e != nil {
			return e
		}
	}

	return nil
}
