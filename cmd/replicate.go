package main

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	ortfodb "github.com/ortfo/db"
	"github.com/spf13/cobra"
)

var force bool

var replicateCmd = &cobra.Command{
	Use:   "replicate <from-filepath> <to-filepath>",
	Short: "Replicate a database directory from a built database file.",
	Long: heredoc.Doc(`Replicate a database from <from-filepath> to <to-filepath>. Note that <to-filepath> must be an empty directory.

	Example: ortfodb replicate ./database.json ./replicated-database/

	WARNING: This command is still kind-of a WIP, it works but there's minimal logging and error handling.
	`),
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		database, err := ortfodb.LoadDatabase(args[0], force)
		if err != nil {
			handleError(fmt.Errorf("while loading given database %s: %w", args[0], err))
		}

		configuration, err := ortfodb.NewConfiguration(flags.Config)
		if err != nil {
			handleError(fmt.Errorf("while loading configuration: %w", err))
		}

		ctx, err := ortfodb.PrepareBuild(args[1], args[0], flags, configuration)
		if err != nil {
			handleError(err)
		}

		err = ctx.ReplicateAll(args[1], database)
		handleError(err)
		ortfodb.ReleaseBuildLock(args[0])
	},
}

func init() {
	replicateCmd.PersistentFlags().BoolVarP(&force, "no-verify", "n", false, "Don't try to validate the built database file before replicating")
	rootCmd.AddCommand(replicateCmd)
}
