// Package ortfodb exposes the various functions used by the ortfodb portfolio database creation command-line tool.
// It is notably used by ortfomk to share some common data between the two complementing programs.
// See https://ewen.works/ortfodb for more information.
package ortfodb

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"path"

	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
)

type Database map[string]AnalyzedWork

type DatabaseMeta struct {
	Partial bool
}

func (w Database) AsSlice() []AnalyzedWork {
	works := make([]AnalyzedWork, 0)
	for _, work := range w {
		works = append(works, work)
	}
	return works
}

// Works gets the mapping of all works (without the #meta "pseudo-work").
func (w Database) Works() map[string]AnalyzedWork {
	works := make(map[string]AnalyzedWork)
	for id, work := range w {
		if id == "#meta" {
			continue
		}
		works[id] = work
	}
	return works
}

// WorksSlice gets the slice of all works in the database (without the #meta "pseudo-work")
func (w Database) WorksSlice() []AnalyzedWork {
	works := make([]AnalyzedWork, 0)
	for id, work := range w {
		if id == "#meta" {
			continue
		}
		works = append(works, work)
	}
	return works
}

// WorksByDate gets all the works sorted by date, with most recent works first.
func (w Database) WorksByDate() []AnalyzedWork {
	return SortWorksByDate(mapValues(w.Works()))
}

func SortWorksByDate(works []AnalyzedWork) []AnalyzedWork {
	worksByDate := make([]AnalyzedWork, 0)
	for _, work := range works {
		worksByDate = append(worksByDate, work)
	}
	sort.Slice(worksByDate, func(i, j int) bool {
		iDate := worksByDate[i].Metadata.CreatedAt()
		jDate := worksByDate[j].Metadata.CreatedAt()

		// if one on them has 9999 as a year, put it at the end instead
		if iDate.Year() == 9999 {
			return false
		}
		if jDate.Year() == 9999 {
			return true
		}

		return iDate.After(jDate)
	})

	return worksByDate
}

// GroupWorksByYear groups works by year, with most recent years first.
func (w Database) GroupWorksByYear() [][]AnalyzedWork {
	worksByDate := w.WorksByDate()
	worksByYear := make([][]AnalyzedWork, 0)
	currentYear := 0
	for _, work := range worksByDate {
		year := work.Metadata.CreatedAt().Year()
		if year != currentYear {
			currentYear = year
			worksByYear = append(worksByYear, make([]AnalyzedWork, 0))
		}
		worksByYear[len(worksByYear)-1] = append(worksByYear[len(worksByYear)-1], work)
	}
	return worksByYear
}

// Meta gets the database meta information
func (w Database) Meta() DatabaseMeta {
	for id, metaWork := range w {
		if id == "#meta" {
			return DatabaseMeta{Partial: metaWork.Partial}
		}
	}
	panic("no meta work found, database has no meta information (no #meta top-level key found)")
}

// Partial returns true if the database results from a partial build.
func (w Database) Partial() bool {
	return w.Meta().Partial
}

func (w Database) Languages() []string {
	langs := make([]string, 0)
	for _, work := range w {
		for lang := range work.Content {
			if lang == "default" {
				continue
			}
			lang = strings.TrimSpace(lang)
			// Check if lang is already in langs
			alreadyInLangs := false
			for _, l := range langs {
				if l == lang {
					alreadyInLangs = true
					break
				}
			}
			if !alreadyInLangs {
				langs = append(langs, lang)
			}
		}
	}
	return langs
}

// RunContext holds several "global" references used throughout all the functions of a command.
type RunContext struct {
	mu sync.Mutex

	Config                *Configuration
	DatabaseDirectory     string
	OutputDatabaseFile    string
	PreviousBuiltDatabase Database
	Flags                 Flags
	BuildMetadata         BuildMetadata
	ProgressInfoFile      string

	TagsRepository         []Tag
	TechnologiesRepository []Technology
}

type Flags struct {
	Scattered        bool
	Silent           bool
	Minified         bool
	Config           string
	NoCache          bool
	WorkersCount     int
	ProgressInfoFile string
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
		ctx.DisplayError("could not release build lockfile %s", err, BuildLockFilepath(outputFilename))
	}
	return err
}

