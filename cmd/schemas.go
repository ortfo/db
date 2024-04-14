package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/invopop/jsonschema"
	ortfodb "github.com/ortfo/db"
	"github.com/spf13/cobra"
)

var schemasCmd = &cobra.Command{
	Use:   "schemas [resource]",
	Short: "Output JSON schemas for ortfodb's various resources",
	Long: heredoc.Doc(`
		Don't pass any resource to get the list of available resources

		Output the JSON schema for:
		- configuration: the configuration file (.ortfodb.yaml)
		- database: the output database file
		- tags: the tags repository file (tags.yaml)
		- technologies: the technologies repository file (technologies.yaml)
		- exporter: the manifest file for an exporter
	`),
	ValidArgs: append(ortfodb.AvailableJSONSchemas, "list"),
	Args:      cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println(strings.Join(ortfodb.AvailableJSONSchemas, "\n"))
			return
		}
		switch args[0] {
		case "configuration":
			printSchema(ortfodb.ConfigurationJSONSchema())
		case "database":
			printSchema(ortfodb.DatabaseJSONSchema())
		case "tags":
			printSchema(ortfodb.TagsRepositoryJSONSchema())
		case "technologies":
			printSchema(ortfodb.TechnologiesRepositoryJSONSchema())
		case "exporter":
			printSchema(ortfodb.ExporterManifestJSONSchema())
		}
	},
}

func init() {
	rootCmd.AddCommand(schemasCmd)
}

func printSchema(schema *jsonschema.Schema) {
	out, err := schema.MarshalJSON()
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(out)
}
