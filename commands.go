package main

import (

	jsoniter "github.com/json-iterator/go"

	"github.com/docopt/docopt-go"
)

// RunCommandBuild runs the command 'build' given parsed CLI args from docopt
func RunCommandBuild(args docopt.Opts) error {
	json := jsoniter.ConfigFastest
	SetJSONNamingStrategy(LowerCaseWithUnderscores)
	// Weird bug if args.String("<database>") is used...
	databaseDirectory := args["<database>"].([]string)[0]
	outputFilename, _ := args.String("<to-filepath>")
	config, validationErrs, err := GetConfigurationFromCLIArgs(args)
	if len(validationErrs) > 0 {
		DisplayValidationErrors(validationErrs)
		return nil
	}
	if err != nil {
		return err
	}
	projects, err := BuildProjectsTree(databaseDirectory)
	if err != nil {
		return err
	}
	works := make([]WorkObject, 0)
	for _, project := range projects {
		description := ParseDescription(project.DescriptionRaw)
		analyzedMediae := AnalyzeAllMedia(description.MediaEmbedDeclarations, project.GetProjectPath(databaseDirectory))
		metadata := description.Metadata
		if config.BuildSteps.ExtractColors.Enabled {
			metadata = StepExtractColors(metadata, project, databaseDirectory, config)
		}
		work := WorkObject{
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
		println(string(worksJSON))
	}
	if err != nil {
		println(err.Error())
	}
	return nil
}

// RunCommandReplicate runs the command 'replicate' given parsed CLI args from docopt
func RunCommandReplicate(args docopt.Opts) error {
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
