package manage

import (
	"context"
	_ "embed" // Neeed for embed. See Docu of Go 1.16
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

// Constants to be used in created docker-compose.yml.
const (
	DockerRegistry     = "ghcr.io/normanjaeckel/openslides"
	OpenSlidesTag      = "4.0.0-dev"
	ExternalHTTPPort   = "8000"
	ExternalManagePort = "9008"
)

// servicesYML contains the definitions for all OpenSlides services.
const servicesYML = `---
proxy:
  image: openslides-proxy
  path: proxy
client:
  image: openslides-client
  path: openslides-client
backend:
  image: openslides-backend
  path: openslides-backend
datastore_reader:
  image: openslides-datastore-reader
  path: openslides-datastore-service
  args:
    module: reader
    port: 9010
datastore_writer:
  image: openslides-datastore-writer
  path: openslides-datastore-service
  args:
    module: writer
    port: 9011
autoupdate:
  image: openslides-autoupdate
  path: openslides-autoupdate-service
auth:
  image: openslides-auth
  path: openslides-auth-service
media:
  image: openslides-media
  path: openslides-media-service
manage:
  image: openslides-manage
  path: openslides-manage-service
permission:
  image: openslides-permission
  path: openslides-permission-service
`

// Service holds the metadate of an OpenSlides service to be used in
// docker-compose.yml template.
type Service struct {
	Image string
	Path  string
	Args  struct {
		MODULE string
		PORT   string
	}
}

// Services provides a map with all OpenSlides services.
func Services() (map[string]Service, error) {
	var s map[string]Service
	if err := yaml.Unmarshal([]byte(servicesYML), &s); err != nil {
		return nil, fmt.Errorf("unmarshalling servivesYML: %w", err)
	}
	return s, nil
}

// createDockerComposeYML creates a docker-compose.yml file in the current working directory
// using a template. In remote mode it uses the GitHub API to fetch the required commit IDs
// of all services. Else it uses relative paths to local code as provided in OpenSlides
// main repository.
func createDockerComposeYML(ctx context.Context, dataPath string, remote bool) error {
	p := path.Join(dataPath, "docker-compose.yml")

	if fileExists(p) {
		return nil
	}

	w, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("creating file `%s`: %w", p, err)
	}
	defer w.Close()

	if err := constructDockerComposeYML(ctx, w, remote); err != nil {
		return fmt.Errorf("writing content to file `%s`: %w", p, err)
	}

	return nil
}

//go:embed docker-compose.yml.tpl
var defaultDockerCompose string

// tplData holds the data used to execute the docker-compose.yml template.
type tplData struct {
	ExternalHTTPPort   string
	ExternalManagePort string
	Service            map[string]string
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

	td := tplData{
		ExternalHTTPPort:   ExternalHTTPPort,
		ExternalManagePort: ExternalManagePort,
	}

	if err := populateServices(ctx, &td, remote); err != nil {
		return fmt.Errorf("populating services to template data: %w", err)
	}

	if err := composeTPL.Execute(w, td); err != nil {
		return fmt.Errorf("writing Docker Compose file: %w", err)
	}

	return nil
}

// populateServices is a small helper function that populates service metadata
// to the given template data.
func populateServices(ctx context.Context, td *tplData, remote bool) error {
	services, err := Services()
	if err != nil {
		return fmt.Errorf("getting services: %w", err)
	}

	td.Service = make(map[string]string, len(services))

	if remote {
		for name, service := range services {
			td.Service[name] = fmt.Sprintf(
				"image: %s/%s:%s",
				DockerRegistry,
				service.Image,
				OpenSlidesTag,
			)
		}
		return nil
	}

	fragment := `image: %s
    build:
      context: ./%s`

	fragmentSuffix := `
      args:
        MODULE: %s
        PORT: %s`

	for name, service := range services {
		s := fmt.Sprintf(
			fragment,
			fmt.Sprintf("%s:%s", service.Image, OpenSlidesTag),
			service.Path,
		)
		if service.Args.MODULE != "" || service.Args.PORT != "" {
			s += fmt.Sprintf(fragmentSuffix, service.Args.MODULE, service.Args.PORT)
		}
		td.Service[name] = s
	}
	return nil
}

//go:embed default_services.env
var defaultServiesEnv []byte

func createEnvFile(dataPath string) error {
	p := path.Join(dataPath, "services.env")

	if fileExists(p) {
		return nil
	}

	if err := os.WriteFile(p, defaultServiesEnv, fs.ModePerm); err != nil {
		return fmt.Errorf("write services file at %s: %w", p, err)
	}
	return nil
}

// fileExists checks if the file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
