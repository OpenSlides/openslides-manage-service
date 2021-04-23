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
	"strings"
)

// CreateFileType provides access to os.Create and enables mocking during testing.
type CreateFileType func(name string) (io.WriteCloser, error)

var CreateFile CreateFileType = func(name string) (io.WriteCloser, error) {
	return os.Create(name)
}

// CreateDockerComposeYML creates a docker-compose.yml file in the current working directory
// using a template. In remote mode it uses the GitHub API to fetch the required commit IDs
// of all services. Else it uses relative paths to local code as provided in OpenSlides
// main repository.
func CreateDockerComposeYML(ctx context.Context, dataPath string, remote bool) error {
	p := path.Join(dataPath, "docker-compose.yml")

	f, err := CreateFile(p)
	if err != nil {
		return fmt.Errorf("creating file `%s`: %w", p, err)
	}
	defer f.Close()

	if err := constructDockerComposeYML(ctx, f, remote); err != nil {
		return fmt.Errorf("writing content to file `%s`: %w", p, err)
	}

	return nil
}

//go:embed docker-compose.yml.tpl
var defaultDockerCompose string

// tplData holds the data used to execute the docker-compose.yml template.
type tplData struct {
	Tag                string
	ExternalHTTPPort   string
	ExternalManagePort string
	Service            map[string]string
}

// Service holds service metadata from GitHub API that are used in the
// docker-compose.yml template.
type Service struct {
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// constructDockerComposeYML writes the populated template to the given writer.
// If remote is true it uses GitHub URIs for the build context. Else it uses relative
// paths to local code as provided in OpenSlides main repository.
func constructDockerComposeYML(ctx context.Context, w io.Writer, remote bool) error {
	composeTPL, err := template.New("compose").Parse(defaultDockerCompose)
	if err != nil {
		return fmt.Errorf("creating Docker Compose template: %w", err)
	}
	composeTPL.Option("missingkey=error")

	const ref = "openslides4-dev"

	td := tplData{
		Tag:                "4.0-dev",
		ExternalHTTPPort:   "8000",
		ExternalManagePort: "9008",
	}

	if err := populateServices(ctx, &td, ref, remote); err != nil {
		return fmt.Errorf("populating services to template data: %w", err)
	}

	if err := composeTPL.Execute(w, td); err != nil {
		return fmt.Errorf("writing Docker Compose file: %w", err)
	}

	return nil
}

// populateServices is a small helper function that populates service metadata
// to the given template data.
func populateServices(ctx context.Context, td *tplData, ref string, remote bool) error {
	services, err := Services(ctx, ref)
	if err != nil {
		return fmt.Errorf("getting services from GitHub API: %w", err)
	}

	td.Service = make(map[string]string, len(services))

	if remote {
		for name, service := range services {
			td.Service[name] = strings.Replace(service.HTMLURL, "/tree/", ".git#", 1)
		}
		td.Service["proxy"] = fmt.Sprintf("https://github.com/OpenSlides/OpenSlides.git#%s:proxy", ref)
		return nil
	}

	for name, service := range services {
		td.Service[name] = "./" + service.Name
	}
	td.Service["proxy"] = "./proxy"
	return nil
}

// Services fetches service definitions from GitHub API.
func Services(ctx context.Context, ref string) (map[string]Service, error) {
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

	var apiBody []Service
	if err := json.NewDecoder(resp.Body).Decode(&apiBody); err != nil {
		return nil, fmt.Errorf("reading and decoding body from GitHub API: %w", err)
	}

	s := map[string]string{
		"openslides-client":             "client",
		"openslides-backend":            "backend",
		"openslides-datastore-service":  "datastore",
		"openslides-autoupdate-service": "autoupdate",
		"openslides-auth-service":       "auth",
		"openslides-media-service":      "media",
		"openslides-manage-service":     "manage",
		"openslides-permission-service": "permission", // TODO: Remove this line after permission service is removed.
	}
	services := make(map[string]Service, len(s))
	for _, apiElement := range apiBody {
		tplName, ok := s[apiElement.Name]
		if ok {
			services[tplName] = apiElement
		}
	}

	return services, nil
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
