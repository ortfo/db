package ortfodb

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/alessio/shellescape.v1"

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

func (e *CustomExporter) VerifyRequiredPrograms() error {
	missingPrograms := make([]string, 0, len(e.manifest.Requires))
	for _, program := range e.manifest.Requires {
		_, err := exec.LookPath(program)
		if err != nil {
			missingPrograms = append(missingPrograms, program)
		}
	}
	if len(missingPrograms) > 0 {
		return fmt.Errorf("intall %s to use the %s exporter", strings.Join(missingPrograms, ", "), e.name)
	}
	return nil
}

func (e *CustomExporter) Name() string {
	return e.name
}

func (e *CustomExporter) Description() string {
	return e.manifest.Description
}

func (e *CustomExporter) OptionsType() any {
	return e.manifest.Data
}

func (e *CustomExporter) Before(ctx *RunContext, opts ExporterOptions) error {
	ctx.LogDebug("Running before commands for %s", e.name)
	err := e.VerifyRequiredPrograms()
	if err != nil {
		return err
	}
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
			commandline := e.renderCommandParts(ctx, []string{command.Run}, additionalData, true)[0]
			if commandline == "" {
				continue
			}
			if verbose && (len(commandline) <= 100 || debugging()) {
				ExporterLogCustom(e, "Running", "yellow", commandline)
			}

			proc := exec.Command("bash", "-c", commandline)
			proc.Dir = e.cwd
			stderr, _ := proc.StderrPipe()
			stdout, _ := proc.StdoutPipe()
			err := proc.Start()
			if err != nil {
				return fmt.Errorf("while starting command %q: %w", commandline, err)
			}

			outputBuffer := new(strings.Builder)
			outputChannel := make(chan string)

			// Goroutine to read from stdout and send lines to the output channel
			go func() {
				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					outputChannel <- scanner.Text()
				}
			}()

			// Goroutine to read from stderr and send lines to the output channel
			go func() {
				scanner := bufio.NewScanner(stderr)
				for scanner.Scan() {
					outputChannel <- scanner.Text()
				}
			}()

			linesPrinterCount := 0

			go func() {
				for line := range outputChannel {
					if linesPrinterCount > 5 {
						// Clear the line fives lines after the first output
						fmt.Print("\033[5A\033[K")
					}
					outputBuffer.WriteString(line + "\n")
					ExporterLogCustomNoFormatting(e, ">", "blue", line)
					if linesPrinterCount > 5 {
						// Go back to last line
						fmt.Print("\033[5B")
					}
					linesPrinterCount++
				}
			}()

			if err = proc.Wait(); err != nil {
				ExporterLogCustomNoFormatting(e, "Error", "red", fmt.Sprintf("While running %s\n%s", commandline, outputBuffer.String()))
				return fmt.Errorf("while running %s: %w", commandline, err)
			}
		} else {
			logParts := e.renderCommandParts(ctx, command.Log, additionalData, true)
			if strings.TrimSpace(logParts[2]) != "" {
				ExporterLogCustom(e, logParts[0], logParts[1], logParts[2])
			}
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
	"escape": func(data string) string {
		return shellescape.Quote(data)
	},
}

func (e *CustomExporter) renderCommandParts(ctx *RunContext, commands []string, additionalData map[string]any, recursive bool) []string {
	output := make([]string, 0, len(commands))
	for _, command := range commands {
		tmpl, err := template.New("top").Funcs(sprig.TxtFuncMap()).Funcs(funcmap).Parse(command)
		if err != nil {
			ctx.DisplayError("custom exporter: while parsing command %s", err, command)
			return []string{}
		}
		var buf strings.Builder
		renderedData := e.data
		if recursive {
			renderedData = e.renderData(ctx)
		}
		err = tmpl.Execute(&buf, merge(additionalData, map[string]any{
			"Data":    renderedData,
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

func (e *CustomExporter) renderData(ctx *RunContext) map[string]any {
	rendered := make(map[string]any)
	for key, value := range e.data {
		switch value := value.(type) {
		case string:
			rendered[key] = e.renderCommandParts(ctx, []string{value}, map[string]any{}, false)[0]
		default:
			rendered[key] = value
		}
	}
	return rendered
}
