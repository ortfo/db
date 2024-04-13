package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/MakeNowJust/heredoc"
	ortfodb "github.com/ortfo/db"
	"github.com/spf13/cobra"
)

var overwrite bool

func init() {
	addCmd.PersistentFlags().BoolVar(&overwrite, "overwrite", false, "Overwrite the description.md file if it already exists")
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add <id>",
	Short: "Add a new project to your portfolio",
	Long:  heredoc.Doc(`Create a new project in the appropriate folder. ID is the work's slug.`),
	Args:  cobra.MinimumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		config, err := ortfodb.NewConfiguration(flags.Config)
		if err != nil {
			handleError(err)
		}
		entries, err := os.ReadDir(config.ProjectsDirectory)
		if err != nil {
			handleError(fmt.Errorf("while listing projects directory at %s: %w", config.ProjectsDirectory, err))
		}

		validWorkIds := make([]string, 0, len(entries))
		for _, entry := range entries {
			if entry.IsDir() {
				validWorkIds = append(validWorkIds, entry.Name())
			}
		}
		return validWorkIds, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		config, err := ortfodb.NewConfiguration(flags.Config)
		if err != nil {
			handleError(err)
		}
		context, err := ortfodb.PrepareBuild(config.ProjectsDirectory, "./fictional.json", flags, config)
		if err != nil {
			handleError(fmt.Errorf("while preparing build: %w", err))
		}

		projectId := args[0]

		// TODO
		metadataItems := []string{}

		descriptionFilepath, err := context.CreateDescriptionFile(projectId, metadataItems, overwrite)
		if err != nil {
			context.ReleaseBuildLock(ortfodb.BuildLockFilepath("./fictional.json"))
			handleError(fmt.Errorf("while creating description file: %w", err))
		}

		err = context.ReleaseBuildLock(ortfodb.BuildLockFilepath("./fictional.json"))
		if err != nil {
			handleError(fmt.Errorf("while releasing build lock: %w", err))
		}

		editor := os.Getenv("EDITOR")
		if editor != "" {
			ortfodb.LogCustom("Opening", "cyan", "%s in %s", descriptionFilepath, editor)
			editorPath, err := exec.LookPath(editor)
			if err != nil {
				handleError(fmt.Errorf("while getting path to %s: %w", editor, err))
			}

			err = syscall.Exec(editorPath, []string{editorPath, descriptionFilepath}, os.Environ())
			if err != nil {
				handleError(fmt.Errorf("while opening with %s: %w", editorPath, err))
			}

		}
	},
}
