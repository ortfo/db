package ortfodb

import (
	"fmt"

	ll "github.com/gwennlbh/label-logger-go"
)

type ImporterOptions map[string]any

type Importer interface {
	Name() string
	Description() string
	List(ctx *RunContext, opts ImporterOptions) ([]string, error)
	Import(ctx *RunContext, opts ImporterOptions, id string) error
	OptionsType() any
}

type ImporterManifest struct {
	// The name of the importer
	Name string `yaml:"name"`

	// Some documentation about the importer
	Description string `yaml:"description"`

	// Commands to run to list work IDs to import. Go text template that receives .Data.
	List []PluginCommand `yaml:"list,omitempty"`

	// Commands to run to import a work. Go text template that receives .Data and .ID, the current work ID.
	Import []PluginCommand `yaml:"import,omitempty"`

	// Initial data
	Data map[string]any `yaml:"data,omitempty"`

	// If true, will show every command that is run
	Verbose bool `yaml:"verbose,omitempty"`

	// List of programs that are required to be available in the PATH for the importer to run.
	Requires []string `yaml:"requires,omitempty"`
}

func (ctx *RunContext) ImporterOptions(importer Importer) (ImporterOptions, error) {
	options := ctx.Config.Importers[importer.Name()]
	err := ValidatePluginOptions(importer, options)
	if err != nil {
		return nil, err
	}

	return options, nil
}

func BuiltinImporters() (importers []Importer) {
	plugins := BuiltinPlugins("importers")

	for _, plugin := range plugins {
		if importer, ok := plugin.(Importer); ok {
			importers = append(importers, importer)
		} else {
			ll.Warn("Plugin %s is not an importer, skipping", plugin.Name())
		}
	}

	return
}

func (ctx *RunContext) FindImporter(name string) (Importer, error) {
	builtins := make([]Plugin, 0)
	for _, importer := range BuiltinImporters() {
		builtins = append(builtins, importer)
	}

	result, err := ctx.FindPlugin(name, builtins, ctx.Config.Importers)
	if err != nil {
		return nil, err
	}

	if importer, ok := result.(Importer); ok {
		return importer, nil
	} else {
		return nil, fmt.Errorf("plugin %q is not an importer", name)
	}
}
