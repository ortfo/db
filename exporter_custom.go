package ortfodb

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/alessio/shellescape.v1"

	jsoniter "github.com/json-iterator/go"
)

type CustomExporter struct {
	data     map[string]any
	name     string
	Manifest ExporterManifest
	verbose  bool
	dryRun   bool
}

func (e *CustomExporter) VerifyRequiredPrograms() error {
	missingPrograms := make([]string, 0, len(e.Manifest.Requires))
	for _, program := range e.Manifest.Requires {
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
	return e.Manifest.Description
}

func (e *CustomExporter) OptionsType() any {
	return e.Manifest.Data
}

func (e *CustomExporter) Before(ctx *RunContext, opts ExporterOptions) error {
	LogDebug("Running before commands for %s", e.name)
	err := e.VerifyRequiredPrograms()
	if err != nil {
		return err
	}
	LogDebug("Setting user-supplied data for exporter %s: %v", e.name, opts)
	e.data = merge(e.Manifest.Data, opts)
	if e.Manifest.Verbose {
		ExporterLogCustom(e, "Debug", "magenta", ".Data for %s is %v", e.name, e.data)
	}
	return e.runCommands(ctx, e.verbose, e.Manifest.Before, map[string]any{})

}

func (e *CustomExporter) Export(ctx *RunContext, opts ExporterOptions, work *AnalyzedWork) error {
	return e.runCommands(ctx, e.verbose, e.Manifest.Work, map[string]any{
		"Work": work,
	})
}

func (e *CustomExporter) After(ctx *RunContext, opts ExporterOptions, db *Database) error {
	return e.runCommands(ctx, e.verbose, e.Manifest.After, map[string]any{
		"Database": db,
	})
}

func (e *CustomExporter) runCommands(ctx *RunContext, verbose bool, commands []ExporterCommand, additionalData map[string]any) error {
	for _, command := range commands {
		if command.Run != "" {
			commandlines_, err := e.renderCommandParts(ctx, []string{command.Run}, additionalData, true)
			if err != nil {
				return fmt.Errorf("while rendering commandline for run instruction: %w", err)
			}

			commandline := commandlines_[0]
			if commandline == "" {
				continue
			}
			if verbose && (len(commandline) <= 100 || debugging()) {
				ExporterLogCustom(e, "Running", "yellow", commandline)
			}

			proc := exec.Command("bash", "-c", commandline)
			LogDebug("exec.Command = %v", commandline)
			proc.Dir = filepath.Dir(ctx.Config.source)
			stderr, _ := proc.StderrPipe()
			stdout, _ := proc.StdoutPipe()
			err = proc.Start()
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

			linesPrintedCount := 0

			go func() {
				for line := range outputChannel {
					if linesPrintedCount > 5 {
						// Clear the line fives lines after the first output
						fmt.Print("\033[5A\033[K")
					}
					outputBuffer.WriteString(line + "\n")
					ExporterLogCustomNoFormatting(e, ">", "blue", line)
					if linesPrintedCount > 5 {
						// Go back to last line
						fmt.Print("\033[5B")
					}
					linesPrintedCount++
				}
			}()

			err = proc.Wait()
			close(outputChannel)
			if err != nil {
				// ExporterLogCustomNoFormatting(e, "Error", "red", fmt.Sprintf("While running %s\n%s", commandline, outputBuffer.String()))
				return fmt.Errorf("while running %s: %w", commandline, err)
			} else {
				// Hide output atfter it's done if there's no errors
				for i := 0; i < 6 && i < linesPrintedCount; i++ {
					if debugging() {
						LogDebug("would clear line %d", i)
					} else {
						fmt.Print("\033[1A\033[K")
					}
				}
			}
		} else {
			logParts, err := e.renderCommandParts(ctx, command.Log, additionalData, true)
			if err != nil {
				return fmt.Errorf("while rendering parts for a log instruction: %w", err)
			}

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

func (e *CustomExporter) renderCommandParts(ctx *RunContext, commands []string, additionalData map[string]any, recursive bool) ([]string, error) {
	output := make([]string, 0, len(commands))
	for _, command := range commands {
		tmpl, err := template.New("top").Funcs(sprig.TxtFuncMap()).Funcs(funcmap).Parse(command)
		if err != nil {
			return []string{}, fmt.Errorf("custom exporter %s: while parsing template part %q: %w", e.Manifest.Name, neutralizeColostring(command), err)
		}
		var buf strings.Builder
		renderedData := e.data
		if recursive {
			renderedData, err = e.renderData(ctx)
			if err != nil {
				return []string{}, fmt.Errorf("while rendering data for command part: %w", err)
			}

		}
		LogDebugNoColor("rendering command part %q, data=%v; renderedData=%v", command, e.data, renderedData)
		completeData := merge(additionalData, map[string]any{
			"Data":    renderedData,
			"Ctx":     ctx,
			"Verbose": e.verbose,
			"DryRun":  e.dryRun,
		})
		LogDebugNoColor("rendering (recursive=%v) part %q with data %v", recursive, command, completeData)
		err = tmpl.Execute(&buf, completeData)
		if err != nil {
			return []string{}, fmt.Errorf("custom exporter: while rendering template part %s: %w", neutralizeColostring(command), err)
		}
		output = append(output, buf.String())
	}
	return output, nil
}

func (e *CustomExporter) renderData(ctx *RunContext) (map[string]any, error) {
	rendered := make(map[string]any)
	for key, value := range e.data {
		switch value := value.(type) {
		case string:
			_rendered, err := e.renderCommandParts(ctx, []string{value}, map[string]any{}, false)
			if err != nil {
				return rendered, fmt.Errorf("while rendering %q: %w", value, err)
			}
			rendered[key] = _rendered[0]

		default:
			rendered[key] = value
		}
	}
	return rendered, nil
}
