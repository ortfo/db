package ortfodb

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/mitchellh/colorstring"
)

var progressbar *uiprogress.Bar
var progressBars *uiprogress.Progress
var currentlyBuildingWorkIDs []string

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

func StartProgressBar(total int) {
	if progressbar != nil {
		panic("progress bar already started")
	}

	progressBars = uiprogress.New()
	progressBars.SetRefreshInterval(1 * time.Millisecond)
	progressbar = progressBars.AddBar(total)
	progressbar.Empty = ' '
	progressbar.Fill = '='
	progressbar.Head = '>'
	progressbar.Width = 30
	progressbar.PrependFunc(func(b *uiprogress.Bar) string {
		return colorstring.Color(
			fmt.Sprintf(
				`[magenta][bold]%15s[reset]`,
				"Building",
			),
		)
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
	progressbar.Incr()
	if progressbar.CompletedPercent() >= 100 {
		progressBars.Bars = nil
		progressBars.Stop()
		// Clear progress bar empty line
		fmt.Print("\r\033[K")
		colorstring.Printf("[bold][green]%15s[reset] compiling to %s in %s\n", "Finished", ctx.OutputDatabaseFile, progressbar.TimeElapsedString())
	}
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
	fmt.Fprintln(progressBars.Bypass(), colorstring.Color(fmt.Sprintf("[bold][%s]%s[reset] %s"+formattedDetails, color, padPhaseVerb(phase), workID)))

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
}
