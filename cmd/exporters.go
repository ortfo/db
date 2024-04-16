package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/mitchellh/colorstring"
	"github.com/mitchellh/mapstructure"
	ortfodb "github.com/ortfo/db"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var exportersCmd = &cobra.Command{
	Use:   "exporters",
	Short: "Commands related to ortfo/db exporters",
}

const exporterStarterFile = `
# yaml-language-server: $schema=https://ortfo.org/exporter.schema.json

name: %s
description: your description here

data:
	verbose: false

requires:
	- echo Hiya!

before:
    - run: echo "Hello, World!"

after:
	- run: '{{ if .Verbose }}echo{{ end }} ls -la .'
    - log: [Finished, green, running %s]
`

var exportersInitCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Create a new exporter",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		os.WriteFile(
			fmt.Sprintf("%s.yaml", args[0]),
			[]byte(strings.ReplaceAll(
				fmt.Sprintf(heredoc.Doc(exporterStarterFile), args[0], args[0]),
				"\t",
				"    "),
			),
			0644)
		ortfodb.LogCustom("Created", "green", fmt.Sprintf("example exporter at [bold]%s.yaml[reset]", args[0]))
	},
}

var exportersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available exporters",
	Run: func(cmd *cobra.Command, args []string) {
		for _, exporter := range ortfodb.BuiltinExporters() {
			printExporter(exporter)
		}
	},
}

var exporterHelpCmd = &cobra.Command{
	Use:   "help <name>",
	Short: "Get help for a specific exporter",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		exporters := make([]string, 0, len(ortfodb.BuiltinExporters()))
		for _, exporter := range ortfodb.BuiltinExporters() {
			exporters = append(exporters, exporter.Name())
		}
		return exporters, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, exporter := range ortfodb.BuiltinExporters() {
			if exporter.Name() == args[0] {
				printExporter(exporter)
				howToAdd(exporter, cmd.Flags())
				return
			}
		}
	},
}

func init() {
	exportersCmd.AddCommand(exportersInitCmd)
	exportersCmd.AddCommand(exportersListCmd)
	exportersCmd.AddCommand(exporterHelpCmd)
	rootCmd.AddCommand(exportersCmd)
}

func exporterDetails(exporter ortfodb.Exporter) (name, description string, requires []string, config map[string]any) {
	switch exporter := exporter.(type) {
	case *ortfodb.CustomExporter:
		return exporter.Name(), exporter.Description(), exporter.Manifest.Requires, exporter.Manifest.Data
	default:
		options := make(map[string]any)
		decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:  &options,
			TagName: "yaml",
		})
		switch exporter.(type) {
		case *ortfodb.SqlExporter:
			decoder.Decode(&ortfodb.SqlExporterOptions{})
		case *ortfodb.LocalizeExporter:
			decoder.Decode(&ortfodb.LocalizeExporterOptions{})
		}

		return exporter.Name(), exporter.Description(), []string{}, options
	}
}

func howToAdd(exporter ortfodb.Exporter, flags *pflag.FlagSet) {
	name, _, _, options := exporterDetails(exporter)
	if len(options) == 0 {
		return
	}
	configFilename, _ := flags.GetString("config")
	if configFilename == "" {
		configFilename = "your ortfodb config file"
	}

	fmt.Printf(colorstring.Color("To add [bold]%s[reset] to your project, add the following to [cyan]%s[reset]:\n\n"), name, configFilename)

	fmt.Printf(colorstring.Color("  [bold][dim][red]exporters:\n[reset]    [bold][red]%s:[reset] [dim]# <- add this alongside your potential other exporters\n"), name)
	for key, defaultValue := range options {
		renderedDefaultValueBytes, _ := json.Marshal(defaultValue)
		renderedDefaultValue := string(renderedDefaultValueBytes)
		if renderedDefaultValue == "null" {
			renderedDefaultValue = ""
		}
		fmt.Printf(colorstring.Color("      [bold][red]%s:[reset] [green]%s[reset]\n"), key, renderedDefaultValue)
	}

	fmt.Println("\nFeel free to change these configuration values. Check out the exporter's documentation to learn more about what they do.")
}

func printExporter(exporter ortfodb.Exporter) {
	name, description, requires, options := exporterDetails(exporter)
	wrappedDescription := wrap(12, terminalWidth(20, 100), description)
	fmt.Println(colorstring.Color(fmt.Sprintf("[bold][blue]%-10s[reset]  %s", name, wrappedDescription)))
	hasDetails := false
	descriptionIsMultiline := strings.Contains(wrappedDescription, "\n")
	if len(requires) > 0 {
		if descriptionIsMultiline {
			fmt.Println()
		}
		fmt.Println(colorstring.Color(fmt.Sprintf("%12s[bold][yellow]Requires[reset]: %s", "", strings.Join(requires, ", "))))
		hasDetails = true
	}
	if len(keys(options)) > 0 {
		if descriptionIsMultiline && !hasDetails {
			fmt.Println()
		}
		fmt.Printf(colorstring.Color("%12s[bold][cyan]Options[reset]:\n"), "")
		optionKeys := keys(options)
		sort.Strings(optionKeys)
		for _, key := range optionKeys {
			// TODO add a way to add descriptions to options
			fmt.Println(colorstring.Color(fmt.Sprintf("%12s[bold][dim]•[reset] [blue]%s[reset] %s", "", key, "")))
		}
		hasDetails = true
	}
	if hasDetails {
		fmt.Println()
	}
}
