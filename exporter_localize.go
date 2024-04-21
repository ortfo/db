package ortfodb

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
)

type LocalizeExporter struct {
}

type LocalizeExporterOptions struct {
	FilenameTemplate string `yaml:"filename_template"`
}

func (e *LocalizeExporter) OptionsType() any {
	return LocalizeExporterOptions{}
}

func (e *LocalizeExporter) Name() string {
	return "localize"
}

func (e *LocalizeExporter) Description() string {
	return "Export separately the database as a single database for each language. The `content` field of each work is localized, meaning it's not an object mapping languages to localized content, but the content directly, in the language."
}

func (e *LocalizeExporter) Before(ctx *RunContext, opts ExporterOptions) error {
	return nil
}

func (e *LocalizeExporter) Export(ctx *RunContext, opts ExporterOptions, work *Work) error {
	return nil
}

func (e *LocalizeExporter) After(ctx *RunContext, opts ExporterOptions, db *Database) error {
	options := GetExporterOptions[LocalizeExporterOptions](e, opts)
	outputFilenameTemplate, err := template.New("filename").Parse(options.FilenameTemplate)
	if err != nil {
		return fmt.Errorf("while parsing output filename template %q: %w", options.FilenameTemplate, err)
	}

	for _, lang := range db.Languages() {
		out := make(map[string]map[string]any)
		for id, work := range db.Works() {
			localizedWork := make(map[string]any)
			mapstructure.Decode(work, localizedWork)
			localizedWork["content"] = work.Content.Localize(lang)
			out[id] = localizedWork
		}
		var outputFilename strings.Builder
		err := outputFilenameTemplate.Execute(&outputFilename, map[string]any{"Lang": lang})
		if err != nil {
			return fmt.Errorf("while computing output database filename template for language %q: %w", lang, err)
		} else if outputFilename.Len() == 0 {
			ExporterLogCustom(e, "Warning", "yellow", "output database filename for language %q is empty, skipping", lang)
			continue
		}

		jsonDatabase, err := jsoniter.ConfigFastest.MarshalIndent(out, "", "  ")
		if err != nil {
			return fmt.Errorf("while marshaling localized database to JSON for language %q: %w", lang, err)
		}

		os.WriteFile(outputFilename.String(), jsonDatabase, 0644)
		ExporterLogCustom(e, "Localized", "green", "database in %s to %s", lang, outputFilename.String())
	}
	return nil
}
