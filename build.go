// Package ortfodb exposes the various functions used by the ortfodb portfolio database creation command-line tool.
// It is notably used by ortfomk to share some common data between the two complementing programs.
// See https://ewen.works/ortfodb for more information.
package ortfodb

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
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
)

type Database map[string]Work

type DatabaseMeta struct {
	// Partial is true if the database was not fully built.
	Partial bool
}

func (w Database) AsSlice() []Work {
	works := make([]Work, 0)
	for _, work := range w {
		works = append(works, work)
	}
	return works
}

// Works gets the mapping of all works
func (w Database) Works() map[string]Work {
	return w
}

// WorksSlice gets the slice of all works in the database
func (w Database) WorksSlice() []Work {
	works := make([]Work, 0)
	for _, work := range w {
		works = append(works, work)
	}
	return works
}

// WorksByDate gets all the works sorted by date, with most recent works first.
func (w Database) WorksByDate() []Work {
	return SortWorksByDate(mapValues(w.Works()))
}

func SortWorksByDate(works []Work) []Work {
	worksByDate := make([]Work, 0)
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
func (w Database) GroupWorksByYear() [][]Work {
	worksByDate := w.WorksByDate()
	worksByYear := make([][]Work, 0)
	currentYear := 0
	for _, work := range worksByDate {
		year := work.Metadata.CreatedAt().Year()
		if year != currentYear {
			currentYear = year
			worksByYear = append(worksByYear, make([]Work, 0))
		}
		worksByYear[len(worksByYear)-1] = append(worksByYear[len(worksByYear)-1], work)
	}
	return worksByYear
}

// Meta gets the database meta information
func (w Database) Meta() DatabaseMeta {
	for _, work := range w {
		return work.Metadata.DatabaseMetadata
	}
	panic("no work in database, cannot get meta information")
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

type PreviouslyBuiltDatabase struct {
	mu *sync.Mutex
	Database
}

func (ctx *RunContext) PreviouslyBuiltDatabase() Database {
	ctx.previousBuiltDatabase.mu = &sync.Mutex{}
	ctx.previousBuiltDatabase.mu.Lock()
	defer ctx.previousBuiltDatabase.mu.Unlock()
	return ctx.previousBuiltDatabase.Database
}

func (ctx *RunContext) PreviouslyBuiltWork(id string) (work Work, found bool) {
	work, found = ctx.PreviouslyBuiltDatabase()[id]
	return
}

func (ctx *RunContext) PreviouslyBuiltMedia(workID string, embedDeclaration Media) (media Media, work Work, found bool) {
	work, found = ctx.PreviouslyBuiltWork(workID)
	if !found {
		return
	}
	for _, localizedContent := range work.Content {
		for _, block := range localizedContent.Blocks {
			if block.Type != "media" {
				continue
			}
			if block.Media.RelativeSource == embedDeclaration.RelativeSource {
				return block.Media, work, true
			}
		}
	}

	return
}

// RunContext holds several "global" references used throughout all the functions of a command.
type RunContext struct {
	mu sync.Mutex

	Config                *Configuration
	DatabaseDirectory     string
	OutputDatabaseFile    string
	previousBuiltDatabase PreviouslyBuiltDatabase
	Flags                 Flags
	ProgressInfoFile      string
	Exporters             []Exporter

	// Number of concurrent goroutines to use to create thumbnails per work
	thumbnailersPerWork int

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
	ExportersToUse   []string
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

func ReleaseBuildLock(outputFilename string) error {
	err := os.Remove(BuildLockFilepath(outputFilename))
	if err != nil {
		DisplayError("could not release build lockfile %s", err, BuildLockFilepath(outputFilename))
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
		previousBuiltDatabase: PreviouslyBuiltDatabase{
			mu:       &sync.Mutex{},
			Database: make(Database),
		},
	}

	thumbnailSizesCount := len(ctx.Config.MakeThumbnails.Sizes)

	if thumbnailSizesCount/2 > flags.WorkersCount {
		LogDebug("ThumbnailSizesCount/2 (%d) > flags.WorkersCount (%d). Using 2 thumbnailers per work.", thumbnailSizesCount/2, flags.WorkersCount)
		ctx.thumbnailersPerWork = 2
	} else {
		LogDebug("Configuration asks for %d thumbnail sizes. setting thumbnail workers count per work to half of that.", thumbnailSizesCount)
		ctx.thumbnailersPerWork = thumbnailSizesCount / 2
	}

	LogDebug("Using %d thumbnailers threads per work", ctx.thumbnailersPerWork)

	if ctx.ProgressInfoFile != "" {
		LogDebug("Removing progress info file %s", ctx.ProgressInfoFile)
		if err := os.Remove(ctx.ProgressInfoFile); err != nil {
			LogDebug("Could not remove progress info file %s: %s", ctx.ProgressInfoFile, err.Error())
		}
	}

	exportersToUse := flags.ExportersToUse
	if len(exportersToUse) == 0 {
		exportersToUse = mapKeys(config.Exporters)
	}

	for _, exporterName := range exportersToUse {
		exporter, err := ctx.FindExporter(exporterName)
		if err != nil {
			return &ctx, fmt.Errorf("while finding exporter %s: %w", exporterName, err)
		}

		ctx.Exporters = append(ctx.Exporters, exporter)
	}

	LogDebug("Running with configuration %#v", &config)

	previousBuiltDatabaseRaw, err := os.ReadFile(outputFilename)
	if err != nil {
		if !os.IsNotExist(err) {
			DisplayError("No previously built database file %s to use", err, outputFilename)
		}
	} else {
		previousDb := Database{}
		err = json.Unmarshal(previousBuiltDatabaseRaw, &previousDb)
		if err != nil {
			DisplayError("Couldn't use previous built database file %s", err, outputFilename)
		}
		ctx.previousBuiltDatabase = PreviouslyBuiltDatabase{Database: previousDb}
	}

	if ctx.Config.IsDefault {
		LogInfo("No configuration file found. The default configuration was used.")
	}

	err = os.MkdirAll(config.Media.At, 0o755)
	if err != nil {
		return &ctx, fmt.Errorf("while creating the media output directory: %w", err)
	}
	if err := AcquireBuildLock(outputFilename); err != nil {
		return &ctx, fmt.Errorf("another ortfo build is in progress (could not acquire build lock): %w", err)
	}

	for _, exporter := range ctx.Exporters {
		options, err := ctx.ExporterOptions(exporter)
		if err != nil {
			return &ctx, err
		}

		LogCustom("Using", "magenta", "exporter [bold]%s[reset]\n[dim]%s", exporter.Name(), exporter.Description())
		err = exporter.Before(&ctx, options)
		if err != nil {
			return &ctx, fmt.Errorf("while running exporter %s before hook: %w", exporter.Name(), err)
		}
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

func (ctx *RunContext) RunExporters(work *Work) error {
	for _, exporter := range ctx.Exporters {
		if debugging {
			LogCustom("Exporting", "magenta", "%s to %s", work.ID, exporter.Name())
		}
		options := ctx.Config.Exporters[exporter.Name()]
		err := exporter.Export(ctx, options, work)
		if err != nil {
			return fmt.Errorf("while exporting %s: %w", exporter.Name(), err)
		}
	}
	return nil
}

func (ctx *RunContext) BuildSome(include string, databaseDirectory string, outputFilename string, flags Flags, config Configuration) (Database, error) {
	defer ReleaseBuildLock(outputFilename)

	type builtItem struct {
		err      error
		work     Work
		workID   string
		reuseOld bool
	}

	// Initialize stuff
	works := ctx.PreviouslyBuiltDatabase()
	// LogDebug("initialized works@%p from previous@%p", works, ctx.previousBuiltDatabase.Database)
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

	if flags.WorkersCount < ctx.thumbnailersPerWork {
		LogWarning("Number of workers (%d) is less than the number of thumbnailers per work (%d). Setting number of workers to %d", ctx.Flags.WorkersCount, ctx.thumbnailersPerWork, ctx.thumbnailersPerWork)
		flags.WorkersCount = ctx.thumbnailersPerWork
	}

	// Build works in parallel
	for i := 0; i < flags.WorkersCount/ctx.thumbnailersPerWork; i++ {
		i := i
		LogDebug("worker #%d: starting", i)
		go func() {
			LogDebug("worker #%d: starting", i)
			for {
				dirEntry := <-workDirectoriesChannel
				workID := dirEntry.Name()
				LogDebug("worker #%d: starting with work %s", i, workID)
				_, presentBefore := ctx.PreviouslyBuiltWork(workID)
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

					ctx.Status(workID, PhaseBuilding)
					newWork, usedCache, err := ctx.Build(string(descriptionRaw), outputFilename, workID)
					if err != nil {
						DisplayError("while building %s", err, workID)
						builtChannel <- builtItem{err: fmt.Errorf("while building %s (%s): %w", workID, ctx.DescriptionFilename(databaseDirectory, workID), err)}
						continue
					}

					ctx.RunExporters(&newWork)
					if usedCache {
						ctx.Status(workID, PhaseUnchanged)
					} else {
						ctx.Status(workID, PhaseBuilt)
					}

					// Update in database
					LogDebug("worker #%d: sending freshly built work %s", i, workID)
					builtChannel <- builtItem{work: newWork, workID: workID}
					continue
					// }
				} else if presentBefore {
					// Nothing to do, old work will be kept as-is.
					LogDebug("worker #%d: nothing to do for work %s", i, workID)
					ctx.Status(workID, PhaseUnchanged)
				} else {
					LogDebug("worker #%d: Build skipped: not included by %s, not present in previous database file.", i, include)
				}
				LogDebug("worker #%d: reusing old work %s", i, workID)
				builtChannel <- builtItem{reuseOld: true, workID: workID}
			}
		}()
	}

	LogDebug("main: filling work directories")
	for _, workDirectory := range workDirectories {
		workDirectoriesChannel <- workDirectory
	}

	// Collect all newly-built works
	LogDebug("main: collecting results")
	for len(builtDirectories) < len(workDirectories) {
		result := <-builtChannel
		LogDebug("main: got result %v", result)
		if result.err != nil {
			LogDebug("main: got error, returning early")
			return Database{}, result.err
		}
		if !result.reuseOld {
			LogDebug("main: updating work %s", result.workID)
			ctx.previousBuiltDatabase.mu.Lock()
			works[result.workID] = result.work
			ctx.previousBuiltDatabase.mu.Unlock()
		}
		ctx.WriteDatabase(works, flags, outputFilename, true)
		builtDirectories = append(builtDirectories, result.workID)
		LogDebug("main: built dirs: %d out of %d", len(builtDirectories), len(workDirectories))
		LogDebug("main: left to build: %v", directoriesLeftToBuild(workDirectoriesNames, builtDirectories))
	}

	for _, exporter := range ctx.Exporters {
		options := ctx.Config.Exporters[exporter.Name()]
		LogDebug("Running exporter %s's after hook with options %#v", exporter.Name(), options)
		err := exporter.After(ctx, options, &works)
		if err != nil {
			DisplayError("while running exporter %s's after hook: %s", err, exporter.Name())
		}

	}

	return works, nil
}

func (ctx *RunContext) WriteDatabase(works Database, flags Flags, outputFilename string, partial bool) {
	LogDebug("Writing database (partial=%v) to %s", partial, outputFilename)
	worksWithDatabaseMetadata := make(Database, 0)
	for id, work := range works {
		work.Metadata.DatabaseMetadata = DatabaseMeta{Partial: partial}
		worksWithDatabaseMetadata[id] = work
	}

	// Compile the database
	var worksJSON []byte
	json := jsoniter.ConfigFastest
	if ctx.Flags.Minified {
		worksJSON, _ = json.Marshal(worksWithDatabaseMetadata)
	} else {
		worksJSON, _ = json.MarshalIndent(worksWithDatabaseMetadata, "", "    ")
	}

	// Output it
	if outputFilename == "-" {
		Println(string(worksJSON))
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
		// Using stat to follow symlinks
		stat, err := os.Stat(dirEntryAbsPath)
		if err != nil {
			continue
		}
		if !stat.IsDir() {
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
			LogDebug("skipping %s as it has no description file: %s does not exist", dirEntry.Name(), descriptionFilename)
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

// Build builds a single work given the database & output folders, as wells as a work ID.
// BuiltAt is set and DescriptionHash are set.
func (ctx *RunContext) Build(descriptionRaw string, outputFilename string, workID string) (work Work, usedCache bool, err error) {
	hash := md5.Sum([]byte(descriptionRaw))
	newDescriptionHash := base64.StdEncoding.EncodeToString(hash[:])

	if oldWork, found := ctx.PreviouslyBuiltWork(workID); found && oldWork.DescriptionHash == newDescriptionHash && !ctx.Flags.NoCache {
		LogDebug("parsing description for %s: using cached work", workID)
		work = oldWork
		usedCache = true
	} else {
		work, err = ParseDescription(ctx, string(descriptionRaw), workID)
		if err != nil {
			return Work{}, false, fmt.Errorf("while parsing description for %s: %w", workID, err)
		}

		work.DescriptionHash = newDescriptionHash
	}

	// Handle mediae
	analyzedMediae := make([]Media, 0)
	for lang, localizedContent := range work.Content {
		for i, block := range localizedContent.Blocks {
			if block.Type != "media" {
				continue
			}
			LogDebug("Handling media %#v", block.Media)
			analyzed, anchor, usedCacheForMedia, err := ctx.HandleMedia(workID, block.ID, block.Media, lang)
			if err != nil {
				return Work{}, false, err
			}

			usedCache = usedCache && usedCacheForMedia
			work.Content[lang].Blocks[i].Media = analyzed
			work.Content[lang].Blocks[i].Anchor = anchor
			analyzedMediae = append(analyzedMediae, analyzed)
		}
	}

	// Extract colors
	extractedColors := ColorPalette{}
	if ctx.Config.ExtractColors.Enabled {
		if work.Metadata.Thumbnail != "" {
		outer:
			for _, m := range analyzedMediae {
				if m.RelativeSource == work.Metadata.Thumbnail {
					extractedColors = m.Colors
					break outer
				}
			}
		} else {
			if len(analyzedMediae) > 0 {
				extractedColors = analyzedMediae[0].Colors
			}
		}
	}
	work.Metadata.Colors = work.Metadata.Colors.MergeWith(extractedColors)

	if !usedCache || work.BuiltAt.IsZero() {
		work.BuiltAt = time.Now()
	}

	// Return the finished work
	return work, usedCache, nil
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
