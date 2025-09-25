package setup

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"io/fs"
	"math/big"
	"os"
	"path"
	"time"

	"github.com/OpenSlides/openslides-manage-service/pkg/config"
	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
	"github.com/spf13/cobra"
)

const (
	subDirPerms  fs.FileMode = 0770
	certCertName             = "cert_crt"
	certKeyName              = "cert_key"
)

const (
	// SetupHelp contains the short help text for the command.
	SetupHelp = "Creates the required files for using Docker Compose or Docker Swarm"

	// SetupHelpExtra contains the long help text for the command without the headline.
	SetupHelpExtra = `This command creates a container configuration YAML file. It also creates the
required secrets and directories for volumes containing persistent database and
SSL certs. Everything is created in the given directory.`

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

// SecretSpec defines how to generate a specific secret
type SecretSpec struct {
	Name      string
	Generator func() ([]byte, error)
}

// defaultSecrets defines all the secrets that should be created by default
var defaultSecrets = []SecretSpec{
	{"auth_token_key", randomSecret},
	{"auth_cookie_key", randomSecret},
	{ManageAuthPasswordFileName, randomSecret},
	{"internal_auth_password", randomSecret},
	{"postgres_password", randomSecret},
	{SuperadminFileName, func() ([]byte, error) { return []byte(DefaultSuperadminPassword), nil }},
}

// Cmd returns the subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup directory",
		Short: SetupHelp,
		Long:  SetupHelp + "\n\n" + SetupHelpExtra,
		Args:  cobra.ExactArgs(1),
	}

	force := cmd.Flags().BoolP("force", "f", false, "do not skip existing files but overwrite them")
	builtinTemplate := config.FlagBuiltinTemplate(cmd)
	tplFileOrDirName := config.FlagTpl(cmd)
	configFileNames := config.FlagConfig(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if *tplFileOrDirName != "" && *builtinTemplate != config.BuiltinTemplateDefault {
			return fmt.Errorf("flag --builtin-template must not be used together with flag --template")
		}

		cfg := config.SetupConfig{
			BaseDir:         args[0],
			Force:           *force,
			BuiltinTemplate: *builtinTemplate,
			CustomTemplate:  *tplFileOrDirName,
			ConfigFiles:     *configFileNames,
		}

		if err := Setup(cfg); err != nil {
			return fmt.Errorf("running Setup(): %w", err)
		}
		return nil
	}
	return cmd
}

// Setup creates one or more (depending on template) files containing the
// deployment definitions and the secrets directory including SSL certs.
func Setup(cfg config.SetupConfig) error {
	// Create YAML config object
	ymlCfg, err := config.NewYmlConfig(cfg.ConfigFiles)
	if err != nil {
		return fmt.Errorf("parsing configuration: %w", err)
	}

	// Create secrets directory first (needed before deployment file generation)
	secrDir := path.Join(cfg.BaseDir, SecretsDirName)
	if err := os.MkdirAll(secrDir, subDirPerms); err != nil {
		return fmt.Errorf("creating secrets directory at %q: %w", cfg.BaseDir, err)
	}

	// Create all secrets (must be done before deployment files since templates may reference them)
	if err := createSecrets(secrDir, cfg.Force, defaultSecrets); err != nil {
		return fmt.Errorf("creating secrets: %w", err)
	}

	// Create certificates if HTTPS is enabled
	if *ymlCfg.EnableLocalHTTPS {
		if err := createCerts(secrDir, cfg.Force); err != nil {
			return fmt.Errorf("creating certificates: %w", err)
		}
	}

	// Create the base directory and the deployment files (after secrets are created)
	if err := config.CreateDirAndFiles(cfg.BaseDir, cfg.Force, cfg.BuiltinTemplate, cfg.CustomTemplate, ymlCfg); err != nil {
		return fmt.Errorf("(re-)creating deployment files: %w", err)
	}

	return nil
}

// createSecrets creates all specified secrets in the given directory
func createSecrets(dir string, force bool, secrets []SecretSpec) error {
	for _, spec := range secrets {
		data, err := spec.Generator()
		if err != nil {
			return fmt.Errorf("generating secret %q: %w", spec.Name, err)
		}
		if err := shared.CreateFile(dir, force, spec.Name, data); err != nil {
			return fmt.Errorf("creating secret file %q: %w", spec.Name, err)
		}
	}
	return nil
}

// randomSecret generates a cryptographically secure random base64-encoded secret
func randomSecret() ([]byte, error) {
	buf := new(bytes.Buffer)
	b64e := base64.NewEncoder(base64.StdEncoding, buf)
	defer b64e.Close()

	if _, err := io.Copy(b64e, io.LimitReader(rand.Reader, 32)); err != nil {
		return nil, fmt.Errorf("writing cryptographically secure random base64 encoded bytes: %w", err)
	}

	return buf.Bytes(), nil
}

// createCerts generates self-signed certificates for local HTTPS
func createCerts(dir string, force bool) error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generating key: %w", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("generating serial number: %w", err)
	}

	templ := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{"OpenSlides"}},
		DNSNames:              []string{"localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(30, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create certificate
	certData, err := x509.CreateCertificate(rand.Reader, &templ, &templ, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("creating certificate data: %w", err)
	}

	// Encode and save certificate
	buf1 := new(bytes.Buffer)
	if err := pem.Encode(buf1, &pem.Block{Type: "CERTIFICATE", Bytes: certData}); err != nil {
		return fmt.Errorf("encoding certificate data: %w", err)
	}
	if err := shared.CreateFile(dir, force, certCertName, buf1.Bytes()); err != nil {
		return fmt.Errorf("creating certificate file %q at %q: %w", certCertName, dir, err)
	}

	// Encode and save private key
	keyData, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("marshalling key: %w", err)
	}
	buf2 := new(bytes.Buffer)
	if err := pem.Encode(buf2, &pem.Block{Type: "PRIVATE KEY", Bytes: keyData}); err != nil {
		return fmt.Errorf("encoding key data: %w", err)
	}
	if err := shared.CreateFile(dir, force, certKeyName, buf2.Bytes()); err != nil {
		return fmt.Errorf("creating key file %q at %q: %w", certKeyName, dir, err)
	}

	return nil
}
