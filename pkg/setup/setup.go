package setup

import (
	"bytes"
	"crypto/rand"
	_ "embed" // Blank import required to use go directive.
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
)

const (
	ymlFileName           = "docker-compose.yml"
	authTokenKeyFileName  = "auth_token_key"
	authCookieKeyFileName = "auth_cookie_key"
	dbDirName             = "db-data"
)

//go:embed default-docker-compose.yml
var defaultDockerComposeYml []byte

//go:embed default-config.yml
var defaultConfig []byte

const (
	// SetupHelp contains the short help text for the command.
	SetupHelp = "Builds the required files for using Docker Compose or Docker Swarm"

	// SetupHelpExtra contains the long help text for the command without the headline.
	SetupHelpExtra = `This command creates a YAML file. It also creates the required secrets and
directories for volumes containing persistent database and SSL certs. Everything
is created in the given directory.`

	// SecretsDirName is the name of the directory for Docker Secrets.
	SecretsDirName = "secrets"

	// SuperadminFileName is the name of the secrets file containing the superadmin password.
	SuperadminFileName = "superadmin"

	// DefaultSuperadminPassword is the password for the first superadmin created with initial data.
	DefaultSuperadminPassword = "superadmin"
)

// Cmd returns the setup subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup directory",
		Short: SetupHelp,
		Long:  SetupHelp + "\n\n" + SetupHelpExtra,
		Args:  cobra.ExactArgs(1),
	}

	force := cmd.Flags().BoolP("force", "f", false, "do not skip existing files but overwrite them")
	tplFile := cmd.Flags().StringP("template", "t", "", "custom YAML template file")
	configFile := cmd.Flags().StringP("config", "c", "", "custom YAML config file")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		dir := args[0]

		var tpl []byte
		if *tplFile != "" {
			fc, err := os.ReadFile(*tplFile)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", *tplFile, err)
			}
			tpl = fc
		}

		var config []byte
		if *configFile != "" {
			fc, err := os.ReadFile(*configFile)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", *configFile, err)
			}
			config = fc
		}

		if err := Setup(dir, *force, tpl, config); err != nil {
			return fmt.Errorf("running Setup(): %w", err)
		}
		return nil
	}
	return cmd
}

// Setup creates YAML file for Docker Compose or Docker Swarm with secrets directory and
// directories for database and SSL certs volumes.
//
// Existing files are skipped unless force is true. A custom template for the YAML file
// and a YAML config can be provided.
func Setup(dir string, force bool, tplContent, cfgContent []byte) error {
	// Create directory
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory at %q: %w", dir, err)
	}

	// Create YAML file
	if err := createYmlFile(dir, ymlFileName, force, tplContent, cfgContent); err != nil {
		return fmt.Errorf("creating YAML file at %q: %w", dir, err)
	}

	// Create secrets directory
	secrDir := path.Join(dir, SecretsDirName)
	if err := os.MkdirAll(secrDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating secrets directory at %q: %w", dir, err)
	}

	// Create auth token key file
	secrToken, err := randomSecret()
	if err != nil {
		return fmt.Errorf("creating random key for auth token: %w", err)
	}
	if err := createFile(secrDir, force, authTokenKeyFileName, secrToken); err != nil {
		return fmt.Errorf("creating secret auth token key file at %q: %w", dir, err)
	}

	// Create auth cookie key file
	secrCookie, err := randomSecret()
	if err != nil {
		return fmt.Errorf("creating random key for auth cookie: %w", err)
	}
	if err := createFile(secrDir, force, authCookieKeyFileName, secrCookie); err != nil {
		return fmt.Errorf("creating secret auth cookie key file at %q: %w", dir, err)
	}

	// Create supereadmin file
	if err := createFile(secrDir, force, SuperadminFileName, []byte(DefaultSuperadminPassword)); err != nil {
		return fmt.Errorf("creating admin file at %q: %w", dir, err)
	}

	// Create database directory
	if err := os.MkdirAll(path.Join(dir, dbDirName), os.ModePerm); err != nil {
		return fmt.Errorf("creating database directory at %q: %w", dir, err)
	}

	return nil
}

