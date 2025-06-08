package ortfodb

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"al.essio.dev/pkg/shellescape"
	"github.com/Masterminds/sprig/v3"
	jsoniter "github.com/json-iterator/go"

	ll "github.com/gwennlbh/label-logger-go"
)

type PluginCommand struct {
	// Run a command in a shell
	Run string `yaml:"run,omitempty"`
	// Log a message. The first argument is the verb, the second is the color, the third is the message.
	Log []string `yaml:"log,omitempty"`
	// Set environment variables
	Env map[string]string `yaml:"env,omitempty"`
}

type CustomPlugin struct {
	data     map[string]any
	name     string
	Manifest PluginManifest
	verbose  bool
	dryRun   bool
}

func (e *CustomPlugin) VerifyRequiredPrograms() error {
	missingPrograms := make([]string, 0, len(e.Manifest.Requires))
	for _, program := range e.Manifest.Requires {
		_, err := exec.LookPath(program)
		if err != nil {
			missingPrograms = append(missingPrograms, program)
		}
	}
	if len(missingPrograms) > 0 {
		return fmt.Errorf("intall %s to use the %s importer", strings.Join(missingPrograms, ", "), e.name)
	}
	return nil
}

func (e *CustomPlugin) Name() string {
	return e.name
}

func (e *CustomPlugin) Description() string {
	return e.Manifest.Description
}

func (e *CustomPlugin) OptionsType() any {
	return e.Manifest.Data
}

// cwd is relative to the source file. Use "." if unsure.
func (e *CustomPlugin) runCommands(ctx *RunContext, verbose bool, cwd string, commands []PluginCommand, additionalData map[string]any) error {
	currentEnv := make(map[string]string)

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
			if verbose && (len(commandline) <= 100 || debugging) {
				PluginLogCustom(e, "Running", "yellow", commandline)
			}

			// check if windows

			var proc *exec.Cmd
			if runtime.GOOS != "windows" {
				proc = exec.Command("bash", "-c", commandline)
			} else {
				// os.WriteFile("command.ps1", []byte(commandline), 0644)
				proc = exec.Command("powershell", "-Command", commandline)
				// proc.SysProcAttr = &syscall.SysProcAttr{
				// 	CmdLine: commandline,
				// }
			}

			ll.Debug("exec.Command = %v", commandline)
			ll.Debug("cwd = %s", cwd)
			stderr, _ := proc.StderrPipe()
			stdout, _ := proc.StdoutPipe()
			proc.Env = os.Environ()
			for key, value := range currentEnv {
				proc.Env = append(proc.Env, fmt.Sprintf("%s=%s", key, value))
			}

			if filepath.IsAbs(cwd) {
				proc.Dir = filepath.Clean(cwd)
			// } else if cwd == "." {
			} else {
				proc.Dir = filepath.Clean(filepath.Join(filepath.Dir(ctx.Config.source), cwd))
			}
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
					PluginLogCustomNoFormatting(e, ">", "blue", line)
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
				// ImporterLogCustomNoFormatting(e, "Error", "red", fmt.Sprintf("While running %s\n%s", commandline, outputBuffer.String()))
				return fmt.Errorf("while running %s: %w", commandline, err)
			} else {
				// Hide output atfter it's done if there's no errors
				for i := 0; i < 6 && i < linesPrintedCount; i++ {
					if debugging {
						ll.Debug("would clear line %d", i)
					} else {
						fmt.Print("\033[1A\033[K")
					}
				}
			}
		} else if len(command.Env) > 0 {
			for key, value := range command.Env {
				renderedParts, err := e.renderCommandParts(ctx, []string{value}, additionalData, false)
				if err != nil {
					return fmt.Errorf("while rendering environment variable %s: %w", key, err)
				}
				currentEnv[key] = renderedParts[0]
				if verbose || debugging {
					PluginLogCustom(e, "Setting", "magenta", "environment variable %s = %q", key, currentEnv[key])
				}
			}
		} else {
			logParts, err := e.renderCommandParts(ctx, command.Log, additionalData, true)
			if err != nil {
				return fmt.Errorf("while rendering parts for a log instruction: %w", err)
			}

			if strings.TrimSpace(logParts[2]) != "" {
				PluginLogCustom(e, logParts[0], logParts[1], logParts[2])
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

func (e *CustomPlugin) renderCommandParts(ctx *RunContext, commands []string, additionalData map[string]any, recursive bool) ([]string, error) {
	output := make([]string, 0, len(commands))
	for _, command := range commands {
		tmpl, err := template.New("top").Funcs(sprig.TxtFuncMap()).Funcs(funcmap).Parse(command)
		if err != nil {
			return []string{}, fmt.Errorf("custom importer %s: while parsing template part %q: %w", e.Manifest.Name, neutralizeColostring(command), err)
		}
		var buf strings.Builder
		renderedData := e.data
		if recursive {
			renderedData, err = e.renderData(ctx)
			if err != nil {
				return []string{}, fmt.Errorf("while rendering data for command part: %w", err)
			}

		}
		ll.DebugNoColor("rendering command part %q, data=%v; renderedData=%v", command, e.data, renderedData)
		completeData := merge(additionalData, map[string]any{
			"Data":    renderedData,
			"Ctx":     ctx,
			"Verbose": e.verbose,
			"DryRun":  e.dryRun,
		})
		ll.DebugNoColor("rendering (recursive=%v) part %q with data %v", recursive, command, completeData)
		err = tmpl.Execute(&buf, completeData)
		if err != nil {
			return []string{}, fmt.Errorf("custom importer: while rendering template part %s: %w", neutralizeColostring(command), err)
		}
		output = append(output, buf.String())
	}
	return output, nil
}

func (e *CustomPlugin) renderData(ctx *RunContext) (map[string]any, error) {
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
