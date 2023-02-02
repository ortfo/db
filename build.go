// Package ortfodb exposes the various functions used by the ortfodb portfolio database creation command-line tool.
// It is notably used by ortfomk to share some common data between the two complementing programs.
// See https://ewen.works/ortfodb for more information.
package ortfodb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"path"

	jsoniter "github.com/json-iterator/go"
)

type Works map[string]AnalyzedWork

// RunContext holds several "global" references used throughout all the functions of a command.
type RunContext struct {
	mu sync.Mutex

	Config *Configuration
	// ID of the work currently being processed.
	CurrentWorkID         string
	DatabaseDirectory     string
	OutputDatabaseFile    string
	PreviousBuiltDatabase Works
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
	Description    ParsedWork
	Ctx            *RunContext
}

// BuildLockFilepath returns the path to the lock file for the given output database file.
func BuildLockFilepath(outputFilename string) string {
	return filepath.Join(filepath.Dir(outputFilename), ".ortfodb-build-lock")
}

// AcquireBuildLock ensures that only one process touches the output database file at the same time.
// An error is returned if the lock could not be acquired
func AcquireBuildLock(outputFilename string) error {
	if _, err := os.Stat(BuildLockFilepath(outputFilename)); os.IsNotExist(err) {
		os.WriteFile(BuildLockFilepath(outputFilename), []byte(""), 0o644)
		return nil
	} else if err == nil {
		return fmt.Errorf("file %s exists", BuildLockFilepath(outputFilename))
	} else {
		return fmt.Errorf("while checking if file %s exists: %w", BuildLockFilepath(outputFilename), err)
	}
}

func (ctx *RunContext) ReleaseBuildLock(outputFilename string) error {
	err := os.Remove(BuildLockFilepath(outputFilename))
	if err != nil {
		ctx.LogError("could not release build lockfile %s: %s", BuildLockFilepath(outputFilename), err)
	}
	return err
}

