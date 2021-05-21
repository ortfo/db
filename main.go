package main

import (
	"errors"

	"github.com/docopt/docopt-go"
	"github.com/mitchellh/colorstring"
)

const CLIUsage = `
Usage:
  portfoliodb [options] <database> build <to-filepath> [--config=FILEPATH] [-msS] [--]
  portfoliodb [options] replicate <from-filepath> <to-directory> [--config=FILEPATH]
  portfoliodb [options] <database> add <fullname> [<metadata-item>...]
  portfoliodb [options] <database> validate <database>

Options:
  -C --config=<filepath>      Use the configuration path at <filepath>. [default: .portfoliodb.yml]
  -m --minified               Output a minifed JSON file
  -s --silent                 Do not write to stdout
  -S --scattered              Operate in scattered mode. See Scattered Mode section for more information.

Examples:
  portfoliodb database build database.json
  portfoliodb database add schoolsyst/presentation -#web -#site --color 268CCE
  portfoliodb replicate database.json replicated-database --config=.portfoliodb.yml

Commands:
  build <from-directory> <to-filepath>
    Scan in <from-directory> for folders with description.md files
    (and potential media files)
    and compile the whole database into a JSON file at <to-filepath>

  replicate <from-filepath> <to-directory>
    The reverse operation of 'build'.
    Note that <to-directory> must be an empty directory

  add <name> [<metadata-item>...]
    Creates a new description.md in the appropriate folder.
    <name> is the work's name.
    You can provide additional metadata items in the form --ITEM_NAME=VALUE,
    eg. 'add phelng --tag=cli --tag=program' will generate ./phelng/description.md,
    with the following contents:
    ---
    collection: null
    ---
    # phelng
    program, cli

  validate <database>
    Make sure that everything is OK in the database:
    Each one of these checks are configurable and deactivable in .portfoliodb.yml:validate.checks,
    the step name is the one in [square brackets] at the beginning of these lines.
    1. [schema compliance] validate compliance to schema for .portfoliodb.yml and .portfoliodb-metadata.yml
    2. [work folder names] check work folder names for url-unsafe characters or case-insensitively non-unique folder names
    3. for each work directory:
        a. [yaml header] check YAML header for unknown keys using .portfoliodb-metadata.yml
        b. [title presence] check presence of work title
        c. [title uniqueness] check uniqueness (case-insensitive) of work title
        d. [tags presence] check if at least one tag is present
        e. [tags knowledge] check absence of unknown tags (using .portfoliodb-metadata.yml)
        f. [working media files] check all local paths for links (audio/video files, image files, other files)
        g. [working urls] check that no http url gives errors

Scattered mode:
  With this mode activated, when building, portfoliodb will go through each folder (non-recursively) of <from-directory>, and, if it finds a .portfoliodb file in the folder, consider the files in that .portfoliodb folder.

  Consider the following directory tree:

  <from-directory>
    project1
      index.html
      src
      dist
      .portfoliodb
        file.png
        description.md
    project2
      .portfoliodb
        file-2.png
      description.md
    otherfolder
      stuff

  Running portfoliodb build --scattered on this tree is equivalent to builing without --scattered on the following tree:

  <from-directory>
    project1
      file.png
      description.md
    project2
      file-2.png
      description.md

  Concretely, it allows you to store your portfoliodb descriptions and supporting files directly in your projects, assuming that your store all of your projects under the same directory.
`

func main() {
	usage := CLIUsage
	args, _ := docopt.ParseDoc(usage)

	if err := dispatchCommand(args); err != nil {
		// Leading \n because previous lines will have \r\033[K in front
		colorstring.Println("\n[red][bold]An error occured[reset]")
		colorstring.Println("\t[red]" + err.Error())
	}
}

func dispatchCommand(args docopt.Opts) error {
	if val, _ := args.Bool("build"); val {
		err := RunCommandBuild(args)
		return err
	}
	if val, _ := args.Bool("replicate"); val {
		err := RunCommandReplicate(args)
		return err
	}
	if val, _ := args.Bool("add"); val {
		return errors.New("command “add” is not implemented yet")
	}
	if val, _ := args.Bool("validate"); val {
		return errors.New("command “validate” is not implemented yet")
	}
	return nil
}

// RunContext holds several "global" references used throughout all the functions of a command
type RunContext struct {
	config         *Configuration
	currentProject *ProjectTreeElement
	progress       struct {
		current int
		total   int
	}
}
