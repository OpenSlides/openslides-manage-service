package config

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing/fstest"
	"text/template"

	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
)

//go:embed templates
var deploymentTemplates embed.FS

const defaultDeploymentTemplate = "docker-compose"

//go:embed default-config.yml
var defaultConfig []byte

const (
	// ConfigHelp contains the short help text for the command.
	ConfigHelp = "(Re)creates the file(s) containing deployment definitions."

	// ConfigHelpExtra contains the long help text for the command without the headline.
	ConfigHelpExtra = `This command (re)creates the deployment file(s) in the given directory.`

	// ConfigCreateDefaultHelp contains the short help text for the command.
	ConfigCreateDefaultHelp = "(Re)creates the default setup configuration YAML file"

	// ConfigCreateDefaultHelpExtra contains the long help text for the command without the headline.
	ConfigCreateDefaultHelpExtra = `This command (re)creates the default setup configuration YAML file in the given directory.`
)

// Static FuncMap available to the templates
var marshalContentFunc = func(ws int, v interface{}) (string, error) {
	y, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	result := "\n"
	for _, line := range strings.Split(string(y), "\n") {
		if len(line) != 0 {
			result += fmt.Sprintf("%s%s\n", strings.Repeat(" ", ws), line)
		}
	}
	result = strings.TrimRight(result, "\n")
	return result, nil
}
var checkFlagFunc = func(v interface{}) (bool, error) {
	f, ok := v.(*bool)
	if !ok {
		return false, fmt.Errorf("using wrong type as argument in checkFlag function, only *bool is allowed")
	}
	return *f, nil
}
var base64EncodeFunc = func(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
var base64DecodeFunc = func(s string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", fmt.Errorf("base64 decoding string (%s): %w", s, err)
	}
	return string(decoded), nil
}
var readFileFunc = func(s string) (string, error) {
	b, err := os.ReadFile(s)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", s, err)
	}
	return string(b), nil
}

var funcMap = template.FuncMap{
	"marshalContent": marshalContentFunc,
	"checkFlag":      checkFlagFunc,
	"base64Encode":   base64EncodeFunc,
	"base64Decode":   base64DecodeFunc,
	"readFile":       readFileFunc,
}

// Cmd returns the subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config directory",
		Short: ConfigHelp,
		Long:  ConfigHelp + "\n\n" + ConfigHelpExtra,
		Args:  cobra.ExactArgs(1),
	}

	tplDirName := FlagTpl(cmd)
	configFileNames := FlagConfig(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		var err error

		if err = ValidateTpl(tplDirName); err != nil {
			return fmt.Errorf("validating configFileNames: %w", err)
		}

		if err = ValidateConfig(configFileNames); err != nil {
			return fmt.Errorf("validating configFileNames: %w", err)
		}

		if err := Config(dir, tplDirName, configFileNames); err != nil {
			return fmt.Errorf("running Config(): %w", err)
		}
		return nil
	}
	return cmd
}

// FlagTpl setups the template flag to the given cobra command.
func FlagTpl(cmd *cobra.Command) *string {
	return cmd.Flags().StringP("template", "t", defaultDeploymentTemplate, "Deployment files template directory (TODO: list presets)")
}

// ValidateTpl validates the provided option.
func ValidateTpl(tplDirName *string) error {
	if _, statErr := os.Stat(*tplDirName); statErr == nil {
		return nil
	}

	var embeddedPath = filepath.Join("templates", *tplDirName)
	if _, statErr := fs.Stat(deploymentTemplates, embeddedPath); statErr == nil {
		return nil
	}
	return fmt.Errorf("template file %q neither exists on disk nor is a known templeate", *tplDirName)
}

// ReadTplFiles reads the provided option and returns template files as a walkable FS
func ReadTplFiles(tplDirName *string, cfg *YmlConfig) (fs.FS, error) {
	var tplDirFs fs.FS = nil
	// If template dir is given and exists, read it
	if fInfo, statErr := os.Stat(*tplDirName); statErr == nil {
		if fInfo.IsDir() {
			tplDirFs = os.DirFS(*tplDirName)
		} else {
			// Be compatible when given a single file
			// -> mock an FS containing only that file
			var err error
			var content []byte
			content, err = os.ReadFile(*tplDirName)
			if err != nil {
				return nil, fmt.Errorf("reading template file %q: %w", *tplDirName, err)
			}
			tplDirFs = fstest.MapFS{fInfo.Name(): {Data: content}}
		}
	} else {
		var err error
		var embeddedPath = filepath.Join("templates", *tplDirName)

		_, err = fs.ReadDir(deploymentTemplates, embeddedPath)
		if err != nil {
			return nil, fmt.Errorf("checking for embedded template dir %q: %w", *tplDirName, err)
		}
		tplDirFs, err = fs.Sub(deploymentTemplates, embeddedPath)
		if err != nil {
			return nil, fmt.Errorf("reading embedded template files %q: %w", *tplDirName, err)
		}
	}
	return tplDirFs, nil
}

