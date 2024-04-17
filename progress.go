package ortfodb

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/mitchellh/colorstring"
)

var progressbar *uiprogress.Bar
var progressBars *uiprogress.Progress
var currentlyBuildingWorkIDs []string
var builtWorksCount int
var worksToBuildCount int

type BuildPhase string

const (
	PhaseThumbnails    BuildPhase = "Thumbnailing"
	PhaseMediaAnalysis BuildPhase = "Analyzing"
	PhaseBuilding      BuildPhase = "Building"
	PhaseBuilt         BuildPhase = "Built"
	PhaseUnchanged     BuildPhase = "Reusing"
)

func padPhaseVerb(phase BuildPhase) string {
	// length of longest phase verb: "Thumbnailing", plus some padding
	return fmt.Sprintf("%15s", phase)
}

func (ctx *RunContext) StartProgressBar(total int) {
	if progressbar != nil {
		panic("progress bar already started")
	}

	worksToBuildCount = total

	if isInteractiveTerminal() || os.Getenv("FORCE_PROGRESS_BAR") == "1" {
		LogDebug("terminal is interactive, starting progress bar")
	} else {
		LogDebug("not starting progress bar because not in an interactive terminal")
		return
	}

	progressBars = uiprogress.New()
	progressBars.SetRefreshInterval(1 * time.Millisecond)
	progressbar = progressBars.AddBar(total)
	progressbar.Empty = ' '
	progressbar.Fill = '='
	progressbar.Head = '>'
	progressbar.Width = 30
	progressbar.PrependFunc(func(b *uiprogress.Bar) string {
		if ShowingColors() {
			return colorstring.Color(
				fmt.Sprintf(
					`[magenta][bold]%15s[reset]`,
					"Building",
				),
			)
		}
		return fmt.Sprintf("%15s", "Building")
	})
	progressbar.AppendFunc(func(b *uiprogress.Bar) string {
		// truncatedCurrentlyBuildingWorkIDs := make([]string, 0, len(currentlyBuildingWorkIDs))
		// for _, id := range currentlyBuildingWorkIDs {
		// 	if len(id) > 5 {
		// 		truncatedCurrentlyBuildingWorkIDs = append(truncatedCurrentlyBuildingWorkIDs, id[:5])
		// 	} else {
		// 		truncatedCurrentlyBuildingWorkIDs = append(truncatedCurrentlyBuildingWorkIDs, id)
		// 	}
		// }

		return fmt.Sprintf("%d/%d", b.Current(), b.Total)
	})
	progressBars.Start()
}

func (ctx *RunContext) IncrementProgress() {
	builtWorksCount++
	if progressbar == nil {
		if builtWorksCount >= worksToBuildCount {
			ctx.showFinishedMessage()
		}
		return
	}

	progressbar.Incr()
	if progressbar.CompletedPercent() >= 100 {
		ctx.showFinishedMessage()
		ctx.StopProgressBar()
	}
}

func (ctx *RunContext) showFinishedMessage() {
	if progressbar == nil {
		LogCustom("Finished", "green", "compiling to %s\n", ctx.OutputDatabaseFile)
	} else {
		LogCustom("Finished", "green", "compiling to %s in %s\n", ctx.OutputDatabaseFile, progressbar.TimeElapsedString())
	}
}

func (ctx *RunContext) StopProgressBar() {
	if progressbar == nil {
		return
	}

	progressBars.Bars = nil
	progressBars.Stop()
	// Clear progress bar empty line
	fmt.Print("\r\033[K")
}

// Status updates the current progress and writes the progress to a file if --write-progress is set.
func (ctx *RunContext) Status(workID string, phase BuildPhase, details ...string) {
	var color string
	switch phase {
	case PhaseBuilt:
		color = "light_green"
	case PhaseUnchanged:
		color = "dim"
	default:
		color = "cyan"
	}
	formattedDetails := ""
	if len(details) > 0 {
		formattedDetails = fmt.Sprintf(" [dim]%s[reset]", strings.Join(details, " "))
	}
	formattedMessage := colorstring.Color(fmt.Sprintf("[bold][%s]%s[reset] %s"+formattedDetails, color, padPhaseVerb(phase), workID))

	if progressBars != nil {
		writer := progressBars.Bypass()
		if !ShowingColors() {
			writer = noAnsiCodesWriter{out: writer}
		}
		fmt.Fprintln(writer, formattedMessage)
	} else {
		Println(formattedMessage)
	}

	if phase == PhaseBuilt || phase == PhaseUnchanged {
		for i, id := range currentlyBuildingWorkIDs {
			if id == workID {
				currentlyBuildingWorkIDs = append(currentlyBuildingWorkIDs[:i], currentlyBuildingWorkIDs[i+1:]...)
				break
			}
		}
		ctx.IncrementProgress()
	} else if phase == PhaseBuilding {
		currentlyBuildingWorkIDs = append(currentlyBuildingWorkIDs, workID)
	}

	if err := ctx.appendToProgressFile(workID, phase, details...); err != nil {
		DisplayWarning("could not append progress info to file", err)
	}
}

// ProgressInfoEvent represents an event that is appended to the progress JSONLines file.
type ProgressInfoEvent struct {
	// WorksDone is the number of works that have been built
	WorksDone int `json:"works_done"`
	// WorksTotal is the total number of works that will be built
	WorksTotal int        `json:"works_total"`
	WorkID     string     `json:"work_id"`
	Phase      BuildPhase `json:"phase"`
	Details    []string   `json:"details"`
}

func (ctx *RunContext) appendToProgressFile(workID string, phase BuildPhase, details ...string) error {
	if ctx.ProgressInfoFile == "" {
		return nil
	}
	event := ProgressInfoEvent{
		WorksDone:  builtWorksCount,
		WorksTotal: worksToBuildCount,
		WorkID:     workID,
		Phase:      phase,
		Details:    details,
	}
	// append JSON marshalled event to file
	file, err := os.OpenFile(ctx.ProgressInfoFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("while opening progress info file at %s: %w", ctx.ProgressInfoFile, err)
	}

	marshaled, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("while converting event to JSON (event is %#v): %w", event, err)
	}

	_, err = file.WriteString(fmt.Sprintf("%s\n", marshaled))
	if err != nil {
		return fmt.Errorf("while appending progress info event to %s: %w", ctx.ProgressInfoFile, err)
	}

	return nil
}
