// Package ortfodb exposes the various functions used by the ortfodb portfolio database creation command-line tool.
// It is notably used by ortfomk to share some common data between the two complementing programs.
// See https://ewen.works/ortfodb for more information.
package ortfodb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"path"

	jsoniter "github.com/json-iterator/go"
)

// RunContext holds several "global" references used throughout all the functions of a command.
type RunContext struct {
	Config *Configuration
	// ID of the work currently being processed.
	CurrentWorkID         string
	DatabaseDirectory     string
	PreviousBuiltDatabase []Work
	Flags                 Flags
	Progress              struct {
		Current int
		Total   int
		// See ProgressFile.Current.Step in progress.go
		Step BuildStep
		// See ProgressFile.Current.Resolution in progress.go
		Resolution int
		// See ProgressFile.Current.File in progress.go
		File string
		Hash string
	}
	BuildMetadata BuildMetadata
	Spinner       Spinner
}

type Flags struct {
	Scattered    bool
	Silent       bool
	Minified     bool
	Config       string
	ProgressFile string
	NoCache      bool
}

// Project represents a project.
type Project struct {
	ID             string
	DescriptionRaw string
	Description    ParsedDescription
	Ctx            *RunContext
}

func buildLockFilepath(outputFilename string) string {
	return filepath.Join(filepath.Dir(outputFilename), ".ortfomk-build-lock")
}

// AcquireBuildLock ensures that only one process touches the output database file at the same time.
// An error is returned if the lock could not be acquired
func AcquireBuildLock(outputFilename string) error {
	if _, err := os.Stat(buildLockFilepath(outputFilename)); os.IsNotExist(err) {
		os.WriteFile(buildLockFilepath(outputFilename), []byte(""), 0o644)
		return nil
	} else if err == nil {
		return fmt.Errorf("file %s exists", buildLockFilepath(outputFilename))
	} else {
		return fmt.Errorf("while checking if file %s exists: %w", buildLockFilepath(outputFilename), err)
	}
}

func (ctx *RunContext) ReleaseBuildLock(outputFilename string) {
	err := os.Remove(buildLockFilepath(outputFilename))
	if err != nil {
		ctx.LogError("could not release build lockfile %s: %s", buildLockFilepath(outputFilename), err)
	}
}

func PrepareBuild(databaseDirectory string, outputFilename string, flags Flags, config Configuration) (RunContext, error) {
	ctx := RunContext{
		Config:            &config,
		Flags:             flags,
		DatabaseDirectory: databaseDirectory,
	}

	previousBuiltDatabaseRaw, err := os.ReadFile(outputFilename)
	if err != nil {
		if !os.IsNotExist(err) {
			ctx.LogError("Couldn't use previous built database file %s: %s", outputFilename, err.Error())
		}
		ctx.PreviousBuiltDatabase = []Work{}
	} else {
		// TODO unmarshal with respect to snake_case -> CamelCase conversion, we are using non-annotated struct fields' data currently.
		err = json.Unmarshal(previousBuiltDatabaseRaw, &ctx.PreviousBuiltDatabase)
		if err != nil {
			ctx.LogError("Couldn't use previous built database file %s: %s", outputFilename, err.Error())
			ctx.PreviousBuiltDatabase = []Work{}
		}
	}

	ctx.Spinner = ctx.CreateSpinner(outputFilename)
	if !flags.Silent {
		err := ctx.Spinner.Start()
		if err != nil {
			panic(err)
		}
	}

	raw, err := os.ReadFile(config.BuildMetadataFilepath)
	if err == nil {
		var metadata BuildMetadata
		err = json.Unmarshal(raw, &metadata)
		if err == nil {
			ctx.BuildMetadata = metadata
		}
	}

	if ctx.Config.IsDefault {
		ctx.LogInfo("No configuration file found. The default configuration was used.")
	}

	err = os.MkdirAll(config.Media.At, 0o755)
	if err != nil {
		return ctx, fmt.Errorf("while creating the media output directory: %w", err)
	}
	if err := AcquireBuildLock(outputFilename); err != nil {
		return ctx, fmt.Errorf("another ortfo build is in progress (could not acquire build lock): %w", err)
	}

	return ctx, nil
}