// FlagConfig setups the config flag to the given cobra command.
func FlagConfig(cmd *cobra.Command) *[]string {
	return cmd.Flags().StringArrayP("config", "c", nil, "custom YAML config file, can be used more then once, ordering is important")
}

// ValidateConfig validates the provided options
func ValidateConfig(configFileNames *[]string) error {
	if len(*configFileNames) > 0 {
		for _, configFileName := range *configFileNames {
			_, statErr := os.Stat(configFileName)
			if statErr != nil {
				return fmt.Errorf("stat file %q: %w", configFileName, statErr)
			}
		}
	}
	return nil
}

// ReadConfigFiles uses the provided option and reads and returns the appropriate FS.
func ReadConfigFiles(configFileNames *[]string) ([][]byte, error) {
	var configFiles [][]byte
	if configFileNames == nil {
		return configFiles, nil
	}

	if len(*configFileNames) > 0 {
		for _, configFileName := range *configFileNames {
			fc, err := os.ReadFile(configFileName)
			if err != nil {
				return nil, fmt.Errorf("reading file %q: %w", configFileName, err)
			}
			configFiles = append(configFiles, fc)
		}
	}
	return configFiles, nil
}

// Config rebuilds the deployment files
//
// A directory containing custom templates for the deployment files can be specified.
// If it doesn't exist it will be match against the embedded default templates.
func Config(dir string, tplDirName *string, configFileNames *[]string) error {
	var err error

	// Create directory
	if err = os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory at %q: %w", dir, err)
	}

	// Create YAML config file
	cfg, err := NewYmlConfig(configFileNames, dir)
	if err != nil {
		return fmt.Errorf("creating new YML config object: %w", err)
	}

	// Create the deployment file(s) from template dir
	if err := CreateDeploymentFilesFromTree(dir, true, tplDirName, cfg); err != nil {
		return fmt.Errorf("creating deployment files at %q: %w", dir, err)
	}

	return nil
}

