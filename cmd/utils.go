package main

import (
	"os"
	"os/signal"

	ll "github.com/ewen-lbh/label-logger-go"
	ortfodb "github.com/ortfo/db"
)

func handleError(err error) {
	if err != nil {
		ll.ErrorDisplay("", err)
		os.Exit(1)
	}
}

func handleControlC(outputFilepath string, context *ortfodb.RunContext) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		for range sig {
			ll.Log("Cancelled", "yellow", "Partial database written to [bold]./%s[reset]", context.OutputDatabaseFile)
			buildLockFilepath := ortfodb.BuildLockFilepath(outputFilepath)
			if _, err := os.Stat(buildLockFilepath); err == nil {
				os.Remove(buildLockFilepath)
			}

			ll.StopProgressBar()
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
