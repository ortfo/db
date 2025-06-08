package ortfodb

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/text/encoding/unicode"

	ll "github.com/gwennlbh/label-logger-go"
)

func (e *CustomPlugin) List(ctx *RunContext, opts ImporterOptions) ([]string, error) {
	ll.Debug("Running before commands for %s, verbose=%v", e.name, e.verbose)
	err := e.VerifyRequiredPrograms()
	if err != nil {
		return []string{}, err
	}
	ll.Debug("Setting user-supplied data for importer %s: %v", e.name, opts)
	e.data = merge(e.Manifest.Data, opts)
	if e.Manifest.Verbose {
		PluginLogCustom(e, "Debug", "magenta", ".Data for %s is %v", e.name, e.data)
	}

	err = e.runCommands(ctx, e.verbose, ".", e.Manifest.Commands["list"], map[string]any{})
	if err != nil {
		return []string{}, err
	}

	outputFile := filepath.Join(filepath.Dir(ctx.Config.source), "output")
	if _, err := os.Stat(outputFile); err != nil && os.IsNotExist(err) {
		return []string{}, fmt.Errorf("could not find output file, write the list of work IDs to import in a text file named output in the working directory, with each line being a work ID: %w", err)
	}

	workIDs := []string{}
	var lines string
	// assume utf16 bom on wind*ws
	if runtime.GOOS == "windows" {
		codec := unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
		file, err := os.Open(outputFile)
		if err != nil {
			return []string{}, fmt.Errorf("could not open output file %s: %w", outputFile, err)
		}
		defer file.Close()

		reader := codec.NewDecoder().Reader(file)
		contents, err := io.ReadAll(reader)
		if err != nil {
			return []string{}, fmt.Errorf("could not read output file %s: %w", outputFile, err)
		}

		lines = string(contents)
	} else {
		contents, err := os.ReadFile(outputFile)
		if err != nil {
			return []string{}, fmt.Errorf("could not read output file %s: %w", outputFile, err)
		}
		lines = string(contents)
	}

	for _, line := range strings.Split(lines, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.ReplaceAll(trimmed, "\uFEFF", "") // Remove any BOM character. I HATE U WINDOWS
		if trimmed == "" {
			continue
		}
		workIDs = append(workIDs, trimmed)
	}

	os.Remove(outputFile) // Clean up the output file after reading

	return workIDs, nil
}

func (e *CustomPlugin) Import(ctx *RunContext, opts ImporterOptions, workID string) error {
	pathToWorkDir := filepath.Join(ctx.DatabaseDirectory, workID)
	err := os.Mkdir(pathToWorkDir, 0755)
	if err != nil {
		return fmt.Errorf("couldn't create database subdirectory for importing work %s: %w", workID, err)
	}

	err = e.runCommands(ctx, e.verbose, pathToWorkDir, e.Manifest.Commands["import"], map[string]any{
		"ID": workID,
	})

	if err != nil {
		return fmt.Errorf("while running import command for work ID %s: %w", workID, err)
	}

	return nil
}
