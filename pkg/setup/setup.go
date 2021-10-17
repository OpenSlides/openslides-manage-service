package setup

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/OpenSlides/openslides-manage-service/pkg/config"
	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
	"github.com/spf13/cobra"
)

const (
	subDirPerms                       fs.FileMode = 0770
	authTokenKeyFileName                          = "auth_token_key"
	authCookieKeyFileName                         = "auth_cookie_key"
	datastorePostgresPasswordFileName             = "datastore_postgres_password"
	mediaPostgresPasswordFileName                 = "media_postgres_password"
	dbDirName                                     = "db-data"
)

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

	// ManageAuthPasswordFileName is the name of the secrets file containing the password for
	// (basic) authorization to the manage service.
	ManageAuthPasswordFileName = "manage_auth_password"
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
	tplFile := config.FlagTpl(cmd)
	configFiles := config.FlagConfig(cmd)

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

		var config [][]byte
		if len(*configFiles) > 0 {
			for _, configFile := range *configFiles {
				fc, err := os.ReadFile(configFile)
				if err != nil {
					return fmt.Errorf("reading file %q: %w", configFile, err)
				}
				config = append(config, fc)
			}
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
// and YAML configs can be provided.
func Setup(dir string, force bool, tplContent []byte, cfgContent [][]byte) error {
	// Create directory
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory at %q: %w", dir, err)
	}

	// Create YAML file
	if err := config.CreateYmlFile(dir, force, tplContent, cfgContent); err != nil {
		return fmt.Errorf("creating YAML file at %q: %w", dir, err)
	}

	// Create secrets directory
	secrDir := path.Join(dir, SecretsDirName)
	if err := os.MkdirAll(secrDir, subDirPerms); err != nil {
		return fmt.Errorf("creating secrets directory at %q: %w", dir, err)
	}

	// Create random secrets (auth token key, auth cookie key, manage auth password)
	if err := createRandomSecrets(secrDir, force); err != nil {
		return fmt.Errorf("creating random secrets: %w", err)
	}

	// Create postgres password files
	// TODO: Make them random too.
	postgresPassword := []byte("openslides")
	if err := shared.CreateFile(secrDir, force, datastorePostgresPasswordFileName, postgresPassword); err != nil {
		return fmt.Errorf("creating secret datastore postgres password file at %q: %w", dir, err)
	}
	if err := shared.CreateFile(secrDir, force, mediaPostgresPasswordFileName, postgresPassword); err != nil {
		return fmt.Errorf("creating secret media postgres password file at %q: %w", dir, err)
	}

	// Create supereadmin file
	if err := shared.CreateFile(secrDir, force, SuperadminFileName, []byte(DefaultSuperadminPassword)); err != nil {
		return fmt.Errorf("creating admin file at %q: %w", dir, err)
	}

	// Create database directory
	// Attention: For unknown reason it is not possible to use perms 0770 here. Docker Compose does not like it ...
	if err := os.MkdirAll(path.Join(dir, dbDirName), 0777); err != nil {
		return fmt.Errorf("creating database directory at %q: %w", dir, err)
	}

	return nil
}

func createRandomSecrets(dir string, force bool) error {
	// Create auth token key file
	secrToken, err := randomSecret()
	if err != nil {
		return fmt.Errorf("creating random key for auth token: %w", err)
	}
	if err := shared.CreateFile(dir, force, authTokenKeyFileName, secrToken); err != nil {
		return fmt.Errorf("creating secret auth token key file at %q: %w", dir, err)
	}

	// Create auth cookie key file
	secrCookie, err := randomSecret()
	if err != nil {
		return fmt.Errorf("creating random key for auth cookie: %w", err)
	}
	if err := shared.CreateFile(dir, force, authCookieKeyFileName, secrCookie); err != nil {
		return fmt.Errorf("creating secret auth cookie key file at %q: %w", dir, err)
	}

	// Create manage auth password file
	authPw, err := randomSecret()
	if err != nil {
		return fmt.Errorf("creating random key for manage auth password: %w", err)
	}
	if err := shared.CreateFile(dir, force, ManageAuthPasswordFileName, authPw); err != nil {
		return fmt.Errorf("creating secret manage auth password file at %q: %w", dir, err)
	}

	return nil
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
