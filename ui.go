package ortfodb

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	ll "github.com/ewen-lbh/label-logger-go"
	"github.com/mitchellh/colorstring"
	"github.com/xeipuuv/gojsonschema"
)

var LogFilePath string
var PrependDateToLogs = false

// noAnsiCodesWriter is an io.Writer that writes to the underlying writer, but strips ANSI color codes beforehand
type noAnsiCodesWriter struct {
	out io.Writer
}

func (w noAnsiCodesWriter) Write(p []byte) (n int, err error) {
	return w.out.Write(stripansicolors(p))
}

// Println is like fmt.Println but automatically strips ANSI color codes if colors are disabled
func Println(a ...interface{}) {
	var writer io.Writer
	writer = os.Stdout
	if !ll.ShowingColors() {
		writer = noAnsiCodesWriter{out: writer}
	}
	fmt.Fprintln(writer, a...)
}

// Printf is like fmt.Printf but automatically strips ANSI color codes if colors are disabled
func Printf(format string, a ...interface{}) {
	var writer io.Writer
	writer = os.Stdout
	if !ll.ShowingColors() {
		writer = noAnsiCodesWriter{out: writer}
	}
	fmt.Fprintf(writer, format, a...)
}

// Print is like fmt.Print but automatically strips ANSI color codes if colors are disabled
func Print(a ...interface{}) {
	var writer io.Writer
	writer = os.Stdout
	if !ll.ShowingColors() {
		writer = noAnsiCodesWriter{out: writer}
	}
	fmt.Fprint(writer, a...)
}

func indentSubsequent(size int, text string) string {
	indentation := strings.Repeat(" ", size)
	return strings.ReplaceAll(text, "\n", "\n"+indentation)
}

func ExporterLogCustom(exporter Exporter, verb string, color string, message string, fmtArgs ...interface{}) {
	if debugging {
		ll.Log(verb, color, fmt.Sprintf("[dim][bold](from exporter %s)[reset] %s", exporter.Name(), message), fmtArgs...)
	} else {
		ll.Log(verb, color, message, fmtArgs...)
	}
}

func ExporterLogCustomNoFormatting(exporter Exporter, verb string, color string, message string) {
	if debugging {
		ll.LogNoFormatting(verb, color, colorstring.Color("[dim][bold](from exporter "+exporter.Name()+")[reset] ")+message)
	} else {
		ll.LogNoFormatting(verb, color, message)
	}
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

func stripansicolors(b []byte) []byte {
	// TODO find a way to do this without converting to string
	s := string(b)
	s = regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(s, "")
	return []byte(s)
}

// neutralizeColorstring strips colorstring syntax from s
func neutralizeColostring(s string) string {
	return string(stripansicolors([]byte(colorstring.Color(s))))
}
