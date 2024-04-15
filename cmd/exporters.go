package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	ortfodb "github.com/ortfo/db"
	"github.com/spf13/cobra"
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

func init() {
	exportersCmd.AddCommand(exportersInitCmd)
	rootCmd.AddCommand(exportersCmd)
}
