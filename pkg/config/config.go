package config

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
)

//go:embed templates
var deploymentTemplates embed.FS

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

// Cmd returns the subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config directory",
		Short: ConfigHelp,
		Long:  ConfigHelp + "\n\n" + ConfigHelpExtra,
		Args:  cobra.ExactArgs(1),
	}

	builtinTemplate := FlagBuiltinTemplate(cmd)
	tplFileOrDirName := FlagTpl(cmd)
	configFileNames := FlagConfig(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if *tplFileOrDirName != "" && *builtinTemplate != BuiltinTemplateDefault {
			return fmt.Errorf("flag --builtin-template must not be used together with flag --template")
		}
		dir := args[0]
		if err := Config(dir, *builtinTemplate, *tplFileOrDirName, *configFileNames); err != nil {
			return fmt.Errorf("running Config(): %w", err)
		}
		return nil
	}
	return cmd
}

// BuiltinTemplateDefault is the default builtin template which is used if the
// user does neigther provide a custom template nor give a builtin template.
const BuiltinTemplateDefault = "docker-compose"

func getBuiltinTemplateMsg() string {
	var allNames []string
	for _, obj := range allBuiltinTemplates {
		allNames = append(allNames, obj.name)
	}
	return fmt.Sprintf(" must be one of the following: %s", strings.Join(allNames, ", "))
}

// FlagBuiltinTemplate setups the builtin-template flag to the given cobra command.
func FlagBuiltinTemplate(cmd *cobra.Command) *string {
	return cmd.Flags().String("builtin-template", BuiltinTemplateDefault, "create files for this builtin deployment variant,"+getBuiltinTemplateMsg())
}

// FlagTpl setups the template flag to the given cobra command.
func FlagTpl(cmd *cobra.Command) *string {
	return cmd.Flags().StringP("template", "t", "", "file or directory for deployment template files, use this to provide a custom template")
}

// FlagConfig setups the config flag to the given cobra command.
func FlagConfig(cmd *cobra.Command) *[]string {
	return cmd.Flags().StringArrayP("config", "c", nil, "custom YAML config file, can be used more then once, ordering is important")
}

// Config rebuilds one or more (depending on template) files containing the
// deployment definitions. The parameters are just the command flags.
func Config(baseDir string, builtinTpl string, tplFileOrDirName string, configFileNames []string) error {
	// Create YAML config object
	cfg, err := NewYmlConfig(configFileNames)
	if err != nil {
		return fmt.Errorf("parsing configuration: %w", err)
	}

	// Create the base directory and the deployment files.
	if err := CreateDirAndFiles(baseDir, true, builtinTpl, tplFileOrDirName, cfg); err != nil {
		return fmt.Errorf("(re-)creating deployment files: %w", err)
	}

	return nil
}

type builtinTemplateFunc struct {
	name string
	fn   func(baseDir string, force bool, cfg *YmlConfig) error
}

var allBuiltinTemplates = []builtinTemplateFunc{
	{name: "docker-compose", fn: builtinTemplateDockerCompose},
	{name: "kubernetes", fn: builtinTemplateKubernetes},
}

func builtinTemplateByName(name string) (builtinTemplateFunc, bool) {
	for _, obj := range allBuiltinTemplates {
		if obj.name == name {
			return obj, true
		}
	}
	return builtinTemplateFunc{}, false
}

// CreateDirAndFiles creates the base directory and (re-)creates the deployment
// files according to the given template. If tplFileOrDirName is empty, the
// given builtin template file or directory is used. Use a truthy value for
// force to override existing files.
func CreateDirAndFiles(baseDir string, force bool, builtinTemplate string, tplFileOrDirName string, cfg *YmlConfig) error {
	if tplFileOrDirName == "" {
		obj, found := builtinTemplateByName(builtinTemplate)
		if !found {
			return fmt.Errorf("unknown builtin template %q,"+getBuiltinTemplateMsg(), builtinTemplate)
		}
		return obj.fn(baseDir, force, cfg)
	}

	fileInfo, err := os.Stat(tplFileOrDirName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("template file or directory %q does not exist", tplFileOrDirName)
		}
		return fmt.Errorf("checking file info of %q: %w", tplFileOrDirName, err)
	}

	if !fileInfo.IsDir() {
		return customTemplateSingleFile(baseDir, force, tplFileOrDirName, cfg)
	}

	return customTemplateDirectory(baseDir, force, tplFileOrDirName, cfg)
}

func builtinTemplateDockerCompose(baseDir string, force bool, cfg *YmlConfig) error {
	// Get default template file
	tplFile, err := deploymentTemplates.ReadFile(path.Join("templates", "docker-compose.yml"))
	if err != nil {
		return fmt.Errorf("reading template file: %w", err)
	}

	// Create directory
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory at %q: %w", baseDir, err)
	}

	// Create deployment file for Docker Compose
	filename := filepath.Join(baseDir, cfg.Filename)
	if err := CreateDeploymentFile(filename, force, tplFile, cfg); err != nil {
		return fmt.Errorf("creating deployment file %q: %w", filename, err)
	}

	return nil
}

