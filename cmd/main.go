package main

import (
	"os"

	ortfodb "github.com/ortfo/db"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ortfodb",
	Short: "Manage your portfolio's database",
	Long: `Scattered mode:
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
  The progress information is written as a JSON Lines file.

  Each line of this file is a JSON object that contains the following properties:

  - works_done: the number of works built
  - works_total: the total number of works to build
  - phase: one of "Thumbnailing", "Analyzing", "Building", "Built", "Reusing"
  - details: free-form additional information as an array of strings

  See ProgressInfoEvent in the documentation.`,
}

var flags ortfodb.Flags

func init() {
	rootCmd.PersistentFlags().StringVarP(&flags.Config, "config", "C", "ortfodb.yaml", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&flags.Scattered, "scattered", "S", false, "Operate in scattered mode. See Scattered Mode section for more information.")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
