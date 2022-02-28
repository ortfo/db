package ortfodb

import (
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
	ctx.Progress.Step = step
	ctx.Progress.Resolution = details.Resolution
	ctx.Progress.File = details.File

	ctx.UpdateSpinner()
	err := ctx.WriteProgressFile()
	if err != nil {
		ctx.LogError("Couldn't write to progress file:", err)
	}
}

// IncrementProgress increments the number of processed works and writes the progress to a file if --write-progress is set.
func (ctx *RunContext) IncrementProgress() error {
	ctx.Progress.Current++

	ctx.UpdateSpinner()
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
