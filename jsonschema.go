package ortfodb

import (
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
)

var AvailableJSONSchemas = []string{"configuration", "database", "tags", "technologies", "exporter"}

var yamlReflector = jsonschema.Reflector{
	FieldNameTag: "yaml",
	KeyNamer:     func(key string) string { return strings.ToLower(key) },
}

func setSchemaId(schema *jsonschema.Schema, name string) {
	schema.ID = jsonschema.ID(fmt.Sprintf("https://raw.githubusercontent.com/ortfo/db/v%s/schemas/%s.schema.json", Version, name))
}

func makeJSONSchema(t any, yaml bool) *jsonschema.Schema {
	selectedReflector := jsonschema.Reflector{}
	if yaml {
		selectedReflector = yamlReflector
	}
	selectedReflector.AddGoComments("github.com/ortfo/db", "./")
	schema := selectedReflector.Reflect(t)
	parts := strings.Split(string(schema.ID), "/")
	base := parts[len(parts)-1]
	setSchemaId(schema, base)
	return schema
}

func ConfigurationJSONSchema() *jsonschema.Schema {
	return makeJSONSchema(&Configuration{}, true)
}

func DatabaseJSONSchema() *jsonschema.Schema {
	return makeJSONSchema(&Database{}, false)
}

type tags []Tag

func TagsRepositoryJSONSchema() *jsonschema.Schema {
	return makeJSONSchema(&tags{}, true)
}

type technologies []Technology

func TechnologiesRepositoryJSONSchema() *jsonschema.Schema {
	return makeJSONSchema(&technologies{}, true)
}

func ExporterManifestJSONSchema() *jsonschema.Schema {
	schema := makeJSONSchema(&ExporterManifest{}, true)
	setSchemaId(schema, "exporter")
	return schema
}