// BuildAll builds the database at outputFilename from databaseDirectory.
// Use LoadConfiguration (and ValidateConfiguration if desired) to get a Configuration.
func BuildAll(databaseDirectory string, outputFilename string, flags Flags, config Configuration) error {
	return BuildSome("*", databaseDirectory, outputFilename, flags, config)
}

func BuildSome(include string, databaseDirectory string, outputFilename string, flags Flags, config Configuration) error {
	ctx, err := PrepareBuild(databaseDirectory, outputFilename, flags, config)
	if err != nil {
		return err
	}

	defer ctx.ReleaseBuildLock(outputFilename)
	ctx.Progress.Total = 1
	works := make([]Work, 0)
	workDirectories, err := ctx.ComputeProgressTotal()
	if err != nil {
		return fmt.Errorf("while computing total number of works to build: %w", err)
	}

	for _, dirEntry := range workDirectories {
		workID := dirEntry.Name()
		presentBefore, oldWork := FindWork(ctx.PreviousBuiltDatabase, workID)
		var included bool
		if include == "*" {
			included = true
		} else {
			included, err = filepath.Match(include, workID)
			if err != nil {

				return fmt.Errorf("while testing include-works pattern %q: %w", include, err)
			}
		}
		if included {
			newWork, err := ctx.Build(databaseDirectory, outputFilename, workID)
			if err != nil {
				ctx.LogError("while building %s: %s", workID, err)
			}
			works = append(works, newWork)
		} else if presentBefore {
			works = append(works, oldWork)
		} else {
			ctx.LogInfo("Skipped building of work %s, as it is neither included in %s nor formerly present in %s.", workID, include, outputFilename)
		}
		ctx.IncrementProgress()
	}

	ctx.WriteDatabase(works, flags, outputFilename)
	return nil
}

func (ctx *RunContext) WriteDatabase(works []Work, flags Flags, outputFilename string) {
	// Compile the database
	var worksJSON []byte
	json := jsoniter.ConfigFastest
	setJSONNamingStrategy(lowerCaseWithUnderscores)
	if ctx.Flags.Minified {
		worksJSON, _ = json.Marshal(works)
	} else {
		worksJSON, _ = json.MarshalIndent(works, "", "    ")
	}

	// Output it
	if outputFilename == "-" {
		fmt.Println(string(worksJSON))
	} else {
		err := writeFile(outputFilename, worksJSON)
		if err != nil {
			println(err.Error())
		}
	}

	ctx.Spinner.Stop()
}

func (ctx *RunContext) ComputeProgressTotal() (workDirectories []fs.DirEntry, err error) {
	databaseFiles, err := os.ReadDir(ctx.DatabaseDirectory)
	if err != nil {
		return
	}
	// Build up workDirectories by filtering through databaseFiles.
	// We do this beforehand to compute ctx.Progress.Total.
	for _, dirEntry := range databaseFiles {
		// TODO: setting to ignore/allow “dotfolders”

		dirEntryAbsPath := path.Join(ctx.DatabaseDirectory, dirEntry.Name())
		if !dirEntry.IsDir() {
			continue
		}
		// Compute the description file's path
		var descriptionFilename string
		if ctx.Flags.Scattered {
			descriptionFilename = path.Join(dirEntryAbsPath, ctx.Config.ScatteredModeFolder, "description.md")
		} else {
			descriptionFilename = path.Join(dirEntryAbsPath, "description.md")
		}
		// If it's not there, this directory is not a project worth scanning.
		if _, err := os.Stat(descriptionFilename); os.IsNotExist(err) {
			continue
		}

		workDirectories = append(workDirectories, dirEntry)
	}

	ctx.Progress.Total = len(workDirectories)
	return
}

