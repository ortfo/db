package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"syscall"

	"github.com/docopt/docopt-go"
	"github.com/mitchellh/colorstring"
	ortfodb "github.com/ortfo/db"
)

var CLIUsage = fmt.Sprintf(`
ortfo/db v%s

Usage:
  ortfodb [options] <database> build to <to-filepath> [--config=FILEPATH] [-msS] [--]
  ortfodb [options] <inside> blog to <to-filepath>
  ortfodb [options] <database> build <include-works> to <to-filepath> [--config=FILEPATH] [-msS] [--]
  ortfodb [options] replicate <from-filepath> <to-directory> [--config=FILEPATH]
  ortfodb [options] <database> add [--overwrite] <id> [<metadata-item>...]
  ortfodb [options] <database> validate
  ortfodb [options] schemas (configuration|database|tags|technologies)

Options:
  -C --config=<filepath>      Use the configuration path at <filepath>. Defaults to ortfodb.yaml.
							  If not provided, and if ortfodb.yaml does not exist, a default configuration
							  will be written to ortfodb.yaml and used.
  -m --minified               Output a minifed JSON file
  -s --silent                 Do not write to stdout
  -S --scattered              Operate in scattered mode. See Scattered Mode section for more information.
  --no-cache				  Disable usage of previous database build as cache for this build (used for media analysis among other things).
  --workers=<count>  	      Use <count> workers to build the database. Defaults to the number of CPU cores.
  --overwrite                 (add command): Overwrite the description.md file if it already exists.

Examples:
  ortfodb database build database.json
  ortfodb database add schoolsyst/presentation -#web -#site --color 268CCE
  ortfodb replicate database.json replicated-database --config=ortfodb.yaml

Commands:
  build <to-filepath>
	Scan in <database> for folders with description.md files
    (and potential media files)
    and compile the whole database into a JSON file at <to-filepath>
	If <to-filepath> is "-", the output will be written to stdout.

  build <include-works> <to-filepath>
	Like build <to-filepath>, but only re-build works that match the glob pattern <include-works>.

  replicate <from-filepath> <to-directory>
    The reverse operation of 'build'.
    Note that <to-directory> must be an empty directory

  add <id> [<metadata-item>...]
    Creates a new description.md in the appropriate folder.
    <id> is the work's slug.
    You can provide additional metadata items in the form ITEM_NAME:VALUE,
    eg. 'add phelng tag:program tag:cli' will generate ./phelng/description.md,
    with the following contents:
    ---
    tags: [program, cli]
	made with: []
	created: ????-??-??
    ---
    # phelng

  validate <database>
    Make sure that everything is OK in the database:
    Each one of these checks are configurable and deactivable in ortfodb.yaml:validate.checks,
    the step name is the one in [square brackets] at the beginning of these lines.
    1. [schema compliance] validate compliance to schema for ortfodb.yaml
    2. [work folder names] check work folder names for url-unsafe characters or case-insensitively non-unique folder names
    3. for each work directory:
        a. [yaml header] check YAML header for unknown keys
        b. [title presence] check presence of work title
        c. [title uniqueness] check uniqueness (case-insensitive) of work title
        d. [tags presence] check if at least one tag is present
        e. [tags knowledge] check absence of unknown tags
        f. [working media files] check all local paths for links (audio/video files, image files, other files)
        g. [working urls] check that no http url gives errors

  schemas (configuration|database|tags|technologies)
    Output the JSON schema for:
	- configuration: the configuration file (.ortfodb.yaml)
	- database: the output database file
	- tags: the tags repository file (tags.yaml)
	- technologies: the technologies repository file (technologies.yaml)

Scattered mode:
  With this mode activated, when building, portfoliodb will go through each folder (non-recursively) of <from-directory>, and, if it finds a .ortfo file in the folder, consider the files in that .ortfo folder.
  (The actual name of .ortfo is configurable, set "scattered mode folder" in ortfodb.yaml to change it)

  Consider the following directory tree:

  <from-directory>
    project1
      index.html
      src
      dist
      .ortfo
        file.png
        description.md
    project2
      .ortfo
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

Build Progress:
  For integration purposes, the current build progress can be written to a file.
  The progress information is written as JSON, and has the following structure:

	total: the total number of works to process.
	processed: the number of works processed so far.
	percent: The current overall progress percentage of the build. Equal to processed/total * 100.
	current: {
		id: The id of the work being built.
		step: The current step. One of: "thumbnail", "color extraction", "description", "media analysis"
		resolution: The resolution of the thumbnail being generated. 0 when step is not "thumbnails"
		file: The file being processed (
			original media when making thumbnails or during media analysis,
			media the colors are being extracted from, or
			the description.md file when parsing description
		)
		language: Unused. Only here for consistency with ortfo/mk's --write-progress
		output: Unused. Only here for consistency with ortfo/mk's --write-progress
	}
`, ortfodb.Version)