func PrepareBuild(databaseDirectory string, outputFilename string, flags Flags, config Configuration) (*RunContext, error) {
	ctx := RunContext{
		Config:             &config,
		Flags:              flags,
		DatabaseDirectory:  databaseDirectory,
		OutputDatabaseFile: outputFilename,
		ProgressInfoFile:   flags.ProgressInfoFile,
	}

	if ctx.ProgressInfoFile != "" {
		ctx.LogDebug("Removing progress info file %s", ctx.ProgressInfoFile)
		if err := os.Remove(ctx.ProgressInfoFile); err != nil {
			ctx.LogDebug("Could not remove progress info file %s: %s", ctx.ProgressInfoFile, err.Error())
		}
	}

	ctx.LogDebug("Running with configuration %#v", &config)

	previousBuiltDatabaseRaw, err := os.ReadFile(outputFilename)
	if err != nil {
		if !os.IsNotExist(err) {
			ctx.DisplayError("No previously built database file %s to use", err, outputFilename)
		}
		ctx.PreviousBuiltDatabase = Database{}
	} else {
		err = json.Unmarshal(previousBuiltDatabaseRaw, &ctx.PreviousBuiltDatabase)
		if err != nil {
			ctx.DisplayError("Couldn't use previous built database file %s", err, outputFilename)
			ctx.PreviousBuiltDatabase = Database{}
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
		return &ctx, fmt.Errorf("while creating the media output directory: %w", err)
	}
	if err := AcquireBuildLock(outputFilename); err != nil {
		return &ctx, fmt.Errorf("another ortfo build is in progress (could not acquire build lock): %w", err)
	}

	return &ctx, nil
}

// BuildAll builds the database at outputFilename from databaseDirectory.
// Use LoadConfiguration (and ValidateConfiguration if desired) to get a Configuration.
func (ctx *RunContext) BuildAll(databaseDirectory string, outputFilename string, flags Flags, config Configuration) (Database, error) {
	return ctx.BuildSome("*", databaseDirectory, outputFilename, flags, config)
}

func directoriesLeftToBuild(all []string, built []string) []string {
	remaining := make([]string, 0)
	for _, dir := range all {
		found := false
		for _, builtDir := range built {
			if dir == builtDir {
				found = true
				break
			}
		}
		if !found {
			remaining = append(remaining, dir)
		}
	}
	return remaining
}

func (ctx *RunContext) BuildSome(include string, databaseDirectory string, outputFilename string, flags Flags, config Configuration) (Database, error) {
	defer ctx.ReleaseBuildLock(outputFilename)

	type builtItem struct {
		err      error
		work     AnalyzedWork
		workID   string
		reuseOld bool
	}

	// Initialize stuff
	works := ctx.PreviousBuiltDatabase
	workDirectories, err := ctx.ComputeProgressTotal()
	if err != nil {
		return Database{}, fmt.Errorf("while computing total number of works to build: %w", err)
	}
	workDirectoriesNames := make([]string, 0)
	for _, dirEntry := range workDirectories {
		workDirectoriesNames = append(workDirectoriesNames, dirEntry.Name())
	}
	workDirectoriesChannel := make(chan os.DirEntry, len(workDirectories))
	builtChannel := make(chan builtItem)
	builtDirectories := make([]string, 0)

	if flags.WorkersCount <= 0 {
		flags.WorkersCount = runtime.NumCPU()
	}

	ctx.StartProgressBar(len(workDirectories))

	// Build works in parallel
	// worker count divided by two because each worker has two workers for thumbnail generation
	for i := 0; i < flags.WorkersCount/2; i++ {
		i := i
		ctx.LogDebug("worker #%d: starting", i)
		go func() {
			ctx.LogDebug("worker #%d: starting", i)
			for {
				dirEntry := <-workDirectoriesChannel
				workID := dirEntry.Name()
				ctx.LogDebug("worker #%d: starting with work %s", i, workID)
				oldWork, presentBefore := ctx.PreviousBuiltDatabase[workID]
				var included bool
				if include == "*" {
					included = true
				} else {
					included, err = filepath.Match(include, workID)
					if err != nil {
						builtChannel <- builtItem{err: fmt.Errorf("while testing include-works pattern %q: %w", include, err)}
						continue
					}
				}
				if included {
					// Get description file name
					descriptionFilename := ctx.DescriptionFilename(databaseDirectory, workID)

					// Get the description's contents
					descriptionRaw, err := os.ReadFile(descriptionFilename)
					if err != nil {
						builtChannel <- builtItem{err: fmt.Errorf("while reading description file %s: %w", descriptionFilename, err)}
						continue
					}

					// Compare with hash of work in old database, to determine if we can skip it
					hash := md5.Sum(descriptionRaw)
					newDescriptionHash := base64.StdEncoding.EncodeToString(hash[:])

					if !flags.NoCache && newDescriptionHash == oldWork.DescriptionHash {
						// Skip it!
						// ctx.LogInfo("%s: Build skipped: description file unmodified", workID)
						ctx.Status(workID, PhaseUnchanged)
					} else {
						ctx.Status(workID, PhaseBuilding)
						// Build it
						newWork, err := ctx.Build(string(descriptionRaw), outputFilename, workID)
						if err != nil {
							ctx.DisplayError("while building %s", err, workID)
							builtChannel <- builtItem{err: fmt.Errorf("while building %s (%s): %w", workID, ctx.DescriptionFilename(databaseDirectory, workID), err)}
							continue
						}
						ctx.Status(workID, PhaseBuilt)

						// Set meta-info
						newWork.BuiltAt = time.Now().String()
						newWork.DescriptionHash = newDescriptionHash

						// Update in database
						ctx.LogDebug("worker #%d: sending freshly built work %s", i, workID)
						builtChannel <- builtItem{work: newWork, workID: workID}
						continue
					}
				} else if presentBefore {
					// Nothing to do, old work will be kept as-is.
					ctx.LogDebug("worker #%d: nothing to do for work %s", i, workID)
				} else {
					ctx.LogDebug("worker #%d: Build skipped: not included by %s, not present in previous database file.", i, include)
				}
				ctx.LogDebug("worker #%d: reusing old work %s", i, workID)
				builtChannel <- builtItem{reuseOld: true, workID: workID}
			}
		}()
	}

	ctx.LogDebug("main: filling work directories")
	for _, workDirectory := range workDirectories {
		workDirectoriesChannel <- workDirectory
	}

	// Collect all newly-built works
	ctx.LogDebug("main: collecting results")
	for len(builtDirectories) < len(workDirectories) {
		result := <-builtChannel
		ctx.LogDebug("main: got result %v", result)
		if result.err != nil {
			ctx.LogDebug("main: got error, returning early")
			return Database{}, result.err
		}
		if !result.reuseOld {
			ctx.LogDebug("main: updating work %s", result.workID)
			works[result.workID] = result.work
		}
		ctx.WriteDatabase(works, flags, outputFilename, true)
		builtDirectories = append(builtDirectories, result.workID)
		ctx.LogDebug("main: built dirs: %d out of %d", len(builtDirectories), len(workDirectories))
		ctx.LogDebug("main: left to build: %v", directoriesLeftToBuild(workDirectoriesNames, builtDirectories))
	}
	return works, nil

}

func (ctx *RunContext) WriteDatabase(works Database, flags Flags, outputFilename string, partial bool) {
	ctx.LogDebug("Writing database (partial=%v) to %s", partial, outputFilename)
	worksWithMetadata := make(map[string]interface{})
	err := mapstructure.Decode(works, &worksWithMetadata)
	if err != nil {
		ctx.DisplayError("while converting works to map", err)
		return
	}
	worksWithMetadata["#meta"] = struct{ Partial bool }{Partial: partial}
	// Compile the database
	var worksJSON []byte
	json := jsoniter.ConfigFastest
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
			ctx.LogDebug("skipping %s as it has no description file: %s does not exist", dirEntry.Name(), descriptionFilename)
			continue
		}

		workDirectories = append(workDirectories, dirEntry)
	}

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
func (ctx *RunContext) Build(descriptionRaw string, outputFilename string, workID string) (AnalyzedWork, error) {
	metadata, localizedBlocks, title, footnotes, _ := ParseDescription[WorkMetadata](ctx, string(descriptionRaw))

	// Handle mediae
	analyzedMediae := make([]Media, 0)
	for lang, blocks := range localizedBlocks {
		for i, block := range blocks {
			if block.Type != "media" {
				continue
			}
			ctx.LogDebug("Handling media %#v", block.Media)
			analyzed, anchor, err := ctx.HandleMedia(workID, block.ID, block.Media, lang)
			if err != nil {
				return AnalyzedWork{}, err
			}

			localizedBlocks[lang][i].Media = analyzed
			localizedBlocks[lang][i].Anchor = anchor
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

	localizedContent := make(map[string]LocalizedContent)

	for lang := range localizedBlocks {
		layout, err := ResolveLayout(metadata, lang, localizedBlocks[lang])
		if err != nil {
			return AnalyzedWork{}, fmt.Errorf("while resolving %s layout of %s: %w", lang, workID, err)
		}

		localizedContent[lang] = LocalizedContent{
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
	ctx.mu.Lock()
	ctx.BuildMetadata.PreviousBuildDate = time.Now()
	ctx.mu.Unlock()
}
