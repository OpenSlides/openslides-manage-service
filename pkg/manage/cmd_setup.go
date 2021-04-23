package manage

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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
- Create local secrets for the auth service and admin password.
- Creates the services.env.
- Runs docker-compose build to build images. TODO
- Runs docker-compose up to create the container. TODO
- Creates initial data and sets admin password.

Then the container are stopped. To start them again, use start command.
`

const appName = "openslides"

// CmdSetup creates docker-compose.yml, secrets and services.env. Also runs
// docker-compose build to build all images.
func CmdSetup(cfg *ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Builds the required files and docker images",
		Long:  setupHelp,
	}

	cwd := cmd.Flags().Bool("cwd", false, "Create required files in currend working directory")
	local := cmd.Flags().Bool("local", false, "Use local code to build images instead of URIs to GitHub. This requires --cwd to be set.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()

		if *local && !*cwd {
			return fmt.Errorf("--local requires --cwd to be set")
		}

		dataPath := path.Join(xdg.DataHome, appName)
		if *cwd {
			p, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting current directory: %w", err)
			}
			dataPath = p
		}

		if err := os.MkdirAll(dataPath, fs.ModePerm); err != nil {
			return fmt.Errorf("creating directory `%s`: %w", dataPath, err)
		}

		if err := CreateDockerComposeYML(ctx, dataPath, !*local); err != nil {
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

	randomSecrets := []string{
		"auth_token_key",
		"auth_cookie_key",
	}
	for _, s := range randomSecrets {
		err := func() error {
			f, err := os.Create(path.Join(dataPath, s))
			if err != nil {
				return fmt.Errorf("creating file `%s`: %w", path.Join(dataPath, s), err)
			}
			defer f.Close()

			// This creates cryptographically secure random bytes. 32 Bytes means
			// 256bit. The output can contain zero bytes.
			b64e := base64.NewEncoder(base64.StdEncoding, f)
			defer b64e.Close()
			if _, err := io.Copy(b64e, io.LimitReader(rand.Reader, 32)); err != nil {
				return fmt.Errorf("writing cryptographically secure random base64 encoded bytes: %w", err)
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}

	if err := os.WriteFile(path.Join(dataPath, "admin"), []byte("admin"), fs.ModePerm); err != nil {
		return fmt.Errorf("writing admin password to secret file: %w", err)
	}

	return nil
}