func main() {
	usage := CLIUsage
	args, _ := docopt.ParseDoc(usage)
	if os.Getenv("DEBUG") == "1" {
		cpuProfileFile, err := os.Create("ortfodb.pprof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(cpuProfileFile)
		defer pprof.StopCPUProfile()
	}

	if err := dispatchCommand(args); err != nil {
		// Start with leading \n because previous lines will have \r\033[K in front
		ortfodb.LogCustom("Error", "red", formatError(err))
		os.Exit(1)
	}
}

func formatError(err error) string {
	output := ""
	errorFragments := strings.Split(err.Error(), ": ")
	for i, fragment := range errorFragments {
		if i > 0 {
			output += strings.Repeat("  ", i-1) + colorstring.Color("[dim][bold]→[reset] ")
		}
		if i == 0 {
			output += colorstring.Color("[red]" + fragment + "[reset]")
		} else if i == len(errorFragments)-1 {
			output += colorstring.Color("[bold]" + fragment + "[reset]")
		} else {
			output += fragment
		}
		if i < len(errorFragments)-1 {
			output += "\n"
		}
	}
	return output
}

func dispatchCommand(args docopt.Opts) error {
	if val, _ := args.Bool("build"); val {
		err := RunCommandBuild(args)
		return err
	}
	if val, _ := args.Bool("blog"); val {
		// err := RunCommandBlog(args)
		return errors.New("command “blog” is not implemented yet")
	}
	if val, _ := args.Bool("replicate"); val {
		handleControlC(args, &ortfodb.RunContext{})
		err := ortfodb.RunCommandReplicate(args)
		return err
	}
	if val, _ := args.Bool("add"); val {
		return RunCommandAdd(args)
	}
	if val, _ := args.Bool("validate"); val {
		return errors.New("command “validate” is not implemented yet")
	}
	if val, _ := args.Bool("schemas"); val {
		return RunCommandSchemas(args)
	}
	return nil
}

// RunCommandBuild runs the command 'build' given parsed CLI args from docopt.
func RunCommandBuild(args docopt.Opts) error {
	flags := ortfodb.Flags{}
	// stupid (docopt).Bind() won't work
	flags.Config, _ = args.String("--config")
	flags.Minified, _ = args.Bool("--minified")
	flags.Scattered, _ = args.Bool("--scattered")
	flags.Silent, _ = args.Bool("--silent")
	flags.NoCache, _ = args.Bool("--no-cache")
	flags.WorkersCount, _ = args.Int("--workers")
	databaseDirectory, _ := args.String("<database>")
	databaseDirectory, err := filepath.Abs(databaseDirectory)
	if err != nil {
		return err
	}
	outputFilename, _ := args.String("<to-filepath>")
	config, err := ortfodb.NewConfiguration(flags.Config, databaseDirectory)
	if err != nil {
		return err
	}
	context, err := ortfodb.PrepareBuild(databaseDirectory, outputFilename, flags, config)
	handleControlC(args, context)
	if err != nil {
		return err
	}
	includeWorksPattern, _ := args.String("<include-works>")
	if includeWorksPattern == "" {
		includeWorksPattern = "*"
	}
	works, err := context.BuildSome(includeWorksPattern, databaseDirectory, outputFilename, flags, config)
	if len(works) > 0 {
		context.WriteDatabase(works, flags, outputFilename, err != nil)
	}
	return err
}

func RunCommandBlog(args docopt.Opts) error {
	inside, _ := args.String("<inside>")
	output, _ := args.String("<to-filepath>")

	blog, err := ortfodb.BuildBlog(inside)
	if err != nil {
		return fmt.Errorf("while building blog: %w", err)
	}

	jsoned, err := json.MarshalIndent(blog, "", "    ")
	if err != nil {
		return fmt.Errorf("while encoding blog to json: %w", err)
	}

	err = os.WriteFile(output, jsoned, 0o644)
	if err != nil {
		return fmt.Errorf("while writing json blog to file: %w", err)
	}
	return nil
}

func RunCommandAdd(args docopt.Opts) error {
	flags := ortfodb.Flags{}
	flags.Config, _ = args.String("--config")
	flags.Minified, _ = args.Bool("--minified")
	flags.Scattered, _ = args.Bool("--scattered")
	flags.Silent, _ = args.Bool("--silent")
	flags.NoCache, _ = args.Bool("--no-cache")
	flags.WorkersCount, _ = args.Int("--workers")
	databaseDirectory, _ := args.String("<database>")
	databaseDirectory, err := filepath.Abs(databaseDirectory)
	if err != nil {
		return err
	}
	outputFilename, _ := args.String("<to-filepath>")
	config, err := ortfodb.NewConfiguration(flags.Config, databaseDirectory)
	if err != nil {
		return err
	}

	context, err := ortfodb.PrepareBuild(databaseDirectory, outputFilename, flags, config)
	if err != nil {
		return fmt.Errorf("while preparing build: %w", err)
	}

	projectId, _ := args.String("<id>")
	descriptionFilepath, err := context.CreateDescriptionFile(projectId, args["<metadata-item>"].([]string), args["--overwrite"].(bool))
	if err != nil {
		context.ReleaseBuildLock(ortfodb.BuildLockFilepath(outputFilename))
		return fmt.Errorf("while creating description file: %w", err)
	}

	err = context.ReleaseBuildLock(ortfodb.BuildLockFilepath(outputFilename))
	if err != nil {
		return fmt.Errorf("while releasing build lock: %w", err)
	}

	editor := os.Getenv("EDITOR")
	if editor != "" {
		ortfodb.LogCustom("Opening", "cyan", "%s in %s", descriptionFilepath, editor)
		editorPath, err := exec.LookPath(editor)
		if err != nil {
			return fmt.Errorf("while getting path to %s: %w", editor, err)
		}

		err = syscall.Exec(editorPath, []string{editorPath, descriptionFilepath}, os.Environ())
		if err != nil {
			return fmt.Errorf("while opening with %s: %w", editorPath, err)
		}

	}

	return nil
}

func RunCommandSchemas(args docopt.Opts) error {
	if val, _ := args.Bool("configuration"); val {
		fmt.Println(ortfodb.ConfigurationJSONSchema())
		return nil
	}
	if val, _ := args.Bool("database"); val {
		fmt.Println(ortfodb.DatabaseJSONSchema())
		return nil
	}
	if val, _ := args.Bool("tags"); val {
		fmt.Println(ortfodb.TagsRepositoryJSONSchema())
		return nil
	}
	if val, _ := args.Bool("technologies"); val {
		fmt.Println(ortfodb.TechnologiesRepositoryJSONSchema())
		return nil
	}
	return errors.New("Unknown schema type")
}

func handleControlC(args docopt.Opts, context *ortfodb.RunContext) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		for range sig {
			ortfodb.LogCustom("Cancelled", "yellow", "Partial database written to [bold]./%s[reset]", context.OutputDatabaseFile)
			toFilepath, argError := args.String("<to-filepath>")
			buildLockFilepath := ortfodb.BuildLockFilepath(toFilepath)
			if _, err := os.Stat(buildLockFilepath); err == nil && argError == nil {
				os.Remove(buildLockFilepath)
			}
	
			context.StopProgressBar()
			os.Exit(1)
		}
	}()
}
