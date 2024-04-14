package main

import (
	"os"
	"os/signal"
	"strings"

	"github.com/mitchellh/colorstring"
	ortfodb "github.com/ortfo/db"
)

func handleError(err error) {
	if err != nil {
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
func handleControlC(outputFilepath string, context *ortfodb.RunContext) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		for range sig {
			ortfodb.LogCustom("Cancelled", "yellow", "Partial database written to [bold]./%s[reset]", context.OutputDatabaseFile)
			buildLockFilepath := ortfodb.BuildLockFilepath(outputFilepath)
			if _, err := os.Stat(buildLockFilepath); err == nil {
				os.Remove(buildLockFilepath)
			}

			context.StopProgressBar()
			os.Exit(1)
		}
	}()
}

func keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
