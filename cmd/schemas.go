package main

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	ortfodb "github.com/ortfo/db"
	"github.com/spf13/cobra"
)

var schemasCmd = &cobra.Command{
	Use:   "schemas <resource>",
	Short: "Output JSON schemas for ortfodb's various resources",
	Long: heredoc.Doc(`
		Output the JSON schema for:
		- configuration: the configuration file (.ortfodb.yaml)
		- database: the output database file
		- tags: the tags repository file (tags.yaml)
		- technologies: the technologies repository file (technologies.yaml)
	`),
	ValidArgs: []string{"configuration", "database", "tags", "technologies"},
	Args:      cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "configuration":
			fmt.Println(ortfodb.ConfigurationJSONSchema())
		case "database":
			fmt.Println(ortfodb.DatabaseJSONSchema())
		case "tags":
			fmt.Println(ortfodb.TagsRepositoryJSONSchema())
		case "technologies":
			fmt.Println(ortfodb.TechnologiesRepositoryJSONSchema())
		case "exporter":
			fmt.Println(ortfodb.ExporterManifestJSONSchema())
		}
	},
}

func init() {
	rootCmd.AddCommand(schemasCmd)
}
