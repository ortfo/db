package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v2"
	"path"

	"github.com/docopt/docopt-go"
)

// ProjectTreeElement represents a project
type ProjectTreeElement struct {
	ID             string
	DescriptionRaw string
	MediaFilepaths []string
	ScatteredMode  bool // Whether the build was run with --scattered
}

// RunCommandBuild runs the command 'build' given parsed CLI args from docopt
func RunCommandBuild(args docopt.Opts) error {
	json := jsoniter.ConfigFastest
	SetJSONNamingStrategy(LowerCaseWithUnderscores)
	// Weird bug if args.String("<database>") is used...
	databaseDirectory := args["<database>"].([]string)[0]
	outputFilename, _ := args.String("<to-filepath>")
	scatteredMode, _ := args.Bool("--scattered")
	config, validationErrs, err := GetConfigurationFromCLIArgs(args)
	if len(validationErrs) > 0 {
		DisplayValidationErrors(validationErrs, "configuration")
		return nil
	}
	if err != nil {
		return err
	}
	var projects []ProjectTreeElement

	if scatteredMode {
		projects, err = BuildProjectsTreeScatteredMode(databaseDirectory)
	} else {
		projects, err = BuildProjectsTree(databaseDirectory)
	}
	if err != nil {
		return err
	}
	// defer fmt.Print("\033[2K\r\n")
	ctx := RunContext{
		config: &config,
		progress: struct {
			current int
			total   int
		}{
			total: len(projects),
		},
	}
	works := make([]Work, 0)
	for _, project := range projects {
		ctx.currentProject = &project
		ctx.progress.current++
		description := ParseDescription(ctx, project.DescriptionRaw)
		var projectPath string
		if scatteredMode {
			projectPath = path.Join(project.GetProjectPath(databaseDirectory), ".portfoliodb")
		} else {
			projectPath = project.GetProjectPath(databaseDirectory)
		}
		analyzedMediae, err := AnalyzeAllMediae(ctx, description.MediaEmbedDeclarations, projectPath)
		if err != nil {
			return err
		}
		metadata := description.Metadata
		if config.ExtractColors.Enabled {
			ctx.Status("Extracting colors")
			metadata = StepExtractColors(metadata, project, databaseDirectory, config)
		}
		if config.MakeThumbnails.Enabled {
			ctx.Status("Making thumbnails")
			metadata, err = StepMakeThumbnails(metadata, project, databaseDirectory, analyzedMediae, config)
			if err != nil {
				return err
			}
		}
		work := Work{
			ID:         project.ID,
			Metadata:   metadata,
			Title:      description.Title,
			Paragraphs: description.Paragraphs,
			Media:      analyzedMediae,
			Links:      description.Links,
			Footnotes:  description.Footnotes,
		}
		works = append(works, work)
	}
	var worksJSON []byte
	if val, _ := args.Bool("--minified"); val {
		worksJSON, _ = json.Marshal(works)
	} else {
		worksJSON, _ = json.MarshalIndent(works, "", "    ")
	}
	err = WriteFile(outputFilename, worksJSON)
	if val, _ := args.Bool("--silent"); !val {
		fmt.Print("\033[2K\r\n")
		println(string(worksJSON))
	}
	if err != nil {
		println(err.Error())
	}
	err = UpdateBuildMetadata(config)
	if err != nil {
		println(err.Error())
	}
	return nil
}

// GetProjectPath returns the project's folder path with regard to databaseDirectory
func (p *ProjectTreeElement) GetProjectPath(databaseDirectory string) string {
	if p.ScatteredMode {
		return path.Join(databaseDirectory, p.ID, ".portfoliodb")
	}
	return path.Join(databaseDirectory, p.ID)
}

// MediaAbsoluteFilepaths is like MediaFilepaths but returns absolute paths with regard to databaseDirectory
func (p *ProjectTreeElement) MediaAbsoluteFilepaths(databaseDirectory string) []string {
	absoluted := make([]string, len(p.MediaFilepaths))
	for _, item := range p.MediaFilepaths {
		absoluted = append(absoluted, path.Join(p.GetProjectPath(databaseDirectory), item))
	}
	return absoluted
}

