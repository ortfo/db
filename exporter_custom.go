package ortfodb

import (
	ll "github.com/gwennlbh/label-logger-go"
)

func (e *CustomPlugin) Before(ctx *RunContext, opts PluginOptions) error {
	ll.Debug("Running before commands for %s", e.name)
	err := e.VerifyRequiredPrograms()
	if err != nil {
		return err
	}
	ll.Debug("Setting user-supplied data for exporter %s: %v", e.name, opts)
	e.data = merge(e.Manifest.Data, opts)
	if e.Manifest.Verbose {
		PluginLogCustom(e, "Debug", "magenta", ".Data for %s is %v", e.name, e.data)
	}

	return e.runCommands(ctx, e.verbose, ".", e.Manifest.Commands["before"], map[string]any{})

}

func (e *CustomPlugin) Export(ctx *RunContext, opts PluginOptions, work *Work) error {

	return e.runCommands(ctx, e.verbose, ".", e.Manifest.Commands["work"], map[string]any{
		"Work": work,
	})
}

func (e *CustomPlugin) After(ctx *RunContext, opts PluginOptions, db *Database) error {

	return e.runCommands(ctx, e.verbose, ".", e.Manifest.Commands["after"], map[string]any{
		"Database": db,
	})
}
