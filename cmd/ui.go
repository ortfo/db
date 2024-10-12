package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/MakeNowJust/heredoc"
	"github.com/mitchellh/colorstring"
	ll "github.com/ewen-lbh/label-logger-go"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

// Most functions in this file are copied from cobra/pflags' source code. All because setting a custom template for the usage line of flags is not possible :/

var usageTemplate = heredoc.Doc(colorstring.Color(`
[bold]Usage:[reset]{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

[bold]Aliases:[reset]
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

[bold]Examples:[reset]
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

[bold]Available Commands:[reset]{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  [blue][bold]{{rpad .Name .NamePadding }}[reset] {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

[bold]{{.Title}}[reset]{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

[bold]Additional Commands:[reset]{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  [blue][bold]{{rpad .Name .NamePadding}}[reset] {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

[bold]Flags:[reset]
{{.LocalFlags | customFlagsUsage | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

[bold]Global Flags:[reset]
{{.InheritedFlags | customFlagsUsage | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

[bold]Additional help topics:[reset]{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

[dim]Use "[bold]{{.CommandPath}} [command] --help[reset][dim]" for more information about a command.[reset]{{end}}
`))

func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

func appendIfNotPresent(s, stringToAppend string) string {
	if strings.Contains(s, stringToAppend) {
		return s
	}
	return s + " " + stringToAppend
}

func rpad(s string, padding int) string {
	formattedString := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(formattedString, s)
}

func lpad(s string, padding int) string {
	formattedString := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(formattedString, s)
}

func customUsage(c *cobra.Command) error {
	// c.mergePersistentFlags()
	t := template.New("top")
	templateFuncs := template.FuncMap{
		"trim":                    strings.TrimSpace,
		"trimRightSpace":          trimRightSpace,
		"trimTrailingWhitespaces": trimRightSpace,
		"appendIfNotPresent":      appendIfNotPresent,
		"rpad":                    rpad,
		"lpad":                    lpad,
		"gt":                      cobra.Gt,
		"eq":                      cobra.Eq,
		"customFlagsUsage":        customFlagsUsage,
	}

	t.Funcs(templateFuncs)
	template.Must(t.Parse(usageTemplate))
	err := t.Execute(c.OutOrStderr(), c)
	if err != nil {
		c.PrintErrln(err)
	}
	return err
}

