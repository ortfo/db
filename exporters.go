package ortfodb

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ll "github.com/ewen-lbh/label-logger-go"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

type ExporterOptions map[string]interface{}

type Exporter interface {
	Name() string
	Description() string
	Before(ctx *RunContext, opts ExporterOptions) error
	Export(ctx *RunContext, opts ExporterOptions, work *Work) error
	After(ctx *RunContext, opts ExporterOptions, built *Database) error
	OptionsType() any
}

type ExporterCommand struct {
	// Run a command in a shell
	Run string `yaml:"run,omitempty"`
	// Log a message. The first argument is the verb, the second is the color, the third is the message.
	Log []string `yaml:"log,omitempty"`
}

type ExporterManifest struct {
	// The name of the exporter
	Name string `yaml:"name"`

	// Some documentation about the exporter
	Description string `yaml:"description"`

	// Commands to run before the build starts. Go text template that receives .Data
	Before []ExporterCommand `yaml:"before,omitempty"`

	// Commands to run after the build finishes. Go text template that receives .Data and .Database, the built database.
	After []ExporterCommand `yaml:"after,omitempty"`

	// Commands to run during the build, for each work. Go text template that receives .Data and .Work, the current work.
	Work []ExporterCommand `yaml:"work,omitempty"`

	// Initial data
	Data map[string]any `yaml:"data,omitempty"`

	// If true, will show every command that is run
	Verbose bool `yaml:"verbose,omitempty"`

	// List of programs that are required to be available in the PATH for the exporter to run.
	Requires []string `yaml:"requires,omitempty"`
}

// ExporterOptions validates then returns the configuration options for the given exporter.
func (ctx *RunContext) ExporterOptions(exporter Exporter) (ExporterOptions, error) {
	options := ctx.Config.Exporters[exporter.Name()]
	err := ValidateExporterOptions(exporter, options)
	if err != nil {
		return nil, err
	}

	return options, nil
}

func ValidateExporterOptions(exporter Exporter, opts ExporterOptions) error {
	validationErrors := ValidateAsJSONSchema(exporter.OptionsType(), true, opts)

	if len(validationErrors) > 0 {
		DisplayValidationErrors(validationErrors, "configuration", "exporters", exporter.Name())
		return fmt.Errorf("the configuration file is invalid. See validation errors above")
	}
	return nil
}

var BuiltinNativeExporters = []Exporter{
	&SqlExporter{},
	&LocalizeExporter{},
	&CustomExporter{},
}

//go:embed exporters/*.yaml
var builtinYAMLExportersFiles embed.FS

func BuiltinExporters() []Exporter {
	exporters := make([]Exporter, 0)
	exporterFiles, err := builtinYAMLExportersFiles.ReadDir("exporters")
	if err != nil {
		panic(fmt.Errorf("error while reading builtin yaml exporters directory (shouldn't happen, it should've been go:embed'd): %w", err))
	}

	exporters = append(exporters, BuiltinNativeExporters...)

	for _, exporterFile := range exporterFiles {
		file := filepath.Join("exporters", exporterFile.Name())
		contents, err := builtinYAMLExportersFiles.ReadFile(file)
		if err != nil {
			panic(fmt.Errorf("error while reading builtin yaml exporter file %s (shouldn't happen, it should've been go:embed'd): %w", file, err))
		}

		exporter, err := LoadExporter(strings.TrimSuffix(exporterFile.Name(), ".yaml"), contents, map[string]any{})
		if err != nil {
			continue
		}

		exporters = append(exporters, exporter)
	}

	return exporters
}

func (ctx *RunContext) FindExporter(name string) (Exporter, error) {
	for _, exporter := range BuiltinExporters() {
		if exporter.Name() == name {
			return exporter, nil
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
			return &CustomExporter{}, fmt.Errorf("while reading local manifest file at %s: %w", name, err)
		}
		return LoadExporter(name, rawManifest, ctx.Config.Exporters[name])
	} else if isValidURL(ensureHttpPrefix(name)) {
		url := ensureHttpPrefix(name)
		ll.Debug("No builtin exporter named %s, attempting download since %s looks like an URL…", name, url)
		return DownloadExporter(name, url, ctx.Config.Exporters[name])
	}
	return nil, fmt.Errorf("no exporter named %s", name)
}

// LoadExporter loads an exporter from a manifest YAML file's contents.
func LoadExporter(name string, manifestRaw []byte, config map[string]any) (*CustomExporter, error) {
	var manifest ExporterManifest
	err := yaml.Unmarshal(manifestRaw, &manifest)
	if err != nil {
		return &CustomExporter{}, fmt.Errorf("while parsing exporter manifest file: %w", err)
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

	exporter := CustomExporter{
		data:     merge(manifest.Data, config),
		name:     name,
		Manifest: manifest,
		verbose:  verbose,
		dryRun:   dryRun,
	}

	return &exporter, nil
}

// DownloadExporter loads an exporter from a URL.
func DownloadExporter(name string, url string, config map[string]any) (*CustomExporter, error) {
	ll.Log("Installing", "cyan", "exporter at %s", url)
	manifestRaw, err := downloadFile(url)
	if err != nil {
		return &CustomExporter{}, fmt.Errorf("while downloading exporter manifest file: %w", err)
	}

	exporter, err := LoadExporter(name, manifestRaw, config)
	if err != nil {
		return &CustomExporter{}, err
	}

	exporter.name = name
	return exporter, nil
}

// GetExporterOptions returns the options for the given exporter.
// Use it to get your options in a nice struct. The struct will be of the same type as the one returned by e.OptionsType().
// Example:
//
//	type MyExporterOptions struct {
//			// Some option
//			Option string `yaml:"option"`
//	}
//
//	func (e *MyExporter) OptionsType() any {
//	 	return MyExporterOptions{}
//	 }
//
//	func (e *MyExporter) After(ctx *ortfodb.RunContext, opts *ortfodb.ExporterOptions, db *ortfodb.Database) error {
//		 options := GetExporterOptions[MyExporterOptions](e, opts)
//		 // Now you can use options as a MyExporterOptions struct
//		}
func GetExporterOptions[ConcreteOptionsType any](e Exporter, opts ExporterOptions) ConcreteOptionsType {
	options := e.OptionsType()
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &options,
		TagName: "yaml",
	})
	decoder.Decode(opts)
	return options.(ConcreteOptionsType)
}
