package manage

import (
	"context"
	_ "embed" // Neeed for embed. See Docu of Go 1.16
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
)

// createDockerComposeYML creates a docker-compose.yml file in the current working directory
// using a template. In non local mode it uses the GitHub API to fetch the required commit IDs
// of all services.
func createDockerComposeYML(ctx context.Context, dataPath string) error {
	p := path.Join(dataPath, "docker-compose.yml")

	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("creating file `%s`: %w", p, err)
	}
	defer f.Close()

	if err := constructDockerComposeYML(ctx, f); err != nil {
		return fmt.Errorf("writing content to file `%s`: %w", p, err)
	}

	return nil
}

//go:embed docker-compose.yml.tpl
var defaultDockerCompose string

// constructDockerComposeYML writes the populated template to the given writer.
func constructDockerComposeYML(ctx context.Context, w io.Writer) error {
	// TODO: Local case

	composeTPL, err := template.New("compose").Parse(defaultDockerCompose)
	if err != nil {
		return fmt.Errorf("creating Docker Compose template: %w", err)
	}
	composeTPL.Option("missingkey=error")

	var tplData struct {
		Tag                string
		ExternalHTTPPort   string
		ExternalManagePort string
		CommitID           map[string]string
		Ref                string
	}

	tplData.Tag = "4.0-dev"
	tplData.ExternalHTTPPort = "8000"
	tplData.ExternalManagePort = "9008"
	tplData.Ref = "openslides4-dev"

	c, err := getCommitIDs(ctx, tplData.Ref)
	if err != nil {
		return fmt.Errorf("getting commit IDs: %w", err)
	}
	tplData.CommitID = c

	if err := composeTPL.Execute(w, tplData); err != nil {
		return fmt.Errorf("writing Docker Compose file: %w", err)
	}
	return nil
}

// getCommitIDs fetches the commit IDs for all services from GitHub API.
func getCommitIDs(ctx context.Context, ref string) (map[string]string, error) {
	addr := "https://api.github.com/repos/OpenSlides/OpenSlides/contents?ref=" + ref
	req, err := http.NewRequestWithContext(ctx, "GET", addr, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request to GitHub API: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request to GitHub API at %s: %w", addr, err)
	}
	defer resp.Body.Close()

	var apiBody []struct {
		Name string `json:"name"`
		SHA  string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiBody); err != nil {
		return nil, fmt.Errorf("reading and decoding body from GitHub API: %w", err)
	}

	services := map[string]string{
		"openslides-client":             "client",
		"openslides-backend":            "backend",
		"openslides-datastore-service":  "datastore",
		"openslides-autoupdate-service": "autoupdate",
		"openslides-auth-service":       "auth",
		"openslides-media-service":      "media",
		"openslides-manage-service":     "manage",
		"openslides-permission-service": "permission", // TODO: Remove this line after permission service is removed.
	}

	commitIDs := make(map[string]string, len(services))
	for _, apiElement := range apiBody {
		tplName, ok := services[apiElement.Name]
		if ok {
			commitIDs[tplName] = apiElement.SHA
		}
	}

	return commitIDs, nil
}

//go:embed default_services.env
var defaultServiesEnv []byte

func createEnvFile(dataPath string) error {
	p := path.Join(dataPath, "services.env")
	if err := os.WriteFile(p, defaultServiesEnv, fs.ModePerm); err != nil {
		return fmt.Errorf("write services file at %s: %w", p, err)
	}
	return nil
}
