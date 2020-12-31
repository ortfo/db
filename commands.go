package main

import (
	"fmt"
	"path"

	jsoniter "github.com/json-iterator/go"

	"github.com/docopt/docopt-go"
)

// RunContext holds several "global" references used throughout all the functions of a command
type RunContext struct {
	config         *Configuration
	currentProject *ProjectTreeElement
	progress       struct {
		current int
		total   int
	}
}

// Status prints the current compilation progress
func (ctx *RunContext) Status(text string) {
	fmt.Print("\033[2K\r")
	fmt.Printf("[%v/%v] %v: %v", ctx.progress.current, ctx.progress.total, ctx.currentProject.ID, text)
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
	defer fmt.Print("\033[2K\r\n")
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
	return nil
}

// RunCommandReplicate runs the command 'replicate' given parsed CLI args from docopt
func RunCommandReplicate(args docopt.Opts) error {
	// TODO: validate database.json with a JSON schema
	var parsedDatabase []Work
	json := jsoniter.ConfigFastest
	SetJSONNamingStrategy(LowerCaseWithUnderscores)
	databaseFilepath, err := args.String("<from-filepath>")
	targetDatabasePath, err := args.String("<to-directory>")
	if err != nil {
		return err
	}
	content, err := ReadFileBytes(databaseFilepath)
	if err != nil {
		return err
	}
	validated, validationErrors, err := ValidateWithJSONSchema(string(content), DatabaseJSONSchema)
	if err != nil {
		return err
	}
	if !validated {
		DisplayValidationErrors(validationErrors, "database JSON")
		return nil
	}
	err = json.Unmarshal(content, &parsedDatabase)
	if err != nil {
		return err
	}
	ctx := RunContext{
		config: &Configuration{},
		progress: struct {
			current int
			total   int
		}{total: len(parsedDatabase)},
	}
	defer fmt.Print("\033[2K\r\n")
	err = ReplicateAll(ctx, targetDatabasePath, parsedDatabase)
	if err != nil {
		return err
	}
	return nil
}

// RunCommandAdd runs the command 'add' given parsed CLI args from docopt
func RunCommandAdd(args docopt.Opts) error {
	return nil
}

// RunCommandValidate runs the command 'validate' given parsed CLI args from docopt
func RunCommandValidate(args docopt.Opts) error {
	return nil
}
