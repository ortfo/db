package ortfodb

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type SqlExporterOptions struct {
	Language string `yaml:"language"`
	Output   string `yaml:"output,omitempty"`
}

type SqlExporter struct {
	result string
}

func (e *SqlExporter) OptionsType() any {
	return SqlExporterOptions{}
}

func (e *SqlExporter) Name() string {
	return "sql"
}

func (e *SqlExporter) Description() string {
	return "Export the database as SQL statements. Rudimentary for now."
}

func (e *SqlExporter) Before(ctx *RunContext, opts ExporterOptions) error {
	e.result = ""
	if !fileExists(e.outputFilename(ctx)) {
		e.result += `CREATE TABLE works (
		id TEXT PRIMARY KEY,
		title TEXT,
		summary TEXT,
		start_date TEXT,
		end_date TEXT,
		tags TEXT,
		technologies TEXT
		);`
	}
	return nil
}

func (e *SqlExporter) Export(ctx *RunContext, opts ExporterOptions, work *AnalyzedWork) error {
	options := SqlExporterOptions{}
	mapstructure.Decode(opts, &options)

	_, summary := work.FirstParagraph(options.Language)
	e.result += fmt.Sprintf(
		"INSERT INTO works (id, title, summary, start_date, end_date, tags, technologies) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s');\n",
		work.ID,
		work.Content.Localize(options.Language).Title,
		summary.Content.Markdown(),
		work.Metadata.Started,
		work.Metadata.Finished,
		strings.Join(work.Metadata.Tags, ","),
		strings.Join(work.Metadata.MadeWith, ","),
	)
	return nil
}

func (e *SqlExporter) outputFilename(ctx *RunContext) string {
	return strings.Replace(ctx.OutputDatabaseFile, ".json", ".sql", 1)
}

func (e *SqlExporter) After(ctx *RunContext, opts ExporterOptions, built *Database) error {
	os.WriteFile(e.outputFilename(ctx), []byte(e.result), 0o644)
	ExporterLogCustom(e, "Exported", "green", "SQL file to %s", e.outputFilename(ctx))
	return nil
}
