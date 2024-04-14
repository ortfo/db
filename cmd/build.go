package main

import (
	"runtime"

	"github.com/MakeNowJust/heredoc"
	ortfodb "github.com/ortfo/db"
	"github.com/spf13/cobra"
)

func init() {
	buildCmd.PersistentFlags().BoolVarP(&flags.Minified, "minified", "m", false, "Output a minifed JSON file")
	buildCmd.PersistentFlags().BoolVarP(&flags.Silent, "silent", "q", false, "Do not write to stdout")
	buildCmd.PersistentFlags().StringVar(&flags.ProgressInfoFile, "write-progress", "", "Write progress information to a file. See https://pkg.go.dev/github.com/ortfo/db#ProgressInfoEvent for more information.")
	buildCmd.PersistentFlags().BoolVar(&flags.NoCache, "no-cache", false, "Disable usage of previous database build as cache for this build (used for media analysis among other things).")
	buildCmd.PersistentFlags().IntVar(&flags.WorkersCount, "workers", runtime.NumCPU(), "Use <count> workers to build the database. Defaults to the number of CPU cores.")
	buildCmd.PersistentFlags().StringArrayVarP(&flags.ExportersToUse, "exporters", "e", []string{}, "Exporters to enable. If not provided, all the exporters configured in the configuration file will be enabled.")
	buildCmd.RegisterFlagCompletionFunc("exporters", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		config, err := ortfodb.NewConfiguration(flags.Config)
		if err != nil {
			handleError(err)
		}
		//TODO omit already enabled exporters
		return keys(config.Exporters), cobra.ShellCompDirectiveNoFileComp
	})
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build <to-filepath> [include-works]",
	Short: "Build the database",
	Long: heredoc.Doc(`Scan in the projects directory for folders with description.md files (and potential media files) and compile the whole database into a JSON file at <to-filepath>.

	If <to-filepath> is "-", the output will be written to stdout.

	If [include-works] is provided, only works that match the pattern will be included in the database.
	`),
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		outputFilename := args[0]
		config, err := ortfodb.NewConfiguration(flags.Config)
		if err != nil {
			handleError(err)
		}

		context, err := ortfodb.PrepareBuild(config.ProjectsDirectory, outputFilename, flags, config)
		if err != nil {
			handleError(err)
		}

		handleControlC(outputFilename, context)

		includeWorksPattern := ""
		if len(args) > 1 {
			includeWorksPattern = args[1]
		} else {
			includeWorksPattern = "*"
		}

		works, err := context.BuildSome(includeWorksPattern, config.ProjectsDirectory, outputFilename, flags, config)

		if len(works) > 0 {
			context.WriteDatabase(works, flags, outputFilename, err != nil)
		}

		if err != nil {
			handleError(err)
		}
	},
}