// CreateDeploymentFilesFromTree walks through the FS containing templates and calls
// CreateDeploymentFile to render them one-by-one.
func CreateDeploymentFilesFromTree(outdir string, force bool, tplDirName *string, cfg *YmlConfig) error {
	var err error
	var tplDirFs fs.FS
	var dirEntries []fs.DirEntry
	if tplDirFs, err = ReadTplFiles(tplDirName, cfg); err != nil {
		return fmt.Errorf("reading template dir: %w", err)
	}

	// Be compatible when fs only contains a single file
	// -> mock an FS containing that file but rename to config.filename
	if dirEntries, err = fs.ReadDir(tplDirFs, "."); err != nil {
		return fmt.Errorf("listing templates directory: %w", err)
	}
	if len(dirEntries) == 1 && !dirEntries[0].IsDir() {
		var content []byte
		content, err = fs.ReadFile(tplDirFs, dirEntries[0].Name())
		if err != nil {
			return fmt.Errorf("reading single template file %q: %w", dirEntries[0].Name(), err)
		}
		tplDirFs = fstest.MapFS{cfg.Filename: {Data: content}}
	}

	if err := fs.WalkDir(tplDirFs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walking templates directory at %q: %w", path, err)
		}

		createFile := filepath.Join(outdir, path)
		if d.IsDir() {
			if err := os.MkdirAll(createFile, os.ModePerm); err != nil {
				return fmt.Errorf("creating directory at %q: %w", path, err)
			}
		} else {
			fc, err := fs.ReadFile(tplDirFs, path)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", path, err)
			}
			CreateDeploymentFile(createFile, force, fc, cfg)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// CreateDeploymentFile builds a single deployment file to the given path. Use a truthy value for force
// to override an existing file.
func CreateDeploymentFile(outfile string, force bool, tplFile []byte, cfg *YmlConfig) error {
	tmpl, err := template.New("Deployment File").Option("missingkey=error").Funcs(funcMap).Parse(string(tplFile))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	var res bytes.Buffer
	if err := tmpl.Execute(&res, cfg); err != nil {
		return fmt.Errorf("executing template %v: %w", tmpl, err)
	}

	if err := shared.CreateFile(filepath.Dir(outfile), force, filepath.Base(outfile), res.Bytes()); err != nil {
		return fmt.Errorf("creating deployment file at %q: %w", outfile, err)
	}

	return nil
}

// YmlConfig contains the (merged) configuration for the creation of the deployment files.
type YmlConfig struct {
	WorkingDirectory string

	Filename  string `yaml:"filename"`
	Url       string `yaml:"url"`
	StackName string `yaml:"stackName"`

	Host string `yaml:"host"`
	Port string `yaml:"port"`

	DisablePostgres  *bool `yaml:"disablePostgres"`
	DisableDependsOn *bool `yaml:"disableDependsOn"`
	EnableLocalHTTPS *bool `yaml:"enableLocalHTTPS"`
	EnableAutoHTTPS  *bool `yaml:"enableAutoHTTPS"`

	Defaults struct {
		ContainerRegistry string `yaml:"containerRegistry"`
		Tag               string `yaml:"tag"`
	} `yaml:"defaults"`

	DefaultEnvironment map[string]string `yaml:"defaultEnvironment"`

	Services map[string]service `yaml:"services"`
}

type service struct {
	ContainerRegistry string            `yaml:"containerRegistry"`
	Tag               string            `yaml:"tag"`
	Environment       map[string]string `yaml:"environment"`
	AdditionalContent json.RawMessage   `yaml:"additionalContent"`
	CustomKeys        map[string]string `yaml:"customKeys"`
}

// nullTransformer is used to fix a problem with mergo
// see https://github.com/imdario/mergo/issues/131#issuecomment-589844203
type nullTransformer struct{}

func (t *nullTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ.Kind() == reflect.Ptr {
		return func(dst, src reflect.Value) error {
			if dst.CanSet() && !src.IsNil() {
				dst.Set(src)
			}
			return nil
		}
	}
	return nil
}

// NewYmlConfig creates a ymlConfig object from all given files. The files were
// merged together with the default config.
func NewYmlConfig(configFileNames *[]string, dir string) (*YmlConfig, error) {
	var err error
	var configFiles [][]byte
	if configFiles, err = ReadConfigFiles(configFileNames); err != nil {
		return nil, fmt.Errorf("reading config files: %w", err)
	}

	allConfigFiles := [][]byte{
		defaultConfig,
	}
	allConfigFiles = append(allConfigFiles, configFiles...)

	// Unmarshal and merge them all
	config := new(YmlConfig)
	for _, configFile := range allConfigFiles {
		c := new(YmlConfig)
		if err := yaml.Unmarshal(configFile, c); err != nil {
			return nil, fmt.Errorf("unmarshaling YAML: %w", err)
		}
		if err := mergo.Merge(config, c, mergo.WithOverride, mergo.WithTransformers(&nullTransformer{})); err != nil {
			return nil, fmt.Errorf("merging config files: %w", err)
		}
	}

	// Fill services
	allServices := []string{
		"proxy",
		"client",
		"backendAction",
		"backendPresenter",
		"backendManage",
		"datastoreReader",
		"datastoreWriter",
		"postgres",
		"autoupdate",
		"auth",
		"vote",
		"search",
		"redis",
		"media",
		"icc",
		"manage",
	}
	if len(config.Services) == 0 {
		config.Services = make(map[string]service, len(allServices))
	}

	for _, name := range allServices {
		_, ok := config.Services[name]
		if !ok {
			config.Services[name] = service{}
		}
		s := config.Services[name]

		if s.ContainerRegistry == "" {
			s.ContainerRegistry = config.Defaults.ContainerRegistry
		}
		if s.Tag == "" {
			s.Tag = config.Defaults.Tag
		}

		config.Services[name] = s
	}

	// Set the WorkingDirectory
	config.WorkingDirectory = dir

	return config, nil
}

// CmdCreateDefault returns the config-create-default subcommand.
func CmdCreateDefault() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config-create-default directory",
		Short: ConfigCreateDefaultHelp,
		Long:  ConfigCreateDefaultHelp + "\n\n" + ConfigCreateDefaultHelpExtra,
		Args:  cobra.ExactArgs(1),
	}

	name := cmd.Flags().StringP("name", "n", "config.yml", "name of the created file")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		dir := args[0]

		if err := configCreateDefault(dir, *name); err != nil {
			return fmt.Errorf("running ConfigCreateDefault(): %w", err)
		}

		return nil
	}
	return cmd
}

// configCreateDefault creates the default setup configuration YAML file.
func configCreateDefault(dir, name string) error {
	// Create directory
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory at %q: %w", dir, err)
	}

	// Create file
	if err := shared.CreateFile(dir, true, name, defaultConfig); err != nil {
		return fmt.Errorf("creating config default file at %q: %w", dir, err)
	}
	return nil
}
