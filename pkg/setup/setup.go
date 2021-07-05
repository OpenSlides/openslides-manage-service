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

	"github.com/spf13/cobra"
)

const (
	ymlFileName           = "docker-compose.yml"
	envFileName           = "services.env"
	secretsDirName        = "secrets"
	authTokenKeyFileName  = "auth_token_key"
	authCookieKeyFileName = "auth_cookie_key"
	adminFileName         = "admin"
	dbDirName             = "db-data"
)

//go:embed default-docker-compose.yml
var defaultDockerComposeYml []byte

//go:embed default-environment.env
var defaultEnvFile []byte

// DefaultAdminPassword is the password for the first admin created with initial data.
const DefaultAdminPassword = "admin"

const (
	// SetupHelp contains the short help text for the setup command.
	SetupHelp = "Builds the required files for using Docker Compose or Docker Swarm."

	// SetupHelpExtra contains the long help text for the setup command without the headline.
	SetupHelpExtra = `This command creates a YAML file with a default .env nearby. It also
creates the required secrets and directories for volumes containing
persistent database and SSL certs. Everything is created in the given
directory.`
)

// Cmd returns the setup subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup directory",
		Short: SetupHelp,
		Long:  SetupHelp + "\n\n" + SetupHelpExtra,
		Args:  cobra.ExactArgs(1),
	}
	force := cmd.Flags().BoolP("force", "f", false, "Do not skip existing files but overwrite them.")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		fn := WithSkip
		if *force {
			fn = WithForce
		}
		if err := fn(dir); err != nil {
			return fmt.Errorf("running Setup(): %w", err)
		}
		return nil
	}
	return cmd
}

// WithSkip creates YAML file for Docker Compose or Docker Swarm with services.env file
// and secrets directory.
//
// Existing files are skipped.
func WithSkip(dir string) error {
	return setup(dir, false)
}

// WithForce creates YAML file for Docker Compose or Docker Swarm with services.env file
// and secrets directory.
//
// Existing files are overwritten.
func WithForce(dir string) error {
	return setup(dir, true)
}

func setup(dir string, force bool) error {
	// Create directory
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory at %q: %w", dir, err)
	}

	// Create YAML file
	if err := createFile(dir, force, ymlFileName, defaultDockerComposeYml); err != nil {
		return fmt.Errorf("creating YAML file at %q: %w", dir, err)
	}

	// Create environment file
	if err := createFile(dir, force, envFileName, defaultEnvFile); err != nil {
		return fmt.Errorf("creating environment file at %q: %w", dir, err)
	}

	// Create secrets directory
	secrDir := path.Join(dir, secretsDirName)
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

	// Create admin file
	if err := createFile(secrDir, force, adminFileName, []byte(DefaultAdminPassword)); err != nil {
		return fmt.Errorf("creating admin file at %q: %w", dir, err)
	}

	// Create database directory
	if err := os.MkdirAll(path.Join(dir, dbDirName), os.ModePerm); err != nil {
		return fmt.Errorf("creating database directory at %q: %w", dir, err)
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
