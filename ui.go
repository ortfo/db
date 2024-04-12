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

func logWriter() io.Writer {
	var writer io.Writer = os.Stderr
	if progressBars != nil {
		writer = progressBars.Bypass()
	}
	return writer
}

func indentSubsequent(size int, text string) string {
	indentation := strings.Repeat(" ", size)
	return strings.ReplaceAll(text, "\n", "\n"+indentation)
}

func LogCustom(verb string, color string, message string, fmtArgs ...interface{}) {
	fmt.Fprintln(logWriter(), colorstring.Color(fmt.Sprintf("[bold][%s]%15s[reset] %s", color, verb, indentSubsequent(15+1, fmt.Sprintf(message, fmtArgs...)))))
}

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

// LogError logs non-fatal errors.
func (ctx *RunContext) LogError(message string, fmtArgs ...interface{}) {
	// colorstring.Fprintf(logWriter(), "[red]          Error[reset] %s\n", fmt.Sprintf(message, fmtArgs...))
	LogCustom("Error", "red", message, fmtArgs...)
}

func (ctx *RunContext) DisplayError(msg string, err error, fmtArgs ...interface{}) {
	ctx.LogError(formatErrors(fmt.Errorf(msg+": %w", append(fmtArgs, err)...)))
}

func (ctx *RunContext) DisplayWarning(msg string, err error, fmtArgs ...interface{}) {
	ctx.LogWarning(formatErrors(fmt.Errorf(msg+": %w", append(fmtArgs, err)...)))
}

// LogInfo logs infos.
func (ctx *RunContext) LogInfo(message string, fmtArgs ...interface{}) {
	LogCustom("Info", "blue", message, fmtArgs...)
}

// LogDebug logs debug information.
func (ctx *RunContext) LogDebug(message string, fmtArgs ...interface{}) {
	if os.Getenv("DEBUG") == "" {
		return
	}
	LogCustom("Debug", "magenta", message, fmtArgs...)
}

// LogWarning logs warnings.
func (ctx *RunContext) LogWarning(message string, fmtArgs ...interface{}) {
	LogCustom("Warning", "yellow", message, fmtArgs...)
}

func formatList(list []string, format string, separator string) string {
	result := ""
	for i, tag := range list {
		sep := separator
		if i == len(list)-1 {
			sep = ""
		}
		result += fmt.Sprintf(format, tag) + sep
	}
	return result
}

// formatErrors returns a string where the error message was split on ': ', and each item is on a new line, indented once more than the previous line.
func formatErrors(err error) string {
	causes := strings.Split(err.Error(), ": ")
	output := ""
	for i, cause := range causes {
		output += strings.Repeat(" ", i) + cause
		if i < len(causes)-1 {
			output += "\n"
		}
	}
	return output
}
