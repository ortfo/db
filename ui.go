package ortfodb

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/mitchellh/colorstring"
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

func ExporterLogCustom(exporter Exporter, verb string, color string, message string, fmtArgs ...interface{}) {
	if os.Getenv("DEBUG") == "1" {
		LogCustom(verb, color, fmt.Sprintf("[dim][bold](from exporter %s)[reset] %s", exporter.Name(), message), fmtArgs...)
	} else {
		LogCustom(verb, color, message, fmtArgs...)
	}
}

func ExporterLogCustomNoFormatting(exporter Exporter, verb string, color string, message string) {
	if os.Getenv("DEBUG") == "1" {
		LogCustomNoFormatting(verb, color, colorstring.Color("[dim][bold](from exporter "+exporter.Name()+")[reset] ")+message)
	} else {
		LogCustomNoFormatting(verb, color, message)
	}
}

func LogCustom(verb string, color string, message string, fmtArgs ...interface{}) {
	LogCustomNoFormatting(verb, color, colorstring.Color(fmt.Sprintf(message, fmtArgs...)))
}

func LogCustomNoFormatting(verb string, color string, message string) {
	fmt.Fprintln(
		logWriter(),
		colorstring.Color(fmt.Sprintf("[bold][%s]%15s[reset]", color, verb))+
			" "+
			indentSubsequent(15+1, message),
	)
}

// DisplayValidationErrors takes in a slice of json schema validation errors and displays them nicely to in the terminal.
func DisplayValidationErrors(errors []gojsonschema.ResultError, filename string, rootPath ...string) {
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
		colorstring.Println("- " + strings.ReplaceAll(displayValidationErrorFieldPath(err.Field(), rootPath...), ".", "[blue][bold]/[reset]"))
		colorstring.Println("    [red]" + err.Description())
	}
}

func displayValidationErrorFieldPath(field string, rootPath ...string) string {
	if field == "(root)" {
		field = ""
	}
	for i, fragment := range rootPath {
		if strings.Contains(fragment, "/") {
			rootPath[i] = fmt.Sprintf("%q", fragment)
		}
	}
	return strings.Join(append(rootPath, field), "/")
}

// LogError logs non-fatal errors.
func (ctx *RunContext) LogError(message string, fmtArgs ...interface{}) {
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

func isInteractiveTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stderr.Fd())
}
