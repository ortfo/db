package ortfodb

import (
	"fmt"

	ll "github.com/gwennlbh/label-logger-go"
)

type Exporter interface {
	Name() string
	Description() string
	Before(ctx *RunContext, opts PluginOptions) error
	Export(ctx *RunContext, opts PluginOptions, work *Work) error
	After(ctx *RunContext, opts PluginOptions, built *Database) error
	OptionsType() any
}

type ExporterManifest struct {
	// The name of the exporter
	Name string `yaml:"name"`

	// Some documentation about the exporter
	Description string `yaml:"description"`

	// Commands to run before the build starts. Go text template that receives .Data
	Before []PluginCommand `yaml:"before,omitempty"`

	// Commands to run after the build finishes. Go text template that receives .Data and .Database, the built database.
	After []PluginCommand `yaml:"after,omitempty"`

	// Commands to run during the build, for each work. Go text template that receives .Data and .Work, the current work.
	Work []PluginCommand `yaml:"work,omitempty"`

	// Initial data
	Data map[string]any `yaml:"data,omitempty"`

	// If true, will show every command that is run
	Verbose bool `yaml:"verbose,omitempty"`

	// List of programs that are required to be available in the PATH for the exporter to run.
	Requires []string `yaml:"requires,omitempty"`
}

// ExporterOptions validates then returns the configuration options for the given exporter.
func (ctx *RunContext) ExporterOptions(exporter Plugin) (PluginOptions, error) {
	options := ctx.Config.Exporters[exporter.Name()]
	err := ValidatePluginOptions(exporter, options)
	if err != nil {
		return nil, err
	}

	return options, nil
}

func BuiltinExporters() (exporters []Exporter) {
	plugins := BuiltinPlugins("exporters", &SqlExporter{}, &LocalizeExporter{})

	for _, plugin := range plugins {
		if exporter, ok := plugin.(Exporter); ok {
			exporters = append(exporters, exporter)
		} else {
			ll.Warn("Plugin %s is not an exporter, skipping", plugin.Name())
		}
	}

	return
}

func (ctx *RunContext) FindExporter(name string) (Exporter, error) {
	builtins := make([]Plugin, 0)
	for _, exporter := range BuiltinExporters() {
		builtins = append(builtins, exporter)
	}

	result, err := ctx.FindPlugin(name, builtins, ctx.Config.Exporters)
	if err != nil {
		return nil, err
	}

	if exporter, ok := result.(Exporter); ok {
		return exporter, nil
	} else {
		return nil, fmt.Errorf("plugin %q is not an exporter", name)
	}
}
