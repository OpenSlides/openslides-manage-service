package manage

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

const startHelp = `Builds images and starts OpenSlides with Docker Compose

This command executes the following steps to start OpenSlides:
- Create a local docker-compose.yml.
- Create local secrets for the auth service.
- Run the docker compose file with the "up" command in daemonized mode.
- TODO ...
`

// CmdStart does ...
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

func createDockerComposeYML(ctx context.Context) error {
	filename := "docker-compose.yml"
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file `%s`: %w", filename, err)
	}
	defer f.Close()

	if err := writeDockerComposeYML(f); err != nil {
		return fmt.Errorf("writing content to file `%s`: %w", filename, err)
	}

	return nil
}

//go:embed docker-compose.yml.tpl
var defaultDockerCompose string

func writeDockerComposeYML(w io.Writer) error {
	// TODO:
	// * Fetch commit hashes for submodules.

	composeTPL, err := template.New("compose").Parse(defaultDockerCompose)
	if err != nil {
		return fmt.Errorf("creating Docker Compose template: %w", err)
	}

	var c struct {
		ExternalHTTPPort   string
		ExternalManagePort string
		CommitID           map[string]string
	}

	c.ExternalHTTPPort = "8000"
	c.ExternalManagePort = "9008"
	c.CommitID = getCommitIDs()

	if err := composeTPL.Execute(w, c); err != nil {
		return fmt.Errorf("writing Docker Compose file: %w", err)
	}
	return nil
}

func getCommitIDs() map[string]string {
	m := make(map[string]string)
	m["proxy"] = "1234567890"
	m["client"] = "99999999999999999"
	return m
}
