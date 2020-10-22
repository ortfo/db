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
	println("outputting database to file", outputFilename)
	_, err := GetConfigurationFromCLIArgs(args)
	projects, err := BuildProjectsTree(databaseDirectory)
	if err != nil {
		return err
	}
	works := make([]WorkObject, 0)
	for _, project := range projects {
		description := ParseDescription(project.DescriptionRaw)
		analyzedMediae := AnalyzeAllMedia(description.MediaEmbedDeclarations, project.GetProjectPath(databaseDirectory))
		work := WorkObject{
			Metadata:   description.Metadata,
			Title:      description.Title,
			Paragraphs: description.Paragraphs,
			Media:      analyzedMediae,
			Links:      description.Links,
			Footnotes:  description.Footnotes,
		}
		works = append(works, work)
	}
	worksJSON, _ := json.Marshal(works)
	println(string(worksJSON))
	err = WriteFile(outputFilename, worksJSON)
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
