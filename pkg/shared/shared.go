package shared

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

// OpenSlidesInstanceConfigurationFileVersion is the hardcoded version of the
// Docker Compose YAML file. The manage service compares this number with the
// value of the OPENSLIDES_INSTANCE_CONFIGURATION_FILE_VERSION environment
// variable.
const OpenSlidesInstanceConfigurationFileVersion = "v001"

// developmentPassword is the password used if environment variable
// OPENSLIDES_DEVELOPMENT is set to one of the following values: 1, t, T, TRUE,
// true, True.
const developmentPassword = "openslides"

// AuthHeader is the name of the header that contains the basic authoriztation password.
const AuthHeader = "authorization"

const fileMode fs.FileMode = 0666

// CreateFile creates a file in the given directory with the given content.
// Use a truthy value for force to override an existing file.
func CreateFile(dir string, force bool, name string, content []byte) error {
	p := path.Join(dir, name)

	pExists, err := fileExists(p)
	if err != nil {
		return fmt.Errorf("checking file existance: %w", err)
	}
	if !force && pExists {
		// No force-mode and file already exists, so skip this file.
		return nil
	}

	if err := os.WriteFile(p, content, fileMode); err != nil {
		return fmt.Errorf("creating and writing to file %q: %w", p, err)
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

// InputOrFileOrStdin takes either a command line input or a filename (which can
// be "-" so we read from stdin) and returns the content.
func InputOrFileOrStdin(input, filename string) ([]byte, error) {
	if input == "" && filename == "" {
		return nil, fmt.Errorf("input and filename must not both be empty")
	}
	if input != "" && filename != "" {
		return nil, fmt.Errorf("input or filename must be empty")
	}

	if input != "" {
		return []byte(input), nil
	}

	p, err := ReadFromFileOrStdin(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return p, nil
}

// ReadFromFileOrStdin reads the given file. If the filename is "-" it reads from stdin instead.
func ReadFromFileOrStdin(filename string) ([]byte, error) {
	var r io.Reader
	if filename == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("opening file %q: %w", filename, err)
		}
		r = f
	}
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading payload: %w", err)
	}
	return content, nil
}

// AuthSecret returns a secret using the secret file as given in
// environment variable. In case of development it uses the development
// password.
func AuthSecret(pwFile string, devEnv string) ([]byte, error) {
	if dev, _ := strconv.ParseBool(devEnv); dev {
		// Error value does not matter here. In case of an error dev is false and
		// this is the expected behavior.
		return []byte(developmentPassword), nil
	}
	pw, err := os.ReadFile(pwFile)
	if err != nil {
		return nil, fmt.Errorf("reading secret file %q: %w", pwFile, err)
	}
	return pw, nil
}

// BasicAuth contains the password used in basic authorization process. The password will be encoded in base64.
// The struct implements https://pkg.go.dev/google.golang.org/grpc@v1.38.0/credentials#PerRPCCredentials
type BasicAuth struct {
	Password []byte
}

// EncPassword returns the password encoded in base 64.
func (a BasicAuth) EncPassword() string {
	return base64.StdEncoding.EncodeToString(a.Password)
}

// GetRequestMetadata gets the current request metadata.
// See https://pkg.go.dev/google.golang.org/grpc@v1.38.0/credentials#PerRPCCredentials
func (a BasicAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": a.EncPassword(),
	}, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security.
// See https://pkg.go.dev/google.golang.org/grpc@v1.38.0/credentials#PerRPCCredentials
func (a BasicAuth) RequireTransportSecurity() bool {
	return false
}

const (
	lvlDebug    = 1
	lvlInfo     = 2
	lvlWarning  = 3
	lvlError    = 4
	lvlCritical = 5
)

// Logger is a logger that provides logging with respect to the log level.
type Logger struct {
	logger *log.Logger
	lvl    int
}

// NewLogger returns a default logger with respect to the given log level.
func NewLogger(level string) (Logger, error) {
	lvl := 0
	switch strings.ToLower(level) {
	case "debug":
		lvl = lvlDebug
	case "info":
		lvl = lvlInfo
	case "warning":
		lvl = lvlWarning
	case "error":
		lvl = lvlError
	case "critical":
		lvl = 5
	default:
		return Logger{}, fmt.Errorf("invalid log level %q", level)
	}
	l := Logger{
		logger: log.Default(),
		lvl:    lvl,
	}
	return l, nil

}

// Debugf calls logger.Printf but only in case of log level debug.
func (l Logger) Debugf(format string, v ...interface{}) {
	if l.lvl == lvlDebug {
		l.logger.Printf("[DEBUG] "+format, v...)
	}
}

// Infof calls logger.Printf but only in case of log level info or lower.
func (l Logger) Infof(format string, v ...interface{}) {
	if l.lvl <= lvlInfo {
		l.logger.Printf("[INFO] "+format, v...)
	}
}

// Warningf calls logger.Printf but only in case of log level warning or lower.
func (l Logger) Warningf(format string, v ...interface{}) {
	if l.lvl <= lvlWarning {
		l.logger.Printf("[WARNING] "+format, v...)
	}
}

// TODO: Add methods for error and critical
