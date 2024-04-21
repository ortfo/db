package ortfodb

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"time"

	"github.com/mattn/go-isatty"
	"github.com/mitchellh/colorstring"
	"github.com/xeipuuv/gojsonschema"
)

var LogFilePath string
var PrependDateToLogs = false
var showingTimingLogs = os.Getenv("DEBUG_TIMING") != ""

func logWriter(original io.Writer) io.Writer {
	writer := original
	if LogFilePath != "" {
		logfile, err := os.OpenFile(LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			writer = io.MultiWriter(writer, logfile)
		}
	}

	if PrependDateToLogs {
		writer = prependDateWriter{out: writer}
	}

	if progressBars != nil {
		writer = progressBars.Bypass()
	}
	if !ShowingColors() {
		writer = noAnsiCodesWriter{out: writer}
	}
	return writer
}

type prependDateWriter struct {
	out io.Writer
}

func (w prependDateWriter) Write(p []byte) (n int, err error) {
	return w.out.Write([]byte(
		fmt.Sprintf("[%s] %s",
			time.Now(),
			strings.TrimLeft(string(p), " "),
		)))
}

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
	if !ShowingColors() {
		writer = noAnsiCodesWriter{out: writer}
	}
	fmt.Fprintln(writer, a...)
}

// Printf is like fmt.Printf but automatically strips ANSI color codes if colors are disabled
func Printf(format string, a ...interface{}) {
	writer := logWriter(os.Stdout)
	fmt.Fprintf(writer, format, a...)
}

// Print is like fmt.Print but automatically strips ANSI color codes if colors are disabled
func Print(a ...interface{}) {
	writer := logWriter(os.Stdout)
	fmt.Fprint(writer, a...)
}

func indentSubsequent(size int, text string) string {
	indentation := strings.Repeat(" ", size)
	return strings.ReplaceAll(text, "\n", "\n"+indentation)
}

func ExporterLogCustom(exporter Exporter, verb string, color string, message string, fmtArgs ...interface{}) {
	if debugging {
		LogCustom(verb, color, fmt.Sprintf("[dim][bold](from exporter %s)[reset] %s", exporter.Name(), message), fmtArgs...)
	} else {
		LogCustom(verb, color, message, fmtArgs...)
	}
}

func ExporterLogCustomNoFormatting(exporter Exporter, verb string, color string, message string) {
	if debugging {
		LogCustomNoFormatting(verb, color, colorstring.Color("[dim][bold](from exporter "+exporter.Name()+")[reset] ")+message)
	} else {
		LogCustomNoFormatting(verb, color, message)
	}
}

func LogCustom(verb string, color string, message string, fmtArgs ...interface{}) {
	LogCustomNoFormatting(verb, color, colorstring.Color(fmt.Sprintf(message, fmtArgs...)))
}

// LogCustomNoColor logs a message without applying colorstring syntax to message.
func LogCustomNoColor(verb string, color string, message string, fmtArgs ...interface{}) {
	LogCustomNoFormatting(verb, color, fmt.Sprintf(message, fmtArgs...))
}

func LogCustomNoFormatting(verb string, color string, message string) {
	fmt.Fprintln(
		logWriter(os.Stderr),
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
func LogError(message string, fmtArgs ...interface{}) {
	LogCustom("Error", "red", message, fmtArgs...)
}

func DisplayError(msg string, err error, fmtArgs ...interface{}) {
	LogError(formatErrors(fmt.Errorf(msg+": %w", append(fmtArgs, err)...)))
}

func DisplayWarning(msg string, err error, fmtArgs ...interface{}) {
	LogWarning(formatErrors(fmt.Errorf(msg+": %w", append(fmtArgs, err)...)))
}

// LogInfo logs infos.
func LogInfo(message string, fmtArgs ...interface{}) {
	LogCustom("Info", "blue", message, fmtArgs...)
}

// LogDebug logs debug information.
func LogDebug(message string, fmtArgs ...interface{}) {
	if os.Getenv("DEBUG") == "" {
		return
	}
	LogCustom("Debug", "magenta", message, fmtArgs...)
}

// LogDebugNoColor logs debug information without applying colorstring syntax to message.
func LogDebugNoColor(message string, fmtArgs ...interface{}) {
	if os.Getenv("DEBUG") == "" {
		return
	}
	LogCustomNoColor("Debug", "magenta", message, fmtArgs...)
}

// LogWarning logs warnings.
func LogWarning(message string, fmtArgs ...interface{}) {
	LogCustom("Warning", "yellow", message, fmtArgs...)
}

// LogTiming logs timing debug logs. Mostly used with TimeTrack
func LogTiming(job string, args []interface{}, timeTaken time.Duration) {
	if !showingTimingLogs {
		return
	}
	formattedArgs := ""
	for i, arg := range args {
		if i > 0 {
			formattedArgs += " "
		}
		formattedArgs += fmt.Sprintf("%v", arg)
	}
	LogCustom("Timing", "dim", "[bold]%-30s[reset][dim]([reset]%-50s[dim])[reset] took [yellow]%s", job, formattedArgs, timeTaken)
}

// TimeTrack logs the time taken for a function to execute, and logs out the time taken.
// Usage: at the top of your function, defer TimeTrack(time.Now(), "your job name")
func TimeTrack(start time.Time, job string, args ...interface{}) {
	if !showingTimingLogs {
		return
	}
	elapsed := time.Since(start)
	LogTiming(job, args, elapsed)
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