func PrepareBuild(databaseDirectory string, outputFilename string, flags Flags, config Configuration) (RunContext, error) {
	ctx := RunContext{
		Config:             &config,
		Flags:              flags,
		DatabaseDirectory:  databaseDirectory,
		OutputDatabaseFile: outputFilename,
	}
	ctx.Spinner = ctx.CreateSpinner(outputFilename)

	previousBuiltDatabaseRaw, err := os.ReadFile(outputFilename)
	if err != nil {
		if !os.IsNotExist(err) {
			ctx.LogError("No previously built database file %s to use: %s", outputFilename, err.Error())
		}
		ctx.PreviousBuiltDatabase = Works{}
	} else {
		// TODO unmarshal with respect to snake_case -> CamelCase conversion, we are using non-annotated struct fields' data currently.
		setJSONNamingStrategy(lowerCaseWithUnderscores)
		err = json.Unmarshal(previousBuiltDatabaseRaw, &ctx.PreviousBuiltDatabase)
		if err != nil {
			ctx.LogError("Couldn't use previous built database file %s: %s", outputFilename, err.Error())
			ctx.PreviousBuiltDatabase = Works{}
		}
	}

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
func (ctx *RunContext) BuildAll(databaseDirectory string, outputFilename string, flags Flags, config Configuration) (Works, error) {
	return ctx.BuildSome("*", databaseDirectory, outputFilename, flags, config)
}

func (ctx *RunContext) BuildSome(include string, databaseDirectory string, outputFilename string, flags Flags, config Configuration) (Works, error) {
	defer ctx.ReleaseBuildLock(outputFilename)
	ctx.Progress.Total = 1
	works := make(map[string]AnalyzedWork)
	workDirectories, err := ctx.ComputeProgressTotal()
	if err != nil {
		return Works{}, fmt.Errorf("while computing total number of works to build: %w", err)
	}

	for _, dirEntry := range workDirectories {
		workID := dirEntry.Name()
		oldWork, presentBefore := ctx.PreviousBuiltDatabase[workID]
		var included bool
		if include == "*" {
			included = true
		} else {
			included, err = filepath.Match(include, workID)
			if err != nil {
				return Works{}, fmt.Errorf("while testing include-works pattern %q: %w", include, err)
			}
		}
		if included {
			newWork, err := ctx.Build(databaseDirectory, outputFilename, workID)
			newWork.Metadata.BuiltAt = time.Now().String()
			if err != nil {
				return works, fmt.Errorf("while building %s (%s): %w", workID, ctx.DescriptionFilename(databaseDirectory, workID), err)
			}
			works[workID] = newWork
		} else if presentBefore {
			works[workID] = oldWork
		} else {
			ctx.LogInfo("Skipped building of work %s, as it is neither included in %s nor formerly present in %s.", workID, include, outputFilename)
		}
		ctx.IncrementProgress()
	}

	return works, nil
}

func (ctx *RunContext) WriteDatabase(works Works, flags Flags, outputFilename string, partial bool) {
	worksWithMetadata := make(map[string]interface{})
	for k, v := range works {
		worksWithMetadata[k] = v
	}
	worksWithMetadata["#meta"] = struct{ Partial bool }{Partial: partial}
	// Compile the database
	var worksJSON []byte
	json := jsoniter.ConfigFastest
	setJSONNamingStrategy(lowerCaseWithUnderscores)
	if ctx.Flags.Minified {
		worksJSON, _ = json.Marshal(worksWithMetadata)
	} else {
		worksJSON, _ = json.MarshalIndent(worksWithMetadata, "", "    ")
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
		if dirEntry.Name() == "../" || dirEntry.Name() == "./" {
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

func ContentBlockByID(id string, blocks []ContentBlock) (block ContentBlock, ok bool) {
	for _, block := range blocks {
		if block.ID == id {
			return block, true
		}
	}
	return ContentBlock{}, false
}

func (ctx *RunContext) DescriptionFilename(databaseDirectory string, workID string) string {
	// Compute the description file's path
	if ctx.Flags.Scattered {
		return path.Join(databaseDirectory, workID, ctx.Config.ScatteredModeFolder, "description.md")
	} else {
		return path.Join(databaseDirectory, workID, "description.md")
	}
}

// Build builds a single work given the database & output folders, as wells as a work ID
func (ctx *RunContext) Build(databaseDirectory string, outputFilename string, workID string) (AnalyzedWork, error) {
	descriptionFilename := ctx.DescriptionFilename(databaseDirectory, workID)

	// Update the UI
	ctx.CurrentWorkID = workID

	// Parse the description
	descriptionRaw, err := os.ReadFile(descriptionFilename)
	if err != nil {
		return AnalyzedWork{}, err
	}

	ctx.Status(StepDescription, ProgressDetails{
		File: descriptionFilename,
	})
	metadata, localizedBlocks, title, footnotes, _ := ctx.ParseDescription(string(descriptionRaw))

	// Handle mediae
	analyzedMediae := make([]Media, 0)
	for lang, blocks := range localizedBlocks {
		for i, block := range blocks {
			if block.Type != "media" {
				continue
			}
			ctx.LogDebug("Handling media %#v", block.Media)
			analyzed, err := ctx.HandleMedia(workID, block.ID, block.Media, lang)
			if err != nil {
				return AnalyzedWork{}, err
			}

			localizedBlocks[lang][i].Media = analyzed
			analyzedMediae = append(analyzedMediae, analyzed)
		}
	}

	// Extract colors
	if ctx.Config.ExtractColors.Enabled && metadata.Colors.Empty() {
		if metadata.Thumbnail != "" {
		outer:
			for _, m := range analyzedMediae {
				if m.RelativeSource == metadata.Thumbnail {
					metadata.Colors = m.Colors
					break outer
				}
			}
		} else {
			if len(analyzedMediae) > 0 {
				metadata.Colors = analyzedMediae[0].Colors
			}
		}
	}

	localizedContent := make(map[string]LocalizedWorkContent)

	for lang := range localizedBlocks {
		layout, err := ResolveLayout(metadata, lang, localizedBlocks[lang])
		if err != nil {
			return AnalyzedWork{}, fmt.Errorf("while resolving %s layout of %s: %w", lang, workID, err)
		}

		localizedContent[lang] = LocalizedWorkContent{
			Layout:    layout,
			Title:     title[lang],
			Footnotes: footnotes[lang],
			Blocks:    localizedBlocks[lang],
		}
	}
	ctx.UpdateBuildMetadata()
	ctx.WriteBuildMetadata()

	// Return the finished work
	return AnalyzedWork{
		ID:       workID,
		Metadata: metadata,
		Content:  localizedContent,
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
