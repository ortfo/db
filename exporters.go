package ortfodb

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type ExporterOptions interface{}

type Exporter interface {
	Name() string
	Before(ctx *RunContext, opts ExporterOptions) error
	Export(ctx *RunContext, opts ExporterOptions, work *AnalyzedWork) error
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

var BuiltinExporters = []Exporter{
	&SqlExporter{},
	&CustomExporter{},
}

// BuiltinYAMLExporters are exporters that can be accessed by their name directly. They should be available for download over the network at the github repository.
// TODO use go:embed instead
var BuiltinYAMLExporters = []string{"ssh-upload"}

func (ctx *RunContext) FindExporter(name string) (Exporter, error) {
	for _, exporter := range BuiltinExporters {
		if exporter.Name() == name {
			return exporter, nil
		}
	}
	for _, builtinName := range BuiltinYAMLExporters {
		if builtinName == name {
			return DownloadExporter(name, fmt.Sprintf("https://raw.githubusercontent.com/ortfodb/exporters/main/exporters/%s.yaml", name), ctx.Config.Exporters[name])
		}
	}
	if strings.HasPrefix(name, "./") {
		rawManifest, err := readFile(filepath.Join(filepath.Dir(ctx.Flags.Config), name))
		if err != nil {
			return &CustomExporter{}, fmt.Errorf("while reading local manifest file at %s: %w", name, err)
		}
		return LoadExporter(name, rawManifest, ctx.Config.Exporters[name])
	} else if isValidURL(ensureHttpPrefix(name)) {
		url := ensureHttpPrefix(name)
		ctx.LogDebug("No builtin exporter named %s, attempting download since %s looks like an URL…", name, url)
		return DownloadExporter(name, url, ctx.Config.Exporters[name])
	}
	return nil, fmt.Errorf("no exporter named %s", name)
}

// LoadExporter loads an exporter from a manifest YAML file's contents.
func LoadExporter(name string, manifestRaw string, config map[string]any) (*CustomExporter, error) {
	var manifest ExporterManifest
	err := yaml.Unmarshal([]byte(manifestRaw), &manifest)
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

	cwd := "."
	if !isValidURL(name) && fileExists(name) {
		cwd = filepath.Dir(name)
	}

	exporter := CustomExporter{
		data:     merge(manifest.Data, config),
		name:     name,
		manifest: manifest,
		verbose:  verbose,
		dryRun:   dryRun,
		cwd:      cwd,
	}
	return &exporter, nil
}

// DownloadExporter loads an exporter from a URL.
func DownloadExporter(name string, url string, config map[string]any) (*CustomExporter, error) {
	LogCustom("Installing", "cyan", "exporter at %s", url)
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