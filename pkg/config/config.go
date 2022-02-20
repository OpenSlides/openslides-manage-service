package config

import (
	"bytes"
	_ "embed" // Blank import required to use go directive.
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
)

//go:embed default-docker-compose.yml
var defaultDockerComposeYml []byte

//go:embed default-config.yml
var defaultConfig []byte

const (
	// ConfigHelp contains the short help text for the command.
	ConfigHelp = "(Re)creates the container configuration YAML file for using Docker Compose or Docker Swarm"

	// ConfigHelpExtra contains the long help text for the command without the headline.
	ConfigHelpExtra = `This command (re)creates the container configuration YAML file in the given directory.`

	// ConfigCreateDefaultHelp contains the short help text for the command.
	ConfigCreateDefaultHelp = "(Re)creates the default setup configuration YAML file"

	// ConfigCreateDefaultHelpExtra contains the long help text for the command without the headline.
	ConfigCreateDefaultHelpExtra = `This command (re)creates the default setup configuration YAML file in the given directory.`
)

// Cmd returns the subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config directory",
		Short: ConfigHelp,
		Long:  ConfigHelp + "\n\n" + ConfigHelpExtra,
		Args:  cobra.ExactArgs(1),
	}

	tplFileName := FlagTpl(cmd)
	configFileNames := FlagConfig(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		dir := args[0]

		var tplFile []byte
		if *tplFileName != "" {
			fc, err := os.ReadFile(*tplFileName)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", *tplFileName, err)
			}
			tplFile = fc
		}

		var configFiles [][]byte
		if len(*configFileNames) > 0 {
			for _, configFileName := range *configFileNames {
				fc, err := os.ReadFile(configFileName)
				if err != nil {
					return fmt.Errorf("reading file %q: %w", configFileName, err)
				}
				configFiles = append(configFiles, fc)
			}
		}

		if err := Config(dir, tplFile, configFiles); err != nil {
			return fmt.Errorf("running Config(): %w", err)
		}
		return nil
	}
	return cmd
}

// FlagTpl setups the template flag to the given cobra command.
func FlagTpl(cmd *cobra.Command) *string {
	return cmd.Flags().StringP("template", "t", "", "custom YAML template file")
}

// FlagConfig setups the config flag to the given cobra command.
func FlagConfig(cmd *cobra.Command) *[]string {
	return cmd.Flags().StringArrayP("config", "c", nil, "custom YAML config file, can be use more then once, ordering is important")
}

// Config rebuilds the YAML file for using Docker Compose or Docker Swarm.
//
// A custom template for the YAML file and YAML configs can be provided.
func Config(dir string, tplFile []byte, configFiles [][]byte) error {
	// Create directory
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory at %q: %w", dir, err)
	}

	// Create YAML file
	cfg, err := NewYmlConfig(configFiles)
	if err != nil {
		return fmt.Errorf("creating new YML config object: %w", err)
	}

	if err := CreateYmlFile(dir, true, tplFile, cfg); err != nil {
		return fmt.Errorf("creating YAML file at %q: %w", dir, err)
	}

	return nil
}

// CreateYmlFile builds the YAML file at the given directory. Use a truthy value for force
// to override an existing file.
func CreateYmlFile(dir string, force bool, tplFile []byte, cfg *YmlConfig) error {
	if tplFile == nil {
		tplFile = defaultDockerComposeYml
	}

	marshalContentFunc := func(ws int, v interface{}) (string, error) {
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
	checkFlagFunc := func(v interface{}) (bool, error) {
		f, ok := v.(*bool)
		if !ok {
			return false, fmt.Errorf("using wrong type as argument in checkFlag function, only *bool is allowed")
		}
		return *f, nil
	}
	funcMap := template.FuncMap{}
	funcMap["marshalContent"] = marshalContentFunc
	funcMap["checkFlag"] = checkFlagFunc

	tmpl, err := template.New("YAML File").Option("missingkey=error").Funcs(funcMap).Parse(string(tplFile))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	var res bytes.Buffer
	if err := tmpl.Execute(&res, cfg); err != nil {
		return fmt.Errorf("executing template %v: %w", tmpl, err)
	}

	if err := shared.CreateFile(dir, force, cfg.Filename, res.Bytes()); err != nil {
		return fmt.Errorf("creating YAML file at %q: %w", dir, err)
	}

	return nil
}

// YmlConfig contains the (merged) configuration for the creation of the Docker
// Compose YAML file.
type YmlConfig struct {
	Filename string `yaml:"filename"`

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
func NewYmlConfig(configFiles [][]byte) (*YmlConfig, error) {
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
		"backend",
		"backendManage",
		"datastoreReader",
		"datastoreWriter",
		"postgres",
		"autoupdate",
		"auth",
		"vote",
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
