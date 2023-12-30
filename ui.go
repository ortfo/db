package ortfodb

import (
	"fmt"
	"io"
	"os"
	"strings"

	// "time"

	// "github.com/mattn/go-isatty"
	"github.com/mitchellh/colorstring"
	// "github.com/theckman/yacspin"
	"github.com/xeipuuv/gojsonschema"
)

// DisplayValidationErrors takes in a slice of json schema validation errors and displays them nicely to in the terminal.
func DisplayValidationErrors(errors []gojsonschema.ResultError, filename string) {
	println("Your " + filename + " file is invalid. Here are the validation errors:\n")
	for _, err := range errors {
		/* FIXME: having a "." in the field name fucks up the display: eg:

		   - 0/media/fr-FR/2/online
		   Invalid type. Expected: boolean, given: string

		   if I replace fr-FR with fr.FR in the JSON:

		   			   ↓
		   - 0/media/fr/FR/2/online
		   Invalid type. Expected: boolean, given: string
		*/
		colorstring.Println("- " + strings.ReplaceAll(err.Field(), ".", "[blue][bold]/[reset]"))
		colorstring.Println("    [red]" + err.Description())
	}
}

// A yacspin spinner or a dummy spinner that does nothing.
// Used to avoid having to check for nil pointers everywhere when --silent is set.
type Spinner interface {
	Start() error
	Stop() error
	StopFail() error
	StopFailColors(colors ...string) error
	StopFailCharacter(char string)
	StopFailMessage(message string)
	Message(string)
	Pause() error
	Unpause() error
}

type dummySpinner struct {
}

func (d dummySpinner) Start() error                          { return nil }
func (d dummySpinner) Stop() error                           { return nil }
func (d dummySpinner) StopFail() error                       { return nil }
func (d dummySpinner) StopFailColors(colors ...string) error { return nil }
func (d dummySpinner) StopFailCharacter(char string)         {}
func (d dummySpinner) StopFailMessage(message string)        {}
func (d dummySpinner) Message(string)                        {}
func (d dummySpinner) Pause() error                          { return nil }
func (d dummySpinner) Unpause() error                        { return nil }

func (ctx *RunContext) CreateSpinner(outputFilename string) Spinner {
	// writer := os.Stdout

	// Don't clog stdout if we're not in a tty
	// if !isatty.IsTerminal(os.Stdout.Fd()) {
	// 	writer = os.Stderr
	// }

	if ctx.Flags.Silent {
		return dummySpinner{}
	}

	return dummySpinner{}

	// spinner, err := yacspin.New(yacspin.Config{
	// 	Writer:            writer,
	// 	Frequency:         100 * time.Millisecond,
	// 	Suffix:            " ",
	// 	Message:           "  0% Warming up",
	// 	CharSet:           yacspin.CharSets[14],
	// 	Colors:            []string{"fgCyan"},
	// 	StopCharacter:     "✓",
	// 	StopColors:        []string{"fgGreen"},
	// 	StopMessage:       colorstring.Color(fmt.Sprintf("Database written to [bold]./%s[reset]", outputFilename)),
	// 	StopFailCharacter: "✗",
	// 	StopFailColors:    []string{"fgRed"},
	// 	ShowCursor:        true, // XXX temporary, as currently the cursors is not shown back when the user Ctrl-Cs
	// })

	// if err != nil {
	// 	ctx.LogError("Couldn't start spinner: %s", err)
	// 	return dummySpinner{}
	// }

	// return spinner
}



// LogError logs non-fatal errors.
func (ctx *RunContext) LogError(message string, fmtArgs ...interface{}) {
	ctx.Spinner.Pause()
	var writer io.Writer = os.Stderr
	if progressBars != nil {
		writer = progressBars.Bypass()
	}
	colorstring.Fprintf(writer, "[red]          Error[reset] %s\n", fmt.Sprintf(message, fmtArgs...))
	ctx.Spinner.Unpause()
}

// LogInfo logs infos.
func (ctx *RunContext) LogInfo(message string, fmtArgs ...interface{}) {
	ctx.Spinner.Pause()
	var writer io.Writer = os.Stderr
	if progressBars != nil {
		writer = progressBars.Bypass()
	}
	colorstring.Fprintf(writer, "[blue]           Info[reset] %s\n", fmt.Sprintf(message, fmtArgs...))
	ctx.Spinner.Unpause()
}

// LogDebug logs debug information.
func (ctx *RunContext) LogDebug(message string, fmtArgs ...interface{}) {
	if os.Getenv("DEBUG") == "" {
		return
	}
	ctx.Spinner.Pause()
	var writer io.Writer = os.Stderr
	if progressBars != nil {
	writer = progressBars.Bypass()
	}
	colorstring.Fprintf(writer, "[magenta]          Debug[reset] %s\n", fmt.Sprintf(message, fmtArgs...))
	ctx.Spinner.Unpause()
}
