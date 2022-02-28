package ortfodb

import (
	"fmt"
	"io/ioutil"

	jsoniter "github.com/json-iterator/go"
)

type BuildStep string

const (
	StepThumbnails      = "thumbnails"
	StepMediaAnalysis   = "media analysis"
	StepDescription     = "description"
	StepColorExtraction = "color extraction"
)

// ProgressFile holds the data that gets written to the progress file as JSON.
type ProgressFile struct {
	Total     int
	Processed int
	Percent   int
	Current   struct {
		ID   string
		Step BuildStep
		// The resolution of the thumbnail being generated. 0 when step is not "thumbnails"
		Resolution int
		// The file being processed:
		//
		// - original media when making thumbnails or during media analysis,
		//
		// - media the colors are being extracted from, or
		//
		// - the description.md file when parsing description
		File string
	}
}

type ProgressDetails struct {
	Resolution int
	File       string
}

// Status updates the current progress and writes the progress to a file if --write-progress is set.
func (ctx *RunContext) Status(step BuildStep, details ProgressDetails) {
	// fmt.Print("\033[2K\r")
	ctx.Progress.Step = step
	ctx.Progress.Resolution = details.Resolution
	ctx.Progress.File = details.File

	var message string
	switch step {
	case StepColorExtraction:
		message = fmt.Sprintf("Extracting colors from %s", details.File)
	case StepDescription:
		message = fmt.Sprintf("Parsing description %s", details.File)
	case StepMediaAnalysis:
		message = fmt.Sprintf("Analyzing media %s", details.File)
	case StepThumbnails:
		message = fmt.Sprintf("Generating thumbnails for %s", details.File)
	}

	fmt.Printf("\033[2K\r[%03d] %s: %s", ctx.ProgressFileData().Percent, ctx.CurrentWorkID, message)

	err := ctx.WriteProgressFile()
	if err != nil {
		fmt.Println("couldn't write progress file:", err)
	}
}

// IncrementProgress increments the number of processed works and writes the progress to a file if --write-progress is set.
func (ctx *RunContext) IncrementProgress() error {
	ctx.Progress.Current++

	return ctx.WriteProgressFile()
}

// WriteProgressFile writes the progress to a file if --write-progress is set.
func (ctx *RunContext) WriteProgressFile() error {
	if ctx.Flags.ProgressFile == "" {
		return nil
	}

	setJSONNamingStrategy(lowerCaseWithUnderscores)
	progressDataJSON, err := jsoniter.Marshal(ctx.ProgressFileData())
	if err != nil {
		return err
	}

	return ioutil.WriteFile(ctx.Flags.ProgressFile, progressDataJSON, 0644)
}