func createYmlFile(dir, name string, force bool, tplContent, cfgContent []byte) error {
	if tplContent == nil {
		tplContent = defaultDockerComposeYml
	}

	cfg, err := newYmlConfig(cfgContent)
	if err != nil {
		return fmt.Errorf("creating new YML config object: %w", err)
	}

	tmpl, err := template.New("YAML File").Option("missingkey=error").Parse(string(tplContent))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}
	var res bytes.Buffer
	if err := tmpl.Execute(&res, cfg); err != nil {
		return fmt.Errorf("executing template %v: %w", tmpl, err)
	}

	if err := createFile(dir, force, name, res.Bytes()); err != nil {
		return fmt.Errorf("creating YAML file %q at %q: %w", name, dir, err)
	}

	return nil
}

func createFile(dir string, force bool, name string, content []byte) error {
	p := path.Join(dir, name)

	pExists, err := fileExists(p)
	if err != nil {
		return fmt.Errorf("checking file existance: %w", err)
	}
	if !force && pExists {
		// No force-mode and file already exists, so skip this file.
		return nil
	}

	w, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", p, err)
	}
	defer w.Close()

	if _, err := w.Write(content); err != nil {
		return fmt.Errorf("writing content to file %q: %w", p, err)
	}
	return nil
}

// fileExists is a small helper function to check if a file already exists. It is not
// save in concurrent usage.
func fileExists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("checking existance of file %s: %w", p, err)
}

func randomSecret() ([]byte, error) {
	buf := new(bytes.Buffer)
	b64e := base64.NewEncoder(base64.StdEncoding, buf)
	defer b64e.Close()

	if _, err := io.Copy(b64e, io.LimitReader(rand.Reader, 32)); err != nil {
		return nil, fmt.Errorf("writing cryptographically secure random base64 encoded bytes: %w", err)
	}

	return buf.Bytes(), nil
}

type ymlConfig struct {
	Defaults struct {
		ContainerRegistry string `yaml:"containerRegistry"`
		Tag               string `yaml:"tag"`
	} `yaml:"defaults"`

	DefaultEnvironment map[string]string `yaml:"defaultEnvironment"`

	Services map[string]service `yaml:"services"`
}

type service struct {
	ContainerRegistry string `yaml:"containerRegistry"`
	Tag               string `yaml:"tag"`
}

func newYmlConfig(data []byte) (*ymlConfig, error) {
	// Unmarshal given YAML data
	c1 := new(ymlConfig)
	if err := yaml.Unmarshal(data, c1); err != nil {
		return nil, fmt.Errorf("unmarshaling YAML: %w", err)
	}

	// Unmarshal default YAML data
	c2 := new(ymlConfig)
	if err := yaml.Unmarshal(defaultConfig, c2); err != nil {
		return nil, fmt.Errorf("unmarshaling YAML: %w", err)
	}

	// Merge default
	if err := mergo.Merge(c1, c2); err != nil {
		return nil, fmt.Errorf("merging default config into given data: %w", err)
	}

	// Fill services
	allServices := []string{
		"proxy",
		"client",
		"backend",
		"datastoreReader",
		"datastoreWriter",
		"postgres", // TODO: Remove me
		"autoupdate",
		"auth",
		"redis", // TODO: Remove me
		"media",
		"icc",
		"manage",
	}
	if len(c1.Services) == 0 {
		c1.Services = make(map[string]service, len(allServices))
	}

	for _, name := range allServices {
		_, ok := c1.Services[name]
		if !ok {
			c1.Services[name] = *new(service)
		}
		s := c1.Services[name]

		if s.ContainerRegistry == "" {
			s.ContainerRegistry = c1.Defaults.ContainerRegistry
		}
		if s.Tag == "" {
			s.Tag = c1.Defaults.Tag
		}

		c1.Services[name] = s
	}

	return c1, nil
}
