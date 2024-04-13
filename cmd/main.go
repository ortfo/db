package main

import (
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/mitchellh/colorstring"
	"github.com/muesli/reflow/indent"
	ortfodb "github.com/ortfo/db"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "ortfodb",
	Short:   "Manage your portfolio's database",
	Long:    `Manage your portfolio's database — See https://github.com/ortfo/db for more information.`,
	Version: ortfodb.Version,
	Example: colorstring.Color(indent.String(heredoc.Doc(`
		[bold][dim]$[reset] [bold]ortfodb[reset] [cyan]--config[reset] [green].ortfodb.yaml[reset] [blue]build[reset] [green]database.json[reset]
		[bold][dim]$[reset] [bold]ortfodb[reset] [blue]add[reset] [green]my-project[reset]`), 2)),
}

var flags ortfodb.Flags

func init() {
	rootCmd.SetUsageFunc(customUsage)
	rootCmd.PersistentFlags().StringVarP(&flags.Config, "config", "c", "ortfodb.yaml", "config file path")
	rootCmd.PersistentFlags().BoolVar(&flags.Scattered, "scattered", false, "Operate in scattered mode. In scattered mode, the description.md files are searched inside `.ortfo' folders in every folder of the database directory, instead of directly in the database directory's folders. See https://github.com/ortfo/")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
