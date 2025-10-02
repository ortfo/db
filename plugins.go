package ortfodb

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ll "github.com/gwennlbh/label-logger-go"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

type Plugin interface {
	Name() string
	Description() string
	OptionsType() any
}

type PluginManifest struct {
	// The name of the importer
	Name string `yaml:"name"`

	// Some documentation about the importer
	Description string `yaml:"description"`

	// Initial data
	Data map[string]any `yaml:"data,omitempty"`

	// If true, will show every command that is run
	Verbose bool `yaml:"verbose,omitempty"`

	// List of programs that are required to be available in the PATH for the importer to run.
	Requires []string `yaml:"requires,omitempty"`

	// Commands of manifest, specific to the plugin type (exporter or importer)
	Commands map[string][]PluginCommand `yaml:",inline"`
}

// LoadPlugin loads an importer from a manifest YAML file's contents.
func LoadPlugin(name string, manifestRaw []byte, config map[string]any) (*CustomPlugin, error) {
	var manifest PluginManifest
	err := yaml.Unmarshal(manifestRaw, &manifest)
	if err != nil {
		return &CustomPlugin{}, fmt.Errorf("while parsing plugin manifest file: %w", err)
	}

	verbose, _ := config["verbose"].(bool)
	dryRun, ok := config["dry run"].(bool)
	if !ok {
		dryRun, ok = config["dry-run"].(bool)
		if !ok {
			dryRun, ok = config["dry_run"].(bool)
			if !ok {
				dryRun, _ = config["dryRun"].(bool)
			}
		}
	}

	importer := CustomPlugin{
		data:     merge(manifest.Data, config),
		name:     name,
		Manifest: manifest,
		verbose:  verbose,
		dryRun:   dryRun,
	}

	return &importer, nil
}

// DownloadPlugin loads an importer from a URL.
func DownloadPlugin(name string, url string, config map[string]any) (*CustomPlugin, error) {
	ll.Log("Installing", "cyan", "importer at %s", url)
	manifestRaw, err := downloadFile(url)
	if err != nil {
		return &CustomPlugin{}, fmt.Errorf("while downloading importer manifest file: %w", err)
	}

	importer, err := LoadPlugin(name, manifestRaw, config)
	if err != nil {
		return &CustomPlugin{}, err
	}

	importer.name = name
	return importer, nil
}

func (ctx *RunContext) FindPlugin(name string, builtins []Plugin, fromConfig map[string]map[string]any) (Plugin, error) {
	for _, plugins := range builtins {
		if plugins.Name() == name {
			return plugins, nil
		}
	}

	if strings.HasPrefix(name, "./") || strings.HasPrefix(name, "/") {
		var manifestPath string
		if filepath.IsAbs(name) {
			manifestPath = name
		} else {
			manifestPath = filepath.Join(filepath.Dir(ctx.Flags.Config), name)
		}

		rawManifest, err := os.ReadFile(manifestPath)
		if err != nil {
			return &CustomPlugin{}, fmt.Errorf("while reading local manifest file at %s: %w", name, err)
		}
		return LoadPlugin(name, rawManifest, fromConfig[name])
	} else if isValidURL(ensureHttpPrefix(name)) {
		url := ensureHttpPrefix(name)
		ll.Debug("No builtin plugin named %s, attempting download since %s looks like an URL…", name, url)
		return DownloadPlugin(name, url, fromConfig[name])
	}
	return nil, fmt.Errorf("no plugin named %s", name)
}

//go:embed importers/*.yaml exporters/*.yaml
var builtinYamlPlugins embed.FS

func BuiltinPlugins(directory string, natives ...Plugin) []Plugin {
	plugins := make([]Plugin, 0)
	pluginFiles, err := builtinYamlPlugins.ReadDir(directory)
	if err != nil {
		panic(fmt.Errorf("error while reading builtin yaml importers directory (shouldn't happen, it should've been go:embed'd): %w", err))
	}

	plugins = append(plugins, natives...)

	for _, exporterFile := range pluginFiles {
		// forces forward slashes for go:embed FS, even when running on Windows
		file := strings.Join([]string{directory, exporterFile.Name()}, "/")
		contents, err := builtinYamlPlugins.ReadFile(file)
		if err != nil {
			panic(fmt.Errorf("error while reading builtin yaml exporter file %s (shouldn't happen, it should've been go:embed'd): %w", file, err))
		}

		importer, err := LoadPlugin(strings.TrimSuffix(exporterFile.Name(), ".yaml"), contents, map[string]any{})
		if err != nil {
			continue
		}

		plugins = append(plugins, importer)
	}

	return plugins
}

type PluginOptions map[string]interface{}

// GetPluginOptions returns the options for the given exporter.
// Use it to get your options in a nice struct. The struct will be of the same type as the one returned by e.OptionsType().
// Example:
//
//	type MyPluginOptions struct {
//			// Some option
//			Option string `yaml:"option"`
//	}
//
//	func (e *MyPlugin) OptionsType() any {
//	 	return MyPluginOptions{}
//	 }
//
//	func (e *MyPlugin) After(ctx *ortfodb.RunContext, opts *ortfodb.PluginOptions, db *ortfodb.Database) error {
//		 options := GetPluginOptions[MyPluginOptions](e, opts)
//		 // Now you can use options as a MyPluginOptions struct
//		}
func GetPluginOptions[ConcreteOptionsType any](e Plugin, opts PluginOptions) ConcreteOptionsType {
	options := e.OptionsType()
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &options,
		TagName: "yaml",
	})
	decoder.Decode(opts)
	return options.(ConcreteOptionsType)
}

func ValidatePluginOptions(exporter Plugin, opts PluginOptions) error {
	validationErrors := ValidateAsJSONSchema(exporter.OptionsType(), true, opts)

	if len(validationErrors) > 0 {
		DisplayValidationErrors(validationErrors, "configuration", "exporters", exporter.Name())
		return fmt.Errorf("the configuration file is invalid. See validation errors above")
	}
	return nil
}