func builtinTemplateKubernetes(baseDir string, force bool, cfg *YmlConfig) error {
	// Get default template directory
	tplDir, err := fs.Sub(deploymentTemplates, path.Join("templates", "kubernetes"))
	if err != nil {
		return fmt.Errorf("retrieving subtree: %w", err)
	}

	// Create directory
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory at %q: %w", baseDir, err)
	}

	// Create the deployment directory for Kubernetes
	if err := CreateDeploymentFilesFromTree(baseDir, force, tplDir, cfg); err != nil {
		return fmt.Errorf("creating deployment files at %q: %w", baseDir, err)
	}

	return nil
}

func customTemplateSingleFile(baseDir string, force bool, tplFilename string, cfg *YmlConfig) error {
	// Get file content
	tplFile, err := os.ReadFile(tplFilename)
	if err != nil {
		return fmt.Errorf("reading file %q: %w", tplFilename, err)
	}

	// Create directory
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory at %q: %w", baseDir, err)
	}

	// Create deployment file
	filename := filepath.Join(baseDir, cfg.Filename)
	if err := CreateDeploymentFile(filename, force, tplFile, cfg); err != nil {
		return fmt.Errorf("creating deployment file %q: %w", filename, err)
	}

	return nil
}

func customTemplateDirectory(baseDir string, force bool, tplDirname string, cfg *YmlConfig) error {
	tplDir := os.DirFS(tplDirname)

	// Create directory
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory at %q: %w", baseDir, err)
	}

	// Create the deployment directory for Kubernetes
	if err := CreateDeploymentFilesFromTree(baseDir, force, tplDir, cfg); err != nil {
		return fmt.Errorf("creating deployment files at %q: %w", baseDir, err)
	}

	return nil
}

// CreateDeploymentFilesFromTree walks through the FS containing templates and
// calls CreateDeploymentFile to render them one-by-one. Use a truthy value for
// force to override existing files.
func CreateDeploymentFilesFromTree(baseDir string, force bool, tplDir fs.FS, cfg *YmlConfig) error {
	if err := fs.WalkDir(tplDir, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walking templates directory at %q: %w", path, err)
		}

		filename := filepath.Join(baseDir, cfg.DeploymentDirectoryName, path)

		if d.IsDir() {
			if err := os.MkdirAll(filename, os.ModePerm); err != nil {
				return fmt.Errorf("creating directory at %q: %w", path, err)
			}
			return nil
		}

		fc, err := fs.ReadFile(tplDir, path)
		if err != nil {
			return fmt.Errorf("reading file %q: %w", path, err)
		}

		if err := CreateDeploymentFile(filename, force, fc, cfg); err != nil {
			return fmt.Errorf("creating deployment file %q: %w", filename, err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("walking templates directory: %w", err)
	}

	return nil
}

// CreateDeploymentFile builds a single deployment file to the given path. Use a
// truthy value for force to override an existing file.
func CreateDeploymentFile(filename string, force bool, tplFile []byte, cfg *YmlConfig) error {
	tmpl, err := template.New("Deployment File").Funcs(funcMap).Parse(string(tplFile))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	var res bytes.Buffer
	if err := tmpl.Execute(&res, cfg); err != nil {
		return fmt.Errorf("executing template %v: %w", tmpl, err)
	}

	if err := shared.CreateFile(filepath.Dir(filename), force, filepath.Base(filename), res.Bytes()); err != nil {
		return fmt.Errorf("creating deployment file at %q: %w", filename, err)
	}

	return nil
}

// YmlConfig contains the (merged) configuration for the creation of the deployment files.
type YmlConfig struct {
	Filename                string `yaml:"filename"`
	DeploymentDirectoryName string `yaml:"deploymentDirectoryName"`

	URL       string `yaml:"url"`
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
func NewYmlConfig(configFileNames []string) (*YmlConfig, error) {
	var configFiles [][]byte

	if len(configFileNames) > 0 {
		for _, configFileName := range configFileNames {
			fc, err := os.ReadFile(configFileName)
			if err != nil {
				return nil, fmt.Errorf("reading file %q: %w", configFileName, err)
			}
			configFiles = append(configFiles, fc)
		}
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

	return config, nil
}

// Static funcMap provides functions to be available to the templates.

var marshalContentFunc = func(ws int, v interface{}) (string, error) {
	y, err := yaml.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("marshalling content: %w", err)
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
var envMapToK8SFunc = func(v map[string]string) interface{} {
	var listOfMaps []map[string]string
	for key, value := range v {
		m := make(map[string]string)
		m["name"] = key
		m["value"] = value
		listOfMaps = append(listOfMaps, m)
	}
	return listOfMaps
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
		return "", fmt.Errorf("base64 decoding string %q: %w", s, err)
	}
	return string(decoded), nil
}
var readFileFunc = func(s string) (string, error) {
	b, err := os.ReadFile(s)
	if err != nil {
		return "", fmt.Errorf("reading file %q: %w", s, err)
	}
	return string(b), nil
}

var funcMap = template.FuncMap{
	"marshalContent": marshalContentFunc,
	"envMapToK8S":    envMapToK8SFunc,
	"checkFlag":      checkFlagFunc,
	"base64Encode":   base64EncodeFunc,
	"base64Decode":   base64DecodeFunc,
	"readFile":       readFileFunc,
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
