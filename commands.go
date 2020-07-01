package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/docopt/docopt-go"
)

// RunCommandBuild runs the command 'build' given parsed CLI args from docopt
func RunCommandBuild(args docopt.Opts) error {
	// Weird bug if args.String("<database>") is used...
	databaseDirectory := args["<database>"].([]string)[0]
	_, err := GetConfigurationFromCLIArgs(args)
	projects, err := BuildProjectsTree(databaseDirectory)
	if err != nil {
		return err
	}
	for _, project := range projects {
		metadata, description := ParseYAMLHeader(project.DescriptionRaw)
		spew.Dump(metadata)
		abbreviationsMap := CollectAbbreviationDeclarations(description)
		description = ReplaceAbbreviations(description, abbreviationsMap)
		description = ConvertMarkdownToHTML(description)
		
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
