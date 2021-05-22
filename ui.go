package ortfodb

import (
	"fmt"
	"strings"

	"github.com/mitchellh/colorstring"
	"github.com/xeipuuv/gojsonschema"
)

// DisplayValidationErrors takes in a slice of json schema validation errors and displays them nicely to in the terminal
func DisplayValidationErrors(errors []gojsonschema.ResultError, name string) {
	println("Your " + name + " file is invalid. Here are the validation errors:\n")
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

// Status prints the current compilation progress
func (ctx *RunContext) Status(text string) {
	fmt.Print("\033[2K\r")
	fmt.Printf("[%v/%v] %v: %v", ctx.Progress.Current, ctx.Progress.Total, ctx.CurrentProject, text)
}