// BuildProjectsTree scans databaseDirectory to return a slice of ProjectTreeElement's, gathering media files and other various information
func BuildProjectsTree(databaseDirectory string) ([]ProjectTreeElement, error) {
	var tree []ProjectTreeElement
	files, err := ioutil.ReadDir(databaseDirectory)
	if err != nil {
		return nil, err
	}
	for _, projectFolder := range files {
		// If it's not a directory, it's not a project folder
		// so it has nothing to do with this
		if !projectFolder.IsDir() {
			continue
		}
		projectFolderPath := path.Join(databaseDirectory, projectFolder.Name())
		// Read the description.md file
		// If description is empty, then the project is not portfoliodb-enabled.
		// See ReadDescriptionFile for more info on why
		descriptionRaw, err := ReadDescriptionFile(projectFolderPath)
		if err != nil {
			return nil, err
		}

		// Build the list of media filepaths
		mediaFilepaths, err := buildMediaFilepaths(projectFolderPath)
		if err != nil {
			return nil, err
		}

		// Append the new project
		tree = append(tree, ProjectTreeElement{
			ID:             projectFolder.Name(),
			DescriptionRaw: descriptionRaw,
			MediaFilepaths: mediaFilepaths,
		})
	}
	return tree, nil
}

func BuildProjectsTreeScatteredMode(projectsDirectory string) ([]ProjectTreeElement, error) {
	var tree []ProjectTreeElement
	files, err := ioutil.ReadDir(projectsDirectory)
	if err != nil {
		return nil, err
	}
	for _, projectFolder := range files {
		// Not a project folder
		if !projectFolder.IsDir() {
			continue
		}
		portfoliodbDirPath := path.Join(projectsDirectory, projectFolder.Name(), ".portfoliodb")
		portfoliodbDir, err := os.Stat(portfoliodbDirPath)
		// Project has no .portfoliodb file/folder
		if err != nil {
			continue
		}
		// .portfoliodb is not a folder
		if !portfoliodbDir.IsDir() {
			continue
		}
		// Read the description.md file
		// If description is empty, then the project is not portfoliodb-enabled.
		// See ReadDescriptionFile for more info on why
		descriptionRaw, err := ReadDescriptionFile(portfoliodbDirPath)
		if err != nil {
			return nil, err
		}
		if descriptionRaw == "" {
			continue
		}
		// Build the lest of media filepaths
		mediaFilepaths, err := buildMediaFilepaths(portfoliodbDirPath)
		if err != nil {
			return nil, err
		}
		// Append the new project
		tree = append(tree, ProjectTreeElement{
			ID:             projectFolder.Name(),
			DescriptionRaw: descriptionRaw,
			MediaFilepaths: mediaFilepaths,
		})
	}
	return tree, nil
}

// ReadDescriptionFile reads the description.md file in directory.
// Returns an empty string if the file is a directory or does not exist.
func ReadDescriptionFile(directory string) (string, error) {
	descriptionFilepath := path.Join(directory, "description.md")
	if !FileExists(descriptionFilepath) {
		return "", nil
	}
	descriptionFile, err := os.Stat(descriptionFilepath)
	if err != nil {
		return "", err
	}
	if descriptionFile.IsDir() {
		return "", nil
	}
	return ReadFile(descriptionFilepath)
}

func buildMediaFilepaths(at string) ([]string, error) {
	mediaFiles, err := ioutil.ReadDir(at)
	var mediaFilepaths []string
	if err != nil {
		return nil, err
	}
	for _, mediaFile := range mediaFiles {
		if mediaFile.Name() == "description.md" {
			continue
		}
		mediaFilepaths = append(mediaFilepaths, mediaFile.Name())
	}
	return mediaFilepaths, nil
}

// UpdateBuildMetadata updates metadata about the latest build in config.BuildMetadataFilepath.
// If the file does not exist, it creates it.
func UpdateBuildMetadata(config Configuration) (err error) {
	var metadata BuildMetadata
	if _, err = os.Stat(config.BuildMetadataFilepath); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(path.Dir(config.BuildMetadataFilepath), os.ModePerm)
		metadata = BuildMetadata{}
	} else {
		metadata, err = GetBuildMetadata(config)
		if err != nil {
			return
		}
	}
	metadata.PreviousBuildDate = time.Now()
	raw, err := yaml.Marshal(&metadata)
	if err != nil {
		return
	}
	err = WriteFile(config.BuildMetadataFilepath, raw)
	return
}

func GetBuildMetadata(config Configuration) (metadata BuildMetadata, err error) {
	raw, err := ReadFileBytes(config.BuildMetadataFilepath)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(raw, &metadata)
	return
}

// NeedsRebuiling returns `true` if the given path has its modified date sooner than the last build's date.
// If any error occurs, the result is true (ie 'this file needs to be rebuilt')
func NeedsRebuiling(absolutePath string, config Configuration) bool {
	metadata, err := GetBuildMetadata(config)
	if err != nil {
		return true
	}
	fileMeta, err := os.Stat(absolutePath)
	if err != nil {
		return true
	}
	return fileMeta.ModTime().After(metadata.PreviousBuildDate)
}