// customFlagsUsage is functionnaly equivalent to pflag.FlagUsages but
// with a different output format.
// Whilst cobra.SetUsageTemplate exists, it is not possible to
// customise the flag usage output without reimplementing the whole
// thing.
func customFlagsUsage(f *pflag.FlagSet) string {
	buf := new(bytes.Buffer)

	cols := 0
	lines := make([]string, 0)

	maxlen := 0
	f.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}

		line := ""
		if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
			line = fmt.Sprintf("  [cyan]-%s, [bold]--%s[reset]", flag.Shorthand, flag.Name)
		} else {
			line = fmt.Sprintf("      [cyan][bold]--%s[reset]", flag.Name)
		}

		varname, usage := pflag.UnquoteUsage(flag)
		if varname != "" {
			line += fmt.Sprintf(" [yellow]%s[reset]", varname)
		}

		if flag.NoOptDefVal != "" {
			line += "[yellow]"
			switch flag.Value.Type() {
			case "string":
				line += fmt.Sprintf("[=\"%s\"]", flag.NoOptDefVal)
			case "bool":
				if flag.NoOptDefVal != "true" {
					line += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
				}
			case "count":
				if flag.NoOptDefVal != "+1" {
					line += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
				}
			default:
				line += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
			}
			line += "[reset]"
		}

		// This special character will be replaced with spacing once the
		// correct alignment is calculated
		line = colorstring.Color(line)
		line += "\x00"
		if len(line) > maxlen {
			maxlen = len(line)
		}

		line += usage
		if !defaultIsZeroValue(flag) {
			line += "[green]"
			if flag.Value.Type() == "string" {
				line += fmt.Sprintf(" (default %q)", flag.DefValue)
			} else {
				line += fmt.Sprintf(" (default %s)", flag.DefValue)
			}
			line += "[reset]"
		}
		if len(flag.Deprecated) != 0 {
			line += fmt.Sprintf(" [red][bold](DEPRECATED: %s)[reset]", flag.Deprecated)
		}

		lines = append(lines, colorstring.Color(line))
	})

	for _, line := range lines {
		sidx := strings.Index(line, "\x00")
		spacing := strings.Repeat(" ", maxlen-sidx)
		// maxlen + 2 comes from + 1 for the \x00 and + 1 for the (deliberate) off-by-one in maxlen-sidx
		fmt.Fprintln(buf, line[:sidx], spacing, wrap(maxlen+2, cols, line[sidx+1:]))
	}

	s := buf.String()
	if !ll.ShowingColors() {
		s = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`).ReplaceAllString(s, "")
	}
	return s
}

// defaultIsZeroValue returns true if the default value for this flag represents
// a zero value.
func defaultIsZeroValue(f *pflag.Flag) bool {
	switch f.Value.Type() {
	case "bool":
		return f.DefValue == "false"
	case "duration":
		// Beginning in Go 1.7, duration zero values are "0s"
		return f.DefValue == "0" || f.DefValue == "0s"
	case "int", "int8", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "count", "float32", "float64":
		return f.DefValue == "0"
	case "string":
		return f.DefValue == ""
	case "ip", "ipMask", "ipNet":
		return f.DefValue == "<nil>"
	case "intSlice", "stringSlice", "stringArray":
		return f.DefValue == "[]"
	default:
		switch f.Value.String() {
		case "false":
			return true
		case "<nil>":
			return true
		case "":
			return true
		case "0":
			return true
		}
		return false
	}
}

// Wraps the string `s` to a maximum width `w` with leading indent
// `i`. The first line is not indented (this is assumed to be done by
// caller). Pass `w` == 0 to do no wrapping
func wrap(i, w int, s string) string {
	if w == 0 {
		return strings.Replace(s, "\n", "\n"+strings.Repeat(" ", i), -1)
	}

	// space between indent i and end of line width w into which
	// we should wrap the text.
	wrap := w - i

	var r, l string

	// Not enough space for sensible wrapping. Wrap as a block on
	// the next line instead.
	if wrap < 24 {
		i = 16
		wrap = w - i
		r += "\n" + strings.Repeat(" ", i)
	}
	// If still not enough space then don't even try to wrap.
	if wrap < 24 {
		return strings.Replace(s, "\n", r, -1)
	}

	// Try to avoid short orphan words on the final line, by
	// allowing wrapN to go a bit over if that would fit in the
	// remainder of the line.
	slop := 5
	wrap = wrap - slop

	// Handle first line, which is indented by the caller (or the
	// special case above)
	l, s = wrapN(wrap, slop, s)
	r = r + strings.Replace(l, "\n", "\n"+strings.Repeat(" ", i), -1)

	// Now wrap the rest
	for s != "" {
		var t string

		t, s = wrapN(wrap, slop, s)
		r = r + "\n" + strings.Repeat(" ", i) + strings.Replace(t, "\n", "\n"+strings.Repeat(" ", i), -1)
	}

	return r

}

// Splits the string `s` on whitespace into an initial substring up to
// `i` runes in length and the remainder. Will go `slop` over `i` if
// that encompasses the entire string (which allows the caller to
// avoid short orphan words on the final line).
func wrapN(i, slop int, s string) (string, string) {
	if i+slop > len(s) {
		return s, ""
	}

	w := strings.LastIndexAny(s[:i], " \t\n")
	if w <= 0 {
		return s, ""
	}
	nlPos := strings.LastIndex(s[:i], "\n")
	if nlPos > 0 && nlPos < w {
		return s[:nlPos], s[nlPos+1:]
	}
	return s[:w], s[w+1:]
}

func terminalWidth(min, max int) (width int) {
	width, _, _ = term.GetSize(0)
	if width < min {
		return min
	}
	if width > max {
		return max
	}
	return
}
