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
		dir := args[0]
		if err := Setup(dir, *force, *builtinTemplate, *tplFileOrDirName, *configFileNames); err != nil {
			return fmt.Errorf("running Setup(): %w", err)
		}
		return nil
	}
	return cmd
}

// Setup creates one or more (depending on template) files containing the
// deployment definitions and the secrets directory including SSL certs. The
// parameters are just the command flags.
func Setup(baseDir string, force bool, builtinTpl string, tplFileOrDirName string, configFileNames []string) error {
	// Create YAML config object
	cfg, err := config.NewYmlConfig(configFileNames)
	if err != nil {
		return fmt.Errorf("parsing configuration: %w", err)
	}

	// Create secrets directory
	secrDir := path.Join(baseDir, SecretsDirName)
	if err := os.MkdirAll(secrDir, subDirPerms); err != nil {
		return fmt.Errorf("creating secrets directory at %q: %w", baseDir, err)
	}

	// Create random secrets
	if err := createRandomSecrets(secrDir, force); err != nil {
		return fmt.Errorf("creating random secrets: %w", err)
	}

	// Create certificates
	if *cfg.EnableLocalHTTPS {
		if err := createCerts(secrDir, force); err != nil {
			return fmt.Errorf("creating certificates: %w", err)
		}
	}

	// Create superadmin file
	if err := shared.CreateFile(secrDir, force, SuperadminFileName, []byte(DefaultSuperadminPassword)); err != nil {
		return fmt.Errorf("creating admin file at %q: %w", baseDir, err)
	}

	// Create the base directory and the the deployment files using the code from the config command.
	if err := config.CreateDirAndFiles(baseDir, force, builtinTpl, tplFileOrDirName, cfg); err != nil {
		return fmt.Errorf("(re-)creating deployment files: %w", err)
	}

	return nil
}

func createRandomSecrets(dir string, force bool) error {
	secs := []struct {
		filename string
	}{
		{"auth_token_key"},
		{"auth_cookie_key"},
		{ManageAuthPasswordFileName},
		{"internal_auth_password"},
		{"postgres_password"},
	}
	for _, s := range secs {
		secrToken, err := randomSecret()
		if err != nil {
			return fmt.Errorf("creating random secret %q: %w", s.filename, err)
		}
		if err := shared.CreateFile(dir, force, s.filename, secrToken); err != nil {
			return fmt.Errorf("creating secret file %q at %q: %w", dir, s.filename, err)
		}
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

	certData, err := x509.CreateCertificate(rand.Reader, &templ, &templ, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("creating certificate data: %w", err)
	}
	buf1 := new(bytes.Buffer)
	if err := pem.Encode(buf1, &pem.Block{Type: "CERTIFICATE", Bytes: certData}); err != nil {
		return fmt.Errorf("encoding certificate data: %w", err)
	}
	if err := shared.CreateFile(dir, force, certCertName, buf1.Bytes()); err != nil {
		return fmt.Errorf("creating certificate file %q at %q: %w", certCertName, dir, err)
	}

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