// Build builds a single work given the database & output folders, as wells as a work ID
func (ctx *RunContext) Build(databaseDirectory string, outputFilename string, workID string) (Work, error) {
	// Compute the description file's path
	var descriptionFilename string
	if ctx.Flags.Scattered {
		descriptionFilename = path.Join(databaseDirectory, workID, ctx.Config.ScatteredModeFolder, "description.md")
	} else {
		descriptionFilename = path.Join(databaseDirectory, workID, "description.md")
	}

	// Update the UI
	ctx.CurrentWorkID = workID

	// Parse the description
	descriptionRaw, err := ioutil.ReadFile(descriptionFilename)
	if err != nil {
		return Work{}, err
	}

	ctx.Status(StepDescription, ProgressDetails{
		File: descriptionFilename,
	})
	description := ctx.ParseDescription(string(descriptionRaw))

	// Handle mediae
	analyzedMediae := make(map[string][]Media)
	for lang, mediae := range description.MediaEmbedDeclarations {
		analyzedMediae[lang] = []Media{}
		for _, media := range mediae {
			analyzed, err := ctx.HandleMedia(workID, media, lang)
			if err != nil {
				ctx.LogError(err.Error())
				continue
			}
			analyzedMediae[lang] = append(analyzedMediae[lang], analyzed)
		}
	}

	// Extract colors
	metadata := description.Metadata
	if _, ok := metadata["colors"]; ctx.Config.ExtractColors.Enabled && !ok {
		if sourceOfChosenThumbnail, ok := metadata["thumbnail"]; ok {
		outer:
			for _, ms := range analyzedMediae {
				for _, m := range ms {
					if m.Source == sourceOfChosenThumbnail {
						metadata["colors"] = m.ExtractedColors
						break outer
					}
				}
			}
		} else {
			for _, ms := range analyzedMediae {
				if len(ms) > 0 {
					metadata["colors"] = ms[0].ExtractedColors
					break
				}
			}
		}
	}

	ctx.UpdateBuildMetadata()
	ctx.WriteBuildMetadata()

	// Return the finished work
	return Work{
		ID:         workID,
		Metadata:   metadata,
		Title:      description.Title,
		Paragraphs: description.Paragraphs,
		Media:      analyzedMediae,
		Links:      description.Links,
		Footnotes:  description.Footnotes,
	}, nil
}

// GetProjectPath returns the project's folder path with regard to databaseDirectory.
func (p *Project) ProjectPath() string {
	if p.Ctx.Flags.Scattered {
		return path.Join(p.Ctx.DatabaseDirectory, p.ID, p.Ctx.Config.ScatteredModeFolder)
	}
	return path.Join(p.Ctx.DatabaseDirectory, p.ID)
}

// ReadDescriptionFile reads the description.md file in directory.
// Returns an empty string if the file is a directory or does not exist.
func ReadDescriptionFile(directory string) (string, error) {
	descriptionFilepath := path.Join(directory, "description.md")
	if !fileExists(descriptionFilepath) {
		return "", nil
	}
	descriptionFile, err := os.Stat(descriptionFilepath)
	if err != nil {
		return "", err
	}
	if descriptionFile.IsDir() {
		return "", nil
	}
	return readFile(descriptionFilepath)
}

// WriteBuildMetadata writes the latest build metadata file.
func (ctx *RunContext) WriteBuildMetadata() error {
	_, err := os.Stat(ctx.Config.BuildMetadataFilepath)

	if errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(filepath.Dir(ctx.Config.BuildMetadataFilepath), os.ModePerm)
	} else if err != nil {
		return fmt.Errorf("while creating parent directories for build metadata file: %w", err)
	}

	raw, err := json.Marshal(ctx.BuildMetadata)
	if err != nil {
		return fmt.Errorf("while marshaling build metadata to JSON: %w", err)
	}

	return os.WriteFile(ctx.Config.BuildMetadataFilepath, []byte(raw), 0644)
}

// UpdateBuildMetadata updates metadata about the latest build.
func (ctx *RunContext) UpdateBuildMetadata() {
	ctx.BuildMetadata.PreviousBuildDate = time.Now()
}
