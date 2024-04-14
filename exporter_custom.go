package ortfodb

import (
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	jsoniter "github.com/json-iterator/go"
)

type CustomExporter struct {
	data     map[string]any
	name     string
	manifest ExporterManifest
	verbose  bool
	dryRun   bool
	cwd      string
}

func (e *CustomExporter) Name() string {
	return e.name
}

func (e *CustomExporter) OptionsType() any {
	return e.manifest.Data
}

func (e *CustomExporter) Before(ctx *RunContext, opts ExporterOptions) error {
	return e.runCommands(ctx, e.manifest.Verbose, e.manifest.Before, map[string]any{})

}

func (e *CustomExporter) Export(ctx *RunContext, opts ExporterOptions, work *AnalyzedWork) error {
	return e.runCommands(ctx, e.verbose, e.manifest.Work, map[string]any{
		"Work": work,
	})
}

func (e *CustomExporter) After(ctx *RunContext, opts ExporterOptions, db *Database) error {

	return e.runCommands(ctx, e.verbose, e.manifest.After, map[string]any{
		"Database": db,
	})
}

func (e *CustomExporter) runCommands(ctx *RunContext, verbose bool, commands []ExporterCommand, additionalData map[string]any) error {
	for _, command := range commands {
		if command.Run != "" {
			commandline := e.renderCommandParts(ctx, []string{command.Run}, additionalData)[0]
			if commandline == "" {
				continue
			}
			if verbose {
				ExporterLogCustom(e, "Running", "yellow", commandline)
			}

			var stderrBuf strings.Builder
			var stdoutBuf strings.Builder

			proc := exec.Command("bash", "-c", commandline)
			proc.Dir = e.cwd
			proc.Stderr = &stderrBuf
			proc.Stdout = &stdoutBuf
			err := proc.Run()

			stdout := strings.TrimSpace(stdoutBuf.String())
			stderr := strings.TrimSpace(stderrBuf.String())

			if stdout != "" {
				ExporterLogCustom(e, ">", "blue", stdout)
			}
			if stderr != "" {
				if !verbose {
					ExporterLogCustom(e, "Error", "red", "While running %s\n%s", commandline, stderr)
				} else {
					ExporterLogCustom(e, "!", "red", stderr)
				}
			}

			if err != nil {
				return fmt.Errorf("while running %s: %w", commandline, err)
			}
		} else {
			logParts := e.renderCommandParts(ctx, command.Log, additionalData)
			ExporterLogCustom(e, logParts[0], logParts[1], logParts[2])
		}
	}
	return nil
}

var funcmap = template.FuncMap{
	"json": func(data any) string {
		bytes, err := jsoniter.ConfigFastest.Marshal(data)
		if err != nil {
			return "{}"
		}
		return string(bytes)
	},
}

func (e *CustomExporter) renderCommandParts(ctx *RunContext, commands []string, additionalData map[string]any) []string {
	output := make([]string, 0, len(commands))
	for _, command := range commands {
		tmpl, err := template.New("top").Funcs(funcmap).Parse(command)
		if err != nil {
			ctx.DisplayError("custom exporter: while parsing command %s", err, command)
			return []string{}
		}
		var buf strings.Builder
		err = tmpl.Execute(&buf, merge(additionalData, map[string]any{
			"Data":    e.data,
			"Ctx":     ctx,
			"Verbose": e.verbose,
			"DryRun":  e.dryRun,
		}))
		if err != nil {
			ctx.DisplayError("custom exporter: while rendering command %s", err, command)
			return []string{}
		}
		output = append(output, buf.String())
	}
	return output
}
