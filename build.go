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
	recurcopy "github.com/plus3it/gorecurcopy"
)

// RunContext holds several "global" references used throughout all the functions of a command.
type RunContext struct {
	Config *Configuration
	// ID of the work currently being processed.
	CurrentWorkID     string
	DatabaseDirectory string
	Flags             Flags
	Progress          struct {
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
}

// Project represents a project.
type Project struct {
	ID             string
	DescriptionRaw string
	Description    ParsedDescription
	Ctx            *RunContext
}

// Build builds the database at outputFilename from databaseDirectory.
// Use LoadConfiguration (and ValidateConfiguration if desired) to get a Configuration.
func Build(databaseDirectory string, outputFilename string, flags Flags, config Configuration) error {
	ctx := RunContext{
		Config:            &config,
		Flags:             flags,
		DatabaseDirectory: databaseDirectory,
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
		return fmt.Errorf("while creating the media output directory: %w", err)
	}

	works := make([]Work, 0)
	workDirectories := make([]fs.DirEntry, 0)
	databaseFiles, err := os.ReadDir(ctx.DatabaseDirectory)
	if err != nil {
		return err
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

	for _, dirEntry := range workDirectories {
		dirEntryAbsPath := path.Join(ctx.DatabaseDirectory, dirEntry.Name())

		workID := dirEntry.Name()

		// Compute the description file's path
		var descriptionFilename string
		if ctx.Flags.Scattered {
			descriptionFilename = path.Join(dirEntryAbsPath, ctx.Config.ScatteredModeFolder, "description.md")
		} else {
			descriptionFilename = path.Join(dirEntryAbsPath, "description.md")
		}

		// Update the UI
		ctx.CurrentWorkID = workID

		// Parse the description
		descriptionRaw, err := ioutil.ReadFile(descriptionFilename)
		if err != nil {
			return err
		}

		ctx.Status(StepDescription, ProgressDetails{
			File: descriptionFilename,
		})
		description := ctx.ParseDescription(string(descriptionRaw))

		// Analyze mediae
		analyzedMediae, err := ctx.AnalyzeAllMediae(description.MediaEmbedDeclarations, dirEntryAbsPath)
		if err != nil {
			return err
		}

		// Copy over the media
		if config.Media.At == "" {
			return errors.New("please specify a destination for the media files in the configuration file (set media.at)")
		}

		for _, mediae := range analyzedMediae {
			for _, media := range mediae {
				var content []byte
				absolutePath := path.Join(dirEntryAbsPath, media.Path)
				if media.ContentType != "directory" {
					content, err = os.ReadFile(absolutePath)
				}
				if err != nil {
					ctx.LogError("could not copy %s to %s: %v", absolutePath, config.Media.At, err)
				}
				err = os.MkdirAll(path.Dir(ctx.AbsolutePathToMedia(media)), 0o755)
				if err != nil {
					return fmt.Errorf("could not create output directory for %s: %w", ctx.AbsolutePathToMedia(media), err)
				}
				if media.ContentType == "directory" {
					err = recurcopy.CopyDirectory(absolutePath, ctx.AbsolutePathToMedia(media))
				} else {
					err = os.WriteFile(ctx.AbsolutePathToMedia(media), content, 0777)
				}
				if err != nil {
					ctx.LogError("could not copy %s to %s: %v", absolutePath, config.Media.At, err)
				}
			}
		}

		// Make thumbnails
		// TODO: do only one loop for media, and do color extraction, thumb creation and copy at once, instead of iterating separately three times
		// TODO: Color extraction comes after since it could take advantage of built thumbs to sample the color:
		// - faster (it takes the smallest image)
		// - for more content types (PDFs and videos cannot be used directly, but thumbnails of them can)
		metadata := description.Metadata
		if config.MakeThumbnails.Enabled {
			metadata, err = ctx.StepMakeThumbnails(metadata, workID, analyzedMediae)
			if err != nil {
				return err
			}
		}

		// Extract colors
		if config.ExtractColors.Enabled {
			// Build up the array of media paths
			// TODO: include thumbnails instead
			mediaPaths := make([]string, 0)
			for _, mediaeInOneLang := range analyzedMediae {
				for _, media := range mediaeInOneLang {
					mediaPaths = append(mediaPaths, ctx.AbsolutePathToMedia(media))
				}
			}
			metadata = ctx.StepExtractColors(metadata, mediaPaths)
		}

		// Return the finished work
		work := Work{
			ID:         workID,
			Metadata:   metadata,
			Title:      description.Title,
			Paragraphs: description.Paragraphs,
			Media:      analyzedMediae,
			Links:      description.Links,
			Footnotes:  description.Footnotes,
		}
		works = append(works, work)
		ctx.IncrementProgress()
	}

	// Compile the database
	var worksJSON []byte
	json := jsoniter.ConfigFastest
	setJSONNamingStrategy(lowerCaseWithUnderscores)
	if flags.Minified {
		worksJSON, _ = json.Marshal(works)
	} else {
		worksJSON, _ = json.MarshalIndent(works, "", "    ")
	}

	// Output it
	if outputFilename == "-" {
		fmt.Println(string(worksJSON))
	} else {
		err = writeFile(outputFilename, worksJSON)
		if err != nil {
			println(err.Error())
		}
	}

	ctx.Spinner.Stop()

	return nil
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
func (ctx *RunContext) UpdateBuildMetadata(hash string, mediaPath string, media Media, builtThumbnailSizes []uint16) {
	if ctx.BuildMetadata.MediaCache == nil {
		ctx.BuildMetadata.MediaCache = make(map[string]CachedMedia)
	}

	if _, ok := ctx.BuildMetadata.MediaCache[hash]; !ok {
		ctx.BuildMetadata.MediaCache[hash] = CachedMedia{
			BuiltThumbnailSizes: noDuplicates(builtThumbnailSizes),
			Media:               media,
			Path:                mediaPath,
		}
	} else {
		newCache := ctx.BuildMetadata.MediaCache[hash]
		newCache.Media = media
		newCache.BuiltThumbnailSizes = noDuplicates(append(ctx.BuildMetadata.MediaCache[hash].BuiltThumbnailSizes, builtThumbnailSizes...))
		ctx.BuildMetadata.MediaCache[hash] = newCache
	}

	ctx.BuildMetadata.PreviousBuildDate = time.Now()
}
