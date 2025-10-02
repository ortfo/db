package ortfodb

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	ll "github.com/gwennlbh/label-logger-go"
)

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

func (phase BuildPhase) String() string {
	return string(phase)
}

func padPhaseVerb(phase BuildPhase) string {
	// length of longest phase verb: "Thumbnailing", plus some padding
	return fmt.Sprintf("%15s", phase)
}

func (ctx *RunContext) StartProgressBar(total int) {
	worksToBuildCount = total
	ll.StartProgressBar(total, "Building", "magenta")
}

func (ctx *RunContext) IncrementProgress() {
	builtWorksCount++

	if BuildIsFinished() {
		ll.Log("Finished", "green", "compiling to %s\n", ctx.OutputDatabaseFile)
		os.Remove(ctx.ProgressInfoFile)
	}

	ll.IncrementProgressBar()
	if BuildIsFinished() {
		ll.StopProgressBar()
	}
}

func BuildIsFinished() bool {
	return ll.ProgressBarFinished() || builtWorksCount >= worksToBuildCount
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
		formattedDetails = strings.Join(details, " ")
	}

	ll.Log(phase.String(), color, "%s%s", workID, formattedDetails)

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
		ll.WarnDisplay("could not append progress info to file", err)
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
		ll.Debug("not writing progress info to file because --write-progress is not set")
		return nil
	}
	event := ProgressInfoEvent{
		WorksDone:  builtWorksCount,
		WorksTotal: worksToBuildCount,
		WorkID:     workID,
		Phase:      phase,
		Details:    details,
	}
	ll.Debug("Appending event %#v to progress file %s", event, ctx.ProgressInfoFile)
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
